package highlighter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lasseh/jink/lexer"
)

// streamFixtures are byte sequences a PTY session really produces.
var streamFixtures = map[string]string{
	"config":          "set interfaces ge-0/0/0 unit 0\nset system host-name r1\n",
	"show table":      "Peer  AS  InPkt  State\n10.0.0.1  65001  12  Establ\n",
	"keystroke echo":  "user@r> s\x08 \x08show route\r\n",
	"prompt redraw":   "user@r> show\r\x1b[Kuser@r> show route",
	"erase and retry": "set\b \b\b \b\b \bdelete system\r\n",
	"erase to end":    "foo\x1b[K+ bar\r\n",
	"window title":    "\x1b]0;user@r: ~\x07user@r> ",
	"title with ST":   "\x1b]0;title\x1b\\set system\n",
	"colored input":   "\x1b[1m\x1b[32muser@r>\x1b[0m show\r\n",
	"cursor moves":    "abc \x1b[10D set \x1b[2Cxyz\n",
	"stray escape":    "abc \x1b",
	"no newline":      "user@r> ",
}

// chunk splits data into pieces of the given size.
func chunk(data string, size int) []string {
	var pieces []string
	for len(data) > size {
		pieces = append(pieces, data[:size])
		data = data[size:]
	}
	return append(pieces, data)
}

// runStream writes pieces to a new Stream one at a time and returns the output.
func runStream(t *testing.T, pieces []string) string {
	t.Helper()
	var out bytes.Buffer
	s := New().NewStream(&out)
	for _, piece := range pieces {
		if _, err := s.Write([]byte(piece)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// escapesOf returns the escape sequences in input, in order.
func escapesOf(input string) []string {
	var escapes []string
	for _, seg := range extractSegments(input) {
		if seg.isEscape {
			escapes = append(escapes, seg.text)
		}
	}
	return escapes
}

// assertPreserved checks that output has the visible text of input, and every
// input escape intact and in order.
func assertPreserved(t *testing.T, input, output string) {
	t.Helper()
	if StripANSI(output) != StripANSI(input) {
		t.Fatalf("visible text changed:\n  input:  %q\n  output: %q", input, output)
	}
	got := escapesOf(output)
	for _, want := range escapesOf(input) {
		found := false
		for len(got) > 0 && !found {
			found = got[0] == want
			got = got[1:]
		}
		if !found {
			t.Fatalf("escape %q lost or out of order:\n  input:  %q\n  output: %q", want, input, output)
		}
	}
}

func TestStreamPreservesBytesAcrossChunkings(t *testing.T) {
	for name, input := range streamFixtures {
		t.Run(name, func(t *testing.T) {
			for size := 1; size <= len(input); size++ {
				assertPreserved(t, input, runStream(t, chunk(input, size)))
			}
		})
	}
}

func TestStreamSplitsAtLinesMatchSingleWrite(t *testing.T) {
	input := "Peer  AS  InPkt  State\n10.0.0.1  65001  12  Establ\nuser@r> show route\nset system\n"
	want := runStream(t, []string{input})
	got := runStream(t, strings.SplitAfter(input, "\n"))
	if got != want {
		t.Errorf("line chunks differ from a single write:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestStreamLexesInLineContext(t *testing.T) {
	theme := DefaultTheme()
	tests := []struct {
		name   string
		pieces []string
		want   string
		token  lexer.TokenType
	}{
		{"word finished in a later write", []string{"user@r> sho", "w route"}, "w", lexer.TokenCommand},
		{"fragment is not the first word", []string{"set foo ", "bar"}, "bar", lexer.TokenValue},
		{"prompt split across writes", []string{"user@", "r> "}, "r", lexer.TokenPromptHostOper},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runStream(t, tt.pieces)
			want := theme.GetColor(tt.token) + tt.want + Reset
			if !strings.HasSuffix(strings.TrimRight(got, " "), want) && !strings.Contains(got, want) {
				t.Errorf("want %q in %q", want, got)
			}
		})
	}
}

func TestStreamEscapeDoesNotStartALine(t *testing.T) {
	got := runStream(t, []string{"foo\x1b[K+ bar\n"})
	diff := DefaultTheme().GetColor(lexer.TokenDiffAdd) + "+"
	if strings.Contains(got, diff) {
		t.Errorf("mid-line + colored as a diff line: %q", got)
	}
}

func TestStreamCarriageReturnStartsALine(t *testing.T) {
	// The space matters: "abc" alone would still read as part of a prompt.
	got := runStream(t, []string{"old text", "\ruser@r> "})
	want := DefaultTheme().GetColor(lexer.TokenPromptUser) + "user" + Reset
	if !strings.Contains(got, want) {
		t.Errorf("prompt after \\r not recognized: %q", got)
	}
}

func TestStreamBackspaceEditsLineContext(t *testing.T) {
	// "shoe" backspaced to "sho" and finished with "w" is the command "show".
	got := runStream(t, []string{"user@r> shoe", "\b", "w"})
	want := DefaultTheme().GetColor(lexer.TokenCommand) + "w" + Reset
	if !strings.Contains(got, want) {
		t.Errorf("want %q in %q", want, got)
	}
}

func TestStreamModeCarriesThroughATable(t *testing.T) {
	// A row alone has too few signals to be recognized as show output.
	row := "10.0.0.1 65001 Establ\n"
	good := DefaultTheme().GetColor(lexer.TokenStateGood) + "Establ" + Reset

	if got := runStream(t, []string{row}); strings.Contains(got, good) {
		t.Fatalf("row without a table should not be show output: %q", got)
	}

	got := runStream(t, []string{"Peer  AS  InPkt  State\n", row})
	if !strings.Contains(got, good) {
		t.Errorf("row inside a table should be show output: %q", got)
	}
}

func TestStreamPromptEndsMode(t *testing.T) {
	row := "10.0.0.1 65001 Establ\n"
	good := DefaultTheme().GetColor(lexer.TokenStateGood) + "Establ" + Reset

	got := runStream(t, []string{"Peer  AS  InPkt  State\n", "user@r> show configuration\n", row})
	if strings.Contains(got, good) {
		t.Errorf("show mode should end at the prompt: %q", got)
	}
}

func TestStreamConfigSignalEndsShowMode(t *testing.T) {
	line := "unknown-word another-word\n"
	value := DefaultTheme().GetColor(lexer.TokenValue) + "another-word" + Reset

	got := runStream(t, []string{"Peer  AS  InPkt  State\n", "set system host-name r1\n", line})
	if !strings.Contains(got, value) {
		t.Errorf("config line should end show mode: %q", got)
	}
}

func TestStreamEscapeAcrossHoldLimitRead(t *testing.T) {
	// A long line in one write keeps an escape whole wherever it falls.
	escapes := []string{"\033[10D", "\033]0;user@router\007"}
	for _, esc := range escapes {
		for offset := 1; offset < len(esc); offset++ {
			input := strings.Repeat("x", escapeHoldLimit+1-offset) + esc + "set system"
			assertPreserved(t, input, runStream(t, []string{input}))
		}
	}
}

func TestStreamOversizedStringIsNotColored(t *testing.T) {
	// A clipboard OSC can exceed the hold limit. It must reach the terminal
	// unchanged, and highlighting must resume after it.
	payload := strings.Repeat("set ", escapeHoldLimit)
	osc := "\033]52;c;" + payload + "\007"
	input := "show " + osc + "set system"

	for _, size := range []int{1000, 3999, len(input)} {
		got := runStream(t, chunk(input, size))
		if !strings.Contains(got, osc) {
			t.Fatalf("chunk size %d: OSC payload was changed", size)
		}
		want := DefaultTheme().GetColor(lexer.TokenCommand) + "set" + Reset
		if !strings.HasPrefix(got[strings.Index(got, osc)+len(osc):], want) {
			t.Errorf("chunk size %d: highlighting did not resume after the OSC", size)
		}
	}
}

func TestStreamOversizedEscapeIsNotHeld(t *testing.T) {
	// An unterminated sequence may never finish, so output must not stall.
	var out bytes.Buffer
	s := New().NewStream(&out)
	input := "\033]52;c;" + strings.Repeat("A", escapeHoldLimit+100)
	if _, err := s.Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if out.String() != input {
		t.Errorf("wrote %d bytes before Flush, want %d", out.Len(), len(input))
	}
}

func TestStreamFlushWritesHeldBytes(t *testing.T) {
	var out bytes.Buffer
	s := New().NewStream(&out)
	if _, err := s.Write([]byte("\x1b[1")); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("unfinished escape was written early: %q", out.String())
	}
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "\x1b[1" {
		t.Errorf("Flush wrote %q", out.String())
	}
}

func TestDetectingStream(t *testing.T) {
	var out bytes.Buffer
	s := New().NewDetectingStream(&out)
	for _, line := range []string{"plain text\n", "\x1b[1mbold\x1b[0m text\n", "set system\n", "plain text\n"} {
		if _, err := s.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}

	got := out.String()
	if !strings.HasPrefix(got, "plain text\n\x1b[1mbold\x1b[0m text\n") {
		t.Errorf("text before JunOS should pass through unchanged: %q", got)
	}
	want := DefaultTheme().GetColor(lexer.TokenCommand) + "set" + Reset
	if !strings.Contains(got, want) {
		t.Errorf("JunOS line should be highlighted: %q", got)
	}
	if strings.HasSuffix(got, "\nplain text\n") {
		t.Errorf("detection should stick for later lines: %q", got)
	}
}
