package lexer

import (
	"regexp"
	"strings"
)

// Constants for lexer configuration
const (
	// parseModeDetectionSampleSize is the number of characters sampled for auto-detection
	parseModeDetectionSampleSize = 500
)

// Lexer tokenizes JunOS configuration text
type Lexer struct {
	input          string
	pos            int
	line           int
	col            int
	parseMode      ParseMode
	detectedMode   bool
	expectingValue bool   // true after keywords like "description" that take a value
	expectingUnit  bool   // true after "unit" keyword to classify numbers as TokenUnit
	lastToken      string // tracks the last non-whitespace token value for context
}

// ParseMode determines which classification rules to use for tokenization.
type ParseMode int

const (
	// ParseModeAuto automatically detects whether input is configuration
	// syntax or show command output based on content heuristics.
	ParseModeAuto ParseMode = iota

	// ParseModeConfig uses configuration syntax classification rules.
	// Use this for JunOS configuration text (set commands, hierarchical config).
	ParseModeConfig

	// ParseModeShow uses show command output classification rules.
	// Use this for output from show commands (bgp summary, interface terse, etc.).
	ParseModeShow
)

// Keyword sets for classification
var (
	commands = map[string]bool{
		"set": true, "delete": true, "deactivate": true, "activate": true,
		"protect": true, "unprotect": true, "edit": true, "show": true,
		"request": true, "run": true, "insert": true, "rename": true,
		"copy": true, "top": true, "up": true, "exit": true, "quit": true,
		"commit": true, "rollback": true, "load": true, "save": true,
		"configure": true, "cli": true, "help": true, "clear": true,
		"restart": true, "start": true, "stop": true, "monitor": true,
		"ping": true, "traceroute": true, "ssh": true, "telnet": true,
	}

	sections = map[string]bool{
		// Core configuration sections
		"system": true, "chassis": true, "interfaces": true,
		"routing-options": true, "routing-instances": true, "protocols": true,
		"policy-options": true, "firewall": true, "security": true,
		"class-of-service": true, "applications": true, "services": true,
		"snmp": true, "forwarding-options": true, "groups": true,
		"apply-groups": true, "apply-groups-except": true,
		"vlans": true, "bridge-domains": true,
		"virtual-chassis": true, "multi-chassis": true, "access": true,
		"ethernet-switching-options": true, "switch-options": true,
		"poe": true, "event-options": true, "accounting-options": true,
		"logical-systems": true, "tenants": true,
		// Data Center / EVPN-VXLAN sections
		"evpn": true, "vxlan": true, "mac-vrf": true, "virtual-switch": true,
		"overlay": true, "underlay": true,
		// Subscriber/BNG sections
		"dynamic-profiles": true, "subscriber-management": true,
		"unified-edge": true, "diameter": true, "aaa": true,
		"address-assignment": true, "access-profile": true,
		// Automation/Telemetry sections
		"openconfig": true, "telemetry": true, "streaming-telemetry": true,
		"grpc": true, "gnmi": true,
	}

	protocols = map[string]bool{
		// Routing protocols
		"ospf": true, "ospf3": true, "bgp": true, "isis": true, "is-is": true,
		"rip": true, "ripng": true, "ldp": true, "rsvp": true,
		"mpls": true, "vpls": true, "evpn": true, "pim": true,
		"igmp": true, "mld": true, "msdp": true, "bfd": true,
		"lacp": true, "lldp": true, "lldp-med": true, "rstp": true,
		"mstp": true, "vstp": true, "stp": true, "vrrp": true,
		"dot1x": true, "oam": true, "cfm": true,
		// Transport protocols
		"tcp": true, "udp": true, "icmp": true, "icmp6": true,
		"icmpv6": true, "gre": true, "ipip": true, "esp": true,
		"ah": true, "sctp": true,
		// Address families
		"inet": true, "inet6": true, "iso": true, "ccc": true,
		"bridge": true, "ethernet-switching": true,
		"inet-vpn": true, "inet6-vpn": true, "l2vpn": true,
		// Services
		"ssh": true, "telnet": true, "ftp": true, "tftp": true,
		"http": true, "https": true, "ntp": true, "dns": true,
		"dhcp": true, "radius": true, "tacplus": true, "syslog": true,
		"netconf": true, "junoscript": true,
		// Data Center / EVPN-VXLAN protocols
		"vxlan": true, "vtep": true, "vni": true, "esi": true,
		"l2circuit": true, "l3vpn": true, "mc-lag": true,
		"igmp-snooping": true, "mld-snooping": true, "l2-learning": true,
		// Segment Routing
		"source-packet-routing": true, "spring": true, "srv6": true,
		"segment-routing": true, "pcep": true, "te": true,
		"sr-te": true, "sr-mpls": true, "sr-policy": true,
		// Subscriber/BNG protocols
		"pppoe": true, "ppp": true, "l2tp": true, "dhcpv6": true,
		"diameter": true, "gx": true, "gy": true,
		"nasreq": true, "subscriber": true,
		// IKE/IPsec
		"ike": true, "ipsec": true,
		// ALGs
		"alg": true, "sip": true, "h323": true, "mgcp": true,
		"sccp": true, "rtsp": true, "pptp": true, "sunrpc": true, "msrpc": true,
		// Automation/Telemetry
		"gnmi": true, "grpc": true, "openconfig": true,
	}

	actions = map[string]bool{
		// Basic firewall actions
		"accept": true, "reject": true, "discard": true, "deny": true,
		"permit": true, "next": true, "next-term": true,
		"count": true, "log": true, "syslog": true, "sample": true,
		"port-mirror": true, "analyzer": true,
		// Routing policy actions
		"next-hop": true, "self": true, "table": true, "policy": true,
		"community": true, "local-preference": true, "metric": true,
		"origin": true, "as-path": true, "as-path-prepend": true, "med": true,
		"preference": true, "tag": true, "color": true, "color2": true,
		"load-balance": true, "install-nexthop": true,
		// CoS actions
		"loss-priority": true, "loss-priority-high": true, "loss-priority-low": true,
		"loss-priority-medium-high": true, "loss-priority-medium-low": true,
		"forwarding-class": true, "forwarding-class-except": true,
		"policer": true, "three-color-policer": true,
		"dscp": true, "traffic-class": true,
		// Security actions
		"tunnel": true, "ipsec-vpn": true,
		"source-nat": true, "destination-nat": true, "static-nat": true,
		// Firewall match conditions - Fragment handling
		"first-fragment": true, "fragment-offset": true, "fragment-offset-except": true,
		"is-fragment": true, "fragment-flags": true,
		// Firewall match conditions - TCP
		"tcp-initial": true, "tcp-established": true, "tcp-flags": true,
		"syn": true, "ack": true, "fin": true, "rst": true, "push": true, "urgent": true,
		// Firewall match conditions - ICMP
		"icmp-type": true, "icmp-type-except": true,
		"icmp-code": true, "icmp-code-except": true,
		// Firewall match conditions - Packet properties
		"packet-length": true, "packet-length-except": true,
		"ttl": true, "ttl-except": true,
		"hop-limit": true, "hop-limit-except": true,
		"payload-protocol": true, "payload-protocol-except": true,
		"traffic-type": true, "traffic-type-except": true,
		// Firewall match conditions - Layer 2
		"source-mac-address": true, "destination-mac-address": true,
		"ether-type": true, "vlan-ether-type": true,
		"user-vlan-id": true, "learn-vlan-id": true,
		"dot1q-tag": true, "dot1q-user-priority": true,
		// Firewall match conditions - Interface/group matching
		"interface": true, "interface-group": true, "interface-group-except": true,
		"interface-set": true, "ifl-number": true,
		"input-interface": true, "output-interface": true,
		// Firewall match conditions - Protocol fields
		"next-header": true, "next-header-except": true,
		"extension-header": true, "extension-header-except": true,
		"ip-options": true, "ip-options-except": true,
		// Firewall match conditions - Advanced
		"flexible-match-mask": true, "flexible-match-range": true,
		"loss-priority-except": true,
		"packet-length-range":  true, "port-except": true,
		"prefix-list-except": true, "source-class": true, "destination-class": true,
		"service-filter-hit": true, "policy-map": true,
	}

	keywords = map[string]bool{
		// System keywords
		"version": true, "host-name": true, "domain-name": true,
		"name-server": true, "root-authentication": true, "login": true,
		"user": true, "class": true, "authentication": true,
		"encrypted-password": true, "ssh-rsa": true, "ssh-dsa": true,
		"ssh-ecdsa": true, "ssh-ed25519": true,
		"description": true, "disable": true, "enable": true,
		"inactive": true, "apply-macro": true, "apply-path": true,
		// Interface keywords
		"unit": true, "family": true, "address": true, "vlan-id": true,
		"vlan-tagging": true, "flexible-vlan-tagging": true,
		"native-vlan-id": true, "mtu": true, "speed": true,
		"duplex": true, "auto-negotiation": true, "no-auto-negotiation": true,
		"gigether-options": true, "ether-options": true,
		"aggregated-ether-options": true, "link-speed": true,
		"minimum-links": true, "lacp": true, "active": true, "passive": true,
		"fast": true, "slow": true, "force-up": true,
		"interface-range": true, "member": true, "members": true,
		"interface-mode": true, "trunk": true, "access": true,
		// System keywords
		"scripts": true, "language": true, "synchronize": true,
		"login-alarms": true, "login-tip": true, "permissions": true,
		"uid": true, "gid": true, "password": true, "format": true,
		"port": true, "root-login": true, "protocol-version": true,
		"auto-snapshot": true, "time-zone": true,
		// Firewall/filter keywords
		"filter": true, "term": true, "from": true, "then": true,
		"source-address": true, "destination-address": true,
		"source-port": true, "destination-port": true,
		"source-prefix-list": true, "destination-prefix-list": true,
		"protocol": true, "prefix-list": true, "prefix-list-filter": true,
		"route-filter": true, "community-count": true, "as-path-group": true,
		// Routing keywords
		"rib-group": true, "rib": true, "static": true, "route": true,
		"qualified-next-hop": true, "preference": true, "tag": true,
		"no-readvertise": true, "retain": true, "no-retain": true,
		"discard": true, "reject": true, "receive": true,
		"aggregate": true, "generate": true, "martians": true,
		"router-id": true, "autonomous-system": true, "confederation": true,
		"instance-type": true, "interface-routes": true,
		"area": true, "interface": true, "neighbor": true, "group": true,
		"type": true, "peer-as": true, "local-as": true, "import": true,
		"export": true, "local-address": true, "authentication-key": true,
		"authentication-type": true, "bfd-liveness-detection": true,
		"minimum-interval": true, "multiplier": true, "hold-time": true,
		"damping": true, "multihop": true, "no-client-reflect": true,
		"cluster": true, "remove-private": true,
		"default-metric": true, "reference-bandwidth": true,
		"traffic-engineering": true, "shortcuts": true, "no-nssa-abr": true,
		"stub": true, "nssa": true, "default-lsa": true, "summaries": true,
		"virtual-link": true, "transit-area": true,
		// MPLS/RSVP keywords
		"label-switched-path": true, "path": true, "primary": true,
		"secondary": true, "standby": true, "bandwidth": true,
		"priority": true, "hop-limit": true, "record": true, "cspf": true,
		"node-link-protection": true, "fast-reroute": true,
		"detour": true, "admin-group": true, "include": true,
		"include-any": true, "exclude": true, "optimize-timer": true,
		"revert-timer": true, "signaled-bandwidth": true,
		// Security keywords
		"zone": true, "security-zone": true, "address-book": true,
		"host-inbound-traffic": true, "system-services": true,
		"policies": true, "policy": true, "match": true, "application": true,
		"source-zone": true, "destination-zone": true, "nat": true,
		"source": true, "destination": true, "pool": true,
		"rule-set": true, "rule": true,
		"translation-type": true, "translated": true,
		"screen": true, "ids-option": true, "icmp": true, "ip": true,
		"tcp-rst": true, "session-close": true, "alarm-threshold": true,
		"flow": true, "tcp-session": true, "tcp-mss": true,
		"allow-dns-reply": true, "allow-embedded-icmp": true,
		// IKE/IPsec keywords
		"ike": true, "gateway": true, "proposal": true, "ipsec": true,
		"vpn": true, "tunnel": true, "establish-tunnels": true,
		"immediately": true, "on-traffic": true, "responder-only": true,
		"bind-interface": true, "ike-policy": true, "ipsec-policy": true,
		"pre-shared-key": true, "ascii-text": true, "certificate": true,
		"local-identity": true, "remote-identity": true,
		"dead-peer-detection": true, "interval": true, "threshold": true,
		"general-ikeid": true, "no-anti-replay": true,
		// SNMP keywords
		"trap-group": true, "trap-options": true, "categories": true,
		"targets": true, "community-name": true, "authorization": true,
		"read-only": true, "read-write": true, "view": true,
		"client-list": true, "interface-list": true,
		"location": true, "contact": true, "community": true,
		// Forwarding-options keywords
		"storm-control-profiles": true, "storm-control": true,
		"analyzer": true, "port-mirroring": true, "helpers": true,
		// Firewall extended keywords
		"ip-version": true, "ip-protocol": true, "ipv4": true, "ipv6": true,
		"ip-destination-address": true, "ip-source-address": true,
		"ip6-destination-address": true, "ip6-source-address": true,
		"router-advertisement": true, "router-solicitation": true,
		"neighbor-advertisement": true, "neighbor-solicitation": true,
		// DHCPv6 client keywords
		"dhcpv6-client": true, "dhcp-client": true,
		"client-type": true, "client-ia-type": true, "ia-na": true, "ia-pd": true,
		"rapid-commit": true, "client-identifier": true,
		"duid-type": true, "duid-llt": true, "duid-ll": true,
		"stateful": true, "stateless": true,
		// Default keyword
		"default": true,
		// Inactive/deactivate prefix
		"inactive:": true,

		// Data Center / EVPN-VXLAN / Segment Routing keywords
		// EVPN core
		"vni": true, "vtep-source-interface": true, "extended-vni-list": true,
		"encapsulation": true, "multicast-mode": true, "ingress-replication": true,
		"route-distinguisher": true, "vrf-target": true,
		"vrf-import": true, "vrf-export": true, "vrf-table-label": true,
		"auto-export": true, "auto-rt": true,
		// EVPN multihoming
		"ethernet-segment": true, "esi": true, "all-active": true,
		"single-active": true, "designated-forwarder-election": true,
		"df-election-type": true, "recovery-timer": true,
		// EVPN/VXLAN specific
		"default-gateway": true, "advertise-default-gateway": true,
		"no-arp-suppression": true, "proxy-arp": true, "proxy-nd": true,
		// VRF/Instance types
		"virtual-router": true, "vrf": true, "layer2-control": true,
		"interconnect": true, "no-vrf-propagate-ttl": true,
		// MC-LAG
		"iccp": true, "peer": true, "liveness-detection": true,
		"redundancy-group": true, "preempt": true,
		// Segment Routing
		"node-segment": true, "index-range": true, "srgb": true, "srlb": true,
		"sid": true, "prefix-segment": true, "adjacency-segment": true,
		"binding-segment": true, "tilfa": true, "ti-lfa": true,
		"post-convergence-lfa": true, "backup-selection": true,
		"segment-list": true, "compute": true, "explicit": true,
		// SR-TE / PCEP
		"sr-te-template": true, "lsp-external-controller": true,
		"pce-controlled": true, "delegate": true, "report": true,
		"stateful-pce": true, "pce-peer": true, "destination-prefix": true,
		// SRv6
		"locator": true, "end-sid": true, "end-x-sid": true, "end-dt": true,
		"source-routing-header": true, "encapsulation-mode": true,

		// === Subscriber/BNG keywords ===
		// Subscriber management
		"demux-source": true, "underlying-interface": true,
		"client-profile": true, "server-profile": true,
		"ppp-options": true, "pppoe-options": true,
		"service-name-table": true, "max-sessions": true,
		"session-limit": true, "service-profile": true,
		// Authentication
		"authentication-order": true, "accounting": true,
		"radius-server": true, "tacplus-server": true,
		"secret": true, "timeout": true, "retry": true,
		// Address assignment
		"network": true, "range": true, "low": true, "high": true,
		"dhcp-attributes": true, "option": true, "option-82": true,
		"relay-option": true, "relay-agent-information": true,
		"subscriber-id": true, "agent-circuit-id": true, "agent-remote-id": true,
		// L2TP
		"lns": true, "lac": true, "l2tp-access-profile": true,
		"receive-window": true, "retransmit-interval": true,
		"maximum-receive-window": true, "tunnel-group": true,
		// CoS for subscribers
		"traffic-control": true, "traffic-control-profile": true,
		"scheduler-map": true, "shaping-rate": true, "guaranteed-rate": true,

		// === Automation/Telemetry keywords ===
		// gNMI/gRPC
		"sensor": true, "sensor-name": true, "resource": true,
		"reporting-rate": true, "polling-interval": true,
		"change-update": true, "on-change": true, "target-defined": true,
		// Streaming telemetry
		"export-profile": true, "local-port": true,
		"remote-address": true, "remote-port": true,
		"transport": true, "encoding": true, "subscription": true,
		// OpenConfig paths
		"xpath": true, "sensor-based-stats": true, "file": true,
		// Automation
		"commit-script": true, "op-script": true, "event-script": true,
		"slax": true, "python": true, "allow-commands": true, "deny-commands": true,
		"extension-service": true, "request-response": true, "notification": true,
	}

	// Keywords that take a value (colored as TokenValue)
	valueKeywords = map[string]bool{
		"description":        true,
		"host-name":          true,
		"domain-name":        true,
		"name-server":        true,
		"encrypted-password": true,
		"authentication-key": true,
		"pre-shared-key":     true,
		"ascii-text":         true,
		"community-name":     true,
		"version":            true,
	}

	// interfacePattern matches JunOS interface naming conventions:
	//   Physical: ge-0/0/0, xe-1/2/3, et-0/0/0 (Gigabit, 10G, 40/100G Ethernet)
	//            fe-0/0/0, so-0/0/0 (Fast Ethernet, SONET)
	//            t1-0/0/0, t3-0/0/0, e1-0/0/0, e3-0/0/0 (T1/T3/E1/E3)
	//            mge-0/0/0 (management), vcp-0/0/0 (virtual chassis)
	//   Channelized: ge-0/0/0:0 (channelized with :N suffix)
	//   Units: ge-0/0/0.100 (logical unit with .N suffix)
	//   Aggregated: ae0, ae15, reth0 (aggregated Ethernet, redundant Ethernet)
	//   Loopback: lo0, lo0.0
	//   Management: em0, me0, vme, fxp0, mxp0, exp0, jsrv (Junos services)
	//   Virtual: irb, irb.100 (integrated routing/bridging)
	//            vlan, vlan.100, gr-0/0/0, ip-0/0/0, vt-0/0/0, lt-0/0/0
	//   Tunnel: st0 (secure tunnel/VPN), gre, ipip
	//   Services: ms-0/0/0, sp-0/0/0, si-0/0/0 (multiservices, services)
	//            lsq-0/0/0, rlsq-0/0/0 (link services queuing)
	//   Internal: pp0, pd0, pe0, lsi, dsc, mtun, pimd, pime, tap, demux, fab
	//   VXLAN: vtep (VXLAN tunnel endpoint)
	//   QFX: fti (flexible tunnel interface)
	//   Special: all (wildcard for all interfaces)
	interfacePattern  = regexp.MustCompile(`^([gx]e|et|so|fe|at|t1|t3|e1|e3|mge|vcp|si|lsq|rlsq)-\d+/\d+/\d+(:\d+)?(\.\d+)?$|^(ae|reth|lo|em|me|irb|vlan|fab|gr|ip|vt|lt|ms|sp|pp|pd|pe|demux|dsc|mtun|pimd|pime|tap|lsi|st|vtep|fti|jsrv|gre|ipip)\d*(\.\d+)?$|^[efm]xp\d+(\.\d+)?$|^vme(\.\d+)?$|^all$`)
	ipv4Pattern       = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	ipv4PrefixPattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
	ipv6Pattern       = regexp.MustCompile(`^[0-9a-fA-F:]+:[0-9a-fA-F:]*$`)
	ipv6PrefixPattern = regexp.MustCompile(`^[0-9a-fA-F:]+:[0-9a-fA-F:]*/\d{1,3}$`)
	macPattern        = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}(/\d{1,2})?$`)
	numberPattern     = regexp.MustCompile(`^\d+[gmkGMK]?$`)
	communityPattern  = regexp.MustCompile(`^\d+:\d+$`)     // BGP community format
	asnPattern        = regexp.MustCompile(`^[Aa][Ss]\d+$`) // AS number format (AS65000)
	unitNumberPattern = regexp.MustCompile(`^\d+$`)         // Plain numbers for unit classification

	// Show output state keywords
	statesGood = map[string]bool{
		"up": true, "establ": true, "established": true,
		"full": true, "master": true, "primary": true,
		"enabled": true, "ok": true, "online": true,
		"running": true, "ready": true, "complete": true,
	}

	statesBad = map[string]bool{
		"down": true, "idle": true, "failed": true,
		"error": true, "offline": true, "disabled": true,
		"unreachable": true, "timeout": true,
		// BGP non-established states
		"active": true, "connect": true,
		"opensent": true, "openconfirm": true,
	}

	statesWarning = map[string]bool{
		// OSPF transitional states
		"init": true, "2way": true, "exstart": true,
		"exchange": true, "loading": true,
		// General
		"flapping": true, "pending": true, "waiting": true,
		"starting": true, "stopping": true,
	}

	statesNeutral = map[string]bool{
		"inactive": true, "standby": true, "backup": true,
		"n/a": true, "none": true,
	}

	columnHeaders = map[string]bool{
		"neighbor": true, "peer": true, "state": true,
		"interface": true, "admin": true, "link": true,
		"proto": true, "local": true, "remote": true,
		"as": true, "inpkt": true, "outpkt": true,
		"flaps": true, "uptime": true, "up/dn": true,
		"mtu": true, "speed": true, "type": true,
		"area": true, "dr": true, "bdr": true,
		"metric": true, "localpref": true, "med": true,
		"nexthop": true, "gateway": true, "flags": true,
		"outq": true, "prefixes": true, "paths": true,
	}

	statusSymbols = map[string]bool{
		"*": true, "+": true, "-": true, ">": true,
		"B": true, "O": true, "I": true, "S": true,
		"L": true, "D": true,
	}

	// Show output regex patterns
	timeDurationPattern  = regexp.MustCompile(`^(\d+[wdhms])+$|^\d+:\d{2}(:\d{2})?$`)
	percentagePattern    = regexp.MustCompile(`^\d+(\.\d+)?%$`)
	byteSizePattern      = regexp.MustCompile(`^\d+(\.\d+)?[KMGTP][Bb]?$`)
	routeProtocolPattern = regexp.MustCompile(`^\[(BGP|OSPF|OSPF3|ISIS|RIP|Static|Direct|Local|Aggregate)/\d+\]$`)
	tableNamePattern     = regexp.MustCompile(`^(inet|inet6|mpls|bgp|iso|l2vpn)\.\d+:?$`)
	tabularPattern       = regexp.MustCompile(`\w+\s{2,}\w+\s{2,}\w+`)

	// Prompt patterns
	// Matches: user@hostname> or user@hostname# (with optional {master:N}[edit ...] prefix)
	// Allows optional command after the prompt character
	// The \n? at the end handles lines that include trailing newlines
	// Group 3 captures leading whitespace/control chars (like \r) to preserve them
	promptPattern = regexp.MustCompile(`^(\{[^}]+\})?(\[edit[^\]]*\])?([\s\x00-\x1f]*)([\w-]+)@([\w.-]+)([>#])(\s*)(.*?)\n?$`)
)

// New creates a new Lexer for the given input.
// The lexer auto-detects whether input is config syntax or show command output.
func New(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

// Tokenize processes the input and returns all tokens.
// If parseMode is Auto (default), it auto-detects whether the input
// is configuration syntax or show command output based on content heuristics.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token

	// Check if the entire input is a prompt line
	if promptTokens := l.tryTokenizePrompt(l.input); promptTokens != nil {
		return promptTokens
	}

	for l.pos < len(l.input) {
		token := l.nextToken()
		if token.Type != TokenText || token.Value != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// tryTokenizePrompt checks if input matches a JunOS prompt and returns tokens if so
func (l *Lexer) tryTokenizePrompt(input string) []Token {
	// Try to match the full prompt pattern
	matches := promptPattern.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}

	var tokens []Token
	col := 1

	// matches[1] = {master:N} prefix (optional)
	// matches[2] = [edit ...] prefix (optional)
	// matches[3] = leading whitespace/control chars like \r (optional)
	// matches[4] = username
	// matches[5] = hostname
	// matches[6] = prompt char (> or #)
	// matches[7] = whitespace between prompt char and command (optional)
	// matches[8] = command after prompt (optional)

	// Add {master:N} prefix if present
	if matches[1] != "" {
		tokens = append(tokens, Token{
			Type:   TokenPromptEdit,
			Value:  matches[1],
			Line:   1,
			Column: col,
		})
		col += len(matches[1])
	}

	// Add [edit ...] context if present
	if matches[2] != "" {
		tokens = append(tokens, Token{
			Type:   TokenPromptEdit,
			Value:  matches[2],
			Line:   1,
			Column: col,
		})
		col += len(matches[2])
	}

	// Preserve leading whitespace/control chars (critical for cursor control like \r)
	if matches[3] != "" {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  matches[3],
			Line:   1,
			Column: col,
		})
		col += len(matches[3])
	}

	// Add username
	tokens = append(tokens, Token{
		Type:   TokenPromptUser,
		Value:  matches[4],
		Line:   1,
		Column: col,
	})
	col += len(matches[4])

	// Add @
	tokens = append(tokens, Token{
		Type:   TokenPromptAt,
		Value:  "@",
		Line:   1,
		Column: col,
	})
	col++

	// Add hostname (different token type based on prompt char)
	isConfig := matches[6] == "#"
	hostTokenType := TokenPromptHostOper
	if isConfig {
		hostTokenType = TokenPromptHostConf
	}
	tokens = append(tokens, Token{
		Type:   hostTokenType,
		Value:  matches[5],
		Line:   1,
		Column: col,
	})
	col += len(matches[5])

	// Add prompt character
	promptTokenType := TokenPromptOper
	if isConfig {
		promptTokenType = TokenPromptConf
	}
	tokens = append(tokens, Token{
		Type:   promptTokenType,
		Value:  matches[6],
		Line:   1,
		Column: col,
	})
	col++

	// Emit captured whitespace after prompt char (group 7)
	if matches[7] != "" {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  matches[7],
			Line:   1,
			Column: col,
		})
		col += len(matches[7])
	}

	// Tokenize command after prompt if present (group 8)
	if matches[8] != "" {
		cmdLexer := New(strings.TrimSpace(matches[8]))
		cmdTokens := cmdLexer.Tokenize()
		for _, tok := range cmdTokens {
			tok.Column = col
			tokens = append(tokens, tok)
			col += len(tok.Value)
		}
	}

	// Preserve trailing newline if present in original input
	if strings.HasSuffix(input, "\n") {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  "\n",
			Line:   1,
			Column: col,
		})
	}

	return tokens
}

// nextToken extracts the next token from the input
func (l *Lexer) nextToken() Token {
	// Skip whitespace but preserve position info
	startLine, startCol := l.line, l.col

	// Check for end of input
	if l.pos >= len(l.input) {
		return Token{Type: TokenText, Value: "", Line: startLine, Column: startCol}
	}

	// Check for diff lines at the start of a line
	if l.col == 1 {
		if tok, ok := l.scanDiffLine(); ok {
			return tok
		}
	}

	ch := l.input[l.pos]

	// Handle different token types
	switch {
	case ch == '#':
		return l.scanComment()
	case ch == '/' && l.peek(1) == '*':
		return l.scanBlockComment()
	case ch == '"':
		isValue := l.expectingValue
		l.expectingValue = false
		token := l.scanString('"')
		if isValue {
			token.Type = TokenValue
		}
		return token
	case ch == '\'':
		isValue := l.expectingValue
		l.expectingValue = false
		token := l.scanString('\'')
		if isValue {
			token.Type = TokenValue
		}
		return token
	case ch == '{' || ch == '}':
		l.expectingValue = false
		return l.scanBrace()
	case ch == ';':
		l.expectingValue = false
		return l.scanSemicolon()
	case ch == '<':
		return l.scanWildcard()
	case ch == '*':
		l.advance()
		return Token{Type: TokenWildcard, Value: "*", Line: startLine, Column: startCol}
	case isWhitespace(ch):
		return l.scanWhitespace()
	default:
		// If we're expecting a value (after description keyword), scan until semicolon
		if l.expectingValue {
			l.expectingValue = false
			return l.scanUnquotedValue()
		}
		return l.scanWord()
	}
}

// scanDiffLine checks if the current line is a diff line and scans it appropriately.
// Returns the token and true if a diff line was found, otherwise false.
// For +/- lines: only colorizes the prefix, letting the rest get normal syntax highlighting.
// For [edit ...] lines: colorizes the entire context header.
func (l *Lexer) scanDiffLine() (Token, bool) {
	startLine, startCol := l.line, l.col
	start := l.pos

	// Check for [edit ...] context line - color the whole line
	if l.input[l.pos] == '[' && l.pos+5 < len(l.input) && l.input[l.pos:l.pos+5] == "[edit" {
		// Scan until end of line
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.advance()
		}
		return Token{
			Type:   TokenDiffContext,
			Value:  l.input[start:l.pos],
			Line:   startLine,
			Column: startCol,
		}, true
	}

	// Check for + or - at start of line (diff add/remove)
	// Must be followed by space to indicate diff line
	if (l.input[l.pos] == '+' || l.input[l.pos] == '-') && l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		// + or - followed by space indicates diff line
		if next == ' ' || next == '\t' {
			tokenType := TokenDiffAdd
			if l.input[l.pos] == '-' {
				tokenType = TokenDiffRemove
			}
			ch := l.input[l.pos]
			l.advance()
			return Token{
				Type:   tokenType,
				Value:  string(ch),
				Line:   startLine,
				Column: startCol,
			}, true
		}
	}

	return Token{}, false
}

// scanComment scans a line comment starting with #
func (l *Lexer) scanComment() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	// Check for annotation (##)
	tokenType := TokenComment
	if l.peek(1) == '#' {
		tokenType = TokenAnnotation
	}

	// Read until end of line
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}

	return Token{
		Type:   tokenType,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanBlockComment scans a /* */ comment
func (l *Lexer) scanBlockComment() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	l.advance() // /
	l.advance() // *

	for l.pos < len(l.input)-1 {
		if l.input[l.pos] == '*' && l.input[l.pos+1] == '/' {
			l.advance() // *
			l.advance() // /
			break
		}
		l.advance()
	}

	return Token{
		Type:   TokenComment,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanString scans a quoted string
func (l *Lexer) scanString(quote byte) Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	l.advance() // opening quote

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == quote {
			l.advance() // closing quote
			break
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.advance() // escape char
		}
		l.advance()
	}

	return Token{
		Type:   TokenString,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanBrace scans { or }
func (l *Lexer) scanBrace() Token {
	startLine, startCol := l.line, l.col
	ch := l.input[l.pos]
	l.advance()

	return Token{
		Type:   TokenBrace,
		Value:  string(ch),
		Line:   startLine,
		Column: startCol,
	}
}

// scanSemicolon scans ;
func (l *Lexer) scanSemicolon() Token {
	startLine, startCol := l.line, l.col
	l.advance()
	return Token{
		Type:   TokenSemicolon,
		Value:  ";",
		Line:   startLine,
		Column: startCol,
	}
}

// scanUnquotedValue scans an unquoted value until semicolon (for keyword values)
func (l *Lexer) scanUnquotedValue() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	// Read until semicolon, newline, or end of input
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ';' || ch == '\n' {
			break
		}
		l.advance()
	}

	value := l.input[start:l.pos]
	// Trim trailing whitespace from the value
	value = strings.TrimRight(value, " \t")

	return Token{
		Type:   TokenValue,
		Value:  value,
		Line:   startLine,
		Column: startCol,
	}
}

// scanWildcard scans <*> style wildcards
func (l *Lexer) scanWildcard() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	l.advance() // <
	for l.pos < len(l.input) && l.input[l.pos] != '>' {
		l.advance()
	}
	if l.pos < len(l.input) {
		l.advance() // >
	}

	return Token{
		Type:   TokenWildcard,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanWhitespace scans whitespace characters
func (l *Lexer) scanWhitespace() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.advance()
	}

	return Token{
		Type:   TokenText,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanWord scans an identifier or keyword
func (l *Lexer) scanWord() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	// Read until whitespace or special character
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if isWhitespace(ch) || ch == '{' || ch == '}' || ch == ';' || ch == '"' || ch == '\'' || ch == '#' {
			break
		}
		l.advance()
	}

	word := l.input[start:l.pos]
	tokenType := l.classifyWord(word)

	return Token{
		Type:   tokenType,
		Value:  word,
		Line:   startLine,
		Column: startCol,
	}
}

// classifyWord determines the token type for a word
func (l *Lexer) classifyWord(word string) TokenType {
	// Auto-detect mode on first classification if needed
	if l.parseMode == ParseModeAuto && !l.detectedMode {
		l.parseMode = l.detectParseMode()
		l.detectedMode = true
	}

	lower := strings.ToLower(word)

	if l.parseMode == ParseModeShow {
		return l.classifyShowWord(word, lower)
	}

	return l.classifyConfigWord(word, lower)
}

// classifyConfigWord handles configuration syntax classification
func (l *Lexer) classifyConfigWord(word, lower string) TokenType {
	// Check if this is a unit number (after "unit" keyword)
	if l.expectingUnit && unitNumberPattern.MatchString(word) {
		l.expectingUnit = false
		return TokenUnit
	}

	// Check for AS number format (AS65000, as65001)
	if asnPattern.MatchString(word) {
		return TokenASN
	}

	// Check keyword maps first
	if commands[lower] {
		l.lastToken = lower
		return TokenCommand
	}
	if sections[lower] {
		l.lastToken = lower
		return TokenSection
	}
	if protocols[lower] {
		l.lastToken = lower
		return TokenProtocol
	}
	if actions[lower] {
		l.lastToken = lower
		return TokenAction
	}
	if keywords[lower] {
		// Set flag for keywords that take a value
		if valueKeywords[lower] {
			l.expectingValue = true
		}
		// Set flag after "unit" keyword to classify next number as TokenUnit
		if lower == "unit" {
			l.expectingUnit = true
		}
		l.lastToken = lower
		return TokenKeyword
	}

	// Fall through to shared patterns
	return l.classifySharedPatterns(word)
}

// classifyShowWord handles show command output classification
func (l *Lexer) classifyShowWord(word, lower string) TokenType {
	// State classification (highest priority for visibility)
	if statesGood[lower] {
		return TokenStateGood
	}
	if statesBad[lower] {
		return TokenStateBad
	}
	if statesWarning[lower] {
		return TokenStateWarning
	}
	if statesNeutral[lower] {
		return TokenStateNeutral
	}

	// Status symbols are single-char route markers (*, +, -, >) or protocol
	// indicators (B, O, I, S, L, D). Limit to 2 chars to avoid matching words.
	if len(word) <= 2 && statusSymbols[word] {
		return TokenStatusSymbol
	}

	// Show-specific patterns
	if timeDurationPattern.MatchString(word) {
		return TokenTimeDuration
	}
	if percentagePattern.MatchString(word) {
		return TokenPercentage
	}
	if byteSizePattern.MatchString(word) {
		return TokenByteSize
	}
	if routeProtocolPattern.MatchString(word) {
		return TokenRouteProtocol
	}
	if tableNamePattern.MatchString(lower) {
		return TokenTableName
	}

	// Column headers
	if columnHeaders[lower] {
		return TokenColumnHeader
	}

	// Fall through to shared patterns (IPs, interfaces, etc.)
	return l.classifySharedPatterns(word)
}

// classifySharedPatterns handles patterns common to both config and show modes
func (l *Lexer) classifySharedPatterns(word string) TokenType {
	// Check patterns - order matters! More specific patterns first.
	if interfacePattern.MatchString(word) {
		return TokenInterface
	}
	if ipv4PrefixPattern.MatchString(word) {
		return TokenIPv4Prefix
	}
	if ipv4Pattern.MatchString(word) {
		return TokenIPv4
	}
	// Check MAC and community BEFORE IPv6 (IPv6 regex is too broad)
	if macPattern.MatchString(word) {
		return TokenMAC
	}
	if communityPattern.MatchString(word) {
		return TokenCommunity
	}
	if ipv6PrefixPattern.MatchString(word) {
		return TokenIPv6Prefix
	}
	if ipv6Pattern.MatchString(word) {
		return TokenIPv6
	}
	if numberPattern.MatchString(word) {
		return TokenNumber
	}

	return TokenIdentifier
}

// Helper methods

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *Lexer) peek(offset int) byte {
	pos := l.pos + offset
	if pos < len(l.input) {
		return l.input[pos]
	}
	return 0
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// detectParseMode analyzes input to determine if it's config or show output.
// Uses heuristics based on common patterns in each format.
func (l *Lexer) detectParseMode() ParseMode {
	// Sample first N chars for detection - enough to see headers/commands
	// without processing entire large configs
	sample := l.input
	if len(sample) > parseModeDetectionSampleSize {
		sample = sample[:parseModeDetectionSampleSize]
	}
	lower := strings.ToLower(sample)

	// Config indicators: set, delete, {, }, ;
	configScore := 0
	configIndicators := []string{"set ", "delete ", "{", "}", ";", "host-name", "policy-statement"}
	for _, ind := range configIndicators {
		if strings.Contains(lower, ind) {
			configScore++
		}
	}

	// Show indicators: states, table names, column patterns
	// Note: these must be fairly specific to avoid false positives
	showScore := 0
	showIndicators := []string{
		"establ", "idle", "2way",
		"inet.0", "inet6.0", "bgp.evpn",
		"flaps", "up/dn",
		"physical interface", "logical interface",
	}
	for _, ind := range showIndicators {
		if strings.Contains(lower, ind) {
			showScore++
		}
	}

	// Tabular data pattern (multiple spaces between words) - strong indicator
	// worth 2 points since tabular output is very characteristic of show commands
	if tabularPattern.MatchString(sample) {
		showScore += 2
	}

	// Require showScore >= 2 to avoid false positives on single words like "up"
	// appearing in config context. Config mode is the safe default
	if showScore >= 2 && showScore > configScore {
		return ParseModeShow
	}
	return ParseModeConfig
}

// IsPrompt checks if the input matches a JunOS CLI prompt pattern.
// Matches formats like "user@router>" or "[edit] user@router#"
func IsPrompt(input string) bool {
	return promptPattern.MatchString(strings.TrimSpace(input))
}

// SetParseMode explicitly sets the parsing mode
func (l *Lexer) SetParseMode(mode ParseMode) {
	l.parseMode = mode
	l.detectedMode = true
}

// GetParseMode returns the current parse mode
func (l *Lexer) GetParseMode() ParseMode {
	return l.parseMode
}
