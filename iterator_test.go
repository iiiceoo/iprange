package iprange

import (
	"math/big"
	"net/netip"
	"reflect"
	"testing"
)

var ipRangesIPIteratorNextTests = []struct {
	name   string
	ranges *IPRanges
	want   []netip.Addr
}{
	{
		name: "IPv4",
		ranges: mustRanges(
			IPv4,
			[2]string{"172.18.0.10", "172.18.0.10"},
			[2]string{"172.18.0.1", "172.18.0.2"},
		),
		want: []netip.Addr{
			netip.MustParseAddr("172.18.0.10"),
			netip.MustParseAddr("172.18.0.1"),
			netip.MustParseAddr("172.18.0.2"),
		},
	},
	{
		name: "IPv6",
		ranges: mustRanges(
			IPv6,
			[2]string{"fd00::a", "fd00::a"},
			[2]string{"fd00::1", "fd00::2"},
		),
		want: []netip.Addr{
			netip.MustParseAddr("fd00::a"),
			netip.MustParseAddr("fd00::1"),
			netip.MustParseAddr("fd00::2"),
		},
	},
	{"zero", &IPRanges{}, nil},
}

func TestIPRangesIPIteratorNext(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIPIteratorNextTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.IPIterator()

			var addrs []netip.Addr
			for {
				addr := iter.Next()
				if !addr.IsValid() {
					break
				}
				addrs = append(addrs, addr)
			}

			if !reflect.DeepEqual(addrs, test.want) {
				t.Fatalf("IPRanges(%v).IPIterator().Next() = %v, want %v", test.ranges, addrs, test.want)
			}
		})
	}
}

var ipRangesIPIteratorNextNTests = []struct {
	name   string
	ranges *IPRanges
	n      *big.Int
	want   []netip.Addr
}{
	{
		name: "IPv4",
		ranges: mustRanges(
			IPv4,
			[2]string{"172.18.0.10", "172.18.0.10"},
			[2]string{"172.18.0.1", "172.18.0.2"},
		),
		n: big.NewInt(0),
		want: []netip.Addr{
			netip.MustParseAddr("172.18.0.10"),
			netip.MustParseAddr("172.18.0.1"),
			netip.MustParseAddr("172.18.0.2"),
		},
	},
	{
		name: "IPv6",
		ranges: mustRanges(
			IPv6,
			[2]string{"fd00::a", "fd00::a"},
			[2]string{"fd00::2", "fd00::3"},
			[2]string{"fd00::1", "fd00::2"},
		),
		n: big.NewInt(2),
		want: []netip.Addr{
			netip.MustParseAddr("fd00::2"),
			netip.MustParseAddr("fd00::1"),
		},
	},
	{
		name:   "nil n",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.2"}),
		n:      nil,
		want: []netip.Addr{
			netip.MustParseAddr("172.18.0.1"),
			netip.MustParseAddr("172.18.0.2"),
		},
	},
	{
		name:   "out of ranges",
		ranges: mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.10"}),
		n:      big.NewInt(2),
		want:   nil,
	},
	{"zero", &IPRanges{}, big.NewInt(1), nil},
}

func TestIPRangesIPIteratorNextN(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIPIteratorNextNTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.IPIterator()

			var addrs []netip.Addr
			for {
				addr := iter.NextN(test.n)
				if !addr.IsValid() {
					break
				}
				addrs = append(addrs, addr)
			}

			if !reflect.DeepEqual(addrs, test.want) {
				t.Fatalf("IPRanges(%v).IPIterator().NextN(%v) = %v, want %v", test.ranges, test.n, addrs, test.want)
			}
		})
	}
}

var ipRangesBlockIteratorNextTests = []struct {
	name      string
	ranges    *IPRanges
	blockSize *big.Int
	want      []*IPRanges
}{
	{
		name: "IPv4",
		ranges: mustRanges(
			IPv4,
			[2]string{"172.18.0.10", "172.18.0.10"},
			[2]string{"172.18.0.1", "172.18.0.2"},
		),
		blockSize: big.NewInt(0),
		want: []*IPRanges{
			mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.10"}),
			mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
			mustRanges(IPv4, [2]string{"172.18.0.2", "172.18.0.2"}),
		},
	},
	{
		name: "IPv6",
		ranges: mustRanges(
			IPv6,
			[2]string{"fd00::a", "fd00::a"},
			[2]string{"fd00::6", "fd00::9"},
		),
		blockSize: big.NewInt(2),
		want: []*IPRanges{
			mustRanges(IPv6, [2]string{"fd00::a", "fd00::a"}, [2]string{"fd00::6", "fd00::6"}),
			mustRanges(IPv6, [2]string{"fd00::7", "fd00::8"}),
			mustRanges(IPv6, [2]string{"fd00::9", "fd00::9"}),
		},
	},
	{"zero", &IPRanges{}, big.NewInt(1), nil},
}

func TestIPRangesBlockIteratorNext(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesBlockIteratorNextTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.BlockIterator(test.blockSize)

			var ranges []*IPRanges
			for {
				r := iter.Next()
				if r == nil {
					break
				}
				ranges = append(ranges, r)
			}

			if !reflect.DeepEqual(ranges, test.want) {
				t.Fatalf("IPRanges(%v).BlockIterator(%v).Next() = %#v, want %#v", test.ranges, test.blockSize, ranges, test.want)
			}
		})
	}
}

var ipRangesBlockIteratorNextNTests = []struct {
	name      string
	ranges    *IPRanges
	blockSize *big.Int
	n         *big.Int
	want      []*IPRanges
}{
	{
		name: "IPv4",
		ranges: mustRanges(
			IPv4,
			[2]string{"172.18.0.10", "172.18.0.10"},
			[2]string{"172.18.0.1", "172.18.0.2"},
		),
		blockSize: big.NewInt(1),
		n:         big.NewInt(0),
		want: []*IPRanges{
			mustRanges(IPv4, [2]string{"172.18.0.10", "172.18.0.10"}),
			mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
			mustRanges(IPv4, [2]string{"172.18.0.2", "172.18.0.2"}),
		},
	},
	{
		name: "IPv6",
		ranges: mustRanges(
			IPv6,
			[2]string{"fd00::a", "fd00::a"},
			[2]string{"fd00::6", "fd00::b"},
		),
		blockSize: big.NewInt(2),
		n:         big.NewInt(2),
		want: []*IPRanges{
			mustRanges(IPv6, [2]string{"fd00::7", "fd00::8"}),
			mustRanges(IPv6, [2]string{"fd00::b", "fd00::b"}),
		},
	},
	{
		name:      "nil n",
		ranges:    mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.3"}),
		blockSize: big.NewInt(1),
		n:         nil,
		want: []*IPRanges{
			mustRanges(IPv4, [2]string{"172.18.0.1", "172.18.0.1"}),
			mustRanges(IPv4, [2]string{"172.18.0.2", "172.18.0.2"}),
			mustRanges(IPv4, [2]string{"172.18.0.3", "172.18.0.3"}),
		},
	},
	{"zero", &IPRanges{}, big.NewInt(1), big.NewInt(1), nil},
}

func TestIPRangesBlockIteratorNextN(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesBlockIteratorNextNTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.BlockIterator(test.blockSize)

			var ranges []*IPRanges
			for {
				r := iter.NextN(test.n)
				if r == nil {
					break
				}
				ranges = append(ranges, r)
			}

			if !reflect.DeepEqual(ranges, test.want) {
				t.Fatalf("IPRanges(%v).BlockIterator(%v).NextN(%v) = %#v, want %#v", test.ranges, test.blockSize, test.n, ranges, test.want)
			}
		})
	}
}
