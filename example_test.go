package iprange_test

import (
	"fmt"
	"log"
	"net"

	"github.com/iiiceoo/iprange"
)

func ExampleParse() {
	v4Ranges, err := iprange.Parse("172.18.0.1", "172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	v6Ranges, err := iprange.Parse("fd00::1-a", "fd00::1-fd00::1:a")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(v4Ranges)
	fmt.Println(v6Ranges)
	// Output:
	// [172.18.0.1 172.18.0.0-172.18.0.255]
	// [fd00::1-fd00::a fd00::1-fd00::1:a]
}

func ExampleIPRanges_Version() {
	v4Ranges, err := iprange.Parse("172.18.0.1", "172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	v6Ranges, err := iprange.Parse("fd00::1-a", "fd00::1-fd00::1:a")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	zero := iprange.IPRanges{}

	fmt.Println(v4Ranges.Version())
	fmt.Println(v6Ranges.Version())
	fmt.Println(zero.Version())
	// Output:
	// IPv4
	// IPv6
	// Unknown
}

func ExampleIPRanges_Contains() {
	ranges, err := iprange.Parse("172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges.Contains(net.ParseIP("172.18.0.1")))
	fmt.Println(ranges.Contains(net.ParseIP("172.19.0.1")))
	fmt.Println(ranges.Contains(net.ParseIP("fd00::1")))
	// Output:
	// true
	// false
	// false
}

func ExampleIPRanges_MergeEqual() {
	ranges1, err := iprange.Parse("172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	ranges2, err := iprange.Parse("172.18.0.100-255", "172.18.0.0-200")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges1.MergeEqual(ranges2))
	// Output:
	// true
}

func ExampleIPRanges_Equal() {
	ranges1, err := iprange.Parse("172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	ranges2, err := iprange.Parse("172.18.0.100-255", "172.18.0.0-200")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges1.Equal(ranges2))
	fmt.Println(ranges1.Equal(ranges1))
	// Output:
	// false
	// true
}

func ExampleIPRanges_Size() {
	ranges, err := iprange.Parse("172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	zero := iprange.IPRanges{}

	fmt.Println(ranges.Size())
	fmt.Println(zero.Size())
	// Output:
	// 256
	// 0
}

func ExampleIPRanges_Merge() {
	ranges, err := iprange.Parse("172.18.1.1", "172.18.0.100-200", "172.18.0.1-150")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges)
	fmt.Println(ranges.Merge())
	// Output:
	// [172.18.1.1 172.18.0.100-172.18.0.200 172.18.0.1-172.18.0.150]
	// [172.18.0.1-172.18.0.200 172.18.1.1]
}

func ExampleIPRanges_Union() {
	ranges1, err := iprange.Parse("172.18.0.20-30", "172.18.0.1-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	ranges2, err := iprange.Parse("172.18.0.5-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges1.Union(ranges2))
	// Output:
	// [172.18.0.1-172.18.0.30]
}

func ExampleIPRanges_Diff() {
	ranges1, err := iprange.Parse("172.18.0.20-30", "172.18.0.1-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	ranges2, err := iprange.Parse("172.18.0.5-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges1.Diff(ranges2))
	// Output:
	// [172.18.0.1-172.18.0.4 172.18.0.26-172.18.0.30]
}

func ExampleIPRanges_Intersect() {
	ranges1, err := iprange.Parse("172.18.0.20-30", "172.18.0.1-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}
	ranges2, err := iprange.Parse("172.18.0.5-25")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges1.Intersect(ranges2))
	// Output:
	// [172.18.0.5-172.18.0.25]
}

func ExampleIPRanges_Iterator() {
	ranges, err := iprange.Parse("172.18.0.1-3")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	iter := ranges.Iterator()
	for {
		ip := iter.Next()
		if ip == nil {
			break
		}
		fmt.Println(ip)
	}
	// Output:
	// 172.18.0.1
	// 172.18.0.2
	// 172.18.0.3
}
