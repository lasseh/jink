package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestCLIHelp tests that --help returns usage information
func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	output, _ := cmd.CombinedOutput()

	outStr := string(output)
	if !strings.Contains(outStr, "jink") {
		t.Error("help should contain program name")
	}
	if !strings.Contains(outStr, "USAGE") {
		t.Error("help should contain USAGE section")
	}
	if !strings.Contains(outStr, "--theme") {
		t.Error("help should mention --theme option")
	}
}

// TestCLIVersion tests that --version returns version information
func TestCLIVersion(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	outStr := string(output)
	if !strings.Contains(outStr, "jink version") {
		t.Error("version output should contain 'jink version'")
	}
}

// TestCLIPipedInput tests highlighting of piped input
func TestCLIPipedInput(t *testing.T) {
	input := "set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24"

	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("piped input failed: %v\nOutput: %s", err, output)
	}

	outStr := string(output)

	// Should contain ANSI codes (highlighting)
	if !strings.Contains(outStr, "\033[") {
		t.Error("output should contain ANSI escape codes for highlighting")
	}

	// Should still contain the key parts of the input
	if !strings.Contains(outStr, "ge-0/0/0") {
		t.Error("output should contain interface name")
	}
	if !strings.Contains(outStr, "192.168.1.1/24") {
		t.Error("output should contain IP address")
	}
}

// TestCLINoHighlight tests the --no-highlight option
func TestCLINoHighlight(t *testing.T) {
	input := "set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24"

	cmd := exec.Command("go", "run", ".", "--no-highlight")
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("no-highlight option failed: %v", err)
	}

	outStr := string(output)

	// Should NOT contain ANSI codes
	if strings.Contains(outStr, "\033[") {
		t.Error("--no-highlight output should not contain ANSI codes")
	}

	// Output should match input exactly
	if strings.TrimSpace(outStr) != input {
		t.Errorf("expected %q, got %q", input, strings.TrimSpace(outStr))
	}
}

// TestCLIThemeOption tests theme selection
func TestCLIThemeOption(t *testing.T) {
	themes := []string{"default", "solarized", "monokai", "nord"}
	input := "set system host-name router"

	for _, theme := range themes {
		t.Run(theme, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".", "-t", theme)
			cmd.Stdin = strings.NewReader(input)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("theme %s failed: %v", theme, err)
			}

			outStr := string(output)
			// All themes should produce highlighted output
			if !strings.Contains(outStr, "\033[") {
				t.Errorf("theme %s should produce ANSI output", theme)
			}
		})
	}
}

// TestCLIShortFlags tests short flag versions
func TestCLIShortFlags(t *testing.T) {
	// Test -h (help)
	cmd := exec.Command("go", "run", ".", "-h")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "USAGE") {
		t.Error("-h should show help")
	}

	// Test -v (version)
	cmd = exec.Command("go", "run", ".", "-v")
	output, _ = cmd.CombinedOutput()
	if !strings.Contains(string(output), "version") {
		t.Error("-v should show version")
	}

	// Test -n (no-highlight)
	input := "set system host-name router"
	cmd = exec.Command("go", "run", ".", "-n")
	cmd.Stdin = strings.NewReader(input)
	output, _ = cmd.CombinedOutput()
	if strings.Contains(string(output), "\033[") {
		t.Error("-n should disable highlighting")
	}
}

// TestCLIMultilineInput tests highlighting of multi-line config
func TestCLIMultilineInput(t *testing.T) {
	input := `set system host-name router
set interfaces ge-0/0/0 unit 0 family inet address 10.0.0.1/24
set protocols ospf area 0.0.0.0 interface ge-0/0/0.0`

	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multiline input failed: %v", err)
	}

	outStr := string(output)

	// Check all lines are present
	if !strings.Contains(outStr, "host-name") {
		t.Error("should contain host-name")
	}
	if !strings.Contains(outStr, "ge-0/0/0") {
		t.Error("should contain interface")
	}
	if !strings.Contains(outStr, "ospf") {
		t.Error("should contain ospf")
	}

	// Check has highlighting
	if !strings.Contains(outStr, "\033[") {
		t.Error("should have ANSI highlighting")
	}
}

// TestCLIHierarchicalConfig tests hierarchical config format
func TestCLIHierarchicalConfig(t *testing.T) {
	input := `system {
    host-name router;
    services {
        ssh;
    }
}`

	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hierarchical config failed: %v", err)
	}

	outStr := string(output)

	// Should be highlighted
	if !strings.Contains(outStr, "\033[") {
		t.Error("hierarchical config should be highlighted")
	}

	// Should preserve structure
	if !strings.Contains(outStr, "system") {
		t.Error("should contain 'system'")
	}
	if !strings.Contains(outStr, "host-name") {
		t.Error("should contain 'host-name'")
	}
}

// TestCLIBinaryBuilds tests that the binary builds correctly
func TestCLIBinaryBuilds(t *testing.T) {
	// Create temp directory for binary
	tmpDir, err := os.MkdirTemp("", "jink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	binaryPath := tmpDir + "/jink-test"

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v\nStderr: %s", err, stderr.String())
	}

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("binary was not created")
	}

	// Test binary works
	testCmd := exec.Command(binaryPath, "--version")
	output, err := testCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary --version failed: %v", err)
	}
	if !strings.Contains(string(output), "jink version") {
		t.Error("binary should output version")
	}
}
