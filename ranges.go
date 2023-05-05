package iprange

import (
	"fmt"
	"math/big"
	"net"
	"sort"
)

// IPRanges is a set of ipRange that uses the starting and ending IP
// addresses to represent any IP range of any size. The following IP
// range formats are valid:
//
//	172.18.0.1              fd00::1
//	172.18.0.0/24           fd00::/64
//	172.18.0.1-10           fd00::1-a
//	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a
//
// Never use Go's append to combine two IPRanges, but rather use its
// interval operation method Union, Diff or Intersect.
type IPRanges []ipRange

// Parse parses a set of IP range format strings as IPRanges, the slice
// of ipRange, which records the starting and ending IP addresses.
//
// The error errInvalidIPRangeFormat wiil be returned when one of IP range
// string is invalid. And dual-stack IP ranges are not allowed, the error
// errDualStackIPRanges occurs when parsing a set of IP range strings, where
// there are both IPv4 and IPv6 addresses.
func Parse(rs ...string) (IPRanges, error) {
	if len(rs) == 0 {
		return nil, fmt.Errorf("%w: []", errInvalidIPRangeFormat)
	}

	if len(rs) == 1 {
		v, err := parse(rs[0])
		if err != nil {
			return nil, err
		}
		return IPRanges{v}, nil
	}

	n := 0
	res := make(IPRanges, 0, len(rs))
	for i, r := range rs {
		v, err := parse(r)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			n = len(v.start.IP)
		}

		if len(v.start.IP) != n {
			return nil, errDualStackIPRanges
		}

		res = append(res, v)
	}

	return res, nil
}

// Contains reports whether IPRanges rr contain net.IP ip. If rr are IPv4
// and ip is IPv6, then it is also considered not contained, and vice versa.
func (rr IPRanges) Contains(ip net.IP) bool {
	for _, r := range rr {
		if r.contains(ip) {
			return true
		}
	}

	return false
}

// MergeEqual reports whether IPRanges rr1 are equal to rr2, but both rr1
// and rr2 are pre-merged, which means they are both ordered and deduplicated.
func (rr1 IPRanges) MergeEqual(rr2 IPRanges) bool {
	return rr1.Merge().Equal(rr2.Merge())
}

// Equal reports whether IPRanges rr1 are equal to rr2.
func (rr1 IPRanges) Equal(rr2 IPRanges) bool {
	n := len(rr1)
	if len(rr2) != n {
		return false
	}

	for i := 0; i < n; i++ {
		if !rr1[i].equal(rr2[i]) {
			return false
		}
	}

	return true
}

// Size calculates the total number of IP addresses that pertain to
// IPRanges rr.
func (rr IPRanges) Size() *big.Int {
	n := big.NewInt(0)
	for _, r := range rr.Merge() {
		n.Add(n, r.size())
	}

	return n
}

// Merge merges the duplicate parts of multiple ipRanges in rr and sort
// them by their respective starting xIP.
func (rr IPRanges) Merge() IPRanges {
	n := len(rr)
	if n <= 1 {
		return rr
	}

	rc := make(IPRanges, n)
	copy(rc, rr)
	sort.Slice(rc, func(i, j int) bool {
		return rc[i].start.cmp(rc[j].start) < 0
	})

	cur := -1
	merge := make(IPRanges, 0, n)
	for _, r := range rc {
		if cur == -1 {
			merge = append(merge, r)
			cur++
			continue
		}

		// When rr is an invalid IP ranges mixed with IPv4 and IPv6, it can cause
		// unpredictable chaos. So don't try to aggregate a set of IPRanges through
		// the Go's append, instead use the method Union.
		if merge[cur].end.cmp(r.start) < 0 {
			merge = append(merge, r)
			cur++
			continue
		}

		if merge[cur].end.cmp(r.end) < 0 {
			merge[cur].end = r.end
		}
	}

	return merge
}

// Union calculates the union of IPRanges rr and rs. The result is always
// merged (ordered and deduplicated), for instance:
//
//	do:  [172.18.0.20-30, 172.18.0.1-25] U [172.18.0.5-25]
//	res: [172.18.0.1-30]
func (rr IPRanges) Union(rs IPRanges) IPRanges {
	return append(rr, rs...).Merge()
}

// Diff calculates the difference of IPRanges rr and rs. The result is
// always merged (ordered and deduplicated), for instance:
//
//	do:  [172.18.0.20-30, 172.18.0.1-25] - [172.18.0.5-25]
//	res: [172.18.0.1-4, 172.18.0.26-30]
func (rr IPRanges) Diff(rs IPRanges) IPRanges {
	if len(rr) == 0 || len(rs) == 0 {
		return rr.Merge()
	}

	om := rr.Merge()
	tm := IPRanges(rs).Merge()
	n1, n2 := len(om), len(tm)
	if n1 == 0 || n2 == 0 {
		return om
	}

	i, j := 0, 0
	res := make(IPRanges, 0, n1+n2)
	for i < n1 && j < n2 {
		// The following are all distributions of the difference sets between two
		// IP range A and B (IP range A - IP range B).
		//
		// For convenience, use symbols to distinguish between two IP ranges:
		//   A: *------*
		//   B: `------`

		// *------*
		//           `------`
		if om[i].end.cmp(tm[j].start) < 0 {
			res = append(res, om[i])
			i++
			continue
		}

		//           *------*
		// `------`
		if om[i].start.cmp(tm[j].end) > 0 {
			j++
			continue
		}

		if om[i].end.cmp(tm[j].end) <= 0 {
			// *------*
			//     `------`
			if om[i].start.cmp(tm[j].start) < 0 {
				res = append(res, ipRange{
					start: om[i].start,
					end:   tm[j].start.prev(),
				})
			}

			//     *--*
			// `----------`
			i++
			continue
		}

		// *----------*
		//     `--`
		if om[i].start.cmp(tm[j].start) < 0 {
			res = append(res, ipRange{
				start: om[i].start,
				end:   tm[j].start.prev(),
			})
		}

		//     *------*
		// `------`
		om[i].start = (tm[j].end.next())
		j++
	}

	if j == n2 && tm[j-1].end.cmp(om[i].end) < 0 {
		res = append(res, om[i])
	}

	if i+1 < n1 {
		res = append(res, om[i+1:]...)
	}

	return res
}

// Intersect calculates the intersection of IPRanges rr and rs. The result
// is always merged (ordered and deduplicated), for instance:
//
//	do:  [172.18.0.20-30, 172.18.0.1-25] ∩ [172.18.0.5-25]
//	res: [172.18.0.5-25]
func (rr IPRanges) Intersect(rs IPRanges) IPRanges {
	if len(rr) == 0 || len(rs) == 0 {
		return rr.Merge()
	}

	om := rr.Merge()
	tm := IPRanges(rs).Merge()
	n1, n2 := len(om), len(tm)
	if n1 == 0 || n2 == 0 {
		return om
	}

	res := make(IPRanges, 0, max(n1, n2))
	for i, j := 0, 0; i < n1 && j < n2; {
		start := maxXIP(om[i].start, tm[j].start)
		end := minXIP(om[i].end, tm[j].end)
		if start.cmp(end) < 0 {
			res = append(res, ipRange{
				start: start,
				end:   end,
			})
		}

		if om[i].end.cmp(tm[j].end) < 0 {
			i++
		} else {
			j++
		}
	}

	return res
}

// max returns the larger of x and y.
func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

// maxXIP returns the larger xIP between x and y.
func maxXIP(x, y xIP) xIP {
	if x.cmp(y) > 0 {
		return x
	}

	return y
}

// minXIP returns the smaller xIP between x and y.
func minXIP(x, y xIP) xIP {
	if x.cmp(y) < 0 {
		return x
	}

	return y
}
