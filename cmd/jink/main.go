package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lasseh/jink/highlighter"
	"github.com/lasseh/jink/terminal"
)

// version is set via ldflags at build time (see Makefile)
var version = "dev"

const usage = `jink - ink your JunOS config

USAGE:
    jink ssh user@router          # Interactive SSH with highlighting
    cat config.conf | jink        # Highlight a config file
    jink -t monokai ssh router    # Use a different theme

OPTIONS:
    -f, --force           Always highlight (skip auto-detection)
    -t, --theme <name>    Color theme (see THEMES below)
    -n, --no-highlight    Disable highlighting (pass-through mode)
    -v, --version         Show version
    -h, --help            Show this help

THEMES:
    default     - Tokyo Night color scheme (default)
    tokyonight  - Tokyo Night color scheme
    vibrant     - Vibrant colors for dark terminals
    solarized   - Solarized Dark color scheme
    monokai     - Monokai-inspired colors
    nord        - Nord color palette
    catppuccin  - Catppuccin Mocha color scheme
    dracula     - Dracula color scheme
    gruvbox     - Gruvbox Dark color scheme
    onedark     - Atom One Dark color scheme

`

func main() {
	// Custom flag handling to support both short and long forms
	var (
		themeName   string
		noHighlight bool
		forceHL     bool
		showVersion bool
		showHelp    bool
		debug       bool
	)

	flag.StringVar(&themeName, "theme", "default", "Color theme")
	flag.StringVar(&themeName, "t", "default", "Color theme (shorthand)")
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
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if showHelp {
		fmt.Print(usage)
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("jink version %s\n", version)
		os.Exit(0)
	}

	// Select theme
	theme := highlighter.ThemeByName(strings.ToLower(themeName))

	args := flag.Args()

	// Enable debug mode
	terminal.SetDebug(debug)

	// If no command provided, read from stdin and highlight
	if len(args) == 0 {
		if err := highlightStdin(theme, noHighlight, forceHL); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run command with PTY terminal
	if err := runWithTerminal(args, theme, noHighlight); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func highlightStdin(theme *highlighter.Theme, disabled bool, force bool) error {
	// Check if stdin is a terminal (no pipe)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Interactive mode - show help
		fmt.Print(usage)
		return nil
	}

	hl := highlighter.NewWithTheme(theme)
	reader := bufio.NewReader(os.Stdin)

	// Track if we've detected JunOS content (sticky detection)
	detectedJunOS := force

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if disabled {
				fmt.Print(line)
			} else if detectedJunOS || force {
				// Force mode or already detected - highlight everything
				fmt.Print(hl.HighlightForced(line))
			} else {
				// Auto-detect mode - check if this looks like JunOS
				highlighted := hl.Highlight(line)
				if highlighted != line {
					// We got highlighting, so it's JunOS - enable for all future lines
					detectedJunOS = true
				}
				fmt.Print(highlighted)
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
