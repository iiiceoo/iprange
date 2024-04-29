package iprange

import (
	"math/big"
	"net"
)

type ipIterator struct {
	ranges     []ipRange
	rangeIndex int
	current    xIP
	freeSize   *big.Int

	one *big.Int
}

// IPIterator generates a new iterator for scanning IP addresses.
func (rr *IPRanges) IPIterator() *ipIterator {
	return &ipIterator{
		ranges: rr.ranges,
		one:    big.NewInt(1),
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
		ii.freeSize = new(big.Int)
		ii.freeSize.Set(ii.ranges[0].size())
		ii.freeSize.Sub(ii.freeSize, ii.one)
		ii.current.IP = ii.ranges[0].start.IP
		return ii.current.IP
	}

	if !ii.current.Equal(ii.ranges[ii.rangeIndex].end.IP) {
		ii.freeSize.Sub(ii.freeSize, ii.one)
		ii.current = ii.current.next()
		return ii.current.IP
	}

	ii.rangeIndex++
	if ii.rangeIndex == n {
		return nil
	}
	ii.freeSize.Set(ii.ranges[ii.rangeIndex].size())
	ii.freeSize.Sub(ii.freeSize, ii.one)
	ii.current = ii.ranges[ii.rangeIndex].start

	return ii.current.IP
}

// NextN returns the next nth IP address. If the ipIterator has been exhausted,
// return nil. If n <= 0, it is equivalent to NextN(1).
func (ii *ipIterator) NextN(n *big.Int) net.IP {
	l := len(ii.ranges)
	if l == 0 {
		return nil
	}

	if n.Sign() <= 0 {
		n = big.NewInt(1)
	}

	if ii.current.IP == nil {
		n = new(big.Int).Set(n)
		for {
			size := ii.ranges[ii.rangeIndex].size()
			if n.Cmp(size) <= 0 {
				ii.freeSize = new(big.Int)
				ii.freeSize.Set(size)
				ii.freeSize.Sub(ii.freeSize, n)
				n.Sub(n, ii.one)
				ii.current.IP = ii.ranges[ii.rangeIndex].start.nextN(n).IP
				return ii.current.IP
			}

			n.Sub(n, size)
			ii.rangeIndex++
			if ii.rangeIndex == l {
				return nil
			}
		}
	}

	if n.Cmp(ii.freeSize) <= 0 {
		ii.freeSize.Sub(ii.freeSize, n)
		ii.current.IP = ii.current.nextN(n).IP
		return ii.current.IP
	}

	ii.rangeIndex++
	if ii.rangeIndex == l {
		return nil
	}
	n = new(big.Int).Set(n)
	n.Sub(n, ii.freeSize)
	ii.freeSize.Set(ii.ranges[ii.rangeIndex].size())
	ii.freeSize.Sub(ii.freeSize, n)
	n.Sub(n, ii.one)
	ii.current.IP = ii.ranges[ii.rangeIndex].start.nextN(n).IP

	return ii.current.IP
}

// Reset resets IP iterator.
func (ii *ipIterator) Reset() {
	ii.rangeIndex = 0
	ii.current.IP = nil
}

type blockIterator struct {
	ranges    *IPRanges
	size      *big.Int
	blockSize *big.Int
	start     *big.Int
	end       *big.Int
}

// BlockIterator generates a new iterator for scanning IP blocks. blockSize
// should be at least 1, which is somewhat equivalent to IPIterator.
func (rr *IPRanges) BlockIterator(blockSize *big.Int) *blockIterator {
	if blockSize == nil || blockSize.Sign() <= 0 {
		blockSize = big.NewInt(1)
	}

	return &blockIterator{
		ranges:    rr,
		size:      rr.Size(),
		blockSize: blockSize,
	}
}

// Next returns the next IP block. If the blockIterator has been exhausted,
// return nil.
func (bi *blockIterator) Next() *IPRanges {
	if bi.size.Sign() == 0 {
		return nil
	}

	if bi.start == nil {
		bi.start = big.NewInt(0)
		bi.end = new(big.Int).Sub(bi.blockSize, big.NewInt(1))
		return bi.ranges.Slice(bi.start, bi.end)
	}

	bi.start.Add(bi.start, bi.blockSize)
	if bi.start.Cmp(bi.size) >= 0 {
		return nil
	}
	bi.end.Add(bi.end, bi.blockSize)

	return bi.ranges.Slice(bi.start, bi.end)
}

// NextN returns the next nth IP block. If the blockIterator has been
// exhausted, return nil.  If n <= 0, it is equivalent to NextN(1).
func (bi *blockIterator) NextN(n *big.Int) *IPRanges {
	if bi.size.Sign() == 0 {
		return nil
	}

	if n.Sign() <= 0 {
		n = big.NewInt(1)
	}

	if bi.start == nil {
		one := big.NewInt(1)
		n = new(big.Int).Sub(n, one)
		bi.start = new(big.Int).Mul(n, bi.blockSize)
		bi.end = new(big.Int).Add(bi.start, bi.blockSize)
		bi.end.Sub(bi.end, one)
		return bi.ranges.Slice(bi.start, bi.end)
	}

	step := new(big.Int).Mul(n, bi.blockSize)
	bi.start.Add(bi.start, step)
	if bi.start.Cmp(bi.size) >= 0 {
		return nil
	}
	bi.end.Add(bi.end, step)

	return bi.ranges.Slice(bi.start, bi.end)
}

// Reset resets IP block iterator.
func (bi *blockIterator) Reset() {
	bi.start = nil
	bi.end = nil
}

type cidrIterator struct {
	ranges     []ipRange
	rangeIndex int

	ipBitLen int
	lastInt  *big.Int
	current  *big.Int

	one *big.Int
}

// CIDRIterator generates a new iterator for scanning CIDR.
func (rr *IPRanges) CIDRIterator() *cidrIterator {
	iter := &cidrIterator{
		ranges: rr.ranges,
		one:    big.NewInt(1),
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
	delta := new(big.Int).Sub(ci.lastInt, ci.current)
	delta.Add(delta, ci.one)

	curIP := intToIP(ci.current)
	nbits := min(righthandZeroBits(curIP), delta.BitLen()-1)

	incr := new(big.Int).Lsh(ci.one, uint(nbits))
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
