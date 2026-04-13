package iprange

import (
	"fmt"
	"math/big"
	"net/netip"
	"slices"
)

// family defines the version of IP.
type family int

// Standard IP version 4 or 6.
const (
	IPv4 family = iota
	IPv6
)

// String implements fmt.Stringer.
func (f family) String() string {
	if f == IPv4 {
		return "IPv4"
	}

	return "IPv6"
}

// IPRanges is a set of ipRange that uses the starting and ending IP addresses
// to represent any IP range of any size. The following IP range formats are
// valid:
//
//	172.18.0.1              fd00::1
//	172.18.0.0/24           fd00::/64
//	172.18.0.1-10           fd00::1-a
//	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a
//
// Dual-stack IP ranges are not allowed, The IP version of an IPRanges can only
// be IPv4 or IPv6.
type IPRanges struct {
	version family
	ranges  []ipRange
}

// Parse parses a set of IP range format strings as IPRanges, the slice of
// ipRange with the same IP version, which records the starting and ending IP
// addresses.
func Parse(formats ...string) (*IPRanges, error) {
	if len(formats) == 0 {
		return &IPRanges{}, nil
	}

	version := IPv4
	ranges := make([]ipRange, 0, len(formats))
	for i, f := range formats {
		r, err := parse(f)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			version = r.version()
		}

		if i > 0 && r.version() != version {
			return nil, errDualStackIPRanges
		}
		ranges = append(ranges, r)
	}

	return &IPRanges{
		version: version,
		ranges:  ranges,
	}, nil
}

// Version returns the IP version of IPRanges:
//
//	0: IPv4
//	1: IPv6
func (rs *IPRanges) Version() family {
	return rs.version
}

// Contains reports whether IPRanges rs contain netip.Addr addr. If rs is IPv4
// and ip is IPv6, then it is also considered not contained, and vice versa.
func (rs *IPRanges) Contains(addr netip.Addr) bool {
	if len(rs.ranges) == 0 {
		return false
	}

	addr = addr.Unmap()
	is4 := addr.Is4()
	if (rs.version == IPv6 && is4) || (rs.version == IPv4 && !is4) {
		return false
	}

	for _, r := range rs.ranges {
		if r.contains(addr) {
			return true
		}
	}

	return false
}

// MergeEqual reports whether IPRanges rs is equal to other, but both rs and
// other are pre-merged, which means they are both ordered and deduplicated.
func (rs *IPRanges) MergeEqual(other *IPRanges) bool {
	if rs.version != other.version {
		return false
	}

	return rs.Merge().Equal(other.Merge())
}

// Equal reports whether IPRanges rs is equal to other.
func (rs *IPRanges) Equal(other *IPRanges) bool {
	if rs.version != other.version {
		return false
	}

	n := len(rs.ranges)
	if len(other.ranges) != n {
		return false
	}

	for i := 0; i < n; i++ {
		if !rs.ranges[i].equal(other.ranges[i]) {
			return false
		}
	}

	return true
}

// Size calculates the total number of IP addresses that pertain to IPRanges rs.
func (rs *IPRanges) Size() *big.Int {
	n := big.NewInt(0)
	buf := new(big.Int)
	for _, r := range rs.ranges {
		n.Add(n, r.size(buf))
	}

	return n
}

// Merge merges the duplicate parts of multiple ipRanges in rs and sort them by
// their respective starting IP.
func (rs *IPRanges) Merge() *IPRanges {
	if len(rs.ranges) <= 1 {
		return rs.Clone()
	}

	newRS := rs.Clone()
	slices.SortFunc(newRS.ranges, func(a, b ipRange) int {
		if cmp := a.start.cmp(b.start); cmp != 0 {
			return cmp
		}
		return b.end.cmp(a.end)
	})

	merged := newRS.ranges[:0]
	for _, r := range newRS.ranges {
		if len(merged) == 0 {
			merged = append(merged, r)
			continue
		}

		last := &merged[len(merged)-1]
		if last.end.next().cmp(r.start) < 0 {
			merged = append(merged, r)
			continue
		}

		if last.end.cmp(r.end) < 0 {
			last.end = r.end
		}
	}

	newRS.ranges = slices.Clip(merged)
	return newRS
}

// IsOverlap reports whether IPRanges rs have overlapping parts.
func (rs *IPRanges) IsOverlap() bool {
	if len(rs.ranges) <= 1 {
		return false
	}

	size := rs.Size()
	merged := rs.Merge().Size()

	return size.Cmp(merged) > 0
}

// Union calculates the union of IPRanges rs and other with the same IP version.
// The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] U [172.18.0.5-25]
//	Output: [172.18.0.1-30]
func (rs *IPRanges) Union(other *IPRanges) *IPRanges {
	if rs.version != other.version {
		return rs.Merge()
	}

	out := rs.Clone()
	out.ranges = append(out.ranges, other.ranges...)

	return out.Merge()
}

// Diff calculates the difference of IPRanges rs and other with the same IP
// version. The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] - [172.18.0.5-25]
//	Output: [172.18.0.1-4, 172.18.0.26-30]
func (rs *IPRanges) Diff(other *IPRanges) *IPRanges {
	if rs.version != other.version {
		return rs.Merge()
	}

	if len(rs.ranges) == 0 || len(other.ranges) == 0 {
		return rs.Merge()
	}

	omr := rs.Merge().ranges
	tmr := other.Merge().ranges
	n1, n2 := len(omr), len(tmr)

	i, j := 0, 0
	var ranges []ipRange
	for i < n1 && j < n2 {
		// The following are all distributions of the difference sets between
		// two IP range A and B (IP range A - IP range B).
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
		omr[i].start = tmr[j].end.next()
		j++
	}

	if j == n2 && i < n1 {
		ranges = append(ranges, omr[i:]...)
	}

	return &IPRanges{
		version: rs.version,
		ranges:  ranges,
	}
}

// Intersect calculates the intersection of IPRanges rs and other with the same
// IP version. The result is always merged (ordered and deduplicated).
//
//	Input:  [172.18.0.20-30, 172.18.0.1-25] ∩ [172.18.0.5-25]
//	Output: [172.18.0.5-25]
func (rs *IPRanges) Intersect(other *IPRanges) *IPRanges {
	if rs.version != other.version {
		return &IPRanges{
			version: rs.version,
		}
	}

	if len(rs.ranges) == 0 || len(other.ranges) == 0 {
		return &IPRanges{
			version: rs.version,
		}
	}

	omr := rs.Merge().ranges
	tmr := other.Merge().ranges

	var ranges []ipRange
	for i, j := 0, 0; i < len(omr) && j < len(tmr); {
		start := maxIP(omr[i].start, tmr[j].start)
		end := minIP(omr[i].end, tmr[j].end)
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

	return &IPRanges{
		version: rs.version,
		ranges:  ranges,
	}
}

// Slice returns a slice of IPRanges, supporting negative indexes.
func (rs *IPRanges) Slice(start, end *big.Int) *IPRanges {
	size := rs.Size()
	out := &IPRanges{version: rs.version}
	if size.Sign() == 0 {
		return out
	}

	if start == nil {
		start = big.NewInt(0)
	}
	if end == nil {
		end = new(big.Int).Sub(size, bigInt[1])
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
			return out
		}
	} else {
		end = new(big.Int).Set(end)
	}

	if start.Cmp(end) > 0 {
		return out
	}

	var ranges []ipRange
	for i := 0; i < len(rs.ranges); i++ {
		rangeSize := rs.ranges[i].size(nil)
		if start.Cmp(rangeSize) >= 0 {
			start.Sub(start, rangeSize)
			end.Sub(end, rangeSize)
			continue
		}

		if end.Cmp(rangeSize) < 0 {
			ranges = append(ranges, ipRange{
				start: rs.ranges[i].start.nextN(start, nil),
				end:   rs.ranges[i].start.nextN(end, nil),
			})
			break
		}

		ranges = append(ranges, ipRange{
			start: rs.ranges[i].start.nextN(start, nil),
			end:   rs.ranges[i].end,
		})
		start.Sub(start, rangeSize)
		end.Sub(end, rangeSize)
	}

	out.ranges = ranges
	return out
}

// Clone creates a deep copy of IPRanges rs.
func (rs *IPRanges) Clone() *IPRanges {
	n := len(rs.ranges)
	if n == 0 {
		return &IPRanges{
			version: rs.version,
		}
	}

	ranges := make([]ipRange, n)
	copy(ranges, rs.ranges)

	return &IPRanges{
		version: rs.version,
		ranges:  ranges,
	}
}

// String implements fmt.Stringer.
func (rs *IPRanges) String() string {
	ss := rs.Strings()
	if len(ss) == 1 {
		return ss[0]
	}

	return fmt.Sprint(ss)
}

// Strings returns a slice of the string representations of the IPRanges rs.
func (rs *IPRanges) Strings() []string {
	ss := make([]string, 0, len(rs.ranges))
	for _, r := range rs.ranges {
		ss = append(ss, r.String())
	}

	return ss
}
