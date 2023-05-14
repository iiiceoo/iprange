package iprange

import (
	"math/big"
	"net"
)

type ipIterator struct {
	ranges     []ipRange
	rangeIndex int
	current    xIP
}

// IPIterator generates a new iterator for scanning IP addresses. Call
// the method Next to get the next IP address.
func (rr *IPRanges) IPIterator() *ipIterator {
	return &ipIterator{
		ranges: rr.ranges,
	}
}

// Next returns the next IP address. If the ipIterator has been exhausted,
// return nil.
func (ii *ipIterator) Next() net.IP {
	n := len(ii.ranges)
	if n == 0 {
		return nil
	}

	if ii.current.IP == nil {
		ii.current.IP = ii.ranges[0].start.IP
		return ii.current.IP
	}

	if !ii.current.Equal(ii.ranges[ii.rangeIndex].end.IP) {
		ii.current = ii.current.next()
		return ii.current.IP
	}

	ii.rangeIndex++
	if ii.rangeIndex == n {
		return nil
	}
	ii.current = ii.ranges[ii.rangeIndex].start

	return ii.current.IP
}

type cidrIterator struct {
	ranges     []ipRange
	rangeIndex int

	ipBitLen int
	lastInt  *big.Int
	current  *big.Int
}

// CIDRIterator generates a new iterator for scanning CIDR. Call the
// method Next to get the next CIDR.
func (rr *IPRanges) CIDRIterator() *cidrIterator {
	iter := &cidrIterator{
		ranges: rr.ranges,
	}

	if len(iter.ranges) != 0 {
		r := iter.ranges[0]
		iter.ipBitLen = len(r.start.IP) * 8
		iter.lastInt = ipToInt(r.end.IP)
		iter.current = ipToInt(r.start.IP)
	}

	return iter
}

// Next returns the next CIDR. If the cidrIterator has been exhausted,
// return nil.
func (ci *cidrIterator) Next() *net.IPNet {
	n := len(ci.ranges)
	if n == 0 {
		return nil
	}

	if ci.current.Cmp(ci.lastInt) <= 0 {
		return ci.next()
	}

	ci.rangeIndex++
	if ci.rangeIndex == n {
		return nil
	}
	ci.lastInt = ipToInt(ci.ranges[ci.rangeIndex].end.IP)
	ci.current = ipToInt(ci.ranges[ci.rangeIndex].start.IP)

	return ci.next()
}

func (ci *cidrIterator) next() *net.IPNet {
	delta := big.NewInt(0)
	delta.Sub(ci.lastInt, ci.current)
	delta.Add(delta, big.NewInt(1))

	curIP := intToIP(ci.current)
	nbits := min(righthandZeroBits(curIP), delta.BitLen()-1)

	incr := big.NewInt(1)
	incr.Lsh(incr, uint(nbits))
	ci.current.Add(ci.current, incr)

	return &net.IPNet{
		IP:   curIP,
		Mask: net.CIDRMask(ci.ipBitLen-nbits, ci.ipBitLen),
	}
}

// righthandZeroBits counts the number of zero bits on the right hand
// side of bb.
func righthandZeroBits(bb []byte) int {
	n := 0
	for i := len(bb) - 1; i >= 0; i-- {
		b := bb[i]
		if b == 0 {
			n += 8
			continue
		}

		for b&1 == 0 {
			n++
			b >>= 1
		}
		break
	}

	return n
}
