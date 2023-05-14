package iprange

import (
	"net"
	"reflect"
	"testing"
)

var ipRangesIPIteratorTests = []struct {
	name   string
	ranges *IPRanges
	want   []net.IP
}{
	{
		"Iterate through the IP addresses in IPv4 IP ranges",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 10).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 10).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 2).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
			},
		},
		[]net.IP{
			net.IPv4(172, 18, 0, 10).To4(),
			net.IPv4(172, 18, 0, 2).To4(),
			net.IPv4(172, 18, 0, 3).To4(),
			net.IPv4(172, 18, 0, 1).To4(),
			net.IPv4(172, 18, 0, 2).To4(),
		},
	},
	{
		"Iterate through the IP addresses in IPv6 IP ranges",
		&IPRanges{
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
		[]net.IP{
			net.ParseIP("fd00::a"),
			net.ParseIP("fd00::2"),
			net.ParseIP("fd00::3"),
			net.ParseIP("fd00::1"),
			net.ParseIP("fd00::2"),
		},
	},
	{
		"Empty IP ranges",
		&IPRanges{},
		nil,
	},
}

func TestIPRangesIPIterator(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIPIteratorTests {
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

			if !reflect.DeepEqual(ips, test.want) {
				t.Fatalf("IPRanges(%v).IPIterator() = %v, want %v", test.ranges, ips, test.want)
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
		"Iterate through the CIDRs in IPv4 IP ranges",
		&IPRanges{
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
		[]*net.IPNet{
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
		"Iterate through the CIDRs in IPv6 IP ranges",
		&IPRanges{
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
		[]*net.IPNet{
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
		"Empty IP ranges",
		&IPRanges{},
		nil,
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

			if !reflect.DeepEqual(ipNets, test.want) {
				t.Fatalf("IPRanges(%v).CIDRIterator() = %v, want %v", test.ranges, ipNets, test.want)
			}
		})
	}
}
