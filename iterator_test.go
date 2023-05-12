package iprange

import (
	"net"
	"reflect"
	"testing"
)

var ipRangesIteratorTests = []struct {
	name   string
	ranges *IPRanges
	want   []net.IP
}{
	{
		"Iterate through IPv4 IP ranges",
		&IPRanges{
			version: IPv4,
			ranges: []ipRange{
				{
					start: xIP{net.IPv4(172, 18, 0, 255).To4()},
					end:   xIP{net.IPv4(172, 18, 1, 1).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 1).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 2).To4()},
				},
				{
					start: xIP{net.IPv4(172, 18, 0, 2).To4()},
					end:   xIP{net.IPv4(172, 18, 0, 3).To4()},
				},
			},
		},
		[]net.IP{
			net.IPv4(172, 18, 0, 1).To4(),
			net.IPv4(172, 18, 0, 2).To4(),
			net.IPv4(172, 18, 0, 3).To4(),
			net.IPv4(172, 18, 0, 255).To4(),
			net.IPv4(172, 18, 1, 0).To4(),
			net.IPv4(172, 18, 1, 1).To4(),
		},
	},
	{
		"Iterate through IPv6 IP ranges",
		&IPRanges{
			version: IPv6,
			ranges: []ipRange{
				{
					start: xIP{net.ParseIP("fd00::ff")},
					end:   xIP{net.ParseIP("fd00::101")},
				},
				{
					start: xIP{net.ParseIP("fd00::1")},
					end:   xIP{net.ParseIP("fd00::2")},
				},
				{
					start: xIP{net.ParseIP("fd00::2")},
					end:   xIP{net.ParseIP("fd00::3")},
				},
			},
		},
		[]net.IP{
			net.ParseIP("fd00::1"),
			net.ParseIP("fd00::2"),
			net.ParseIP("fd00::3"),
			net.ParseIP("fd00::ff"),
			net.ParseIP("fd00::100"),
			net.ParseIP("fd00::101"),
		},
	},
	{
		"Empty IP ranges",
		&IPRanges{},
		nil,
	},
}

func TestIPRangesIterator(t *testing.T) {
	t.Parallel()
	for _, test := range ipRangesIteratorTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			iter := test.ranges.Iterator()

			var ips []net.IP
			for {
				ip := iter.Next()
				if ip == nil {
					break
				}
				ips = append(ips, ip)
			}

			if !reflect.DeepEqual(ips, test.want) {
				t.Fatalf("IPRanges(%v).Iterator() = %v, want %v", test.ranges, ips, test.want)
			}
		})
	}
}
