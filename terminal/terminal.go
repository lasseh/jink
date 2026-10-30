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
	"golang.org/x/term"

	"github.com/lasseh/jink/highlighter"
)

// readBufferSize is the size of the read buffer from the PTY.
const readBufferSize = 32 * 1024

var (
	debug   bool
	debugMu sync.RWMutex
)

// SetDebug enables or disables debug output to stderr.
func SetDebug(enabled bool) {
	debugMu.Lock()
	defer debugMu.Unlock()
	debug = enabled
}

// IsDebug returns whether debug mode is enabled.
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

// New creates a new Terminal for the given command.
func New(name string, args ...string) *Terminal {
	// The wrapped command is long-lived; its lifecycle (start, resize signals,
	// wait) is owned by Run(). Threading a context through this constructor
	// would change the public API without giving us cancellation we can act on.
	cmd := exec.Command(name, args...) //nolint:noctx // lifecycle managed by Run()
	return &Terminal{
		cmd:         cmd,
		highlighter: highlighter.New(),
		enabled:     true,
	}
}

// SetTheme changes the highlighting theme.
func (t *Terminal) SetTheme(theme *highlighter.Theme) {
	t.highlighter.SetTheme(theme)
}

// SetEnabled enables or disables highlighting.
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

// processOutput reads from the PTY and writes the output, highlighted unless
// disabled. Each read is written straight away, so prompts and keystroke echoes
// appear immediately.
func (t *Terminal) processOutput(r io.Reader, w io.Writer) {
	var stream *highlighter.Stream
	if t.enabled {
		stream = t.highlighter.NewStream(w)
		w = stream
	}

	buf := make([]byte, readBufferSize)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if IsDebug() {
				fmt.Fprintf(os.Stderr, "\n[DEBUG] Read %d bytes: %q\n", n, buf[:n])
			}
			if _, err := w.Write(buf[:n]); err != nil && IsDebug() {
				fmt.Fprintf(os.Stderr, "[DEBUG] Write error: %v\n", err)
			}
		}

		if err != nil {
			if IsDebug() && err != io.EOF {
				fmt.Fprintf(os.Stderr, "[DEBUG] Read error: %v\n", err)
			}
			break
		}
	}

	if stream != nil {
		if err := stream.Flush(); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Write error: %v\n", err)
		}
	}
}
