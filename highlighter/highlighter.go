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

// NewWithTheme creates a new Highlighter with a specific theme
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
	if !h.looksLikeJunOS(cleaned) {
		return input
	}

	return h.highlightTokensCleaned(cleaned)
}

// HighlightForced applies syntax highlighting without checking if input looks like JunOS.
func (h *Highlighter) HighlightForced(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}
	return h.highlightTokens(input)
}

// highlightTokens tokenizes and colorizes the input while preserving cursor control sequences
func (h *Highlighter) highlightTokens(input string) string {
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
			highlighted := h.highlightTokensCleaned(seg.text)
			buf.WriteString(highlighted)
		}
	}
	return buf.String()
}

// highlightTokensCleaned tokenizes and colorizes already-cleaned input
func (h *Highlighter) highlightTokensCleaned(cleaned string) string {
	lex := lexer.New(cleaned)
	tokens := lex.Tokenize()
	return h.renderTokens(tokens)
}

// renderTokens applies theme colors to a slice of tokens and returns the colorized string
func (h *Highlighter) renderTokens(tokens []lexer.Token) string {
	h.mu.RLock()
	theme := h.theme
	h.mu.RUnlock()

	var buf bytes.Buffer
	for _, token := range tokens {
		color := theme.GetColor(token.Type)
		if color != "" {
			buf.WriteString(color)
			buf.WriteString(token.Value)
			buf.WriteString(Reset)
		} else {
			buf.WriteString(token.Value)
		}
	}
	return buf.String()
}

// HighlightLine highlights a single line (useful for streaming)
func (h *Highlighter) HighlightLine(line string) string {
	return h.Highlight(line)
}

// HighlightLines highlights multiple lines preserving line structure
func (h *Highlighter) HighlightLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = h.Highlight(line)
	}
	return result
}

// JunOS detection indicator lists
var (
	configIndicators = []string{
		"set ", "delete ", "show ", "edit ",
		"interfaces {", "system {", "protocols {",
		"routing-options", "policy-options", "firewall {",
		"security {", "groups {", "vlans {",
		"ge-", "xe-", "et-", "ae", "lo0",
		"family inet", "unit ", "vlan-id",
		"ospf", "bgp", "neighbor", "group",
	}

	showIndicators = []string{
		"establ", "idle", "full", "2way",
		"inet.0", "inet6.0", "mpls.0", "bgp.evpn",
		"bgp summary", "ospf neighbor", "interface terse",
		"physical interface", "logical interface",
		"routing table", "flaps", "up/dn",
		"state:", "admin link", "outq",
	}

	commandPrefixes = []string{"set ", "delete ", "show ", "edit ", "request ", "##"}
)

// looksLikeJunOS performs a quick check to see if text appears to be JunOS config or show output
func (h *Highlighter) looksLikeJunOS(input string) bool {
	if h.isPromptLine(input) {
		return true
	}

	lower := strings.ToLower(input)

	if h.hasConfigIndicators(lower) {
		return true
	}

	if h.hasShowIndicators(lower) {
		return true
	}

	if h.hasStructuralPatterns(input, lower) {
		return true
	}

	if h.startsWithCommand(input) {
		return true
	}

	return false
}

// isPromptLine checks if the input looks like a JunOS CLI prompt
func (h *Highlighter) isPromptLine(input string) bool {
	// Check for JunOS CLI prompts (user@hostname> or user@hostname#, possibly with command)
	if lexer.IsPrompt(input) {
		return true
	}

	// Check for prompt-like patterns even with commands after them
	// Pattern: something@something followed by > or #
	if !strings.Contains(input, "@") {
		return false
	}
	if !strings.Contains(input, ">") && !strings.Contains(input, "#") {
		return false
	}

	// Verify it looks like user@host pattern
	atIdx := strings.Index(input, "@")
	if atIdx <= 0 || atIdx >= len(input)-1 {
		return false
	}

	// Check there's alphanumeric before and after @
	beforeAt := input[atIdx-1]
	afterAt := input[atIdx+1]
	validBefore := isAlphanumericOrDash(beforeAt)
	validAfter := isAlphanumeric(afterAt)

	return validBefore && validAfter
}

// hasConfigIndicators checks for common JunOS config keywords/patterns
func (h *Highlighter) hasConfigIndicators(lower string) bool {
	for _, indicator := range configIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}

// hasShowIndicators checks for show command output patterns
func (h *Highlighter) hasShowIndicators(lower string) bool {
	for _, indicator := range showIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}

// hasStructuralPatterns checks for typical JunOS structure patterns
func (h *Highlighter) hasStructuralPatterns(input, lower string) bool {
	if !strings.Contains(input, "{\n") && !strings.Contains(input, ";") {
		return false
	}
	// Could be config, but verify with more checks
	return strings.Contains(lower, "version ") ||
		strings.Contains(input, "#") ||
		strings.Contains(lower, "host-name")
}

// startsWithCommand checks if line starts with common JunOS commands
func (h *Highlighter) startsWithCommand(input string) bool {
	trimmed := strings.TrimSpace(input)
	lowerTrimmed := strings.ToLower(trimmed)
	for _, prefix := range commandPrefixes {
		if strings.HasPrefix(lowerTrimmed, prefix) {
			return true
		}
	}
	return false
}

// isAlphanumeric checks if a byte is a letter or digit
func isAlphanumeric(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

// isAlphanumericOrDash checks if a byte is a letter, digit, or dash
func isAlphanumericOrDash(ch byte) bool {
	return isAlphanumeric(ch) || ch == '-'
}

// HighlightShowOutput highlights show command output specifically using show mode.
func (h *Highlighter) HighlightShowOutput(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}

	lex := lexer.New(input)
	lex.SetParseMode(lexer.ParseModeShow)
	tokens := lex.Tokenize()
	return h.renderTokens(tokens)
}

// segment represents either an escape sequence or text content
type segment struct {
	text     string
	isEscape bool
}

// CSI sequence byte range constants
const (
	csiParamStart = 0x20 // Space - start of parameter/intermediate bytes
	csiParamEnd   = 0x3F // ? - end of parameter bytes
	csiFinalStart = 0x40 // @ - start of final bytes
	csiFinalEnd   = 0x7E // ~ - end of final bytes
	csiIntermEnd  = 0x2F // / - end of intermediate bytes
	escapeChar    = '\033'
	csiBracket    = '['
)

// isCSIParamByte checks if byte is a CSI parameter or intermediate byte (0x20-0x3F)
func isCSIParamByte(b byte) bool {
	return b >= csiParamStart && b <= csiParamEnd
}

// isCSIFinalByte checks if byte is a CSI final byte (0x40-0x7E)
func isCSIFinalByte(b byte) bool {
	return b >= csiFinalStart && b <= csiFinalEnd
}

// isCSIIntermediateByte checks if byte is a CSI intermediate byte (0x20-0x2F)
func isCSIIntermediateByte(b byte) bool {
	return b >= csiParamStart && b <= csiIntermEnd
}

// skipCSISequence skips a CSI sequence starting at position i (after \033[)
// Returns the new position after the sequence
func skipCSISequence(input string, i int) int {
	// Skip parameter bytes (0x30-0x3F) and intermediate bytes (0x20-0x2F)
	for i < len(input) && isCSIParamByte(input[i]) {
		i++
	}
	// Skip the final byte (0x40-0x7E)
	if i < len(input) && isCSIFinalByte(input[i]) {
		i++
	}
	return i
}

// skipOtherEscapeSequence skips non-CSI escape sequences
// Returns the new position after the sequence
func skipOtherEscapeSequence(input string, i int) int {
	// Skip intermediate bytes (0x20-0x2F)
	for i < len(input) && isCSIIntermediateByte(input[i]) {
		i++
	}
	// Skip final byte
	if i < len(input) {
		i++
	}
	return i
}

// extractSegments splits input into escape sequences and text segments
// This allows us to preserve cursor control sequences while highlighting text
func extractSegments(input string) []segment {
	var segments []segment
	var textBuf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar && i+1 < len(input) && input[i+1] == csiBracket {
			// Flush any accumulated text
			if textBuf.Len() > 0 {
				segments = append(segments, segment{text: textBuf.String(), isEscape: false})
				textBuf.Reset()
			}

			// Extract CSI sequence
			start := i
			i = skipCSISequence(input, i+2) // +2 to skip \033[
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
// Handles both SGR codes (colors, ending in 'm') and CSI sequences (cursor control, etc.)
func StripANSI(input string) string {
	var buf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar && i+1 < len(input) && input[i+1] == csiBracket {
			// CSI sequence: \033[ followed by params and a final byte
			i = skipCSISequence(input, i+2) // +2 to skip \033[
			continue
		}
		if input[i] == escapeChar {
			// Other escape sequence (OSC, etc.)
			i = skipOtherEscapeSequence(input, i+1) // +1 to skip \033
			continue
		}
		buf.WriteByte(input[i])
		i++
	}

	return buf.String()
}

// HasANSI checks if the input contains ANSI escape codes
func HasANSI(input string) bool {
	return strings.Contains(input, "\033[")
}
