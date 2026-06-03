package lexer

import "testing"

// FuzzTokenize feeds arbitrary input through the lexer in every parse mode and
// asserts it never panics or hangs, and that emitted tokens carry sane columns.
// The lexer parses untrusted data (router output, piped files), so robustness
// against malformed bytes matters. It runs only under the -fuzz flag, e.g.
// "go test -run '^$' -fuzz FuzzTokenize ./lexer/".
func FuzzTokenize(f *testing.F) {
	seeds := []string{
		"",
		"set system host-name r1;",
		"user@router> show interfaces",
		"{master:0}[edit]\nadmin@core-sw1#",
		"[edit interfaces]",
		"+ unit 0;",
		"- mtu 1500;",
		"192.168.1.0/24 ge-0/0/0.100 2001:db8::1 00:11:22:33:44:55",
		"inet.0: 5 destinations\n10.0.0.0/8  *[BGP/170] 1w2d 99.9%",
		"description \"unterminated string",
		"\x00\x1b[2J\r\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	check := func(t *testing.T, tokens []Token, mode ParseMode, in string) {
		for _, tk := range tokens {
			if tk.Value != "" && tk.Column < 1 {
				t.Fatalf("non-positive column %d for token %q (mode %d, input %q)",
					tk.Column, tk.Value, mode, in)
			}
		}
	}

	f.Fuzz(func(t *testing.T, in string) {
		// Auto-detection path.
		check(t, New(in).Tokenize(), ParseModeAuto, in)

		// Explicit modes.
		for _, m := range []ParseMode{ParseModeConfig, ParseModeShow} {
			lex := New(in)
			lex.SetParseMode(m)
			check(t, lex.Tokenize(), m, in)
		}

		_ = IsPrompt(in) // must not panic on arbitrary input
	})
}
