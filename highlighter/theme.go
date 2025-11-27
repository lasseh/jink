package highlighter

import (
	"strconv"

	"github.com/lasseh/jink/lexer"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// 256-color mode
	Color256Prefix = "\033[38;5;"
	Color256Suffix = "m"
)

// Color256 returns an ANSI escape for 256-color mode
func Color256(n int) string {
	return Color256Prefix + strconv.Itoa(n) + Color256Suffix
}

// RGB returns an ANSI escape for true color mode
func RGB(r, g, b int) string {
	return "\033[38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m"
}

// Palette defines the semantic colors used to build a theme.
// Each theme provides its own palette, and buildTheme maps these to token types.
type Palette struct {
	// Base colors
	Foreground string // default text, braces, identifiers
	Comment    string // comments, semicolons, dim text

	// Accent colors (semantic mapping to JunOS elements)
	Command   string // set, delete, show (bold)
	Section   string // system, interfaces (bold)
	Protocol  string // ospf, bgp, tcp
	Action    string // accept, reject (bold)
	Interface string // ge-0/0/0, ae0 (bold)
	IP        string // IP addresses
	Number    string // numbers, units
	String    string // quoted strings
	Keyword   string // other keywords
	Operator  string // operators
	ASN       string // AS numbers
	Community string // BGP communities
	Value     string // values after keywords
	Wildcard  string // wildcards (typically red)
	MAC       string // MAC addresses

	// State colors (for show output)
	StateGood    string // up, Establ (bold green)
	StateBad     string // down, Idle (bold red)
	StateWarning string // 2Way, ExStart (bold yellow)

	// Show output extras
	Duration      string // time durations
	RouteProtocol string // [BGP/170] (bold)
	TableName     string // inet.0 (bold)

	// Prompt colors
	PromptUser     string // username
	PromptAt       string // @ separator
	PromptHostOper string // hostname (operational mode)
	PromptHostConf string // hostname (config mode)
	PromptOper     string // > prompt
	PromptConf     string // # prompt
	PromptEdit     string // [edit ...] prefix
}

// buildTheme creates a Theme from a Palette by mapping semantic colors to token types.
func buildTheme(p Palette) *Theme {
	return &Theme{
		colors: map[lexer.TokenType]string{
			// Config tokens
			lexer.TokenCommand:    Bold + p.Command,
			lexer.TokenSection:    Bold + p.Section,
			lexer.TokenProtocol:   p.Protocol,
			lexer.TokenAction:     Bold + p.Action,
			lexer.TokenInterface:  Bold + p.Interface,
			lexer.TokenIPv4:       p.IP,
			lexer.TokenIPv4Prefix: p.IP,
			lexer.TokenIPv6:       p.IP,
			lexer.TokenIPv6Prefix: p.IP,
			lexer.TokenMAC:        p.MAC,
			lexer.TokenNumber:     p.Number,
			lexer.TokenString:     p.String,
			lexer.TokenComment:    Italic + p.Comment,
			lexer.TokenAnnotation: Italic + p.Comment,
			lexer.TokenBrace:      p.Foreground,
			lexer.TokenSemicolon:  p.Comment,
			lexer.TokenWildcard:   p.Wildcard,
			lexer.TokenIdentifier: p.Foreground,
			lexer.TokenKeyword:    p.Keyword,
			lexer.TokenOperator:   p.Operator,
			lexer.TokenUnit:       p.Number,
			lexer.TokenASN:        p.ASN,
			lexer.TokenCommunity:  p.Community,
			lexer.TokenValue:      p.Value,
			lexer.TokenText:       "",

			// Show output tokens
			lexer.TokenStateGood:     Bold + p.StateGood,
			lexer.TokenStateBad:      Bold + p.StateBad,
			lexer.TokenStateWarning:  Bold + p.StateWarning,
			lexer.TokenStateNeutral:  Dim + p.Comment,
			lexer.TokenColumnHeader:  Bold + p.Foreground,
			lexer.TokenStatusSymbol:  Bold + p.Protocol,
			lexer.TokenTimeDuration:  p.Duration,
			lexer.TokenPercentage:    p.StateGood,
			lexer.TokenByteSize:      p.Protocol,
			lexer.TokenRouteProtocol: Bold + p.RouteProtocol,
			lexer.TokenTableName:     Bold + p.TableName,

			// Prompt tokens
			lexer.TokenPromptUser:     p.PromptUser,
			lexer.TokenPromptAt:       p.PromptAt,
			lexer.TokenPromptHostOper: p.PromptHostOper,
			lexer.TokenPromptHostConf: p.PromptHostConf,
			lexer.TokenPromptOper:     p.PromptOper,
			lexer.TokenPromptConf:     p.PromptConf,
			lexer.TokenPromptEdit:     p.PromptEdit,

			// Diff tokens (git-style: green for add, red for remove)
			lexer.TokenDiffAdd:     Bold + p.StateGood,
			lexer.TokenDiffRemove:  Bold + p.StateBad,
			lexer.TokenDiffContext: Bold + p.Protocol,
		},
	}
}

// Theme defines ANSI color mappings for each token type.
// Use ThemeByName() to get a theme by name, or create custom themes
// by modifying an existing theme with SetColor().
type Theme struct {
	colors map[lexer.TokenType]string
}

// DefaultTheme returns the default theme (Tokyo Night)
func DefaultTheme() *Theme {
	return TokyoNightTheme()
}

// TokyoNightTheme returns a Tokyo Night inspired theme
func TokyoNightTheme() *Theme {
	foreground := RGB(192, 202, 245) // #c0caf5
	comment := RGB(86, 95, 137)      // #565f89
	red := RGB(247, 118, 142)        // #f7768e
	green := RGB(158, 206, 106)      // #9ece6a
	yellow := RGB(224, 175, 104)     // #e0af68
	blue := RGB(122, 162, 247)       // #7aa2f7
	magenta := RGB(187, 154, 247)    // #bb9af7
	cyan := RGB(125, 207, 255)       // #7dcfff
	orange := RGB(255, 158, 100)     // #ff9e64
	purple := RGB(157, 124, 216)     // #9d7cd8
	teal := RGB(115, 218, 202)       // #73daca

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        magenta,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             teal,
		Number:         purple,
		String:         green,
		Keyword:        yellow,
		Operator:       blue,
		ASN:            orange,
		Community:      magenta,
		Value:          cyan,
		Wildcard:       red,
		MAC:            cyan,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		TableName:      blue,
		PromptUser:     Color256(32),
		PromptAt:       Color256(163),
		PromptHostOper: Color256(35),
		PromptHostConf: Color256(128),
		PromptOper:     Color256(128),
		PromptConf:     Color256(35),
		PromptEdit:     Dim + comment,
	})
}

// VibrantTheme returns a vibrant color theme (original default)
func VibrantTheme() *Theme {
	return buildTheme(Palette{
		Foreground:     White,
		Comment:        Dim + BrightBlack,
		Command:        BrightYellow,
		Section:        BrightBlue,
		Protocol:       BrightCyan,
		Action:         BrightGreen,
		Interface:      BrightMagenta,
		IP:             BrightGreen,
		Number:         BrightCyan,
		String:         BrightYellow,
		Keyword:        Yellow,
		Operator:       BrightWhite,
		ASN:            BrightMagenta,
		Community:      Magenta,
		Value:          BrightCyan,
		Wildcard:       BrightRed,
		MAC:            Cyan,
		StateGood:      BrightGreen,
		StateBad:       BrightRed,
		StateWarning:   BrightYellow,
		Duration:       BrightMagenta,
		RouteProtocol:  Magenta,
		TableName:      BrightBlue,
		PromptUser:     Bold + BrightGreen,
		PromptAt:       White,
		PromptHostOper: Bold + BrightCyan,
		PromptHostConf: Bold + BrightMagenta,
		PromptOper:     Bold + BrightGreen,
		PromptConf:     Bold + BrightRed,
		PromptEdit:     BrightYellow,
	})
}

// SolarizedDarkTheme returns a Solarized Dark theme
func SolarizedDarkTheme() *Theme {
	base01 := Color256(240) // comments
	base0 := Color256(244)  // body text
	yellow := Color256(136)
	orange := Color256(166)
	red := Color256(160)
	magenta := Color256(125)
	violet := Color256(61)
	blue := Color256(33)
	cyan := Color256(37)
	green := Color256(64)

	return buildTheme(Palette{
		Foreground:     base0,
		Comment:        base01,
		Command:        yellow,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      magenta,
		IP:             green,
		Number:         cyan,
		String:         yellow,
		Keyword:        orange,
		Operator:       base0,
		ASN:            magenta,
		Community:      violet,
		Value:          cyan,
		Wildcard:       red,
		MAC:            cyan,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  violet,
		TableName:      blue,
		PromptUser:     Bold + green,
		PromptAt:       base0,
		PromptHostOper: Bold + cyan,
		PromptHostConf: Bold + magenta,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
		PromptEdit:     yellow,
	})
}

// MonokaiTheme returns a Monokai-inspired theme
func MonokaiTheme() *Theme {
	pink := Color256(197)
	green := Color256(148)
	orange := Color256(208)
	purple := Color256(141)
	cyan := Color256(81)
	yellow := Color256(186)
	gray := Color256(242)
	white := Color256(231)
	red := Color256(196)

	return buildTheme(Palette{
		Foreground:     white,
		Comment:        gray,
		Command:        pink,
		Section:        cyan,
		Protocol:       purple,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         purple,
		String:         yellow,
		Keyword:        orange,
		Operator:       pink,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		Wildcard:       pink,
		MAC:            cyan,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		TableName:      cyan,
		PromptUser:     Bold + green,
		PromptAt:       white,
		PromptHostOper: Bold + cyan,
		PromptHostConf: Bold + orange,
		PromptOper:     Bold + green,
		PromptConf:     Bold + pink,
		PromptEdit:     yellow,
	})
}

// NordTheme returns a Nord theme
func NordTheme() *Theme {
	nord4 := Color256(252)  // snow storm - main text
	nord7 := Color256(109)  // frost - cyan
	nord8 := Color256(110)  // frost - light blue
	nord9 := Color256(68)   // frost - blue
	nord11 := Color256(167) // aurora - red
	nord12 := Color256(173) // aurora - orange
	nord13 := Color256(179) // aurora - yellow
	nord14 := Color256(108) // aurora - green
	nord15 := Color256(139) // aurora - purple
	nordComment := Color256(60)

	return buildTheme(Palette{
		Foreground:     nord4,
		Comment:        nordComment,
		Command:        nord13,
		Section:        nord9,
		Protocol:       nord8,
		Action:         nord14,
		Interface:      nord15,
		IP:             nord14,
		Number:         nord15,
		String:         nord13,
		Keyword:        nord12,
		Operator:       nord9,
		ASN:            nord12,
		Community:      nord15,
		Value:          nord8,
		Wildcard:       nord11,
		MAC:            nord7,
		StateGood:      nord14,
		StateBad:       nord11,
		StateWarning:   nord13,
		Duration:       nord12,
		RouteProtocol:  nord15,
		TableName:      nord9,
		PromptUser:     Bold + nord14,
		PromptAt:       nord4,
		PromptHostOper: Bold + nord7,
		PromptHostConf: Bold + nord12,
		PromptOper:     Bold + nord14,
		PromptConf:     Bold + nord11,
		PromptEdit:     nord13,
	})
}

// CatppuccinMochaTheme returns a Catppuccin Mocha theme
// https://github.com/catppuccin/catppuccin
func CatppuccinMochaTheme() *Theme {
	text := RGB(205, 214, 244)     // #cdd6f4
	subtext0 := RGB(166, 173, 200) // #a6adc8
	overlay0 := RGB(108, 112, 134) // #6c7086
	red := RGB(243, 139, 168)      // #f38ba8
	peach := RGB(250, 179, 135)    // #fab387
	yellow := RGB(249, 226, 175)   // #f9e2af
	green := RGB(166, 227, 161)    // #a6e3a1
	teal := RGB(148, 226, 213)     // #94e2d5
	sky := RGB(137, 220, 235)      // #89dceb
	sapphire := RGB(116, 199, 236) // #74c7ec
	blue := RGB(137, 180, 250)     // #89b4fa
	lavender := RGB(180, 190, 254) // #b4befe
	mauve := RGB(203, 166, 247)    // #cba6f7
	pink := RGB(245, 194, 231)     // #f5c2e7

	return buildTheme(Palette{
		Foreground:     text,
		Comment:        overlay0,
		Command:        mauve,
		Section:        blue,
		Protocol:       sapphire,
		Action:         green,
		Interface:      peach,
		IP:             teal,
		Number:         lavender,
		String:         green,
		Keyword:        yellow,
		Operator:       sky,
		ASN:            peach,
		Community:      pink,
		Value:          sky,
		Wildcard:       red,
		MAC:            sky,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       peach,
		RouteProtocol:  mauve,
		TableName:      blue,
		PromptUser:     Bold + green,
		PromptAt:       subtext0,
		PromptHostOper: Bold + sapphire,
		PromptHostConf: Bold + peach,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
		PromptEdit:     yellow,
	})
}

// DraculaTheme returns the popular Dracula color scheme
// https://draculatheme.com
func DraculaTheme() *Theme {
	foreground := RGB(248, 248, 242) // #f8f8f2
	comment := RGB(98, 114, 164)     // #6272a4
	cyan := RGB(139, 233, 253)       // #8be9fd
	green := RGB(80, 250, 123)       // #50fa7b
	orange := RGB(255, 184, 108)     // #ffb86c
	pink := RGB(255, 121, 198)       // #ff79c6
	purple := RGB(189, 147, 249)     // #bd93f9
	red := RGB(255, 85, 85)          // #ff5555
	yellow := RGB(241, 250, 140)     // #f1fa8c

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        pink,
		Section:        purple,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         purple,
		String:         yellow,
		Keyword:        orange,
		Operator:       pink,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		Wildcard:       red,
		MAC:            cyan,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		TableName:      purple,
		PromptUser:     Bold + green,
		PromptAt:       foreground,
		PromptHostOper: Bold + cyan,
		PromptHostConf: Bold + orange,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
		PromptEdit:     yellow,
	})
}

// GruvboxDarkTheme returns the Gruvbox Dark color scheme
// https://github.com/morhetz/gruvbox
func GruvboxDarkTheme() *Theme {
	foreground := RGB(235, 219, 178) // #ebdbb2
	comment := RGB(146, 131, 116)    // #928374
	red := RGB(251, 73, 52)          // #fb4934
	green := RGB(184, 187, 38)       // #b8bb26
	yellow := RGB(250, 189, 47)      // #fabd2f
	blue := RGB(131, 165, 152)       // #83a598
	purple := RGB(211, 134, 155)     // #d3869b
	aqua := RGB(142, 192, 124)       // #8ec07c
	orange := RGB(254, 128, 25)      // #fe8019

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        yellow,
		Section:        blue,
		Protocol:       aqua,
		Action:         green,
		Interface:      orange,
		IP:             aqua,
		Number:         purple,
		String:         green,
		Keyword:        orange,
		Operator:       foreground,
		ASN:            orange,
		Community:      purple,
		Value:          aqua,
		Wildcard:       red,
		MAC:            aqua,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		TableName:      blue,
		PromptUser:     Bold + green,
		PromptAt:       foreground,
		PromptHostOper: Bold + aqua,
		PromptHostConf: Bold + orange,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
		PromptEdit:     yellow,
	})
}

// OneDarkTheme returns the Atom One Dark color scheme
// https://github.com/atom/one-dark-syntax
func OneDarkTheme() *Theme {
	foreground := RGB(171, 178, 191) // #abb2bf
	comment := RGB(92, 99, 112)      // #5c6370
	red := RGB(224, 108, 117)        // #e06c75
	green := RGB(152, 195, 121)      // #98c379
	yellow := RGB(229, 192, 123)     // #e5c07b
	blue := RGB(97, 175, 239)        // #61afef
	purple := RGB(198, 120, 221)     // #c678dd
	cyan := RGB(86, 182, 194)        // #56b6c2
	orange := RGB(209, 154, 102)     // #d19a66

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        purple,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         orange,
		String:         green,
		Keyword:        yellow,
		Operator:       foreground,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		Wildcard:       red,
		MAC:            cyan,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		TableName:      blue,
		PromptUser:     Bold + green,
		PromptAt:       foreground,
		PromptHostOper: Bold + cyan,
		PromptHostConf: Bold + orange,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
		PromptEdit:     yellow,
	})
}

// GetColor returns the color string for a token type
func (t *Theme) GetColor(tokenType lexer.TokenType) string {
	if color, ok := t.colors[tokenType]; ok {
		return color
	}
	return ""
}

// ThemeNames returns a list of available theme names.
func ThemeNames() []string {
	return []string{"tokyonight", "vibrant", "solarized", "monokai", "nord", "catppuccin", "dracula", "gruvbox", "onedark"}
}

// ThemeByName returns a theme by its name. Returns DefaultTheme for unknown names.
// Supported names: tokyonight, vibrant, solarized, monokai, nord, catppuccin, dracula, gruvbox, onedark
func ThemeByName(name string) *Theme {
	switch name {
	case "tokyonight", "tokyo-night", "tokyo":
		return TokyoNightTheme()
	case "vibrant":
		return VibrantTheme()
	case "solarized":
		return SolarizedDarkTheme()
	case "monokai":
		return MonokaiTheme()
	case "nord":
		return NordTheme()
	case "catppuccin", "catppuccin-mocha", "mocha":
		return CatppuccinMochaTheme()
	case "dracula":
		return DraculaTheme()
	case "gruvbox", "gruvbox-dark":
		return GruvboxDarkTheme()
	case "onedark", "one-dark":
		return OneDarkTheme()
	default:
		return DefaultTheme()
	}
}

// SetColor allows customizing a color for a token type
func (t *Theme) SetColor(tokenType lexer.TokenType, color string) {
	t.colors[tokenType] = color
}
