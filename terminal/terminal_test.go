package terminal

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/lasseh/jink/highlighter"
)

func TestSetDebug(t *testing.T) {
	// Reset to known state
	SetDebug(false)

	if IsDebug() {
		t.Error("expected debug to be false after SetDebug(false)")
	}

	SetDebug(true)
	if !IsDebug() {
		t.Error("expected debug to be true after SetDebug(true)")
	}

	SetDebug(false)
	if IsDebug() {
		t.Error("expected debug to be false after SetDebug(false)")
	}
}

func TestDebugConcurrency(t *testing.T) {
	// Test that SetDebug and IsDebug are safe for concurrent use
	var wg sync.WaitGroup

	// Run multiple goroutines toggling and reading debug
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetDebug(true)
			SetDebug(false)
		}()
		go func() {
			defer wg.Done()
			_ = IsDebug()
		}()
	}

	wg.Wait()
	// If we get here without race detector complaints, the test passes
}

func TestNew(t *testing.T) {
	term := New("echo", "hello")
	if term == nil {
		t.Fatal("New() returned nil")
	}
	if term.cmd == nil {
		t.Error("cmd should not be nil")
	}
	if term.highlighter == nil {
		t.Error("highlighter should not be nil")
	}
	if !term.enabled {
		t.Error("highlighting should be enabled by default")
	}
}

func TestSetTheme(t *testing.T) {
	term := New("echo", "test")
	theme := highlighter.MonokaiTheme()
	term.SetTheme(theme)
	// No panic means success - theme is private
}

func TestSetEnabled(t *testing.T) {
	term := New("echo", "test")

	if !term.enabled {
		t.Error("should be enabled by default")
	}

	term.SetEnabled(false)
	if term.enabled {
		t.Error("should be disabled after SetEnabled(false)")
	}

	term.SetEnabled(true)
	if !term.enabled {
		t.Error("should be enabled after SetEnabled(true)")
	}
}

func TestWriteOutput(t *testing.T) {
	term := New("echo", "test")

	tests := []struct {
		name     string
		enabled  bool
		input    string
		contains string
	}{
		{
			name:     "highlighting enabled with JunOS",
			enabled:  true,
			input:    "set interfaces ge-0/0/0",
			contains: "\033[", // Should have ANSI codes
		},
		{
			name:     "highlighting disabled",
			enabled:  false,
			input:    "set interfaces ge-0/0/0",
			contains: "set interfaces", // Should be unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term.SetEnabled(tt.enabled)
			var buf bytes.Buffer
			term.writeOutput(&buf, []byte(tt.input))
			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("output %q should contain %q", output, tt.contains)
			}
		})
	}
}

func TestProcessOutputBasic(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false) // Disable for simpler testing

	input := "line1\nline2\nline3\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != input {
		t.Errorf("expected %q, got %q", input, output.String())
	}
}

func TestProcessOutputPartialLine(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	// Simulate a prompt without newline
	input := "user@router> "
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != input {
		t.Errorf("expected %q, got %q", input, output.String())
	}
}

func TestProcessOutputLargeBuffer(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	// Create a line larger than lineFlushLimit
	largeLine := strings.Repeat("x", lineFlushLimit+100)
	reader := strings.NewReader(largeLine)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != largeLine {
		t.Errorf("expected %d chars, got %d chars", len(largeLine), len(output.String()))
	}
}

func TestProcessOutputWithHighlighting(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(true)

	input := "set interfaces ge-0/0/0\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	// Output should contain ANSI codes
	if !strings.Contains(output.String(), "\033[") {
		t.Error("output should contain ANSI escape codes when highlighting is enabled")
	}

	// Content should be preserved
	stripped := highlighter.StripANSI(output.String())
	if stripped != input {
		t.Errorf("stripped output %q should equal input %q", stripped, input)
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are reasonable
	if readBufferSize <= 0 {
		t.Error("readBufferSize should be positive")
	}
	if lineBufferSize <= 0 {
		t.Error("lineBufferSize should be positive")
	}
	if lineFlushLimit <= 0 {
		t.Error("lineFlushLimit should be positive")
	}
	if lineFlushLimit >= readBufferSize {
		t.Error("lineFlushLimit should be less than readBufferSize")
	}
}

// mockReader allows testing error handling in processOutput
type mockReader struct {
	data    []byte
	pos     int
	errAt   int
	errOnce error
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	if m.errAt >= 0 && m.pos >= m.errAt && m.errOnce != nil {
		err := m.errOnce
		m.errOnce = nil
		return 0, err
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func TestProcessOutputHandlesErrors(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	// Test that processOutput handles EOF gracefully
	reader := &mockReader{data: []byte("test\n"), errAt: -1}
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != "test\n" {
		t.Errorf("expected 'test\\n', got %q", output.String())
	}
}
