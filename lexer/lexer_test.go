package lexer

import (
	"testing"
)

func TestTokenizeCommands(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"set", TokenCommand},
		{"delete", TokenCommand},
		{"deactivate", TokenCommand},
		{"activate", TokenCommand},
		{"edit", TokenCommand},
		{"show", TokenCommand},
		{"request", TokenCommand},
		{"commit", TokenCommand},
		{"rollback", TokenCommand},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeSections(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"system", TokenSection},
		{"interfaces", TokenSection},
		{"protocols", TokenSection},
		{"routing-options", TokenSection},
		{"policy-options", TokenSection},
		{"firewall", TokenSection},
		{"security", TokenSection},
		{"vlans", TokenSection},
		{"groups", TokenSection},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeProtocols(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"ospf", TokenProtocol},
		{"bgp", TokenProtocol},
		{"isis", TokenProtocol},
		{"ldp", TokenProtocol},
		{"mpls", TokenProtocol},
		{"tcp", TokenProtocol},
		{"udp", TokenProtocol},
		{"icmp", TokenProtocol},
		{"inet", TokenProtocol},
		{"inet6", TokenProtocol},
		{"lldp", TokenProtocol},
		{"lacp", TokenProtocol},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeActions(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"accept", TokenAction},
		{"reject", TokenAction},
		{"discard", TokenAction},
		{"deny", TokenAction},
		{"permit", TokenAction},
		{"next-hop", TokenAction},
		{"community", TokenAction},
		{"log", TokenAction},
		{"count", TokenAction},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeInterfaces(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		// Physical interfaces
		{"ge-0/0/0", TokenInterface},
		{"ge-0/0/1", TokenInterface},
		{"xe-0/0/0", TokenInterface},
		{"xe-1/2/3", TokenInterface},
		{"et-0/0/0", TokenInterface},
		{"fe-0/0/0", TokenInterface},
		// With units
		{"ge-0/0/0.0", TokenInterface},
		{"ge-0/0/0.100", TokenInterface},
		{"xe-1/0/0.999", TokenInterface},
		// Aggregated/logical
		{"ae0", TokenInterface},
		{"ae1", TokenInterface},
		{"ae15", TokenInterface},
		{"lo0", TokenInterface},
		{"lo0.0", TokenInterface},
		{"irb", TokenInterface},
		{"irb.100", TokenInterface},
		{"em0", TokenInterface},
		{"me0", TokenInterface},
		{"vlan", TokenInterface},
		{"vlan.100", TokenInterface},
		// Special
		{"all", TokenInterface},
		// Channelized
		{"ge-0/0/0:0", TokenInterface},
		{"ge-0/0/0:1", TokenInterface},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv4(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.1", TokenIPv4},
		{"10.0.0.1", TokenIPv4},
		{"172.16.0.1", TokenIPv4},
		{"255.255.255.255", TokenIPv4},
		{"0.0.0.0", TokenIPv4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv4Prefix(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.0/24", TokenIPv4Prefix},
		{"10.0.0.0/8", TokenIPv4Prefix},
		{"172.16.0.0/12", TokenIPv4Prefix},
		{"0.0.0.0/0", TokenIPv4Prefix},
		{"192.168.1.1/32", TokenIPv4Prefix},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv6(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"2001:db8::1", TokenIPv6},
		{"::1", TokenIPv6},
		{"fe80::1", TokenIPv6},
		{"2001:db8:85a3::8a2e:370:7334", TokenIPv6},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv6Prefix(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"2001:db8::/32", TokenIPv6Prefix},
		{"::/0", TokenIPv6Prefix},
		{"fe80::/10", TokenIPv6Prefix},
		{"2001:db8::1/128", TokenIPv6Prefix},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{"line comment", "# this is a comment", TokenComment},
		{"annotation", "## Last commit: 2024-01-01", TokenAnnotation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
			if tokens[0].Value != tt.input {
				t.Errorf("expected value %q, got %q", tt.input, tokens[0].Value)
			}
		})
	}
}

func TestTokenizeBlockComment(t *testing.T) {
	input := "/* block comment */"
	l := New(input)
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenComment {
		t.Errorf("expected TokenComment, got %v", tokens[0].Type)
	}
}

func TestTokenizeStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"double quoted", `"hello world"`},
		{"single quoted", `'hello world'`},
		{"with spaces", `"Uplink to ISP"`},
		{"empty", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenString {
				t.Errorf("expected TokenString, got %v", tokens[0].Type)
			}
			if tokens[0].Value != tt.input {
				t.Errorf("expected value %q, got %q", tt.input, tokens[0].Value)
			}
		})
	}
}

func TestTokenizeBraces(t *testing.T) {
	input := "{}"
	l := New(input)
	tokens := l.Tokenize()
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenBrace || tokens[0].Value != "{" {
		t.Errorf("expected opening brace, got %v %q", tokens[0].Type, tokens[0].Value)
	}
	if tokens[1].Type != TokenBrace || tokens[1].Value != "}" {
		t.Errorf("expected closing brace, got %v %q", tokens[1].Type, tokens[1].Value)
	}
}

func TestTokenizeSemicolon(t *testing.T) {
	input := ";"
	l := New(input)
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenSemicolon {
		t.Errorf("expected TokenSemicolon, got %v", tokens[0].Type)
	}
}

func TestTokenizeWildcards(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"<*>"},
		{"<ge-*>"},
		{"*"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenWildcard {
				t.Errorf("expected TokenWildcard, got %v", tokens[0].Type)
			}
		})
	}
}

func TestTokenizeNumbers(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"100"},
		{"1000"},
		{"65535"},
		{"10g"},
		{"100m"},
		{"1G"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenNumber {
				t.Errorf("expected TokenNumber for %q, got %v", tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeCommunity(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"65000:100"},
		{"65001:1"},
		{"100:200"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenCommunity {
				t.Errorf("expected TokenCommunity for %q, got %v", tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeASN(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"AS65000", TokenASN},
		{"AS1", TokenASN},
		{"as65001", TokenASN},
		{"As12345", TokenASN},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeUnit(t *testing.T) {
	// Unit numbers should be classified as TokenUnit after "unit" keyword
	input := "unit 100"
	l := New(input)
	tokens := l.Tokenize()

	// Should have: unit, space, 100
	foundUnit := false
	for i, tok := range tokens {
		if tok.Type == TokenKeyword && tok.Value == "unit" {
			// Next non-whitespace should be TokenUnit
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j].Type != TokenText {
					if tokens[j].Type != TokenUnit {
						t.Errorf("expected TokenUnit after 'unit', got %v", tokens[j].Type)
					}
					foundUnit = true
					break
				}
			}
			break
		}
	}
	if !foundUnit {
		t.Error("did not find TokenUnit in output")
	}
}

func TestTokenizeFullLine(t *testing.T) {
	input := "set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24;"

	l := New(input)
	tokens := l.Tokenize()

	expected := []struct {
		tokenType TokenType
		value     string
	}{
		{TokenCommand, "set"},
		{TokenText, " "},
		{TokenSection, "interfaces"},
		{TokenText, " "},
		{TokenInterface, "ge-0/0/0"},
		{TokenText, " "},
		{TokenKeyword, "unit"},
		{TokenText, " "},
		{TokenUnit, "0"},
		{TokenText, " "},
		{TokenKeyword, "family"},
		{TokenText, " "},
		{TokenProtocol, "inet"},
		{TokenText, " "},
		{TokenKeyword, "address"},
		{TokenText, " "},
		{TokenIPv4Prefix, "192.168.1.1/24"},
		{TokenSemicolon, ";"},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.tokenType {
			t.Errorf("token %d: expected type %v, got %v (value: %q)", i, exp.tokenType, tokens[i].Type, tokens[i].Value)
		}
		if tokens[i].Value != exp.value {
			t.Errorf("token %d: expected value %q, got %q", i, exp.value, tokens[i].Value)
		}
	}
}

func TestTokenizeHierarchicalConfig(t *testing.T) {
	input := `system {
    host-name router;
}`

	l := New(input)
	tokens := l.Tokenize()

	// Should have: system, space, {, newline+spaces, host-name, space, router, ;, newline, }
	hasSection := false
	hasKeyword := false
	hasBraces := false

	for _, tok := range tokens {
		if tok.Type == TokenSection && tok.Value == "system" {
			hasSection = true
		}
		if tok.Type == TokenKeyword && tok.Value == "host-name" {
			hasKeyword = true
		}
		if tok.Type == TokenBrace {
			hasBraces = true
		}
	}

	if !hasSection {
		t.Error("expected to find section 'system'")
	}
	if !hasKeyword {
		t.Error("expected to find keyword 'host-name'")
	}
	if !hasBraces {
		t.Error("expected to find braces")
	}
}

func TestTokenizeKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"host-name", TokenKeyword},
		{"description", TokenKeyword},
		{"address", TokenKeyword},
		{"unit", TokenKeyword},
		{"family", TokenKeyword},
		{"vlan-id", TokenKeyword},
		{"neighbor", TokenKeyword},
		{"group", TokenKeyword},
		{"filter", TokenKeyword},
		{"term", TokenKeyword},
		{"from", TokenKeyword},
		{"then", TokenKeyword},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeMAC(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"00:11:22:33:44:55"},
		{"aa:bb:cc:dd:ee:ff"},
		{"AA:BB:CC:DD:EE:FF"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenMAC {
				t.Errorf("expected TokenMAC for %q, got %v", tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	input := "set\ndelete"
	l := New(input)
	tokens := l.Tokenize()

	// set is on line 1
	if tokens[0].Line != 1 {
		t.Errorf("expected line 1 for 'set', got %d", tokens[0].Line)
	}

	// Find 'delete' token (skip whitespace)
	for _, tok := range tokens {
		if tok.Value == "delete" {
			if tok.Line != 2 {
				t.Errorf("expected line 2 for 'delete', got %d", tok.Line)
			}
			break
		}
	}
}

func TestEmptyInput(t *testing.T) {
	l := New("")
	tokens := l.Tokenize()
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", len(tokens))
	}
}

func TestWhitespaceOnly(t *testing.T) {
	l := New("   \t\n  ")
	tokens := l.Tokenize()
	// Should get whitespace tokens
	for _, tok := range tokens {
		if tok.Type != TokenText {
			t.Errorf("expected TokenText for whitespace, got %v", tok.Type)
		}
	}
}

// ==================== Show Output Tests ====================

func TestTokenizeStatesGood(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"Up", TokenStateGood},
		{"up", TokenStateGood},
		{"Establ", TokenStateGood},
		{"establ", TokenStateGood},
		{"established", TokenStateGood},
		{"Established", TokenStateGood},
		{"Full", TokenStateGood},
		{"full", TokenStateGood},
		{"Master", TokenStateGood},
		{"master", TokenStateGood},
		{"Primary", TokenStateGood},
		{"Enabled", TokenStateGood},
		{"ok", TokenStateGood},
		{"online", TokenStateGood},
		{"running", TokenStateGood},
		{"ready", TokenStateGood},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesBad(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"Down", TokenStateBad},
		{"down", TokenStateBad},
		{"Idle", TokenStateBad},
		{"idle", TokenStateBad},
		{"Active", TokenStateBad}, // BGP Active state = not established
		{"active", TokenStateBad},
		{"Connect", TokenStateBad},
		{"connect", TokenStateBad},
		{"OpenSent", TokenStateBad},
		{"opensent", TokenStateBad},
		{"OpenConfirm", TokenStateBad},
		{"failed", TokenStateBad},
		{"error", TokenStateBad},
		{"offline", TokenStateBad},
		{"disabled", TokenStateBad},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesWarning(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"Init", TokenStateWarning},
		{"init", TokenStateWarning},
		{"2Way", TokenStateWarning},
		{"2way", TokenStateWarning},
		{"ExStart", TokenStateWarning},
		{"exstart", TokenStateWarning},
		{"Exchange", TokenStateWarning},
		{"exchange", TokenStateWarning},
		{"Loading", TokenStateWarning},
		{"loading", TokenStateWarning},
		{"flapping", TokenStateWarning},
		{"pending", TokenStateWarning},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesNeutral(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"inactive", TokenStateNeutral},
		{"standby", TokenStateNeutral},
		{"backup", TokenStateNeutral},
		{"n/a", TokenStateNeutral},
		{"none", TokenStateNeutral},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeTimeDurations(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"1w", TokenTimeDuration},
		{"2d", TokenTimeDuration},
		{"3h", TokenTimeDuration},
		{"45m", TokenTimeDuration},
		{"30s", TokenTimeDuration},
		{"1w2d", TokenTimeDuration},
		{"1w2d3h", TokenTimeDuration},
		{"0:45", TokenTimeDuration},
		{"0:45:30", TokenTimeDuration},
		{"12:00:00", TokenTimeDuration},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeTableNames(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"inet.0", TokenTableName},
		{"inet.0:", TokenTableName},
		{"inet6.0", TokenTableName},
		{"mpls.0", TokenTableName},
		{"bgp.0", TokenTableName},
		{"l2vpn.0", TokenTableName},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatusSymbols(t *testing.T) {
	// Note: "*" is handled specially as TokenWildcard in the scanner
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"+", TokenStatusSymbol},
		{"-", TokenStatusSymbol},
		{">", TokenStatusSymbol},
		{"B", TokenStatusSymbol},
		{"O", TokenStatusSymbol},
		{"S", TokenStatusSymbol},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeColumnHeaders(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"Neighbor", TokenColumnHeader},
		{"neighbor", TokenColumnHeader},
		{"Peer", TokenColumnHeader},
		{"State", TokenColumnHeader},
		{"Interface", TokenColumnHeader},
		{"Admin", TokenColumnHeader},
		{"Link", TokenColumnHeader},
		{"Flaps", TokenColumnHeader},
		{"OutQ", TokenColumnHeader},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestParseModeDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParseMode
	}{
		{
			name:     "config set style",
			input:    "set interfaces ge-0/0/0 unit 0 family inet",
			expected: ParseModeConfig,
		},
		{
			name:     "config hierarchical",
			input:    "system {\n    host-name router;\n}",
			expected: ParseModeConfig,
		},
		{
			name:     "bgp summary with states",
			input:    "Peer                     AS      InPkt     OutPkt    State\n10.0.0.1              65001       1234       5678    Establ",
			expected: ParseModeShow,
		},
		{
			name:     "interface terse",
			input:    "Interface               Admin Link Proto    Local\nge-0/0/0                up    up",
			expected: ParseModeShow,
		},
		{
			name:     "ospf neighbor state",
			input:    "Address          Interface              State\n10.0.0.2         ge-0/0/0.0             Full",
			expected: ParseModeShow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			mode := l.detectParseMode()
			if mode != tt.expected {
				t.Errorf("expected mode %v, got %v", tt.expected, mode)
			}
		})
	}
}

func TestSetParseMode(t *testing.T) {
	l := New("up")

	// Default mode should be Auto
	if l.GetParseMode() != ParseModeAuto {
		t.Errorf("expected default mode ParseModeAuto, got %v", l.GetParseMode())
	}

	// Set to Show mode
	l.SetParseMode(ParseModeShow)
	if l.GetParseMode() != ParseModeShow {
		t.Errorf("expected ParseModeShow, got %v", l.GetParseMode())
	}

	// In show mode, "up" should be StateGood
	tokens := l.Tokenize()
	if len(tokens) != 1 || tokens[0].Type != TokenStateGood {
		t.Errorf("expected TokenStateGood for 'up' in show mode")
	}
}

func TestShowModePreservesSharedPatterns(t *testing.T) {
	// IPs and interfaces should still work in show mode
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.1", TokenIPv4},
		{"10.0.0.0/24", TokenIPv4Prefix},
		{"ge-0/0/0", TokenInterface},
		{"ae0", TokenInterface},
		{"65001", TokenNumber},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q in show mode, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}
