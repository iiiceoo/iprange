package iprange

import (
	"errors"
	"math/big"
	"net"
	"testing"
)

var parseTests = []struct {
	name string
	rs   []string
	want *IPRanges
	err  error
}{
	{
		"IPv4 IP ranges",
		[]string{
			"172.18.0.1",
			"172.18.0.0/24",
			"172.18.0.1-10",
			"172.18.0.1-172.18.1.10",
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 1)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 10)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 1, 10)},
				},
			},
		},
		nil,
	},
	{
		"IPv6 IP ranges",
		[]string{
			"fd00::1",
			"fd00::/64",
			"fd00::1-a",
			"fd00::1-fd00::1:a",
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ffff:ffff:ffff:ffff")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1:a")},
				},
			},
		},
		nil,
	},
	{"Empty IP ranges", []string{}, nil, errInvalidIPRangeFormat},
	{"Empty IP range", []string{""}, nil, errInvalidIPRangeFormat},
	{"Invalid CIDR", []string{"172.18.0.0/33"}, nil, errInvalidIPRangeFormat},
	{"Invalid start IP address", []string{"172.18.0.a-10"}, nil, errInvalidIPRangeFormat},
	{"Invalid start IP address", []string{"172.18.0.a-172.18.0.10"}, nil, errInvalidIPRangeFormat},
	{"Invalid end IP address", []string{"172.18.0.1-a"}, nil, errInvalidIPRangeFormat},
	{"Invalid end IP address", []string{"172.18.0.1-172.18.0.a"}, nil, errInvalidIPRangeFormat},
	{"Start IP address > end IP address", []string{"172.18.0.10-1"}, nil, errInvalidIPRangeFormat},
	{"Start IP address > end IP address", []string{"172.18.0.10-172.18.0.1"}, nil, errInvalidIPRangeFormat},
	{"Invalid IP address", []string{"172.18.0.a"}, nil, errInvalidIPRangeFormat},
	{"Dual-stack IP ranges", []string{"172.18.0.1", "fd00::/64"}, nil, errDualStackIPRanges},
}

func TestParse(t *testing.T) {
	t.Parallel()
	for _, test := range parseTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ranges, err := Parse(test.rs...)
			if err != nil {
				if !errors.Is(err, test.err) {
					t.Fatalf("Parse(%q) err %q, want %q", test.rs, err, test.err)
				}
				return
			}
			if !ranges.Equal(test.want) {
				t.Fatalf("Parse(%q) = %v, want %v", test.rs, ranges, test.want)
			}
		})
	}
}

var ipRangesVersionTests = []struct {
	name   string
	ranges *IPRanges
	want   family
}{
	{
		"IPv4",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 1)},
				},
			},
		},
		IPv4,
	},
	{
		"IPv6",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
			},
		},
		IPv6,
	},
	{
		"Unknown",
		&IPRanges{},
		Unknown,
	},
}

func TestIPRangesVersion(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesVersionTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			version := test.ranges.Version()
			if version != test.want {
				t.Fatalf("IPRanges(%v).Version() = %v, want %v", test.ranges, version, test.want)
			}
		})
	}
}

var ipRangesContainsTests = []struct {
	name   string
	ranges *IPRanges
	ip     net.IP
	want   bool
}{
	{
		"Contains IP address",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 3)},
				},
			},
		},
		net.IPv4(172, 18, 0, 1),
		true,
	},
	{
		"Contains IP address",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		net.ParseIP("fd00::2"),
		true,
	},
	{
		"Not contains IP address",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 3)},
				},
			},
		},
		net.IPv4(172, 18, 0, 0),
		false,
	},
	{
		"Not contains IP address",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		net.ParseIP("fd00::0"),
		false,
	},
	{
		"Different IP versions",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 3)},
				},
			},
		},
		net.ParseIP("fd00::2"),
		false,
	},
	{
		"Different IP versions",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		net.IPv4(172, 18, 0, 1),
		false,
	},
	{
		"Different IP versions",
		&IPRanges{},
		net.IPv4(172, 18, 0, 1),
		false,
	},
	{
		"Invalid IP address",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1)},
					end:   xIP{net.IPv4(172, 18, 0, 3)},
				},
			},
		},
		nil,
		false,
	},
}

func TestIPRangesContains(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesContainsTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			contains := test.ranges.Contains(test.ip)
			if contains != test.want {
				t.Fatalf("IPRanges(%v).Contains(%v) = %v, want %v", test.ranges, test.ip, contains, test.want)
			}
		})
	}
}

var ipRangesMergeEqualTests = []struct {
	name    string
	rangesX *IPRanges
	rangesY *IPRanges
	want    bool
}{
	{
		"X == Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
			},
		},
		true,
	},
	{
		"X == Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::aa")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::dd")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		true,
	},
	{
		"X == Y",
		&IPRanges{},
		&IPRanges{},
		true,
	},
	{
		"Different IP version",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		false,
	},
}

func TestIPRangesMergeEqual(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesMergeEqualTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			equal := test.rangesX.MergeEqual(test.rangesY)
			if equal != test.want {
				t.Fatalf("IPRanges(%v).MergeEqual(%v) = %v, want %v", test.rangesX, test.rangesY, equal, test.want)
			}
		})
	}
}

var ipRangesEqualTests = []struct {
	name    string
	rangesX *IPRanges
	rangesY *IPRanges
	want    bool
}{
	{
		"X == Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		true,
	},
	{
		"X == Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::aa")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::dd")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::aa")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::dd")},
				},
			},
		},
		true,
	},
	{
		"X == Y",
		&IPRanges{},
		&IPRanges{},
		true,
	},
	{
		"X != Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
			},
		},
		false,
	},
	{
		"X != Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::aa")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::dd")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::aa")},
				},
				{
					start: xIP{net.ParseIP("fd00::ab")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		false,
	},
	{
		"Different IP version",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		false,
	},
}

func TestIPRangesEqual(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesEqualTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			equal := test.rangesX.Equal(test.rangesY)
			if equal != test.want {
				t.Fatalf("IPRanges(%v).Equal(%v) = %v, want %v", test.rangesX, test.rangesY, equal, test.want)
			}
		})
	}
}

var size, _ = big.NewInt(0).SetString("18446744073709551615", 10)

var ipRangesSizeTests = []struct {
	name   string
	ranges *IPRanges
	want   *big.Int
}{
	{
		"IPv4 IP ranges size",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		big.NewInt(256),
	},
	{
		"IPv6 IP range size",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::ffff:ffff:ffff:ffff")},
				},
			},
		},
		size,
	},
	{
		"Zero",
		&IPRanges{},
		big.NewInt(0),
	},
}

func TestIPRangesSize(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesSizeTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			size := test.ranges.Size()
			if size.Cmp(test.want) != 0 {
				t.Fatalf("IPRanges(%v).Size() = %v, want %v", test.ranges, size, test.want)
			}
		})
	}
}

var ipRangesMergeTests = []struct {
	name   string
	ranges *IPRanges
	want   *IPRanges
}{
	{
		"Multiple IP ranges",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 210)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220)},
					end:   xIP{net.IPv4(172, 18, 0, 230)},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 210)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220)},
					end:   xIP{net.IPv4(172, 18, 0, 230)},
				},
			},
		},
	},
	{
		"Single IP range",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
	},
}

func TestIPRangesMerge(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesMergeTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			merged := test.ranges.Merge()
			if !merged.Equal(test.want) {
				t.Fatalf("IPRanges(%v).Merge() = %v, want %v", test.ranges, merged, test.want)
			}
		})
	}
}

var ipRangesStringTests = []struct {
	name   string
	ranges *IPRanges
	want   string
}{
	{
		"IP ranges string",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100)},
					end:   xIP{net.IPv4(172, 18, 0, 255)},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0)},
					end:   xIP{net.IPv4(172, 18, 0, 200)},
				},
			},
		},
		"[172.18.0.100-172.18.0.255 172.18.0.0-172.18.0.200]",
	},
	{
		"Single IP address string",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
			},
		},
		"[fd00::1]",
	},
	{
		"Zero string",
		&IPRanges{},
		"[]",
	},
}

func TestIPRangesString(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesStringTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s := test.ranges.String()
			if s != test.want {
				t.Fatalf("IPRanges(%v).String() = %v, want %v", test.ranges, s, test.want)
			}
		})
	}
}
