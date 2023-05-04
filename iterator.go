package iprange

import "net"

type rangesIterator struct {
	ranges     IPRanges
	rangeIndex int
	current    xIP
}

func (rr IPRanges) Iterator() *rangesIterator {
	return &rangesIterator{
		ranges: rr.Merge(),
	}
}

func (ri *rangesIterator) Next() net.IP {
	n := len(ri.ranges)
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
