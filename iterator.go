package iprange

import "net"

type rangesIterator struct {
	ranges     []ipRange
	rangeIndex int
	current    xIP
}

// Iterator generates a new iterator for IPRanges rr, which stores the merged
// IP ranges (ordered and deduplicated) and always points the cursor to the
// first IP address. Call Next to iterate through the IPRanges.
func (rr *IPRanges) Iterator() *rangesIterator {
	return &rangesIterator{
		ranges: rr.Merge().ranges,
	}
}

// Next returns the next IP address. If the rangesIterator has been exhausted,
// return nil.
func (ri *rangesIterator) Next() net.IP {
	n := len(ri.ranges)
	// ri.ranges is an empty slice or ri.current equals to the last IP address.
	if n == ri.rangeIndex {
		return nil
	}

	r := ri.ranges[ri.rangeIndex]
	if ri.current.IP == nil {
		ri.current.IP = r.start.IP
		return ri.current.IP
	}

	if !ri.current.Equal(r.end.IP) {
		ri.current = ri.current.next()
		return ri.current.IP
	}

	ri.rangeIndex++
	if ri.rangeIndex == n {
		return nil
	}
	r = ri.ranges[ri.rangeIndex]
	ri.current = r.start

	return ri.current.IP
}
