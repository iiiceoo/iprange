package iprange

import (
	"errors"
	"math/big"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var parseTests = []struct {
	name string
	rs   []string
	want *IPRanges
	err  error
}{
	{
		name: "IPv4",
		rs: []string{
			"172.18.0.1",
			"172.18.0.0/24",
			"172.18.0.1-10",
			"172.18.0.1-172.18.1.10",
		},
		want: &IPRanges{
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
		err: nil,
	},
	{
		name: "IPv6",
		rs: []string{
			"fd00::1",
			"fd00::/64",
			"fd00::1-a",
			"fd00::1-fd00::1:a",
		},
		want: &IPRanges{
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
		err: nil,
	},
	{"empty", []string{}, &IPRanges{}, nil},
	{"empty", []string{""}, nil, errInvalidIPRangeFormat},
	{"invalid CIDR", []string{"172.18.0.0/33"}, nil, errInvalidIPRangeFormat},
	{"invalid start", []string{"172.18.0.a"}, nil, errInvalidIPRangeFormat},
	{"invalid start", []string{"172.18.0.a-10"}, nil, errInvalidIPRangeFormat},
	{"invalid start", []string{"172.18.0.a-172.18.0.10"}, nil, errInvalidIPRangeFormat},
	{"invalid end", []string{"172.18.0.1-a"}, nil, errInvalidIPRangeFormat},
	{"invalid end", []string{"172.18.0.1-172.18.0.a"}, nil, errInvalidIPRangeFormat},
	{"start exceeds end", []string{"172.18.0.10-1"}, nil, errInvalidIPRangeFormat},
	{"start exceeds end", []string{"172.18.0.10-172.18.0.1"}, nil, errInvalidIPRangeFormat},
	{"dual-stack", []string{"172.18.0.1", "fd00::/64"}, nil, errDualStackIPRanges},
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
			if !cmp.Equal(ranges, test.want) {
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
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 1).To4()},
				},
			},
		},
		want: IPv4,
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
			},
		},
		want: IPv6,
	},
	{
		name:   "unknown",
		ranges: &IPRanges{},
		want:   Unknown,
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
		name: "IPv4 contain",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		ip:   net.IPv4(172, 18, 0, 1),
		want: true,
	},
	{
		name: "IPv6 contain",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		ip:   net.ParseIP("fd00::2"),
		want: true,
	},
	{
		name: "IPv4 not contain",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		ip:   net.IPv4(172, 18, 0, 0),
		want: false,
	},
	{
		name: "IPv6 not contain",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		ip:   net.ParseIP("fd00::0"),
		want: false,
	},
	{
		name: "diff version",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		ip:   net.ParseIP("fd00::2"),
		want: false,
	},
	{
		name: "diff version",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		ip:   net.IPv4(172, 18, 0, 1),
		want: false,
	},
	{
		name:   "diff version",
		ranges: &IPRanges{},
		ip:     net.IPv4(172, 18, 0, 1),
		want:   false,
	},
	{
		name: "invalid IP",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		ip:   nil,
		want: false,
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
		name: "IPv4",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
			},
		},
		want: true,
	},
	{
		name: "IPv6",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		want: true,
	},
	{
		name:    "zero",
		rangesX: &IPRanges{},
		rangesY: &IPRanges{},
		want:    true,
	},
	{
		name: "diff version",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
			},
		},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		want: false,
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
		name: "IPv4 equal",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: true,
	},
	{
		name: "IPv6 equal",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: true,
	},
	{
		name:    "zero",
		rangesX: &IPRanges{},
		rangesY: &IPRanges{},
		want:    true,
	},
	{
		name: "IPv4 not equal",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
			},
		},
		want: false,
	},
	{
		name: "IPv6 not equal",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: false,
	},
	{
		name: "diff version",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 255).To4()},
				},
			},
		},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		want: false,
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
		name: "IPv4",
		ranges: &IPRanges{
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
		want: big.NewInt(357),
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::ffff:ffff:ffff:ffff")},
				},
			},
		},
		want: size,
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   big.NewInt(0),
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
		name: "multiple",
		ranges: &IPRanges{
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
					start: xIP{net.IPv4(172, 18, 0, 211).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 211).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 230).To4()},
				},
			},
		},
		want: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 211).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 220).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 230).To4()},
				},
			},
		},
	},
	{
		name: "one",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		want: &IPRanges{
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
			if !cmp.Equal(merged, test.want) {
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
		name: "IPv4",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 5).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
			},
		},
		want: &IPRanges{
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
		name: "IPv6",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::5")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
			},
		},
		want: &IPRanges{
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
		name: "diff version",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		want: &IPRanges{
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
			if !cmp.Equal(union, test.want) {
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
		name: "IPv4",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: &IPRanges{
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
		name: "IPv6",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: &IPRanges{
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
		name: "diff version",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		want: &IPRanges{
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
		name:    "zero-",
		rangesX: &IPRanges{version: IPv6},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		want: &IPRanges{version: IPv6},
	},
	{
		name: "-zero",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		rangesY: &IPRanges{version: IPv6},
		want: &IPRanges{
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
			if !cmp.Equal(difference, test.want) {
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
		name: "IPv4",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 12).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 15).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 15).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 18).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
			},
		},
	},
	{
		name: "IPv6",
		rangesX: &IPRanges{
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
		rangesY: &IPRanges{
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
		want: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::c")},
				},
				{
					start: xIP{net.ParseIP("fd00::f")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::14")},
				},
			},
		},
	},
	{
		name: "diff version",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		want: &IPRanges{
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
		name:    "zero-",
		rangesX: &IPRanges{version: IPv6},
		rangesY: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::5")},
				},
			},
		},
		want: &IPRanges{version: IPv6},
	},
	{
		name: "-zero",
		rangesX: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 20).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		rangesY: &IPRanges{version: IPv6},
		want: &IPRanges{
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
			if !cmp.Equal(intersection, test.want) {
				t.Fatalf("IPRanges(%v).Intersect(%v) = %v, want %v", test.rangesX, test.rangesY, intersection, test.want)
			}
		})
	}
}

var ipRangesSliceTests = []struct {
	name   string
	ranges *IPRanges
	start  *big.Int
	end    *big.Int
	want   *IPRanges
}{
	{
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
			},
		},
		start: big.NewInt(0),
		end:   big.NewInt(2),
		want: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 12).To4()},
				},
			},
		},
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::f")},
					end:   xIP{net.ParseIP("fd00::f")},
				},
				{
					start: xIP{net.ParseIP("fd00::8")},
					end:   xIP{net.ParseIP("fd00::9")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::16")},
				},
			},
		},
		start: big.NewInt(1),
		end:   big.NewInt(3),
		want: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::8")},
					end:   xIP{net.ParseIP("fd00::9")},
				},
				{
					start: xIP{net.ParseIP("fd00::12")},
					end:   xIP{net.ParseIP("fd00::12")},
				},
			},
		},
	},
	{
		name: "negative index",
		ranges: &IPRanges{
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
			},
		},
		start: big.NewInt(1),
		end:   big.NewInt(-2),
		want: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 11).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 25).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 29).To4()},
				},
			},
		},
	},
	{
		name: "start < 0 && end > size",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
			},
		},
		start: big.NewInt(-100),
		end:   big.NewInt(100),
		want: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
			},
		},
	},
	{
		name: "end out of ranges",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
			},
		},
		start: big.NewInt(-100),
		end:   big.NewInt(-100),
		want:  &IPRanges{version: IPv4},
	},
	{
		name: "start out of ranges",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
			},
		},
		start: big.NewInt(6),
		end:   big.NewInt(6),
		want:  &IPRanges{version: IPv4},
	},
	{
		name: "start > end",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 5).To4()},
				},
			},
		},
		start: big.NewInt(-1),
		end:   big.NewInt(0),
		want:  &IPRanges{version: IPv4},
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   &IPRanges{},
	},
}

func TestIPRangesSlice(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesSliceTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s := test.ranges.Slice(test.start, test.end)
			if !cmp.Equal(s, test.want) {
				t.Fatalf("IPRanges(%v).Slice(%v, %v) = %v, want %v", test.ranges, test.start, test.end, s, test.want)
			}
		})
	}
}

var ipRangesIsOverlapTests = []struct {
	name   string
	ranges *IPRanges
	want   bool
}{
	{
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 20).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 15).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 25).To4()},
				},
			},
		},
		want: true,
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::ab")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::aa")},
				},
			},
		},
		want: false,
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   false,
	},
}

func TestIPRangesIsOverlap(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIsOverlapTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			overlap := test.ranges.IsOverlap()
			if !cmp.Equal(overlap, test.want) {
				t.Fatalf("IPRanges(%v).IsOverlap() = %v, want %v", test.ranges, overlap, test.want)
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
		name: "range",
		ranges: &IPRanges{
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
		want: "[172.18.0.100-172.18.0.255 172.18.0.0-172.18.0.200]",
	},
	{
		name: "single",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
			},
		},
		want: "[fd00::1]",
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   "[]",
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

var ipRangesStringsTests = []struct {
	name   string
	ranges *IPRanges
	want   []string
}{
	{
		name: "range",
		ranges: &IPRanges{
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
		want: []string{
			"172.18.0.100-172.18.0.255",
			"172.18.0.0-172.18.0.200",
		},
	},
	{
		name: "CIDR",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::")},
					end:   xIP{net.ParseIP("fd00::ff")},
				},
			},
		},
		want: []string{"fd00::/120"},
	},
	{
		name: "single",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::1")},
				},
			},
		},
		want: []string{"fd00::1"},
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   nil,
	},
}

func TestIPRangesStrings(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesStringsTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ss := test.ranges.Strings()
			if !cmp.Equal(ss, test.want) {
				t.Fatalf("IPRanges(%v).Strings() = %v, want %v", test.ranges, ss, test.want)
			}
		})
	}
}
