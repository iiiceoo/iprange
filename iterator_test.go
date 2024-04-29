package iprange

import (
	"math/big"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var ipRangesIPIteratorNextTests = []struct {
	name   string
	ranges *IPRanges
	want   []net.IP
}{
	{
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
			},
		},
		want: []net.IP{
			net.IPv4(172, 18, 0, 10).To4(),
			net.IPv4(172, 18, 0, 1).To4(),
			net.IPv4(172, 18, 0, 2).To4(),
		},
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::2")},
				},
			},
		},
		want: []net.IP{
			net.ParseIP("fd00::a"),
			net.ParseIP("fd00::1"),
			net.ParseIP("fd00::2"),
		},
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   nil,
	},
}

func TestIPRangesIPIteratorNext(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIPIteratorNextTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.IPIterator()

			var ips []net.IP
			for {
				ip := iter.Next()
				if ip == nil {
					break
				}
				ips = append(ips, ip)
			}

			if !cmp.Equal(ips, test.want) {
				t.Fatalf("IPRanges(%v).IPIterator().Next() = %v, want %v", test.ranges, ips, test.want)
			}
		})
	}
}

var ipRangesIPIteratorNextNTests = []struct {
	name   string
	ranges *IPRanges
	n      *big.Int
	want   []net.IP
}{
	{
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
			},
		},
		n: big.NewInt(0),
		want: []net.IP{
			net.IPv4(172, 18, 0, 10).To4(),
			net.IPv4(172, 18, 0, 1).To4(),
			net.IPv4(172, 18, 0, 2).To4(),
		},
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
				{
					start: xIP{net.ParseIP("fd00::2")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::2")},
				},
			},
		},
		n: big.NewInt(2),
		want: []net.IP{
			net.ParseIP("fd00::2"),
			net.ParseIP("fd00::1"),
		},
	},
	{
		name: "out of ranges",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
			},
		},
		n:    big.NewInt(2),
		want: nil,
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		n:      big.NewInt(1),
		want:   nil,
	},
}

func TestIPRangesIPIteratorNextN(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIPIteratorNextNTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.IPIterator()

			var ips []net.IP
			for {
				ip := iter.NextN(test.n)
				if ip == nil {
					break
				}
				ips = append(ips, ip)
			}

			if !cmp.Equal(ips, test.want) {
				t.Fatalf("IPRanges(%v).IPIterator().NextN(%v) = %v, want %v", test.ranges, test.n, ips, test.want)
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
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
			},
		},
		blockSize: big.NewInt(0),
		want: []*IPRanges{
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 10).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
					},
				},
			},
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 1).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 1).To4()},
					},
				},
			},
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 2).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
					},
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
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
				{
					start: xIP{net.ParseIP("fd00::6")},
					end:   xIP{net.ParseIP("fd00::9")},
				},
			},
		},
		blockSize: big.NewInt(2),
		want: []*IPRanges{
			{
				version: IPv6,
				ranges: []ipRange{
					{
						start: xIP{net.ParseIP("fd00::a")},
						end:   xIP{net.ParseIP("fd00::a")},
					},
					{
						start: xIP{net.ParseIP("fd00::6")},
						end:   xIP{net.ParseIP("fd00::6")},
					},
				},
			},
			{
				version: IPv6,
				ranges: []ipRange{
					{
						start: xIP{net.ParseIP("fd00::7")},
						end:   xIP{net.ParseIP("fd00::8")},
					},
				},
			},
			{
				version: IPv6,
				ranges: []ipRange{
					{
						start: xIP{net.ParseIP("fd00::9")},
						end:   xIP{net.ParseIP("fd00::9")},
					},
				},
			},
		},
	},
	{
		name:      "zero",
		ranges:    &IPRanges{},
		blockSize: big.NewInt(1),
		want:      nil,
	},
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

			if !cmp.Equal(ranges, test.want) {
				t.Fatalf("IPRanges(%v).BlockIterator(%v).Next() = %v, want %v", test.ranges, test.blockSize, ranges, test.want)
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
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
			},
		},
		blockSize: big.NewInt(1),
		n:         big.NewInt(0),
		want: []*IPRanges{
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 10).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
					},
				},
			},
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 1).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 1).To4()},
					},
				},
			},
			{
				version: IPv4,
				ranges: []ipRange{
					{
						start: xIP{net.IPv4(172, 18, 0, 2).To4()},
						end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
					},
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
					start: xIP{net.ParseIP("fd00::a")},
					end:   xIP{net.ParseIP("fd00::a")},
				},
				{
					start: xIP{net.ParseIP("fd00::6")},
					end:   xIP{net.ParseIP("fd00::b")},
				},
			},
		},
		blockSize: big.NewInt(2),
		n:         big.NewInt(2),
		want: []*IPRanges{
			{
				version: IPv6,
				ranges: []ipRange{
					{
						start: xIP{net.ParseIP("fd00::7")},
						end:   xIP{net.ParseIP("fd00::8")},
					},
				},
			},
			{
				version: IPv6,
				ranges: []ipRange{
					{
						start: xIP{net.ParseIP("fd00::b")},
						end:   xIP{net.ParseIP("fd00::b")},
					},
				},
			},
		},
	},
	{
		name:      "zero",
		ranges:    &IPRanges{},
		blockSize: big.NewInt(1),
		n:         big.NewInt(1),
		want:      nil,
	},
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

			if !cmp.Equal(ranges, test.want) {
				t.Fatalf("IPRanges(%v).BlockIterator(%v).NextN(%v) = %v, want %v", test.ranges, test.blockSize, test.n, ranges, test.want)
			}
		})
	}
}

var ipRangesCIDRIteratorTests = []struct {
	name   string
	ranges *IPRanges
	want   []*net.IPNet
}{
	{
		name: "IPv4",
		ranges: &IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 1, 0).To4()},
					end:   xIP{net.IPv4(172, 18, 1, 255).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		want: []*net.IPNet{
			{
				IP:   net.IPv4(172, 18, 1, 0).To4(),
				Mask: net.CIDRMask(24, 32),
			},
			{
				IP:   net.IPv4(172, 18, 0, 1).To4(),
				Mask: net.CIDRMask(32, 32),
			},
			{
				IP:   net.IPv4(172, 18, 0, 2).To4(),
				Mask: net.CIDRMask(31, 32),
			},
		},
	},
	{
		name: "IPv6",
		ranges: &IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::0")},
					end:   xIP{net.ParseIP("fd00::ffff:ffff:ffff:ffff")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		want: []*net.IPNet{
			{
				IP:   net.ParseIP("fd00::0"),
				Mask: net.CIDRMask(64, 128),
			},
			{
				IP:   net.ParseIP("fd00::1"),
				Mask: net.CIDRMask(128, 128),
			},
			{
				IP:   net.ParseIP("fd00::2"),
				Mask: net.CIDRMask(127, 128),
			},
		},
	},
	{
		name:   "zero",
		ranges: &IPRanges{},
		want:   nil,
	},
}

func TestIPRangesCIDRIterator(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesCIDRIteratorTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.CIDRIterator()

			var ipNets []*net.IPNet
			for {
				ipNet := iter.Next()
				if ipNet == nil {
					break
				}
				ipNets = append(ipNets, ipNet)
			}

			if !cmp.Equal(ipNets, test.want) {
				t.Fatalf("IPRanges(%v).CIDRIterator() = %v, want %v", test.ranges, ipNets, test.want)
			}
		})
	}
}
