package highlighter

import (
	"bytes"
	"io"
	"slices"
	"strings"

	"github.com/lasseh/jink/lexer"
)

const (
	// escapeHoldLimit caps an unfinished escape sequence held for the next
	// write. One longer than this may never finish, so it isn't held.
	escapeHoldLimit = 4000

	// lineContextLimit caps the text kept for lexing the rest of a line, which
	// bounds the work per write on very long lines.
	lineContextLimit = 1024
)

// Stream highlights a byte stream that arrives in arbitrary chunks, such as PTY
// output. It owns the boundaries a chunk can fall on: escape sequences are kept
// whole, text is lexed knowing what came earlier on its line, and the parse mode
// carries from line to line, so a row of a show table is lexed as part of the
// table.
//
// Text is written as soon as it arrives, because an interactive session needs
// every keystroke echoed. A word split across writes therefore gets its color
// from what was known at the time, so colors can differ between chunkings of
// the same bytes when a split falls inside a word. The bytes themselves never
// change, apart from the color codes added around them.
//
// A Stream is not safe for concurrent use.
type Stream struct {
	h            *Highlighter
	w            io.Writer
	requireJunOS bool            // leave text uncolored until it looks like JunOS
	held         []byte          // unfinished escape sequence
	inString     bool            // inside a string sequence too long to hold
	line         []byte          // text already written on the current line
	mode         lexer.ParseMode // mode carried over from earlier lines
}

// NewStream returns a Stream that highlights everything written to it and
// writes the result to w.
func (h *Highlighter) NewStream(w io.Writer) *Stream {
	return &Stream{h: h, w: w}
}

// NewDetectingStream returns a Stream that passes text through unchanged until
// a line looks like JunOS, and highlights from that line on.
func (h *Highlighter) NewDetectingStream(w io.Writer) *Stream {
	return &Stream{h: h, w: w, requireJunOS: true}
}

// Write highlights p and writes the result to the underlying writer. An
// escape sequence cut off at the end of p is held until the next Write or Flush.
func (s *Stream) Write(p []byte) (int, error) {
	data := p
	if len(s.held) > 0 {
		data = slices.Concat(s.held, p)
		s.held = nil
	}

	var out bytes.Buffer
	if s.inString {
		data = s.passString(&out, data)
	}

	text := string(data)
	k := incompleteEscapeStart(text)
	if k < 0 {
		s.writeComplete(&out, text)
	} else {
		s.writeComplete(&out, text[:k])
		rest := text[k:]
		if len(rest) > escapeHoldLimit {
			out.WriteString(rest)
			s.inString = isStringEscape(rest)
		} else {
			s.held = []byte(rest)
		}
	}

	if _, err := s.w.Write(out.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Flush writes any held bytes unchanged. Call it when the stream ends.
func (s *Stream) Flush() error {
	if len(s.held) == 0 {
		return nil
	}
	_, err := s.w.Write(s.held)
	s.held = nil
	return err
}

// passString copies the rest of a string sequence too long to hold and returns
// what follows it. The terminal is still consuming that sequence, so none of it
// can be colored. A closing ESC is left for the escape scanner.
func (s *Stream) passString(out *bytes.Buffer, data []byte) []byte {
	i := bytes.IndexAny(data, "\a\x1b")
	if i < 0 {
		out.Write(data)
		return nil
	}
	s.inString = false
	if data[i] == belChar {
		i++
	}
	out.Write(data[:i])
	return data[i:]
}

// writeComplete highlights data that holds only complete escape sequences.
func (s *Stream) writeComplete(out *bytes.Buffer, data string) {
	if !HasANSI(data) {
		s.writeText(out, data)
		return
	}
	for _, seg := range extractSegments(data) {
		if seg.isEscape {
			out.WriteString(seg.text)
		} else {
			s.writeText(out, seg.text)
		}
	}
}

// writeText highlights text between escape sequences. Characters that move the
// cursor are written as they are, and tracked to keep the line context accurate.
func (s *Stream) writeText(out *bytes.Buffer, text string) {
	for text != "" {
		i := strings.IndexAny(text, "\r\n\b")
		if i < 0 {
			s.writeRun(out, text)
			return
		}
		s.writeRun(out, text[:i])
		out.WriteByte(text[i])
		if text[i] != '\b' {
			s.line = s.line[:0]
		} else if len(s.line) > 0 {
			s.line = s.line[:len(s.line)-1]
		}
		text = text[i+1:]
	}
}

// writeRun highlights a run of text within one line. It lexes the whole line so
// far, so the run is classified in context, and renders only the new part.
func (s *Stream) writeRun(out *bytes.Buffer, run string) {
	if run == "" {
		return
	}

	line := string(s.line) + run
	if s.requireJunOS && !lexer.IsJunOS(line) {
		out.WriteString(run)
		s.remember(run)
		return
	}
	s.requireJunOS = false

	// A prompt ends the previous command's output, so its mode ends here too.
	if lexer.IsPrompt(line) {
		s.mode = lexer.ParseModeAuto
	} else if mode := lexer.DetectMode(line); mode != lexer.ParseModeAuto {
		s.mode = mode
	}

	out.WriteString(s.h.renderTokens(tokenize(line, s.mode), len(s.line)))
	s.remember(run)
}

// remember adds run to the line context.
func (s *Stream) remember(run string) {
	s.line = append(s.line, run...)
	if len(s.line) > lineContextLimit {
		s.line = s.line[:0]
	}
}

// isStringEscape reports whether sequence is a string sequence, which runs
// until BEL or ST.
func isStringEscape(sequence string) bool {
	return len(sequence) > 1 && strings.IndexByte("]PX^_", sequence[1]) >= 0
}
