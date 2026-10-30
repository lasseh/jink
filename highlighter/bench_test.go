package highlighter

import (
	"io"
	"strings"
	"testing"
)

var benchConfig = `set system host-name router1
set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24
set protocols ospf area 0.0.0.0 interface ge-0/0/0.0
set protocols bgp group ibgp neighbor 10.0.0.2 peer-as 65000`

// benchEscapes carries cursor/color control sequences, exercising the
// segment-split path; the plain config above exercises the escape-free fast path.
var benchEscapes = "\x1b[1m\x1b[32muser@router>\x1b[0m show interfaces ge-0/0/0\r\n"

// sink prevents the compiler from eliminating the benchmarked calls.
var sink string

func BenchmarkHighlightForcedConfig(b *testing.B) {
	h := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = h.HighlightForced(benchConfig)
	}
}

func BenchmarkHighlightAutoConfig(b *testing.B) {
	h := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = h.Highlight(benchConfig)
	}
}

func BenchmarkHighlightForcedWithEscapes(b *testing.B) {
	h := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = h.HighlightForced(benchEscapes)
	}
}

func BenchmarkStreamLines(b *testing.B) {
	lines := strings.SplitAfter(benchConfig, "\n")
	s := New().NewStream(io.Discard)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			if _, err := s.Write([]byte(line)); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStripANSI(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = StripANSI(benchEscapes)
	}
}
