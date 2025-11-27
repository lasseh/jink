package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/lasseh/jink/highlighter"
)

const sampleConfig = `## Last commit: 2024-01-15 10:30:00 UTC by admin
version 21.4R3.5;
system {
    host-name core-router-01;
    domain-name example.com;
    root-authentication {
        encrypted-password "$6$abc123...";
    }
    services {
        ssh;
        netconf {
            ssh;
        }
    }
    syslog {
        host 10.0.0.100 {
            any any;
        }
    }
    ntp {
        server 10.0.0.1;
    }
}
interfaces {
    ge-0/0/0 {
        description "Uplink to ISP";
        unit 0 {
            family inet {
                address 203.0.113.1/30;
            }
            family inet6 {
                address 2001:db8::1/64;
            }
        }
    }
    ge-0/0/1 {
        description "LAN";
        unit 0 {
            family ethernet-switching {
                vlan {
                    members vlan100;
                }
            }
        }
    }
    ae0 {
        description "LACP bundle to switch";
        aggregated-ether-options {
            lacp {
                active;
            }
        }
        unit 0 {
            family inet {
                address 192.168.1.1/24;
            }
        }
    }
    lo0 {
        unit 0 {
            family inet {
                address 10.255.255.1/32;
            }
        }
    }
    irb {
        unit 100 {
            family inet {
                address 10.100.0.1/24;
            }
        }
    }
}
routing-options {
    router-id 10.255.255.1;
    autonomous-system 65001;
    static {
        route 0.0.0.0/0 next-hop 203.0.113.2;
    }
}
protocols {
    ospf {
        area 0.0.0.0 {
            interface ge-0/0/0.0 {
                interface-type p2p;
            }
            interface lo0.0 {
                passive;
            }
        }
    }
    bgp {
        group external {
            type external;
            peer-as 65000;
            neighbor 203.0.113.2 {
                description "ISP BGP peer";
                import import-policy;
                export export-policy;
            }
        }
        group internal {
            type internal;
            local-address 10.255.255.1;
            neighbor 10.255.255.2;
            neighbor 10.255.255.3;
        }
    }
    lldp {
        interface all;
    }
}
policy-options {
    prefix-list internal-networks {
        192.168.0.0/16;
        10.0.0.0/8;
    }
    policy-statement import-policy {
        term accept-default {
            from {
                route-filter 0.0.0.0/0 exact;
            }
            then accept;
        }
        term reject-rest {
            then reject;
        }
    }
    policy-statement export-policy {
        term advertise-internal {
            from {
                prefix-list internal-networks;
            }
            then {
                community add my-community;
                accept;
            }
        }
    }
    community my-community members 65001:100;
}
firewall {
    family inet {
        filter protect-re {
            term accept-ssh {
                from {
                    source-prefix-list {
                        internal-networks;
                    }
                    protocol tcp;
                    destination-port ssh;
                }
                then accept;
            }
            term accept-icmp {
                from {
                    protocol icmp;
                }
                then {
                    policer icmp-policer;
                    accept;
                }
            }
            term deny-rest {
                then {
                    count denied-packets;
                    log;
                    discard;
                }
            }
        }
    }
}
vlans {
    vlan100 {
        vlan-id 100;
        l3-interface irb.100;
    }
}
`

const sampleSetConfig = `set system host-name core-router-01
set system domain-name example.com
set system services ssh
set system services netconf ssh
set system syslog host 10.0.0.100 any any
set system ntp server 10.0.0.1

set interfaces ge-0/0/0 description "Uplink to ISP"
set interfaces ge-0/0/0 unit 0 family inet address 203.0.113.1/30
set interfaces ge-0/0/0 unit 0 family inet6 address 2001:db8::1/64
set interfaces ge-0/0/1 description "LAN"
set interfaces ae0 description "LACP bundle to switch"
set interfaces ae0 aggregated-ether-options lacp active
set interfaces ae0 unit 0 family inet address 192.168.1.1/24
set interfaces lo0 unit 0 family inet address 10.255.255.1/32

set routing-options router-id 10.255.255.1
set routing-options autonomous-system 65001
set routing-options static route 0.0.0.0/0 next-hop 203.0.113.2

set protocols ospf area 0.0.0.0 interface ge-0/0/0.0 interface-type p2p
set protocols ospf area 0.0.0.0 interface lo0.0 passive
set protocols bgp group external type external
set protocols bgp group external peer-as 65000
set protocols bgp group external neighbor 203.0.113.2 description "ISP BGP peer"
set protocols bgp group external neighbor 203.0.113.2 import import-policy
set protocols bgp group external neighbor 203.0.113.2 export export-policy
set protocols bgp group internal type internal
set protocols bgp group internal local-address 10.255.255.1
set protocols bgp group internal neighbor 10.255.255.2
set protocols bgp group internal neighbor 10.255.255.3
set protocols lldp interface all

set policy-options prefix-list internal-networks 192.168.0.0/16
set policy-options prefix-list internal-networks 10.0.0.0/8
set policy-options policy-statement import-policy term accept-default from route-filter 0.0.0.0/0 exact
set policy-options policy-statement import-policy term accept-default then accept
set policy-options policy-statement import-policy term reject-rest then reject
set policy-options community my-community members 65001:100

set firewall family inet filter protect-re term accept-ssh from source-prefix-list internal-networks
set firewall family inet filter protect-re term accept-ssh from protocol tcp
set firewall family inet filter protect-re term accept-ssh from destination-port ssh
set firewall family inet filter protect-re term accept-ssh then accept
set firewall family inet filter protect-re term accept-icmp from protocol icmp
set firewall family inet filter protect-re term accept-icmp then accept
set firewall family inet filter protect-re term deny-rest then count denied-packets
set firewall family inet filter protect-re term deny-rest then log
set firewall family inet filter protect-re term deny-rest then discard

delete system services ftp
deactivate interfaces ge-0/0/2

set vlans vlan100 vlan-id 100
set vlans vlan100 l3-interface irb.100
`

const sampleBGPSummary = `Peer                     AS      InPkt     OutPkt    OutQ   Flaps Last Up/Dwn State|#Active/Received/Accepted/Damped...
10.0.0.1              65001      12345      12340       0       2     1w2d3h Establ
  inet.0: 150/200/180/0
  inet6.0: 50/60/55/0
10.0.0.2              65002       8234       8230       0       0    3d12:30 Establ
  inet.0: 2500/3000/2800/0
192.168.1.1           65003        100        105       0      15       5:30 Active
203.0.113.5           65004          0          0       0       3     2w1d4h Idle
172.16.0.1            65005       5000       4998       0       1    12:45:00 Connect
`

const sampleOSPFNeighbors = `Address          Interface              State     ID               Pri  Dead
10.0.0.2         ge-0/0/0.0             Full      10.255.255.2     128    35
10.0.0.6         ge-0/0/1.0             Full      10.255.255.3     128    38
10.0.0.10        ae0.0                  2Way      10.255.255.4       1    32
10.0.0.14        ge-0/0/2.0             Init      10.255.255.5     128    40
10.0.0.18        xe-0/1/0.0             ExStart   10.255.255.6     128    37
172.16.0.2       et-0/0/0.0             Down      0.0.0.0            0     0
`

const sampleInterfaceTerse = `Interface               Admin Link Proto    Local                 Remote
ge-0/0/0                up    up
ge-0/0/0.0              up    up   inet     203.0.113.1/30
                                   inet6    2001:db8::1/64
ge-0/0/1                up    down
ge-0/0/1.0              up    down inet     192.168.1.1/24
xe-0/1/0                up    up
xe-0/1/0.0              up    up   inet     10.0.0.1/30
ae0                     up    up
ae0.0                   up    up   inet     172.16.0.1/24
lo0                     up    up
lo0.0                   up    up   inet     10.255.255.1/32
                                            127.0.0.1/32
irb                     up    up
irb.100                 up    up   inet     10.100.0.1/24
`

const sampleRouteTable = `inet.0: 25 destinations, 30 routes (25 active, 0 holddown, 0 hidden)
+ = Active Route, - = Last Active, * = Both

0.0.0.0/0          *[Static/5] 2w3d 12:30:45
                    > to 203.0.113.2 via ge-0/0/0.0
10.0.0.0/24        *[Direct/0] 1d 05:20:00
                    > via ge-0/0/1.0
10.0.0.1/32        *[Local/0] 1d 05:20:00
                      Local via ge-0/0/1.0
10.255.255.0/24    *[OSPF/10] 3d 08:15:30, metric 20
                    > to 10.0.0.2 via ge-0/0/0.0
172.16.0.0/16      *[BGP/170] 5d 14:22:10, localpref 100
                      AS path: 65002 65003 I, validation-state: valid
                    > to 10.0.0.1 via ge-0/0/0.0
192.168.0.0/16     *[Aggregate/130] 2w0d 00:00:00
                      Reject
`

const sampleChassisHardware = `Hardware inventory:
Item             Version  Part number  Serial number     Description
Chassis                                JN12345678        MX480
Midplane         REV 01   750-028467   ABCD1234          MX480 Midplane
FPC 0            REV 01   750-031089   FPC01234          MPC Type 2 3D
  CPU            REV 01   711-029089   CPU01234          MEMORY 2048MB
  PIC 0                   BUILTIN      BUILTIN           4x 10GE(LAN) SFP+
    Xcvr 0       REV 01   740-021308   XC001234          SFP+-10G-SR
    Xcvr 1       REV 01   740-021308   XC001235          SFP+-10G-LR
Routing Engine 0 REV 01   750-031093   RE001234          RE-S-1800x4
Power Supply 0   REV 02   740-024283   PS001234          DC 40A Power Supply
Fan Tray 0       REV 01   760-029763   FAN01234          Fan Tray
`

func main() {
	var (
		themeName  string
		setFormat  bool
		showAll    bool
		showOutput bool
	)

	flag.StringVar(&themeName, "theme", "default", "Theme: default, solarized, monokai, nord")
	flag.StringVar(&themeName, "t", "default", "Theme (shorthand)")
	flag.BoolVar(&setFormat, "set", false, "Show set-style config instead of hierarchical")
	flag.BoolVar(&setFormat, "s", false, "Show set-style config (shorthand)")
	flag.BoolVar(&showAll, "all", false, "Show all themes")
	flag.BoolVar(&showAll, "a", false, "Show all themes (shorthand)")
	flag.BoolVar(&showOutput, "show", false, "Show 'show' command output demo (BGP, OSPF, interfaces, routes)")
	flag.BoolVar(&showOutput, "o", false, "Show command output demo (shorthand)")

	flag.Parse()

	if showAll {
		showAllThemes()
		return
	}

	if showOutput {
		showShowOutputDemo(themeName)
		return
	}

	theme := highlighter.ThemeByName(strings.ToLower(themeName))
	hl := highlighter.NewWithTheme(theme)

	config := sampleConfig
	if setFormat {
		config = sampleSetConfig
	}

	fmt.Printf("\n=== JunOS Syntax Highlighting Demo (Theme: %s) ===\n\n", themeName)
	fmt.Println(hl.Highlight(config))
}

func showAllThemes() {
	themes := []struct {
		name  string
		theme *highlighter.Theme
	}{
		{"tokyonight (default)", highlighter.TokyoNightTheme()},
		{"vibrant", highlighter.VibrantTheme()},
		{"solarized", highlighter.SolarizedDarkTheme()},
		{"monokai", highlighter.MonokaiTheme()},
		{"nord", highlighter.NordTheme()},
		{"catppuccin", highlighter.CatppuccinMochaTheme()},
		{"dracula", highlighter.DraculaTheme()},
		{"gruvbox", highlighter.GruvboxDarkTheme()},
		{"onedark", highlighter.OneDarkTheme()},
	}

	// Short sample for comparison
	sample := `set system host-name router-01
set interfaces ge-0/0/0 unit 0 family inet address 192.168.1.1/24
set protocols bgp group external neighbor 10.0.0.1 peer-as 65000
set firewall family inet filter protect term 1 then accept
# This is a comment
`

	for _, t := range themes {
		hl := highlighter.NewWithTheme(t.theme)
		fmt.Printf("\n=== Theme: %s ===\n", t.name)
		fmt.Println(hl.Highlight(sample))
	}
}

func showShowOutputDemo(themeName string) {
	theme := highlighter.ThemeByName(strings.ToLower(themeName))
	hl := highlighter.NewWithTheme(theme)

	fmt.Printf("\n=== JunOS Show Output Highlighting Demo (Theme: %s) ===\n", themeName)

	fmt.Println("\n--- show bgp summary ---")
	fmt.Println(hl.HighlightShowOutput(sampleBGPSummary))

	fmt.Println("\n--- show ospf neighbor ---")
	fmt.Println(hl.HighlightShowOutput(sampleOSPFNeighbors))

	fmt.Println("\n--- show interfaces terse ---")
	fmt.Println(hl.HighlightShowOutput(sampleInterfaceTerse))

	fmt.Println("\n--- show route ---")
	fmt.Println(hl.HighlightShowOutput(sampleRouteTable))

	fmt.Println("\n--- show chassis hardware ---")
	fmt.Println(hl.HighlightShowOutput(sampleChassisHardware))
}
