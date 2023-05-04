# iprange

[![GoDoc](https://godoc.org/github.com/iiiceoo/iprange?status.svg)](https://godoc.org/github.com/iiiceoo/iprange)

*The iprange package parses IPv4/IPv6 address from strings in IP range format and handle interval mathematics between multiple IP ranges.*

## IP range formats

- `172.18.0.1` / `fd00::1`
- `172.18.0.0/24` / `fd00::/64`
- `172.18.0.1-10` / `fd00::1-a`
- `172.18.0.1-172.18.1.10` / `fd00::1-fd00::1:a`

## Get started

```go
package main

import (
	"fmt"

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
		fmt.Printf("failed to parse IP ranges: %v\n", err)
		return
	}
	fmt.Printf("IP ranges: %+v\n", ranges)

	// Merge IP ranges.
	merged := ranges.Merge()
	fmt.Printf("Merged IP ranges: %+v\n", merged)

	// Interval mathematics of IP ranges.
	another, _ := iprange.Parse("172.18.0.0/24")
	diffSet := merged.Diff(another)
	fmt.Printf("Difference set between two IP ranges: %+v\n", diffSet)

	// Iterate through IP ranges.
	fmt.Println("Scan the difference set:")
	it := diffSet.Iterator()
	for {
		ip := it.Next()
		if ip == nil {
			break
		}
		fmt.Println(ip)
	}
}
```

## License

iprange is [MIT-Licensed](https://github.com/iiiceoo/iprange/blob/main/LICENSE).
