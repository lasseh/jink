package lexer

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

const showGoldenFile = "testdata/show_mode.golden"

// showModeCorpus is the input shape of external importers: taillight and
// taillight-tui call New, SetParseMode(ParseModeShow) and Tokenize on syslog
// messages. Their token output is a compatibility contract, so a change to
// this golden file is a change for them.
var showModeCorpus = []string{
	"BGP peer 10.0.0.1 (External AS 65001) changed state from Established to Idle",
	"bgp_recv: peer 2001:db8::1 (Internal AS 64512) received unexpected EOF",
	"RPD_BGP_NEIGHBOR_STATE_CHANGED: BGP peer 192.0.2.1 (External AS 65002) changed state from OpenConfirm to Established (event RecvKeepAlive)",
	"SNMP_TRAP_LINK_DOWN: ifIndex 526, ifAdminStatus up(1), ifOperStatus down(2), ifName ge-0/0/1",
	"mib2d: SNMP_TRAP_LINK_UP: ifIndex 601, ifName xe-0/1/0.100",
	"Interface ae0 flapped 12 times in 5m",
	"ospf neighbor 172.16.0.2 (realm ospf-v2 et-0/0/0.0 area 0.0.0.0) state changed from Full to Down",
	"route 10.1.0.0/16 via 192.0.2.254 [BGP/170] 1w2d 03:04:05 AS path: 65001 I",
	"prefix 2001:db8:1::/48 community 65001:100 from inet6.0",
	"inet.0: 912345 destinations, 1824690 routes (912345 active, 0 holddown, 0 hidden)",
	"MAC 00:11:22:33:44:55 moved from ge-0/0/1.0 to ge-0/0/2.0 on vlan-id 100",
	"CPU usage 95% memory 12.5GB of 16GB",
	"chassisd: FPC 0 temperature 85 degrees, status Warning",
	"session opened for user \"admin\" by (uid=0)",
	"login failed for 'root' from 198.51.100.7 port 22",
	"Peer                     AS      InPkt     OutPkt    OutQ   Flaps Last Up/Dwn State",
	"10.0.0.1              65001       1234       1235       0       0    1:02:03 Establ",
	"ge-0/0/0                up    up   inet     192.168.1.1/24",
	"user@router>",
	"user@router> ",
	"user@router> show bgp summary",
	"user@router> show route ",
	"admin@edge-1.example.net# set system host-name edge-1",
	"[edit interfaces]\nuser@router#",
	"{master:0}[edit]user@router# commit check",
	"\ruser@router> show chassis alarms\n",
	"+ set system host-name new",
	"- set system host-name old",
	"[edit system]",
	"# comment line with 10.0.0.1",
	"/* block */ peer 10.0.0.2 Active",
	"description \"link to core\"; unit 0",
	"policy-statement EXPORT { term 1 { then accept; } }",
	"Connect ExStart 2Way Init Loading",
	"",
	"   ",
	"\ttabbed 10.0.0.3\r\n",
	"multi\nline 10.0.0.4 down\nthird up",
}

func dumpShowTokens(input string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "INPUT %q\n", input)
	lex := New(input)
	lex.SetParseMode(ParseModeShow)
	for _, tok := range lex.Tokenize() {
		fmt.Fprintf(&b, "%d:%d %s %q\n", tok.Line, tok.Column, tok.Type, tok.Value)
	}
	return b.String()
}

func TestShowModeGolden(t *testing.T) {
	var b strings.Builder
	for _, input := range showModeCorpus {
		b.WriteString(dumpShowTokens(input))
	}
	got := b.String()

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(showGoldenFile, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(showGoldenFile)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		gotLines := strings.Split(got, "\n")
		wantLines := strings.Split(string(want), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("show-mode tokens changed at line %d:\n  got:  %s\n  want: %s", i+1, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("show-mode tokens changed: got %d lines, want %d", len(gotLines), len(wantLines))
	}
}
