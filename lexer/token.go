package lexer

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenText       TokenType = iota
	TokenCommand              // set, delete, edit, show, request
	TokenSection              // system, interfaces, protocols, etc.
	TokenProtocol             // ospf, bgp, tcp, udp, etc.
	TokenAction               // accept, reject, deny, permit
	TokenInterface            // ge-0/0/0, xe-1/0/0, ae0, lo0
	TokenIPv4                 // 192.168.1.1
	TokenIPv4Prefix           // 192.168.1.0/24
	TokenIPv6                 // 2001:db8::1
	TokenIPv6Prefix           // 2001:db8::/32
	TokenMAC                  // 00:11:22:33:44:55
	TokenNumber               // 100, 1000m, 10g
	TokenString               // "quoted string"
	TokenComment              // # comment or /* */
	TokenAnnotation           // ## annotation
	TokenBrace                // { }
	TokenSemicolon            // ;
	TokenWildcard             // <*>, *
	TokenIdentifier           // generic identifier
	TokenKeyword              // other important keywords
	TokenOperator             // operators like +, -, etc.
	TokenUnit                 // unit numbers
	TokenASN                  // AS numbers
	TokenCommunity            // BGP communities
	TokenValue                // Values after keywords (host-name, description, etc.)

	// Show output semantic tokens.
	TokenStateGood    // up, Establ, Full, Master (green)
	TokenStateBad     // down, Idle, Active, Connect (red)
	TokenStateWarning // 2Way, ExStart, Exchange, Loading (yellow)
	TokenStateNeutral // inactive, standby, backup (dim)

	// Show output structural tokens.
	TokenColumnHeader  // Table column headers
	TokenStatusSymbol  // *, +, -, > (route markers)
	TokenTimeDuration  // 1d 2:30:45, 1w2d, 0:05:10
	TokenPercentage    // 50%, 99.9%
	TokenByteSize      // 1.5G, 500M, 10K
	TokenRouteProtocol // [BGP/170], [OSPF/10], [Static/5]
	TokenTableName     // inet.0, inet6.0, mpls.0

	// Prompt tokens.
	TokenPromptUser     // username in prompt
	TokenPromptAt       // @ separator
	TokenPromptHostOper // hostname in operational mode
	TokenPromptHostConf // hostname in configuration mode
	TokenPromptOper     // > (operational mode prompt char)
	TokenPromptConf     // # (configuration mode prompt char)
	TokenPromptEdit     // [edit ...] context indicator

	// Diff tokens (show | compare output).
	TokenDiffAdd     // + lines (added) - green
	TokenDiffRemove  // - lines (removed) - red
	TokenDiffContext // [edit ...] context headers - cyan/blue
)

// Token represents a single lexical token.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// tokenTypeNames maps each token type to its display name. Keyed by the
// constant (not positional), so reordering the const block can't desync it.
var tokenTypeNames = [...]string{
	TokenText:           "Text",
	TokenCommand:        "Command",
	TokenSection:        "Section",
	TokenProtocol:       "Protocol",
	TokenAction:         "Action",
	TokenInterface:      "Interface",
	TokenIPv4:           "IPv4",
	TokenIPv4Prefix:     "IPv4Prefix",
	TokenIPv6:           "IPv6",
	TokenIPv6Prefix:     "IPv6Prefix",
	TokenMAC:            "MAC",
	TokenNumber:         "Number",
	TokenString:         "String",
	TokenComment:        "Comment",
	TokenAnnotation:     "Annotation",
	TokenBrace:          "Brace",
	TokenSemicolon:      "Semicolon",
	TokenWildcard:       "Wildcard",
	TokenIdentifier:     "Identifier",
	TokenKeyword:        "Keyword",
	TokenOperator:       "Operator",
	TokenUnit:           "Unit",
	TokenASN:            "ASN",
	TokenCommunity:      "Community",
	TokenValue:          "Value",
	TokenStateGood:      "StateGood",
	TokenStateBad:       "StateBad",
	TokenStateWarning:   "StateWarning",
	TokenStateNeutral:   "StateNeutral",
	TokenColumnHeader:   "ColumnHeader",
	TokenStatusSymbol:   "StatusSymbol",
	TokenTimeDuration:   "TimeDuration",
	TokenPercentage:     "Percentage",
	TokenByteSize:       "ByteSize",
	TokenRouteProtocol:  "RouteProtocol",
	TokenTableName:      "TableName",
	TokenPromptUser:     "PromptUser",
	TokenPromptAt:       "PromptAt",
	TokenPromptHostOper: "PromptHostOper",
	TokenPromptHostConf: "PromptHostConf",
	TokenPromptOper:     "PromptOper",
	TokenPromptConf:     "PromptConf",
	TokenPromptEdit:     "PromptEdit",
	TokenDiffAdd:        "DiffAdd",
	TokenDiffRemove:     "DiffRemove",
	TokenDiffContext:    "DiffContext",
}

// String returns a string representation of the token type.
func (t TokenType) String() string {
	if int(t) >= 0 && int(t) < len(tokenTypeNames) {
		if name := tokenTypeNames[t]; name != "" {
			return name
		}
	}
	return "Unknown"
}
