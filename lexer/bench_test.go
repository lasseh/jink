package lexer

import "testing"

var benchConfig = `set system host-name router1
set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24
set interfaces ae0 unit 100 family inet address 10.0.0.1/30
set protocols ospf area 0.0.0.0 interface ge-0/0/0.0
set protocols bgp group ibgp neighbor 10.0.0.2 peer-as 65000
set firewall family inet filter PROTECT term ALLOW from source-address 10.0.0.0/8
set firewall family inet filter PROTECT term ALLOW then accept`

var benchShow = `Peer               AS      InPkt     OutPkt    OutQ   Flaps Last Up/Dwn State
10.0.0.2        65000      12345      12346       0       0     1w2d3h Establ
10.0.0.3        65001          0          0       0       5         45 Idle
inet.0: 152 destinations, 200 routes (152 active, 0 holddown)
10.0.0.0/8         *[BGP/170] 1w2d 00:30:45, localpref 100, from 10.0.0.2`

// sinkTokens prevents the compiler from eliminating the benchmarked calls.
var sinkTokens []Token

func BenchmarkTokenizeConfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkTokens = New(benchConfig).Tokenize()
	}
}

func BenchmarkTokenizeShow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lex := New(benchShow)
		lex.SetParseMode(ParseModeShow)
		sinkTokens = lex.Tokenize()
	}
}
