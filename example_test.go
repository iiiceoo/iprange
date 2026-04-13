package iprange_test

import (
	"fmt"
	"log"
	"math/big"
	"net/netip"

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
	// [172.18.0.1 172.18.0.0/24]
	// [fd00::1-fd00::a fd00::1-fd00::1:a]
}

func ExampleIPRanges_Contains() {
	ranges, err := iprange.Parse("172.18.0.0/24")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	fmt.Println(ranges.Contains(netip.MustParseAddr("172.18.0.1")))
	fmt.Println(ranges.Contains(netip.MustParseAddr("172.19.0.1")))
	fmt.Println(ranges.Contains(netip.MustParseAddr("fd00::1")))
	// Output:
	// true
	// false
	// false
}

func ExampleIPRanges_Union() {
	ranges1, _ := iprange.Parse("172.18.0.20-30", "172.18.0.1-25")
	ranges2, _ := iprange.Parse("172.18.0.5-25")

	fmt.Println(ranges1.Union(ranges2))
	fmt.Println(ranges1)
	// Output:
	// 172.18.0.1-172.18.0.30
	// [172.18.0.20-172.18.0.30 172.18.0.1-172.18.0.25]
}

func ExampleIPRanges_IPIterator() {
	ranges, _ := iprange.Parse("172.18.0.1-3")

	iter := ranges.IPIterator()
	for {
		addr := iter.Next()
		if !addr.IsValid() {
			break
		}
		fmt.Println(addr)
	}

	iter.Reset()
	for {
		addr := iter.NextN(big.NewInt(2))
		if !addr.IsValid() {
			break
		}
		fmt.Println(addr)
	}

	// Output:
	// 172.18.0.1
	// 172.18.0.2
	// 172.18.0.3
	// 172.18.0.2
}

func ExampleIPRanges_BlockIterator() {
	ranges, err := iprange.Parse("172.18.0.0-4")
	if err != nil {
		log.Fatalf("error parsing IP ranges: %v", err)
	}

	iter := ranges.BlockIterator(big.NewInt(2))
	for {
		ip := iter.Next()
		if ip == nil {
			break
		}
		fmt.Println(ip)
	}

	iter.Reset()
	n := big.NewInt(3)
	for {
		ip := iter.NextN(n)
		if ip == nil {
			break
		}
		fmt.Println(ip)
	}

	// Output:
	// 172.18.0.0/31
	// 172.18.0.2/31
	// 172.18.0.4
	// 172.18.0.4
}
