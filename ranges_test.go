package iprange

import (
	"math/big"
	"net"
	"testing"
)

var parseTests = []struct {
	name          string
	rs            []string
	want          *IPRanges
	err           error
	isWantedError func(err error) bool
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 1).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 1, 10).To4()},
				},
			},
		},
		nil,
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
		nil,
	},
	{"Empty IP ranges", []string{}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Empty IP range", []string{""}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid CIDR", []string{"172.18.0.0/33"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid start IP address", []string{"172.18.0.a-10"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid start IP address", []string{"172.18.0.a-172.18.0.10"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid end IP address", []string{"172.18.0.1-a"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid end IP address", []string{"172.18.0.1-172.18.0.a"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Start IP address > end IP address", []string{"172.18.0.10-1"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Start IP address > end IP address", []string{"172.18.0.10-172.18.0.1"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Invalid IP address", []string{"172.18.0.a"}, nil, errInvalidIPRangeFormat, IsInvalidIPRangeFormat},
	{"Dual-stack IP ranges", []string{"172.18.0.1", "fd00::/64"}, nil, errDualStackIPRanges, IsDualStackIPRanges},
}

func TestParse(t *testing.T) {
	t.Parallel()
	for _, test := range parseTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ranges, err := Parse(test.rs...)
			if err != nil {
				if !test.isWantedError(err) {
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 1).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
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
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
				},
			},
		},
		big.NewInt(357),
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
					start: xIP{net.IPv4(172, 18, 0, 100).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 210).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 200).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 230).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 210).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 230).To4()},
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

var ipRangesUnionTests = []struct {
	name    string
	rangesX *IPRanges
	rangesY *IPRanges
	want    *IPRanges
}{
	{
		"X U Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 5).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
	{
		"X U Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::14")},
					end:   xIP{net.ParseIP("fd00::19")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::5")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
				{
					start: xIP{net.ParseIP("fd00::14")},
					end:   xIP{net.ParseIP("fd00::19")},
				},
			},
		},
	},
	{
		"Different IP version",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
}

func TestIPRangesUnion(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesUnionTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			union := test.rangesX.Union(test.rangesY)
			if !union.Equal(test.want) {
				t.Fatalf("IPRanges(%v).Union(%v) = %v, want %v", test.rangesX, test.rangesY, union, test.want)
			}
		})
	}
}

var ipRangesDiffTests = []struct {
	name    string
	rangesX *IPRanges
	rangesY *IPRanges
	want    *IPRanges
}{
	{
		"X - Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 25).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 30).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 40).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 50).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 15).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 8).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 12).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 18).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 22).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 13).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 14).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 16).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 17).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 25).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 30).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 40).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 50).To4()},
				},
			},
		},
	},
	{
		"X - Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::14")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
				{
					start: xIP{net.ParseIP("fd00::19")},
					end:   xIP{net.ParseIP("fd00::1e")},
				},
				{
					start: xIP{net.ParseIP("fd00::28")},
					end:   xIP{net.ParseIP("fd00::32")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::f")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
				{
					start: xIP{net.ParseIP("fd00::8")},
					end:   xIP{net.ParseIP("fd00::c")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::16")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
				{
					start: xIP{net.ParseIP("fd00::d")},
					end:   xIP{net.ParseIP("fd00::e")},
				},
				{
					start: xIP{net.ParseIP("fd00::10")},
					end:   xIP{net.ParseIP("fd00::11")},
				},
				{
					start: xIP{net.ParseIP("fd00::19")},
					end:   xIP{net.ParseIP("fd00::1e")},
				},
				{
					start: xIP{net.ParseIP("fd00::28")},
					end:   xIP{net.ParseIP("fd00::32")},
				},
			},
		},
	},
	{
		"Different IP version",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
	{
		"Zero - Y",
		&IPRanges{version: IPv6},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		&IPRanges{version: IPv6},
	},
	{
		"X - Zero",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		&IPRanges{version: IPv6},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
}

func TestIPRangesDiff(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesDiffTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			difference := test.rangesX.Diff(test.rangesY)
			if !difference.Equal(test.want) {
				t.Fatalf("IPRanges(%v).Diff(%v) = %v, want %v", test.rangesX, test.rangesY, difference, test.want)
			}
		})
	}
}

var ipRangesIntersectTests = []struct {
	name    string
	rangesX *IPRanges
	rangesY *IPRanges
	want    *IPRanges
}{
	{
		"X ∩ Y",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 25).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 30).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 40).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 50).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 15).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 8).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 12).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 18).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 22).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 12).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 18).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
			},
		},
	},
	{
		"X ∩ Y",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::f")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
				{
					start: xIP{net.ParseIP("fd00::8")},
					end:   xIP{net.ParseIP("fd00::c")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::16")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::14")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
				{
					start: xIP{net.ParseIP("fd00::19")},
					end:   xIP{net.ParseIP("fd00::1e")},
				},
				{
					start: xIP{net.ParseIP("fd00::28")},
					end:   xIP{net.ParseIP("fd00::32")},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::c")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::14")},
				},
			},
		},
	},
	{
		"Different IP version",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
	{
		"Zero ∩ Y",
		&IPRanges{version: IPv6},
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		&IPRanges{version: IPv6},
	},
	{
		"X ∩ Zero",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		&IPRanges{version: IPv6},
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
	},
}

func TestIPRangesIntersect(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIntersectTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			intersection := test.rangesX.Intersect(test.rangesY)
			if !intersection.Equal(test.want) {
				t.Fatalf("IPRanges(%v).Intersect(%v) = %v, want %v", test.rangesX, test.rangesY, intersection, test.want)
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
