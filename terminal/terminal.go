package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/lasseh/jink/highlighter"
	"golang.org/x/term"
)

// Buffer size constants
const (
	readBufferSize = 32 * 1024 // Size of the read buffer from PTY
	lineBufferSize = 4096      // Initial capacity for line buffer
	lineFlushLimit = 4000      // Flush line buffer when it exceeds this size
)

var (
	debug   bool
	debugMu sync.RWMutex
)

// SetDebug enables or disables debug output to stderr
func SetDebug(enabled bool) {
	debugMu.Lock()
	defer debugMu.Unlock()
	debug = enabled
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	debugMu.RLock()
	defer debugMu.RUnlock()
	return debug
}

// Terminal wraps a command in a PTY and applies syntax highlighting to its output.
type Terminal struct {
	cmd         *exec.Cmd
	pty         *os.File
	highlighter *highlighter.Highlighter
	enabled     bool
}

// New creates a new Terminal for the given command
func New(name string, args ...string) *Terminal {
	cmd := exec.Command(name, args...)
	return &Terminal{
		cmd:         cmd,
		highlighter: highlighter.New(),
		enabled:     true,
	}
}

// SetTheme changes the highlighting theme
func (t *Terminal) SetTheme(theme *highlighter.Theme) {
	t.highlighter.SetTheme(theme)
}

// SetEnabled enables or disables highlighting
func (t *Terminal) SetEnabled(enabled bool) {
	t.enabled = enabled
}

// Run starts the command and processes its output with highlighting.
func (t *Terminal) Run() error {
	// Start the command with a PTY
	ptmx, err := pty.Start(t.cmd)
	if err != nil {
		return fmt.Errorf("starting pty: %w", err)
	}
	t.pty = ptmx
	defer func() {
		if err := ptmx.Close(); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Error closing pty: %v\n", err)
		}
	}()

	// Handle terminal resize with proper cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	sigDone := make(chan struct{})
	go func() {
		defer close(sigDone)
		for range sigCh {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil && IsDebug() {
				fmt.Fprintf(os.Stderr, "[DEBUG] Error resizing pty: %v\n", err)
			}
		}
	}()
	// Cleanup signal handler when done
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
		<-sigDone // Wait for goroutine to exit
	}()

	// Trigger initial resize
	sigCh <- syscall.SIGWINCH

	// Put terminal into raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("setting raw mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Error restoring terminal: %v\n", err)
		}
	}()

	// Create channel for coordination
	done := make(chan struct{})

	// Copy stdin to PTY
	go func() {
		if _, err := io.Copy(ptmx, os.Stdin); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Error copying stdin: %v\n", err)
		}
	}()

	// Copy PTY to stdout with highlighting
	go func() {
		t.processOutput(ptmx, os.Stdout)
		close(done)
	}()

	// Wait for command to finish
	<-done
	if err := t.cmd.Wait(); err != nil {
		return fmt.Errorf("command finished: %w", err)
	}
	return nil
}

// processOutput reads from the PTY and writes highlighted output.
// Both complete lines and partial lines (prompts) are highlighted.
// Cursor control characters (like \r) are preserved to allow command-line editing.
func (t *Terminal) processOutput(r io.Reader, w io.Writer) {
	buf := make([]byte, readBufferSize)
	lineBuf := make([]byte, 0, lineBufferSize)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			data := buf[:n]

			if IsDebug() {
				fmt.Fprintf(os.Stderr, "\n[DEBUG] Read %d bytes: %q\n", n, data)
			}

			// Process byte by byte
			for i := 0; i < n; i++ {
				b := data[i]
				lineBuf = append(lineBuf, b)

				// Flush on newline or when buffer gets large
				if b == '\n' || len(lineBuf) > lineFlushLimit {
					t.writeOutput(w, lineBuf)
					lineBuf = lineBuf[:0]
				}
			}

			// Flush partial lines (prompts) - also highlighted
			// Cursor control chars like \r are preserved by the lexer
			if len(lineBuf) > 0 {
				t.writeOutput(w, lineBuf)
				lineBuf = lineBuf[:0]
			}
		}

		if err != nil {
			if IsDebug() && err != io.EOF {
				fmt.Fprintf(os.Stderr, "[DEBUG] Read error: %v\n", err)
			}
			break
		}
	}
}

// writeOutput writes data to the writer, optionally highlighting it.
func (t *Terminal) writeOutput(w io.Writer, data []byte) {
	var output string
	if t.enabled {
		output = t.highlighter.HighlightForced(string(data))
		if IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Highlight: %q -> %q\n", data, output)
		}
	} else {
		output = string(data)
	}

	if _, err := w.Write([]byte(output)); err != nil && IsDebug() {
		fmt.Fprintf(os.Stderr, "[DEBUG] Write error: %v\n", err)
	}
}
