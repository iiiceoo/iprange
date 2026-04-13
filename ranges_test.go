package iprange

import (
	"errors"
	"math/big"
	"net/netip"
	"reflect"
	"testing"
)

func mustRanges(version family, pairs ...[2]string) *IPRanges {
	ranges := make([]ipRange, 0, len(pairs))
	for _, pair := range pairs {
		ranges = append(ranges, ipRange{
			start: ip{addr: netip.MustParseAddr(pair[0])},
			end:   ip{addr: netip.MustParseAddr(pair[1])},
		})
	}

	return &IPRanges{
		version: version,
		ranges:  ranges,
	}
}

var parseTests = []struct {
	name string
	rs   []string
	want *IPRanges
	err  error
}{
	{
		name: "IPv4",
		rs:   []string{"172.18.0.1", "172.18.0.0/24", "172.18.0.1-10", "172.18.0.1-172.18.1.10"},
		want: mustRanges(IPv4,
			[2]string{"172.18.0.1", "172.18.0.1"},
			[2]string{"172.18.0.0", "172.18.0.255"},
			[2]string{"172.18.0.1", "172.18.0.10"},
			[2]string{"172.18.0.1", "172.18.1.10"},
		),
	},
	{
		name: "IPv6",
		rs:   []string{"fd00::1", "fd00::/64", "fd00::1-a", "fd00::1-fd00::1:a"},
		want: mustRanges(IPv6,
			[2]string{"fd00::1", "fd00::1"},
			[2]string{"fd00::", "fd00::ffff:ffff:ffff:ffff"},
			[2]string{"fd00::1", "fd00::a"},
			[2]string{"fd00::1", "fd00::1:a"},
		),
	},
	{"empty", []string{}, &IPRanges{}, nil},
	{"empty string", []string{""}, nil, errInvalidIPRangeFormat},
	{"invalid CIDR", []string{"172.18.0.0/33"}, nil, errInvalidIPRangeFormat},
	{"invalid start", []string{"172.18.0.a"}, nil, errInvalidIPRangeFormat},
	{"invalid short end", []string{"172.18.0.a-10"}, nil, errInvalidIPRangeFormat},
	{"invalid end", []string{"172.18.0.1-a"}, nil, errInvalidIPRangeFormat},
	{"start exceeds end", []string{"172.18.0.10-1"}, nil, errInvalidIPRangeFormat},
	{"mixed range version", []string{"172.18.0.1-fd00::1"}, nil, errInvalidIPRangeFormat},
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
			if !reflect.DeepEqual(ranges, test.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", test.rs, ranges, test.want)
			}
		})
	}
}

func TestErrorHelpers(t *testing.T) {
	t.Parallel()

	if !IsInvalidIPRangeFormat(errInvalidIPRangeFormat) {
		t.Fatal("IsInvalidIPRangeFormat() = false, want true")
	}
	if IsInvalidIPRangeFormat(errDualStackIPRanges) {
		t.Fatal("IsInvalidIPRangeFormat() = true, want false")
	}
	if !IsDualStackIPRanges(errDualStackIPRanges) {
		t.Fatal("IsDualStackIPRanges() = false, want true")
	}
	if IsDualStackIPRanges(errInvalidIPRangeFormat) {
		t.Fatal("IsDualStackIPRanges() = true, want false")
	}
}

var familyStringTests = []struct {
	name   string
	family family
	want   string
}{
	{"IPv4", IPv4, "IPv4"},
	{"IPv6", IPv6, "IPv6"},
}

func TestFamilyString(t *testing.T) {
	t.Parallel()
	for _, test := range familyStringTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.family.String(); got != test.want {
				t.Fatalf("family(%v).String() = %q, want %q", test.family, got, test.want)
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
		name:   "IPv4",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		want:   IPv4,
	},
	{
		name:   "IPv6",
		ranges: mustRanges(IPv6, [2]string{"fd00::1", "fd00::1"}),
		want:   IPv6,
	},
	{"zero", &IPRanges{}, IPv4},
}

func TestIPRangesVersion(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesVersionTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.Version(); got != test.want {
				t.Fatalf("IPRanges(%v).Version() = %v, want %v", test.ranges, got, test.want)
			}
		})
	}
}

var ipRangesContainsTests = []struct {
	name   string
	ranges *IPRanges
	addr   netip.Addr
	want   bool
}{
	{
		name:   "IPv4 contain",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		addr:   netip.MustParseAddr("172.18.0.1"),
		want:   true,
	},
	{
		name:   "IPv6 contain",
		ranges: mustRanges(IPv6, [2]string{"fd00::1", "fd00::3"}),
		addr:   netip.MustParseAddr("fd00::2"),
		want:   true,
	},
	{
		name:   "IPv4 not contain",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		addr:   netip.MustParseAddr("172.18.0.4"),
		want:   false,
	},
	{
		name:   "IPv6 not contain",
		ranges: mustRanges(IPv6, [2]string{"fd00::1", "fd00::3"}),
		addr:   netip.MustParseAddr("fd00::0"),
		want:   false,
	},
	{
		name:   "different version",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		addr:   netip.MustParseAddr("fd00::2"),
		want:   false,
	},
	{
		name:   "unmapped IPv4",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		addr:   netip.MustParseAddr("::ffff:172.18.0.2"),
		want:   true,
	},
	{"zero", &IPRanges{}, netip.Addr{}, false},
}

func TestIPRangesContains(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesContainsTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.Contains(test.addr); got != test.want {
				t.Fatalf("IPRanges(%v).Contains(%v) = %v, want %v", test.ranges, test.addr, got, test.want)
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
		name:    "IPv4",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.100", "172.18.0.255"}, [2]string{"172.18.0.0", "172.18.0.200"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.0", "172.18.0.255"}),
		want:    true,
	},
	{
		name:    "IPv6",
		rangesX: mustRanges(IPv6, [2]string{"fd00::aa", "fd00::ff"}, [2]string{"fd00::", "fd00::dd"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::", "fd00::ff"}),
		want:    true,
	},
	{"zero", &IPRanges{}, &IPRanges{}, true},
	{
		name:    "diff version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.0", "172.18.0.255"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::", "fd00::ff"}),
		want:    false,
	},
}

func TestIPRangesMergeEqual(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesMergeEqualTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.rangesX.MergeEqual(test.rangesY); got != test.want {
				t.Fatalf("IPRanges(%v).MergeEqual(%v) = %v, want %v", test.rangesX, test.rangesY, got, test.want)
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
		name:    "equal",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		want:    true,
	},
	{
		name:    "different len",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		rangesY: mustRanges(IPv4),
		want:    false,
	},
	{
		name:    "different range",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.2", "172.18.0.2"}),
		want:    false,
	},
	{
		name:    "different version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::1", "fd00::1"}),
		want:    false,
	},
}

func TestIPRangesEqual(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesEqualTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.rangesX.Equal(test.rangesY); got != test.want {
				t.Fatalf("IPRanges(%v).Equal(%v) = %v, want %v", test.rangesX, test.rangesY, got, test.want)
			}
		})
	}
}

var ipRangesSizeTests = []struct {
	name   string
	ranges *IPRanges
	want   *big.Int
}{
	{
		name:   "IPv4",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		want:   big.NewInt(3),
	},
	{
		name:   "IPv6",
		ranges: mustRanges(IPv6, [2]string{"fd00::1", "fd00::3"}),
		want:   big.NewInt(3),
	},
	{
		name:   "multiple",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}, [2]string{"172.18.0.10", "172.18.0.12"}),
		want:   big.NewInt(6),
	},
	{"zero", &IPRanges{}, big.NewInt(0)},
}

func TestIPRangesSize(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesSizeTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.Size(); got.Cmp(test.want) != 0 {
				t.Fatalf("IPRanges(%v).Size() = %v, want %v", test.ranges, got, test.want)
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
		name:   "single",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
	},
	{
		name:   "merge overlap",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.100", "172.18.0.255"}, [2]string{"172.18.0.0", "172.18.0.200"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.0", "172.18.0.255"}),
	},
	{
		name:   "merge adjacent",
		ranges: mustRanges(IPv6, [2]string{"fd00::1", "fd00::2"}, [2]string{"fd00::3", "fd00::4"}),
		want:   mustRanges(IPv6, [2]string{"fd00::1", "fd00::4"}),
	},
	{
		name:   "contained",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.0", "172.18.0.10"}, [2]string{"172.18.0.2", "172.18.0.4"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.0", "172.18.0.10"}),
	},
	{
		name:   "same start different end",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}, [2]string{"172.18.0.1", "172.18.0.5"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
	},
	{
		name:   "multiple overlaps",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.10"}, [2]string{"172.18.0.5", "172.18.0.15"}, [2]string{"172.18.0.20", "172.18.0.25"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.15"}, [2]string{"172.18.0.20", "172.18.0.25"}),
	},
	{
		name:   "unordered input",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.20"}, [2]string{"172.18.0.1", "172.18.0.5"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}, [2]string{"172.18.0.10", "172.18.0.20"}),
	},
	{
		name:   "duplicate ranges",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}, [2]string{"172.18.0.1", "172.18.0.5"}),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
	},
}

func TestIPRangesMerge(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesMergeTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			original := test.ranges.Clone()
			merged := test.ranges.Merge()
			if !reflect.DeepEqual(merged, test.want) {
				t.Fatalf("IPRanges(%v).Merge() = %#v, want %#v", test.ranges, merged, test.want)
			}
			if !reflect.DeepEqual(test.ranges, original) {
				t.Fatalf("IPRanges(%v).Merge() mutated receiver", original)
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
		name:   "overlap",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.20"}, [2]string{"172.18.0.15", "172.18.0.25"}),
		want:   true,
	},
	{
		name:   "adjacent only",
		ranges: mustRanges(IPv6, [2]string{"fd00::", "fd00::aa"}, [2]string{"fd00::ab", "fd00::ff"}),
		want:   false,
	},
	{"single", mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}), false},
	{"zero", &IPRanges{}, false},
}

func TestIPRangesIsOverlap(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIsOverlapTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.IsOverlap(); got != test.want {
				t.Fatalf("IPRanges(%v).IsOverlap() = %v, want %v", test.ranges, got, test.want)
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
		name:    "same version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.20", "172.18.0.30"}, [2]string{"172.18.0.1", "172.18.0.25"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.5", "172.18.0.25"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.30"}),
	},
	{
		name:    "different version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.20", "172.18.0.30"}, [2]string{"172.18.0.1", "172.18.0.25"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::1", "fd00::5"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.30"}),
	},
}

func TestIPRangesUnion(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesUnionTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			union := test.rangesX.Union(test.rangesY)
			if !reflect.DeepEqual(union, test.want) {
				t.Fatalf("IPRanges(%v).Union(%v) = %#v, want %#v", test.rangesX, test.rangesY, union, test.want)
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
		name:    "subset middle",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.20", "172.18.0.30"}, [2]string{"172.18.0.1", "172.18.0.25"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.5", "172.18.0.25"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.4"}, [2]string{"172.18.0.26", "172.18.0.30"}),
	},
	{
		name:    "disjoint",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.12"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
	},
	{
		name:    "full cover",
		rangesX: mustRanges(IPv6, [2]string{"fd00::1", "fd00::5"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::", "fd00::10"}),
		want:    &IPRanges{version: IPv6},
	},
	{
		name:    "left after right",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.12"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.12"}),
	},
	{
		name:    "trim tail only",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.10"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.5", "172.18.0.20"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.4"}),
	},
	{
		name:    "different version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::1", "fd00::5"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
	},
	{"zero left", &IPRanges{version: IPv4}, mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}), &IPRanges{version: IPv4}},
	{"zero right", mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}), &IPRanges{version: IPv4}, mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"})},
}

func TestIPRangesDiff(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesDiffTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			diff := test.rangesX.Diff(test.rangesY)
			if !reflect.DeepEqual(diff, test.want) {
				t.Fatalf("IPRanges(%v).Diff(%v) = %#v, want %#v", test.rangesX, test.rangesY, diff, test.want)
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
		name:    "same version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.20", "172.18.0.30"}, [2]string{"172.18.0.1", "172.18.0.25"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.5", "172.18.0.25"}),
		want:    mustRanges(IPv4, [2]string{"172.18.0.5", "172.18.0.25"}),
	},
	{
		name:    "disjoint",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		rangesY: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.12"}),
		want:    &IPRanges{version: IPv4},
	},
	{
		name:    "different version",
		rangesX: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		rangesY: mustRanges(IPv6, [2]string{"fd00::1", "fd00::5"}),
		want:    &IPRanges{version: IPv4},
	},
	{"zero", &IPRanges{version: IPv6}, mustRanges(IPv6, [2]string{"fd00::1", "fd00::5"}), &IPRanges{version: IPv6}},
}

func TestIPRangesIntersect(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIntersectTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			intersection := test.rangesX.Intersect(test.rangesY)
			if !reflect.DeepEqual(intersection, test.want) {
				t.Fatalf("IPRanges(%v).Intersect(%v) = %#v, want %#v", test.rangesX, test.rangesY, intersection, test.want)
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
		name:   "IPv4",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.20"}),
		start:  big.NewInt(0),
		end:    big.NewInt(2),
		want:   mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.12"}),
	},
	{
		name:   "IPv6",
		ranges: mustRanges(IPv6, [2]string{"fd00::f", "fd00::f"}, [2]string{"fd00::8", "fd00::9"}, [2]string{"fd00::12", "fd00::16"}),
		start:  big.NewInt(1),
		end:    big.NewInt(3),
		want:   mustRanges(IPv6, [2]string{"fd00::8", "fd00::9"}, [2]string{"fd00::12", "fd00::12"}),
	},
	{
		name:   "negative index",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.20"}, [2]string{"172.18.0.1", "172.18.0.5"}, [2]string{"172.18.0.25", "172.18.0.30"}),
		start:  big.NewInt(1),
		end:    big.NewInt(-2),
		want:   mustRanges(IPv4, [2]string{"172.18.0.11", "172.18.0.20"}, [2]string{"172.18.0.1", "172.18.0.5"}, [2]string{"172.18.0.25", "172.18.0.29"}),
	},
	{
		name:   "start < 0 && end > size",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		start:  big.NewInt(-100),
		end:    big.NewInt(100),
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
	},
	{
		name:   "end out of ranges",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		start:  big.NewInt(-100),
		end:    big.NewInt(-100),
		want:   &IPRanges{version: IPv4},
	},
	{
		name:   "start out of ranges",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		start:  big.NewInt(6),
		end:    big.NewInt(6),
		want:   &IPRanges{version: IPv4},
	},
	{
		name:   "start > end",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.5"}),
		start:  big.NewInt(-1),
		end:    big.NewInt(0),
		want:   &IPRanges{version: IPv4},
	},
	{
		name:   "nil indexes",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		start:  nil,
		end:    nil,
		want:   mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
	},
	{"zero", &IPRanges{}, nil, nil, &IPRanges{}},
}

func TestIPRangesSlice(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesSliceTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s := test.ranges.Slice(test.start, test.end)
			if !reflect.DeepEqual(s, test.want) {
				t.Fatalf("IPRanges(%v).Slice(%v, %v) = %#v, want %#v", test.ranges, test.start, test.end, s, test.want)
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
		name:   "ranges",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.100", "172.18.0.255"}, [2]string{"172.18.0.0", "172.18.0.200"}),
		want:   "[172.18.0.100-172.18.0.255 172.18.0.0-172.18.0.200]",
	},
	{
		name:   "range",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.100", "172.18.0.255"}),
		want:   "172.18.0.100-172.18.0.255",
	},
	{
		name:   "CIDR",
		ranges: mustRanges(IPv6, [2]string{"fd00::", "fd00::ff"}),
		want:   "fd00::/120",
	},
	{"single", mustRanges(IPv6, [2]string{"fd00::1", "fd00::1"}), "fd00::1"},
	{"zero", &IPRanges{}, "[]"},
}

func TestIPRangesString(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesStringTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.String(); got != test.want {
				t.Fatalf("IPRanges(%v).String() = %v, want %v", test.ranges, got, test.want)
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
		name:   "range",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.100", "172.18.0.255"}, [2]string{"172.18.0.0", "172.18.0.200"}),
		want:   []string{"172.18.0.100-172.18.0.255", "172.18.0.0-172.18.0.200"},
	},
	{
		name:   "CIDR",
		ranges: mustRanges(IPv6, [2]string{"fd00::", "fd00::ff"}),
		want:   []string{"fd00::/120"},
	},
	{"single", mustRanges(IPv6, [2]string{"fd00::1", "fd00::1"}), []string{"fd00::1"}},
	{"zero", &IPRanges{}, []string{}},
}

func TestIPRangesStrings(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesStringsTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.ranges.Strings(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("IPRanges(%v).Strings() = %v, want %v", test.ranges, got, test.want)
			}
		})
	}
}
