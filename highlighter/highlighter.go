package highlighter

import (
	"bytes"
	"strings"
	"sync"

	"github.com/lasseh/jink/lexer"
)

// Highlight is a convenience function that highlights JunOS config/output using the default theme.
// For more control, create a Highlighter instance with New() or NewWithTheme().
func Highlight(input string) string {
	return New().Highlight(input)
}

// Highlighter applies ANSI color codes to JunOS configuration and show command output.
// It supports multiple color themes and can be toggled on/off at runtime.
// All methods are safe for concurrent use.
type Highlighter struct {
	theme   *Theme
	enabled bool
	mu      sync.RWMutex
}

// New creates a new Highlighter with the default theme (Tokyo Night).
func New() *Highlighter {
	return &Highlighter{
		theme:   DefaultTheme(),
		enabled: true,
	}
}

// NewWithTheme creates a new Highlighter with a specific theme.
func NewWithTheme(theme *Theme) *Highlighter {
	return &Highlighter{
		theme:   theme,
		enabled: true,
	}
}

// SetTheme changes the highlighting theme.
func (h *Highlighter) SetTheme(theme *Theme) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.theme = theme
}

// Enable turns highlighting on.
func (h *Highlighter) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = true
}

// Disable turns highlighting off.
func (h *Highlighter) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = false
}

// IsEnabled returns whether highlighting is enabled.
func (h *Highlighter) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

// Toggle switches highlighting on/off and returns the new state.
func (h *Highlighter) Toggle() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = !h.enabled
	return h.enabled
}

// Highlight applies syntax highlighting to the input text.
// Returns input unchanged if highlighting is disabled, input is empty,
// or input doesn't look like JunOS config/output (uses heuristic detection).
func (h *Highlighter) Highlight(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}

	// Strip ANSI codes before detection check - routers may send colored output
	// that interferes with pattern matching
	cleaned := StripANSI(input)

	// Check if this looks like JunOS config (simple heuristics)
	if !lexer.IsJunOS(cleaned) {
		return input
	}

	return h.renderTokens(tokenize(cleaned, lexer.ParseModeAuto), 0)
}

// HighlightForced applies syntax highlighting without checking if input looks like JunOS.
func (h *Highlighter) HighlightForced(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}
	return h.highlightTokens(input, lexer.ParseModeAuto)
}

// highlightTokens tokenizes and colorizes the input while preserving cursor control sequences.
func (h *Highlighter) highlightTokens(input string, mode lexer.ParseMode) string {
	// Fast path: most lines contain no escape sequences, so skip the segment
	// split (one slice alloc + extra buffer copies) and tokenize directly.
	if !HasANSI(input) {
		return h.renderTokens(tokenize(input, mode), 0)
	}

	// Extract cursor control sequences and text segments separately
	// We need to preserve cursor movement/clear sequences for proper terminal rendering
	segments := extractSegments(input)

	var buf bytes.Buffer
	for _, seg := range segments {
		if seg.isEscape {
			// Pass through escape sequences unchanged
			buf.WriteString(seg.text)
		} else {
			// Highlight text segments
			buf.WriteString(h.renderTokens(tokenize(seg.text, mode), 0))
		}
	}
	return buf.String()
}

// tokenize lexes text, leaving mode detection to the lexer for ParseModeAuto.
func tokenize(text string, mode lexer.ParseMode) []lexer.Token {
	lex := lexer.New(text)
	if mode != lexer.ParseModeAuto {
		lex.SetParseMode(mode)
	}
	return lex.Tokenize()
}

// renderTokens applies theme colors to a slice of tokens and returns the
// colorized string, leaving out the first skip bytes of token text.
func (h *Highlighter) renderTokens(tokens []lexer.Token, skip int) string {
	h.mu.RLock()
	theme := h.theme
	h.mu.RUnlock()

	// Hold the theme lock for the whole render rather than per token.
	theme.mu.RLock()
	defer theme.mu.RUnlock()

	var buf bytes.Buffer
	for _, token := range tokens {
		value := token.Value
		if skip >= len(value) {
			skip -= len(value)
			continue
		}
		value = value[skip:]
		skip = 0

		color := theme.colors[token.Type]
		if color != "" {
			buf.WriteString(color)
			buf.WriteString(value)
			buf.WriteString(Reset)
		} else {
			buf.WriteString(value)
		}
	}
	return buf.String()
}

// HighlightLine highlights a single line (useful for streaming).
func (h *Highlighter) HighlightLine(line string) string {
	return h.Highlight(line)
}

// HighlightLines highlights multiple lines preserving line structure.
func (h *Highlighter) HighlightLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = h.Highlight(line)
	}
	return result
}

// HighlightShowOutput highlights show command output specifically using show mode.
func (h *Highlighter) HighlightShowOutput(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}

	return h.highlightTokens(input, lexer.ParseModeShow)
}

// segment represents either an escape sequence or text content.
type segment struct {
	text     string
	isEscape bool
}

// CSI sequence byte range constants.
const (
	csiParamStart = 0x20 // Space - start of parameter/intermediate bytes
	csiParamEnd   = 0x3F // ? - end of parameter bytes
	csiFinalStart = 0x40 // @ - start of final bytes
	csiFinalEnd   = 0x7E // ~ - end of final bytes
	csiIntermEnd  = 0x2F // / - end of intermediate bytes
	escFinalStart = 0x30 // 0 - start of non-CSI escape final bytes
	escapeChar    = '\033'
	csiBracket    = '['
	belChar       = '\a'
	stFinal       = '\\' // ESC \ is the String Terminator (ST)
)

// isCSIParamByte checks if byte is a CSI parameter or intermediate byte (0x20-0x3F).
func isCSIParamByte(b byte) bool {
	return b >= csiParamStart && b <= csiParamEnd
}

// isCSIFinalByte checks if byte is a CSI final byte (0x40-0x7E).
func isCSIFinalByte(b byte) bool {
	return b >= csiFinalStart && b <= csiFinalEnd
}

// isCSIIntermediateByte checks if byte is a CSI intermediate byte (0x20-0x2F).
func isCSIIntermediateByte(b byte) bool {
	return b >= csiParamStart && b <= csiIntermEnd
}

// escapeEnd returns the position just past the escape sequence starting at
// input[i] (an ESC byte), and whether the sequence is complete. It handles
// CSI (ESC [), string sequences terminated by BEL or ST (OSC ESC ], DCS ESC P,
// SOS ESC X, PM ESC ^, APC ESC _), and short escapes like ESC =.
// An incomplete sequence runs to the end of input.
func escapeEnd(input string, i int) (int, bool) {
	j := i + 1
	if j >= len(input) {
		return len(input), false
	}

	switch input[j] {
	case csiBracket:
		// Parameter and intermediate bytes, then one final byte
		j++
		for j < len(input) && isCSIParamByte(input[j]) {
			j++
		}
		if j >= len(input) {
			return len(input), false
		}
		if isCSIFinalByte(input[j]) {
			j++
		}
		return j, true
	case ']', 'P', 'X', '^', '_':
		// String payload until BEL or ST
		for j++; j < len(input); j++ {
			if input[j] == belChar {
				return j + 1, true
			}
			if input[j] == escapeChar {
				if j+1 >= len(input) {
					return len(input), false
				}
				if input[j+1] == stFinal {
					return j + 2, true
				}
				// Any other escape aborts the string
				return j, true
			}
		}
		return len(input), false
	default:
		// Intermediate bytes, then one final byte
		for j < len(input) && isCSIIntermediateByte(input[j]) {
			j++
		}
		if j >= len(input) {
			return len(input), false
		}
		if input[j] < escFinalStart || input[j] > csiFinalEnd {
			// Not a valid final byte (e.g. another ESC or \r); the ESC
			// stands alone and the next byte is scanned normally.
			return i + 1, true
		}
		return j + 1, true
	}
}

// extractSegments splits input into escape sequences and text segments
// This allows us to preserve cursor control sequences while highlighting text.
func extractSegments(input string) []segment {
	var segments []segment
	var textBuf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar {
			// Flush any accumulated text
			if textBuf.Len() > 0 {
				segments = append(segments, segment{text: textBuf.String(), isEscape: false})
				textBuf.Reset()
			}

			start := i
			i, _ = escapeEnd(input, i)
			segments = append(segments, segment{text: input[start:i], isEscape: true})
			continue
		}
		textBuf.WriteByte(input[i])
		i++
	}

	// Flush remaining text
	if textBuf.Len() > 0 {
		segments = append(segments, segment{text: textBuf.String(), isEscape: false})
	}

	return segments
}

// StripANSI removes ANSI escape codes from text.
// Handles CSI sequences (colors, cursor control), OSC and other string
// sequences (window titles), and short two-byte escapes.
func StripANSI(input string) string {
	var buf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar {
			i, _ = escapeEnd(input, i)
			continue
		}
		buf.WriteByte(input[i])
		i++
	}

	return buf.String()
}

// incompleteEscapeStart returns the index where an unfinished escape sequence
// at the end of input begins, or -1 if input ends cleanly. Streaming callers
// hold those bytes back until the next read completes the sequence.
func incompleteEscapeStart(input string) int {
	i := 0
	for i < len(input) {
		if input[i] != escapeChar {
			i++
			continue
		}
		end, complete := escapeEnd(input, i)
		if !complete {
			return i
		}
		i = end
	}
	return -1
}

// HasANSI checks if the input contains ANSI escape codes.
func HasANSI(input string) bool {
	return strings.IndexByte(input, escapeChar) >= 0
}
