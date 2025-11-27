package highlighter

import (
	"strings"
	"testing"

	"github.com/lasseh/jink/lexer"
)

func TestNew(t *testing.T) {
	h := New()
	if h == nil {
		t.Fatal("New() returned nil")
	}
	if !h.IsEnabled() {
		t.Error("new highlighter should be enabled by default")
	}
}

func TestNewWithTheme(t *testing.T) {
	theme := MonokaiTheme()
	h := NewWithTheme(theme)
	if h == nil {
		t.Fatal("NewWithTheme() returned nil")
	}
	if !h.IsEnabled() {
		t.Error("new highlighter should be enabled by default")
	}
}

func TestSetTheme(t *testing.T) {
	h := New()
	h.SetTheme(NordTheme())
	// Just verify no panic - theme is private
}

func TestEnableDisable(t *testing.T) {
	h := New()

	if !h.IsEnabled() {
		t.Error("should be enabled by default")
	}

	h.Disable()
	if h.IsEnabled() {
		t.Error("should be disabled after Disable()")
	}

	h.Enable()
	if !h.IsEnabled() {
		t.Error("should be enabled after Enable()")
	}
}

func TestToggle(t *testing.T) {
	h := New()

	// Initially enabled
	if !h.IsEnabled() {
		t.Error("should be enabled by default")
	}

	// Toggle off
	result := h.Toggle()
	if result {
		t.Error("Toggle() should return false after disabling")
	}
	if h.IsEnabled() {
		t.Error("should be disabled after first toggle")
	}

	// Toggle on
	result = h.Toggle()
	if !result {
		t.Error("Toggle() should return true after enabling")
	}
	if !h.IsEnabled() {
		t.Error("should be enabled after second toggle")
	}
}

func TestHighlightEmpty(t *testing.T) {
	h := New()
	result := h.Highlight("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestHighlightDisabled(t *testing.T) {
	h := New()
	h.Disable()

	input := "set interfaces ge-0/0/0"
	result := h.Highlight(input)
	if result != input {
		t.Errorf("disabled highlighter should return input unchanged, got %q", result)
	}
}

func TestHighlightNonJunOS(t *testing.T) {
	h := New()

	// Random text that doesn't look like JunOS config
	input := "Hello, this is just some random text"
	result := h.Highlight(input)

	// Should return unchanged (no highlighting)
	if result != input {
		t.Errorf("non-JunOS text should be returned unchanged")
	}
}

func TestHighlightBasic(t *testing.T) {
	h := New()

	input := "set interfaces ge-0/0/0"
	result := h.Highlight(input)

	// Should contain ANSI codes
	if !strings.Contains(result, "\033[") {
		t.Error("highlighted output should contain ANSI escape codes")
	}

	// Should still contain the original text
	stripped := StripANSI(result)
	if stripped != input {
		t.Errorf("stripped output should match input, got %q", stripped)
	}
}

func TestHighlightLine(t *testing.T) {
	h := New()

	input := "set system host-name router"
	result := h.HighlightLine(input)

	if !HasANSI(result) {
		t.Error("HighlightLine should produce ANSI output for JunOS config")
	}
}

func TestHighlightLines(t *testing.T) {
	h := New()

	input := []string{
		"set system host-name router",
		"set interfaces ge-0/0/0 unit 0",
	}
	result := h.HighlightLines(input)

	if len(result) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(result))
	}

	for i, line := range result {
		if !HasANSI(line) {
			t.Errorf("line %d should have ANSI codes", i)
		}
	}
}

func TestLooksLikeJunOS(t *testing.T) {
	h := New()

	positives := []string{
		"set interfaces ge-0/0/0",
		"delete system services",
		"show configuration",
		"edit protocols bgp",
		"interfaces {",
		"system {",
		"protocols {",
		"routing-options {",
		"family inet",
		"ge-0/0/0 {",
		"unit 0 {",
		"## Last commit",
		"ospf area 0.0.0.0",
		"bgp group external",
	}

	for _, input := range positives {
		if !h.looksLikeJunOS(input) {
			t.Errorf("should recognize %q as JunOS config", input)
		}
	}

	negatives := []string{
		"Hello world",
		"This is plain text",
		"SELECT * FROM users",
		"function main() {}",
		"import os",
	}

	for _, input := range negatives {
		if h.looksLikeJunOS(input) {
			t.Errorf("should NOT recognize %q as JunOS config", input)
		}
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\033[31mred\033[0m", "red"},
		{"\033[1m\033[94mblue\033[0m", "blue"},
		{"\033[38;5;208morange\033[0m", "orange"},
		{"no codes here", "no codes here"},
		{"", ""},
		// Cursor control sequences (the bug fix)
		{"\033[Khello", "hello"},                 // Clear to end of line
		{"\033[2Khello", "hello"},                // Clear entire line
		{"\033[Ahello", "hello"},                 // Cursor up
		{"\033[1;1Hhello", "hello"},              // Cursor position
		{"\033[K\rprompt> cmd", "\rprompt> cmd"}, // Clear + carriage return + text
		{"before\033[Kafter", "beforeafter"},     // Mid-string clear
		{"\033[32mgreen\033[K\033[0m", "green"},  // Color + clear + reset
	}

	for _, tt := range tests {
		result := StripANSI(tt.input)
		if result != tt.expected {
			t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractSegments(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
		description   string
	}{
		{"plain text", 1, "plain text has one segment"},
		{"\033[Khello", 2, "clear + text"},
		{"\033[31mred\033[0m", 3, "color + text + reset"},
		{"hello\033[Kworld", 3, "text + clear + text"},
		{"", 0, "empty string"},
	}

	for _, tt := range tests {
		result := extractSegments(tt.input)
		if len(result) != tt.expectedCount {
			t.Errorf("extractSegments(%q): expected %d segments, got %d (%s)",
				tt.input, tt.expectedCount, len(result), tt.description)
		}
	}
}

func TestHighlightForcedPreservesEscapeSequences(t *testing.T) {
	h := New()

	// Simulate what JunOS sends when pressing up arrow for command recall:
	// Clear line + carriage return + prompt with recalled command
	input := "\033[Kuser@router> show route"

	result := h.HighlightForced(input)

	// The cursor control sequence should be preserved at the start
	if !strings.HasPrefix(result, "\033[K") {
		t.Errorf("HighlightForced should preserve cursor control sequences, got: %q", result)
	}

	// Should still contain highlighting (ANSI color codes)
	if !strings.Contains(result, "\033[") {
		t.Error("result should contain ANSI codes")
	}
}

func TestHasANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"plain text", false},
		{"\033[31mred\033[0m", true},
		{"\033[1mtext", true},
		{"no escape", false},
		{"", false},
	}

	for _, tt := range tests {
		result := HasANSI(tt.input)
		if result != tt.expected {
			t.Errorf("HasANSI(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme == nil {
		t.Fatal("DefaultTheme() returned nil")
	}

	// Check that important token types have colors
	tokenTypes := []lexer.TokenType{
		lexer.TokenCommand,
		lexer.TokenSection,
		lexer.TokenProtocol,
		lexer.TokenAction,
		lexer.TokenInterface,
		lexer.TokenIPv4,
		lexer.TokenIPv4Prefix,
		lexer.TokenString,
		lexer.TokenComment,
	}

	for _, tt := range tokenTypes {
		color := theme.GetColor(tt)
		if color == "" {
			t.Errorf("DefaultTheme should have color for %v", tt)
		}
	}
}

func TestSolarizedTheme(t *testing.T) {
	theme := SolarizedDarkTheme()
	if theme == nil {
		t.Fatal("SolarizedDarkTheme() returned nil")
	}

	// Spot check a few colors
	if theme.GetColor(lexer.TokenCommand) == "" {
		t.Error("should have command color")
	}
}

func TestMonokaiTheme(t *testing.T) {
	theme := MonokaiTheme()
	if theme == nil {
		t.Fatal("MonokaiTheme() returned nil")
	}

	if theme.GetColor(lexer.TokenSection) == "" {
		t.Error("should have section color")
	}
}

func TestNordTheme(t *testing.T) {
	theme := NordTheme()
	if theme == nil {
		t.Fatal("NordTheme() returned nil")
	}

	if theme.GetColor(lexer.TokenInterface) == "" {
		t.Error("should have interface color")
	}
}

func TestTokyoNightTheme(t *testing.T) {
	theme := TokyoNightTheme()
	if theme == nil {
		t.Fatal("TokyoNightTheme() returned nil")
	}

	// Check that important token types have colors
	tokenTypes := []lexer.TokenType{
		lexer.TokenCommand,
		lexer.TokenSection,
		lexer.TokenProtocol,
		lexer.TokenAction,
		lexer.TokenInterface,
		lexer.TokenIPv4,
		lexer.TokenString,
	}

	for _, tt := range tokenTypes {
		color := theme.GetColor(tt)
		if color == "" {
			t.Errorf("TokyoNightTheme should have color for %v", tt)
		}
	}
}

func TestVibrantTheme(t *testing.T) {
	theme := VibrantTheme()
	if theme == nil {
		t.Fatal("VibrantTheme() returned nil")
	}

	if theme.GetColor(lexer.TokenCommand) == "" {
		t.Error("should have command color")
	}
}

func TestDefaultThemeIsTokyoNight(t *testing.T) {
	// Verify DefaultTheme returns the same colors as TokyoNightTheme
	defaultTheme := DefaultTheme()
	tokyoTheme := TokyoNightTheme()

	// Check a few colors match
	if defaultTheme.GetColor(lexer.TokenCommand) != tokyoTheme.GetColor(lexer.TokenCommand) {
		t.Error("DefaultTheme should match TokyoNightTheme")
	}
	if defaultTheme.GetColor(lexer.TokenSection) != tokyoTheme.GetColor(lexer.TokenSection) {
		t.Error("DefaultTheme should match TokyoNightTheme for sections")
	}
}

func TestThemeSetColor(t *testing.T) {
	theme := DefaultTheme()

	newColor := "\033[35m" // magenta
	theme.SetColor(lexer.TokenCommand, newColor)

	if theme.GetColor(lexer.TokenCommand) != newColor {
		t.Error("SetColor should update the color")
	}
}

func TestThemeGetColorUnknown(t *testing.T) {
	theme := DefaultTheme()

	// TokenType 999 doesn't exist
	color := theme.GetColor(lexer.TokenType(999))
	if color != "" {
		t.Error("unknown token type should return empty string")
	}
}

func TestColor256(t *testing.T) {
	result := Color256(208)
	expected := "\033[38;5;208m"
	if result != expected {
		t.Errorf("Color256(208) = %q, want %q", result, expected)
	}
}

func TestRGB(t *testing.T) {
	result := RGB(255, 128, 0)
	expected := "\033[38;2;255;128;0m"
	if result != expected {
		t.Errorf("RGB(255,128,0) = %q, want %q", result, expected)
	}
}

func TestHighlightPreservesContent(t *testing.T) {
	h := New()

	configs := []string{
		"set system host-name core-router-01",
		"set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24",
		"set protocols bgp group external neighbor 10.0.0.1 peer-as 65000",
		"set firewall family inet filter protect term 1 then accept",
		"# This is a comment",
		`set interfaces ge-0/0/0 description "Uplink to ISP"`,
	}

	for _, config := range configs {
		result := h.Highlight(config)
		stripped := StripANSI(result)
		if stripped != config {
			t.Errorf("content not preserved:\ninput:    %q\nstripped: %q", config, stripped)
		}
	}
}

func TestHighlightHierarchicalConfig(t *testing.T) {
	h := New()

	input := `system {
    host-name router;
    services {
        ssh;
    }
}`

	result := h.Highlight(input)

	// Should have highlighting
	if !HasANSI(result) {
		t.Error("hierarchical config should be highlighted")
	}

	// Content should be preserved
	stripped := StripANSI(result)
	if stripped != input {
		t.Errorf("content not preserved")
	}
}

func TestHighlightSetStyleConfig(t *testing.T) {
	h := New()

	input := `set system host-name router
set interfaces ge-0/0/0 unit 0 family inet address 10.0.0.1/24
set protocols ospf area 0.0.0.0 interface ge-0/0/0.0`

	result := h.Highlight(input)

	if !HasANSI(result) {
		t.Error("set-style config should be highlighted")
	}

	stripped := StripANSI(result)
	if stripped != input {
		t.Errorf("content not preserved")
	}
}
