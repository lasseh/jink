package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"

	"github.com/lasseh/jink/highlighter"
	"github.com/lasseh/jink/lexer"
	"github.com/lasseh/jink/terminal"
)

// version is set via ldflags at build time (see Makefile).
var version = "dev"

const usageHeader = `jink - ink your JunOS config

USAGE:
    jink ssh user@router          # Interactive SSH with highlighting
    cat config.conf | jink        # Highlight a config file
    jink -t monokai ssh router    # Use a different theme

OPTIONS:
    -f, --force           Always highlight (skip auto-detection)
    -t, --theme <name>    Color theme (see THEMES below)
    -n, --no-highlight    Disable highlighting (pass-through mode)
        --color <when>    When to colorize: auto (default), always, never
                          auto = only when stdout is a terminal and NO_COLOR is unset
    -v, --version         Show version
    -h, --help            Show this help

THEMES:
`

// usage returns the help text, with the theme list from the registry.
func usage() string {
	var b strings.Builder
	b.WriteString(usageHeader)
	for _, t := range highlighter.Themes() {
		fmt.Fprintf(&b, "    %-11s - %s\n", t.Name, t.Description)
	}
	b.WriteString("\n")
	return b.String()
}

func main() {
	// Custom flag handling to support both short and long forms
	var (
		themeName   string
		colorWhen   string
		noHighlight bool
		forceHL     bool
		showVersion bool
		showHelp    bool
		debug       bool
	)

	flag.StringVar(&themeName, "theme", "default", "Color theme")
	flag.StringVar(&themeName, "t", "default", "Color theme (shorthand)")
	flag.StringVar(&colorWhen, "color", "auto", "When to colorize: auto, always, never")
	flag.BoolVar(&noHighlight, "no-highlight", false, "Disable highlighting")
	flag.BoolVar(&noHighlight, "n", false, "Disable highlighting (shorthand)")
	flag.BoolVar(&forceHL, "force", false, "Force highlighting (skip detection)")
	flag.BoolVar(&forceHL, "f", false, "Force highlighting (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showVersion, "v", false, "Show version (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help (shorthand)")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.BoolVar(&debug, "d", false, "Enable debug output (shorthand)")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage())
	}

	flag.Parse()

	if showHelp {
		fmt.Print(usage())
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("jink version %s\n", version)
		os.Exit(0)
	}

	// Select theme
	theme, ok := highlighter.LookupTheme(themeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown theme %q (available: %s)\n",
			themeName, strings.Join(highlighter.ThemeNames(), ", "))
		os.Exit(1)
	}

	args := flag.Args()

	// Enable debug mode
	terminal.SetDebug(debug)

	// Decide whether to colorize: explicit flags win, otherwise auto-detect
	// based on stdout being a terminal and the NO_COLOR convention.
	_, noColorSet := os.LookupEnv("NO_COLOR")
	stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))
	disabled := !resolveHighlight(colorWhen, noHighlight, stdoutIsTTY, noColorSet)

	// If no command provided, read from stdin and highlight
	if len(args) == 0 {
		if err := highlightStdin(theme, disabled, forceHL); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run command with PTY terminal
	if err := runWithTerminal(args, theme, disabled); err != nil {
		// A non-zero exit from the wrapped command is its own outcome, not a
		// jink error: propagate the code silently. Report anything else.
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCodeFor(err))
	}
}

// resolveHighlight decides whether to colorize output. --no-highlight and
// --color=never disable it; --color=always forces it on (even into a pipe or
// when NO_COLOR is set); the default "auto" colorizes only when stdout is a
// terminal and NO_COLOR is unset (https://no-color.org).
func resolveHighlight(when string, noHighlight, stdoutIsTTY, noColorSet bool) bool {
	if noHighlight {
		return false
	}
	switch when {
	case "never":
		return false
	case "always":
		return true
	default: // "auto"
		return stdoutIsTTY && !noColorSet
	}
}

// exitCodeFor maps a runWithTerminal error to a process exit code: the wrapped
// command's own code when it ran and exited non-zero, 1 for any other failure,
// 0 for nil.
func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func highlightStdin(theme *highlighter.Theme, disabled, force bool) error {
	// Check if stdin is a terminal (no pipe)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Interactive mode - show help
		fmt.Print(usage())
		return nil
	}

	hl := highlighter.NewWithTheme(theme)
	reader := bufio.NewReader(os.Stdin)

	// Track if we've detected JunOS content (sticky detection)
	detectedJunOS := force

	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			switch {
			case disabled:
				fmt.Print(line)
			case detectedJunOS || force:
				// Force mode or already detected - highlight everything
				fmt.Print(hl.HighlightForced(line))
			default:
				// Auto-detect mode - once a line looks like JunOS, enable for
				// all future lines
				detectedJunOS = lexer.IsJunOS(highlighter.StripANSI(line))
				fmt.Print(hl.Highlight(line))
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func runWithTerminal(args []string, theme *highlighter.Theme, disabled bool) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	t := terminal.New(args[0], args[1:]...)
	t.SetTheme(theme)
	t.SetEnabled(!disabled)

	return t.Run()
}
