package iprange

import (
	"math/big"
	"net/netip"
)

type IPIterator struct {
	ranges     []ipRange
	rangeIndex int
	current    ip
	freeSize   *big.Int

	buf *big.Int
}

// IPIterator generates a new iterator for scanning IP addresses.
func (rs *IPRanges) IPIterator() *IPIterator {
	return &IPIterator{
		ranges: rs.ranges,
		buf:    new(big.Int),
	}
}

// Next returns the next IP address. If the IPIterator has been exhausted,
// return the zero Addr.
func (ii *IPIterator) Next() netip.Addr {
	n := len(ii.ranges)
	if n == 0 {
		return netip.Addr{}
	}

	if !ii.current.addr.IsValid() {
		ii.freeSize = new(big.Int)
		ii.freeSize.Set(ii.ranges[0].size(ii.buf))
		ii.freeSize.Sub(ii.freeSize, bigInt[1])
		ii.current = ii.ranges[0].start
		return ii.current.addr
	}

	if ii.current.cmp(ii.ranges[ii.rangeIndex].end) != 0 {
		ii.freeSize.Sub(ii.freeSize, bigInt[1])
		ii.current = ii.current.next()
		return ii.current.addr
	}

	ii.rangeIndex++
	if ii.rangeIndex == n {
		return netip.Addr{}
	}

	ii.freeSize.Set(ii.ranges[ii.rangeIndex].size(ii.buf))
	ii.freeSize.Sub(ii.freeSize, bigInt[1])
	ii.current = ii.ranges[ii.rangeIndex].start

	return ii.current.addr
}

// NextN returns the next nth IP address. If the IPIterator has been exhausted,
// return the zero Addr. If n <= 0, it is equivalent to NextN(1).
func (ii *IPIterator) NextN(n *big.Int) netip.Addr {
	l := len(ii.ranges)
	if l == 0 {
		return netip.Addr{}
	}

	if n == nil || n.Sign() <= 0 {
		n = big.NewInt(1)
	}

	if !ii.current.addr.IsValid() {
		n = new(big.Int).Set(n)
		for {
			size := ii.ranges[ii.rangeIndex].size(ii.buf)
			if n.Cmp(size) <= 0 {
				ii.freeSize = new(big.Int)
				ii.freeSize.Set(size)
				ii.freeSize.Sub(ii.freeSize, n)
				n.Sub(n, bigInt[1])
				ii.current = ii.ranges[ii.rangeIndex].start.nextN(n, ii.buf)
				return ii.current.addr
			}

			n.Sub(n, size)
			ii.rangeIndex++
			if ii.rangeIndex == l {
				return netip.Addr{}
			}
		}
	}

	if n.Cmp(ii.freeSize) <= 0 {
		ii.freeSize.Sub(ii.freeSize, n)
		ii.current = ii.current.nextN(n, ii.buf)
		return ii.current.addr
	}

	ii.rangeIndex++
	if ii.rangeIndex == l {
		return netip.Addr{}
	}
	n = new(big.Int).Set(n)
	n.Sub(n, ii.freeSize)
	ii.freeSize.Set(ii.ranges[ii.rangeIndex].size(ii.buf))
	ii.freeSize.Sub(ii.freeSize, n)
	n.Sub(n, bigInt[1])
	ii.current = ii.ranges[ii.rangeIndex].start.nextN(n, ii.buf)

	return ii.current.addr
}

// Reset resets IP iterator.
func (ii *IPIterator) Reset() {
	ii.rangeIndex = 0
	ii.current = ip{}
}

type BlockIterator struct {
	ranges    *IPRanges
	size      *big.Int
	blockSize *big.Int
	start     *big.Int
	end       *big.Int
}

// BlockIterator generates a new iterator for scanning IP blocks. blockSize
// should be at least 1, which is somewhat equivalent to IPIterator.
func (rs *IPRanges) BlockIterator(blockSize *big.Int) *BlockIterator {
	if blockSize == nil || blockSize.Sign() <= 0 {
		blockSize = big.NewInt(1)
	}

	return &BlockIterator{
		ranges:    rs,
		size:      rs.Size(),
		blockSize: blockSize,
	}
}

// Next returns the next IP block. If the BlockIterator has been exhausted,
// return nil.
func (bi *BlockIterator) Next() *IPRanges {
	if bi.size.Sign() == 0 {
		return nil
	}

	if bi.start == nil {
		bi.start = big.NewInt(0)
		bi.end = new(big.Int).Sub(bi.blockSize, bigInt[1])
		return bi.ranges.Slice(bi.start, bi.end)
	}

	bi.start.Add(bi.start, bi.blockSize)
	if bi.start.Cmp(bi.size) >= 0 {
		return nil
	}
	bi.end.Add(bi.end, bi.blockSize)

	return bi.ranges.Slice(bi.start, bi.end)
}

// NextN returns the next nth IP block. If the BlockIterator has been
// exhausted, return nil. If n <= 0, it is equivalent to NextN(1).
func (bi *BlockIterator) NextN(n *big.Int) *IPRanges {
	if bi.size.Sign() == 0 {
		return nil
	}

	if n == nil || n.Sign() <= 0 {
		n = big.NewInt(1)
	}

	if bi.start == nil {
		n = new(big.Int).Sub(n, bigInt[1])
		bi.start = new(big.Int).Mul(n, bi.blockSize)
		bi.end = new(big.Int).Add(bi.start, bi.blockSize)
		bi.end.Sub(bi.end, bigInt[1])
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
func (bi *BlockIterator) Reset() {
	bi.start = nil
	bi.end = nil
}
