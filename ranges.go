package iprange

import (
	"fmt"
	"net"
	"sort"
)

type IPRanges []IPRange

func ParseRanges(rs ...string) (IPRanges, error) {
	if len(rs) == 0 {
		return nil, fmt.Errorf("%w: []", errInvalidIPRangeFormat)
	}

	if len(rs) == 1 {
		v, err := Parse(rs[0])
		if err != nil {
			return nil, err
		}
		return IPRanges{v}, nil
	}

	n := 0
	res := make(IPRanges, 0, len(rs))
	for i, r := range rs {
		v, err := Parse(r)
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

func (rr IPRanges) Contains(ip net.IP) bool {
	for _, r := range rr {
		if r.Contains(ip) {
			return true
		}
	}

	return false
}

func (rr1 IPRanges) MergeEqual(rr2 IPRanges) bool {
	return rr1.Merge().Equal(rr2.Merge())
}

func (rr1 IPRanges) Equal(rr2 IPRanges) bool {
	n := len(rr1)
	if len(rr2) != n {
		return false
	}

	for i := 0; i < n; i++ {
		if !rr1[i].Equal(rr2[i]) {
			return false
		}
	}

	return true
}

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
		if normalizeIP(r.start.IP) == nil {
			continue
		}

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

func (rr IPRanges) Union(rs ...IPRange) IPRanges {
	return append(rr, rs...).Merge()
}

func (rr IPRanges) Diff(rs ...IPRange) IPRanges {
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
				res = append(res, IPRange{
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
			res = append(res, IPRange{
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

func (rr IPRanges) Intersect(rs ...IPRange) IPRanges {
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
			res = append(res, IPRange{
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

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func maxXIP(x, y xIP) xIP {
	if x.cmp(y) > 0 {
		return x
	}

	return y
}

func minXIP(x, y xIP) xIP {
	if x.cmp(y) < 0 {
		return x
	}

	return y
}
