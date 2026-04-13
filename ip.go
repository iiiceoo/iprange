package iprange

import (
	"math/big"
	"net/netip"
)

// ip wraps netip.Addr in order to expand the method of it.
type ip struct {
	addr netip.Addr
}

// version returns IP version.
func (ip ip) version() family {
	if ip.addr.Is4() {
		return IPv4
	}

	return IPv6
}

// next returns the IP following i. If there is none, it returns the zero IP.
func (ip ip) next() ip {
	ip.addr = ip.addr.Next()
	return ip
}

// nextN returns the next nth IP following i. If there is none, it returns the
// zero IP. buf is an optional big.Int used to avoid memory allocations.
func (ip ip) nextN(n, buf *big.Int) ip {
	if n == nil || n.Sign() <= 0 {
		return ip
	}

	if n.IsUint64() && n.Uint64() == 1 {
		return ip.next()
	}

	if buf == nil {
		buf = new(big.Int)
	}

	buf.SetBytes(ip.addr.AsSlice())
	buf.Add(buf, n)
	ip.addr = intToAddr(ip.version(), buf)

	return ip
}

// prev returns the IP before i. If there is none, it returns the zero IP.
func (ip ip) prev() ip {
	ip.addr = ip.addr.Prev()
	return ip
}

// cmp returns an integer comparing two IPs:
//
//	-1: ip <  other
//	 0: ip == other
//	+1: ip >  other
func (ip ip) cmp(other ip) int {
	return ip.addr.Compare(other.addr)
}

// intToAddr converts big.Int to netip.Addr.
func intToAddr(version family, i *big.Int) netip.Addr {
	var b [16]byte
	if version == IPv4 {
		i.FillBytes(b[0:4])
		return netip.AddrFrom4([4]byte(b[0:4]))
	}

	i.FillBytes(b[:])
	return netip.AddrFrom16(b)
}

// maxIP returns the numerically higher IP.
func maxIP(x, y ip) ip {
	if x.cmp(y) > 0 {
		return x
	}

	return y
}

// minIP returns the numerically lower IP.
func minIP(x, y ip) ip {
	if x.cmp(y) < 0 {
		return x
	}

	return y
}
