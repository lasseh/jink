package lexer

import (
	"slices"
	"strings"
)

// Detection signals. All of them are matched against lowercased text.
var (
	// modeConfigSignals and modeShowSignals are scored against each other to
	// pick a parse mode, so they must be specific: a false hit flips the
	// colors of a whole block.
	modeConfigSignals = []string{"set ", "delete ", "{", "}", ";", "host-name", "policy-statement"}
	modeShowSignals   = []string{
		"establ", "idle", "2way",
		"inet.0", "inet6.0", "bgp.evpn",
		"flaps", "up/dn",
		"physical interface", "logical interface",
	}

	// junosSignals only decide whether text is JunOS at all, where any single
	// hit is enough, so they can be broader.
	junosSignals = slices.Concat([]string{
		"set ", "delete ", "show ", "edit ",
		"interfaces {", "system {", "protocols {",
		"routing-options", "policy-options", "firewall {",
		"security {", "groups {", "vlans {",
		"ge-", "xe-", "et-", "ae", "lo0",
		"family inet", "unit ", "vlan-id",
		"ospf", "bgp", "neighbor", "group",

		"full", "mpls.0",
		"bgp summary", "ospf neighbor", "interface terse",
		"routing table", "state:", "admin link", "outq",
	}, modeShowSignals)

	commandPrefixes = []string{"set ", "delete ", "show ", "edit ", "request ", "##"}
)

// IsJunOS reports whether text looks like JunOS configuration or show output.
// It expects text without escape sequences.
func IsJunOS(text string) bool {
	if looksLikePrompt(text) {
		return true
	}

	lower := strings.ToLower(text)

	if containsAny(lower, junosSignals) {
		return true
	}

	if hasConfigStructure(text, lower) {
		return true
	}

	trimmed := strings.TrimSpace(lower)
	return slices.ContainsFunc(commandPrefixes, func(prefix string) bool {
		return strings.HasPrefix(trimmed, prefix)
	})
}

// DetectMode picks a parse mode from the signals in text, or returns
// ParseModeAuto when text doesn't favor either mode. A lone show signal isn't
// enough, because words like "up" also appear in configuration.
func DetectMode(text string) ParseMode {
	// Sample first N chars for detection - enough to see headers/commands
	// without processing entire large configs
	sample := text
	if len(sample) > parseModeDetectionSampleSize {
		sample = sample[:parseModeDetectionSampleSize]
	}
	lower := strings.ToLower(sample)

	configScore := countSignals(lower, modeConfigSignals)
	showScore := countSignals(lower, modeShowSignals)

	// Tabular data pattern (multiple spaces between words) - strong indicator
	// worth 2 points since tabular output is very characteristic of show commands
	if tabularPattern.MatchString(sample) {
		showScore += 2
	}

	switch {
	case showScore >= 2 && showScore > configScore:
		return ParseModeShow
	case configScore > 0:
		return ParseModeConfig
	default:
		return ParseModeAuto
	}
}

// looksLikePrompt checks for a prompt line, or for the user@host> shape with
// other text around it.
func looksLikePrompt(text string) bool {
	if IsPrompt(text) {
		return true
	}

	if !strings.ContainsAny(text, ">#") {
		return false
	}

	atIdx := strings.Index(text, "@")
	if atIdx <= 0 || atIdx >= len(text)-1 {
		return false
	}

	// Check there's alphanumeric before and after @
	before, after := text[atIdx-1], text[atIdx+1]
	return (isAlphanumeric(before) || before == '-') && isAlphanumeric(after)
}

// hasConfigStructure checks for typical JunOS structure patterns.
func hasConfigStructure(text, lower string) bool {
	if !strings.Contains(text, "{\n") && !strings.Contains(text, ";") {
		return false
	}
	// Could be config, but verify with more checks
	return strings.Contains(lower, "version ") ||
		strings.Contains(text, "#") ||
		strings.Contains(lower, "host-name")
}

func containsAny(text string, signals []string) bool {
	return slices.ContainsFunc(signals, func(signal string) bool {
		return strings.Contains(text, signal)
	})
}

func countSignals(text string, signals []string) int {
	count := 0
	for _, signal := range signals {
		if strings.Contains(text, signal) {
			count++
		}
	}
	return count
}

// isAlphanumeric checks if a byte is a letter or digit.
func isAlphanumeric(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}
