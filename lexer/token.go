package lexer

// TokenType represents the type of a lexical token
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

	// Show output semantic tokens
	TokenStateGood    // up, Establ, Full, Master (green)
	TokenStateBad     // down, Idle, Active, Connect (red)
	TokenStateWarning // 2Way, ExStart, Exchange, Loading (yellow)
	TokenStateNeutral // inactive, standby, backup (dim)

	// Show output structural tokens
	TokenColumnHeader  // Table column headers
	TokenStatusSymbol  // *, +, -, > (route markers)
	TokenTimeDuration  // 1d 2:30:45, 1w2d, 0:05:10
	TokenPercentage    // 50%, 99.9%
	TokenByteSize      // 1.5G, 500M, 10K
	TokenRouteProtocol // [BGP/170], [OSPF/10], [Static/5]
	TokenTableName     // inet.0, inet6.0, mpls.0

	// Prompt tokens
	TokenPromptUser     // username in prompt
	TokenPromptAt       // @ separator
	TokenPromptHostOper // hostname in operational mode
	TokenPromptHostConf // hostname in configuration mode
	TokenPromptOper     // > (operational mode prompt char)
	TokenPromptConf     // # (configuration mode prompt char)
	TokenPromptEdit     // [edit ...] context indicator

	// Diff tokens (show | compare output)
	TokenDiffAdd     // + lines (added) - green
	TokenDiffRemove  // - lines (removed) - red
	TokenDiffContext // [edit ...] context headers - cyan/blue
)

// Token represents a single lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case TokenText:
		return "Text"
	case TokenCommand:
		return "Command"
	case TokenSection:
		return "Section"
	case TokenProtocol:
		return "Protocol"
	case TokenAction:
		return "Action"
	case TokenInterface:
		return "Interface"
	case TokenIPv4:
		return "IPv4"
	case TokenIPv4Prefix:
		return "IPv4Prefix"
	case TokenIPv6:
		return "IPv6"
	case TokenIPv6Prefix:
		return "IPv6Prefix"
	case TokenMAC:
		return "MAC"
	case TokenNumber:
		return "Number"
	case TokenString:
		return "String"
	case TokenComment:
		return "Comment"
	case TokenAnnotation:
		return "Annotation"
	case TokenBrace:
		return "Brace"
	case TokenSemicolon:
		return "Semicolon"
	case TokenWildcard:
		return "Wildcard"
	case TokenIdentifier:
		return "Identifier"
	case TokenKeyword:
		return "Keyword"
	case TokenOperator:
		return "Operator"
	case TokenUnit:
		return "Unit"
	case TokenASN:
		return "ASN"
	case TokenCommunity:
		return "Community"
	case TokenValue:
		return "Value"
	case TokenStateGood:
		return "StateGood"
	case TokenStateBad:
		return "StateBad"
	case TokenStateWarning:
		return "StateWarning"
	case TokenStateNeutral:
		return "StateNeutral"
	case TokenColumnHeader:
		return "ColumnHeader"
	case TokenStatusSymbol:
		return "StatusSymbol"
	case TokenTimeDuration:
		return "TimeDuration"
	case TokenPercentage:
		return "Percentage"
	case TokenByteSize:
		return "ByteSize"
	case TokenRouteProtocol:
		return "RouteProtocol"
	case TokenTableName:
		return "TableName"
	case TokenPromptUser:
		return "PromptUser"
	case TokenPromptAt:
		return "PromptAt"
	case TokenPromptHostOper:
		return "PromptHostOper"
	case TokenPromptHostConf:
		return "PromptHostConf"
	case TokenPromptOper:
		return "PromptOper"
	case TokenPromptConf:
		return "PromptConf"
	case TokenPromptEdit:
		return "PromptEdit"
	case TokenDiffAdd:
		return "DiffAdd"
	case TokenDiffRemove:
		return "DiffRemove"
	case TokenDiffContext:
		return "DiffContext"
	default:
		return "Unknown"
	}
}
