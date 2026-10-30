package highlighter

import "testing"

// FuzzHighlight feeds arbitrary bytes through the public highlighting and ANSI
// helpers, asserting none panic and that StripANSI behaves as a pure removal
// (never grows the input and is idempotent). Input arrives from routers and
// piped files, so it cannot be trusted to be well-formed. It runs only under
// the -fuzz flag, e.g. "go test -run '^$' -fuzz FuzzHighlight ./highlighter/".
func FuzzHighlight(f *testing.F) {
	seeds := []string{
		"",
		"set system host-name r1;",
		"user@router> show interfaces",
		"\x1b[31mcolored\x1b[0m text",
		"inet.0: 5 destinations\n10.0.0.0/8  *[BGP/170]",
		"\x1b]0;window title\x07",
		"\x00\x1b[2J\r\n",
		"\x1b", // lone escape
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		h := New()
		// None of these may panic on arbitrary bytes.
		_ = Highlight(in)
		_ = h.Highlight(in)
		_ = h.HighlightForced(in)
		_ = h.HighlightShowOutput(in)
		_ = HasANSI(in)

		stripped := StripANSI(in)
		if len(stripped) > len(in) {
			t.Fatalf("StripANSI grew input: %d > %d (%q)", len(stripped), len(in), in)
		}
		if again := StripANSI(stripped); again != stripped {
			t.Fatalf("StripANSI not idempotent: %q -> %q -> %q", in, stripped, again)
		}
	})
}

// FuzzStream feeds arbitrary bytes through a Stream in arbitrary chunk sizes and
// asserts the visible text and escape sequences survive. It runs only under the
// -fuzz flag, e.g. "go test -run '^$' -fuzz FuzzStream ./highlighter/".
func FuzzStream(f *testing.F) {
	for _, input := range streamFixtures {
		f.Add(input, uint8(1))
		f.Add(input, uint8(3))
	}
	f.Add("\x1b", uint8(1))
	f.Add("a\x1b\x1b[1m\x1b]0;\x1b", uint8(2))

	f.Fuzz(func(t *testing.T, in string, size uint8) {
		pieces := chunk(in, int(size)+1)
		assertPreserved(t, in, runStream(t, pieces))
	})
}
