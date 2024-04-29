package iprange

import (
	"fmt"
	"math/big"
	"net"
	"sort"

	"github.com/brunoga/deep"
)

// family defines the version of IP.
type family int

// Standard IP version 4 or 6. Unknown represents an invalid IP version,
// which is commonly used in the zero value of an IPRanges struct or to
// distinguish an invalid xIP.
const (
	Unknown family = iota
	IPv4
	IPv6
)

// String implements fmt.Stringer.
func (f family) String() string {
	if f == IPv4 {
		return "IPv4"
	}
	if f == IPv6 {
		return "IPv6"
	}

	return "Unknown"
}

// IPRanges is a set of ipRange that uses the starting and ending IP
// addresses to represent any IP range of any size. The following IP
// range formats are valid:
//
//	172.18.0.1              fd00::1
//	172.18.0.0/24           fd00::/64
//	172.18.0.1-10           fd00::1-a
//	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a
//
// Dual-stack IP ranges are not allowed, The IP version of an IPRanges
// can only be IPv4, IPv6, or unknown (zero value).
type IPRanges struct {
	version family
	ranges  []ipRange
}

// Parse parses a set of IP range format strings as IPRanges, the slice
// of ipRange with the same IP version, which records the starting and
// ending IP addresses.
//
// The error errInvalidIPRangeFormat wiil be returned when one of IP range
// string is invalid. And dual-stack IP ranges are not allowed, the error
// errDualStackIPRanges occurs when parsing a set of IP range strings, where
// there are both IPv4 and IPv6 addresses.
func Parse(rs ...string) (*IPRanges, error) {
	if len(rs) == 0 {
		return &IPRanges{}, nil
	}

	version := Unknown
	ranges := make([]ipRange, 0, len(rs))
	for i, r := range rs {
		v, err := parse(r)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			version = IPv4
			if v.start.To4() == nil {
				version = IPv6
			}
		}

		if v.start.version() != version {
			return nil, errDualStackIPRanges
		}
		ranges = append(ranges, *v)
	}

	return &IPRanges{
		version: version,
		ranges:  ranges,
	}, nil
}

// Version returns the IP version of IPRanges:
//
//	1: IPv4
//	2: IPv6
//	0: zero value of IPRanges
//
// Do not compare family with a regular int value, which is confusing.
// Use predefined const such as IPv4, IPv6, or Unknown.
func (rr *IPRanges) Version() family {
	return rr.version
}

// Contains reports whether IPRanges rr contain net.IP ip. If rr is IPv4
// and ip is IPv6, then it is also considered not contained, and vice versa.
func (rr *IPRanges) Contains(ip net.IP) bool {
	w := xIP{ip}
	if w.version() != rr.version {
		return false
	}

	for _, r := range rr.ranges {
		if r.contains(ip) {
			return true
		}
	}

	return false
}

// MergeEqual reports whether IPRanges rr is equal to rr2, but both rr and
// rr2 are pre-merged, which means they are both ordered and deduplicated.
func (rr *IPRanges) MergeEqual(rr2 *IPRanges) bool {
	if rr.version != rr2.version {
		return false
	}

	rr = rr.DeepCopy().Merge()
	rr2 = rr2.DeepCopy().Merge()

	return rr.Equal(rr2)
}

// Equal reports whether IPRanges rr is equal to rr2.
func (rr *IPRanges) Equal(rr2 *IPRanges) bool {
	if rr.version != rr2.version {
		return false
	}

	n := len(rr.ranges)
	if len(rr2.ranges) != n {
		return false
	}

	for i := 0; i < n; i++ {
		if !rr.ranges[i].equal(&rr2.ranges[i]) {
			return false
		}
	}

	return true
}

// Size calculates the total number of IP addresses that pertain to
// IPRanges rr.
func (rr *IPRanges) Size() *big.Int {
	n := big.NewInt(0)
	for _, r := range rr.ranges {
		n.Add(n, r.size())
	}

	return n
}

// Merge merges the duplicate parts of multiple ipRanges in rr and sort
// them by their respective starting xIP.
func (rr *IPRanges) Merge() *IPRanges {
	if len(rr.ranges) <= 1 {
		return rr
	}

	sort.Slice(rr.ranges, func(i, j int) bool {
		return rr.ranges[i].start.cmp(rr.ranges[j].start) < 0
	})

	cur := -1
	merged := make([]ipRange, 0, len(rr.ranges))
	for _, r := range rr.ranges {
		if cur == -1 {
			merged = append(merged, r)
			cur++
			continue
		}

		if merged[cur].end.next().cmp(r.start) == 0 {
			merged[cur].end = r.end
			continue
		}

		if merged[cur].end.cmp(r.start) < 0 {
			merged = append(merged, r)
			cur++
			continue
		}

		if merged[cur].end.cmp(r.end) < 0 {
			merged[cur].end = r.end
		}
	}
	rr.ranges = merged

	return rr
}

// IsOverlap reports whether IPRanges rr have overlapping parts.
func (rr *IPRanges) IsOverlap() bool {
	n := len(rr.ranges)
	if n <= 1 {
		return false
	}

	rs := rr.DeepCopy().ranges
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].start.cmp(rs[j].start) < 0
	})

	for i := 0; i < n-1; i++ {
		if rs[i].end.cmp(rs[i+1].start) >= 0 {
			return true
		}
	}

	return false
}

// Union calculates the union of IPRanges rr and rs with the same IP
// version. The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] U [172.18.0.5-25]
//	Output: [172.18.0.1-30]
func (rr *IPRanges) Union(rs *IPRanges) *IPRanges {
	if rr.version != rs.version {
		return rr.Merge()
	}
	rr.ranges = append(rr.ranges, rs.ranges...)

	return rr.Merge()
}

// Diff calculates the difference of IPRanges rr and rs with the same IP
// version. The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] - [172.18.0.5-25]
//	Output: [172.18.0.1-4, 172.18.0.26-30]
func (rr *IPRanges) Diff(rs *IPRanges) *IPRanges {
	if rr.version != rs.version {
		return rr.Merge()
	}

	if len(rr.ranges) == 0 || len(rs.ranges) == 0 {
		return rr.Merge()
	}

	rs = rs.DeepCopy()
	omr := rr.Merge().ranges
	tmr := rs.Merge().ranges
	n1, n2 := len(omr), len(tmr)
	ranges := make([]ipRange, 0, n1+n2)

	i, j := 0, 0
	for i < n1 && j < n2 {
		// The following are all distributions of the difference sets between two
		// IP range A and B (IP range A - IP range B).
		//
		// For convenience, use symbols to distinguish between two IP ranges:
		//   A: *------*
		//   B: `------`

		// *------*
		//           `------`
		if omr[i].end.cmp(tmr[j].start) < 0 {
			ranges = append(ranges, omr[i])
			i++
			continue
		}

		//           *------*
		// `------`
		if omr[i].start.cmp(tmr[j].end) > 0 {
			j++
			continue
		}

		if omr[i].end.cmp(tmr[j].end) <= 0 {
			// *------*
			//     `------`
			if omr[i].start.cmp(tmr[j].start) < 0 {
				ranges = append(ranges, ipRange{
					start: omr[i].start,
					end:   tmr[j].start.prev(),
				})
			}

			//     *--*
			// `----------`
			i++
			continue
		}

		// *----------*
		//     `--`
		if omr[i].start.cmp(tmr[j].start) < 0 {
			ranges = append(ranges, ipRange{
				start: omr[i].start,
				end:   tmr[j].start.prev(),
			})
		}

		//     *------*
		// `------`
		omr[i].start = (tmr[j].end.next())
		j++
	}

	if j == n2 && tmr[j-1].end.cmp(omr[i].end) < 0 {
		ranges = append(ranges, omr[i])
	}

	if i+1 < n1 {
		ranges = append(ranges, omr[i+1:]...)
	}
	rr.ranges = ranges

	return rr
}

// Intersect calculates the intersection of IPRanges rr and rs with the
// same IP version. The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] ∩ [172.18.0.5-25]
//	Output: [172.18.0.5-25]
func (rr *IPRanges) Intersect(rs *IPRanges) *IPRanges {
	if rr.version != rs.version {
		return rr.Merge()
	}

	if len(rr.ranges) == 0 || len(rs.ranges) == 0 {
		return rr.Merge()
	}

	rs = rs.DeepCopy()
	omr := rr.Merge().ranges
	tmr := rs.Merge().ranges
	n1, n2 := len(omr), len(tmr)
	ranges := make([]ipRange, 0, max(n1, n2))

	for i, j := 0, 0; i < n1 && j < n2; {
		start := maxXIP(omr[i].start, tmr[j].start)
		end := minXIP(omr[i].end, tmr[j].end)
		if start.cmp(end) <= 0 {
			ranges = append(ranges, ipRange{
				start: start,
				end:   end,
			})
		}

		if omr[i].end.cmp(tmr[j].end) < 0 {
			i++
		} else {
			j++
		}
	}
	rr.ranges = ranges

	return rr
}

// Slice returns a slice of IPRanges, supporting negative indexes.
func (rr *IPRanges) Slice(start, end *big.Int) *IPRanges {
	size := rr.Size()
	version := rr.version
	rs := &IPRanges{version: version}
	if size.Sign() == 0 {
		return rs
	}

	if start.Sign() < 0 {
		start = new(big.Int).Add(start, size)
		if start.Sign() < 0 {
			start = big.NewInt(0)
		}
	} else {
		start = new(big.Int).Set(start)
	}

	if end.Sign() < 0 {
		end = new(big.Int).Add(end, size)
		if end.Sign() < 0 {
			return rs
		}
	} else {
		end = new(big.Int).Set(end)
	}

	if start.Cmp(end) > 0 {
		return rs
	}

	var ranges []ipRange
	for i := 0; i < len(rr.ranges); i++ {
		size := rr.ranges[i].size()
		if start.Cmp(size) >= 0 {
			start.Sub(start, size)
			end.Sub(end, size)
			continue
		}

		if start.Sign() >= 0 {
			if end.Cmp(size) < 0 {
				ranges = append(ranges, ipRange{
					start: rr.ranges[i].start.nextN(start),
					end:   rr.ranges[i].start.nextN(end),
				})
				break
			}

			ranges = append(ranges, ipRange{
				start: rr.ranges[i].start.nextN(start),
				end:   rr.ranges[i].end,
			})
			start.Sub(start, size)
			end.Sub(end, size)
			continue
		}

		if end.Cmp(size) >= 0 {
			ranges = append(ranges, ipRange{
				start: rr.ranges[i].start,
				end:   rr.ranges[i].end,
			})
			end.Sub(end, size)
			continue
		}

		ranges = append(ranges, ipRange{
			start: rr.ranges[i].start,
			end:   rr.ranges[i].start.nextN(end),
		})
		break
	}
	rs.ranges = ranges

	return rs
}

func (rr *IPRanges) DeepCopy() *IPRanges {
	return deep.MustCopy(rr)
}

// String implements fmt.Stringer.
func (rr *IPRanges) String() string {
	ss := rr.Strings()
	if len(ss) == 1 {
		return ss[0]
	}

	return fmt.Sprint(ss)
}

// Strings returns a slice of the string representations of the IPRanges rr.
func (rr *IPRanges) Strings() []string {
	if len(rr.ranges) == 0 {
		return nil
	}

	ss := make([]string, 0, len(rr.ranges))
	for _, r := range rr.ranges {
		ss = append(ss, r.String())
	}

	return ss
}
