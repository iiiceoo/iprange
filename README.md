# iprange

[![GoDoc](https://godoc.org/github.com/iiiceoo/iprange?status.svg)](https://godoc.org/github.com/iiiceoo/iprange)
[![codecov](https://codecov.io/gh/iiiceoo/iprange/branch/main/graph/badge.svg?token=7STDXD53G0)](https://codecov.io/gh/iiiceoo/iprange)

*The package iprange parses IPv4/IPv6 addresses from strings in IP range format and handles interval mathematics between multiple IP ranges.*

## IP range formats

- `172.18.0.1` / `fd00::1`
- `172.18.0.0/24` / `fd00::/64`
- `172.18.0.1-10` / `fd00::1-a`
- `172.18.0.1-172.18.1.10` / `fd00::1-fd00::1:a`

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/iiiceoo/iprange"
)

func main() {
	// Parse IP ranges.
	ranges, err := iprange.Parse(
		"172.18.0.1",
		"172.18.0.0/24",
		"172.18.0.1-10",
		"172.18.0.1-172.18.1.10",
	)
	if err != nil {
		log.Fatalf("failed to parse IP ranges: %v\n", err)
	}
	fmt.Printf("%s IP ranges: %s\n", ranges.Version(), ranges)

	// Merge IP ranges.
	merged := ranges.Merge()
	fmt.Printf("Merged IP ranges: %s\n", merged)

	// Interval mathematics of IP ranges.
	another, _ := iprange.Parse("172.18.0.0/24")
	diffSet := merged.Diff(another)
	fmt.Printf("The difference between two IP ranges: %s\n", diffSet)

	// Iterate through IP ranges.
	fmt.Printf("Scan all %d IP addresses in the set:\n", diffSet.Size())
	iter := diffSet.Iterator()
	for {
		ip := iter.Next()
		if ip == nil {
			break
		}
		fmt.Println(ip)
	}
}
```
