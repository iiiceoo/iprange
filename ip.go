package iprange

import (
	"math/big"
	"net"
)

var bigInt = [...]*big.Int{
	big.NewInt(0),
	big.NewInt(1),
}

// xIP wraps net.IP in order to expand the method of net.IP.
type xIP struct {
	net.IP
}

// version returns the IP version of xIP:
//
//	1: IPv4
//	2: IPv6
//	0: not an IP: Unknown
func (ip xIP) version() family {
	nIP := normalizeIP(ip.IP)
	if nIP == nil {
		return Unknown
	}

	if len(nIP) == net.IPv4len {
		return IPv4
	}

	return IPv6
}

// next returns the next IP address of xIP.
func (ip xIP) next() xIP {
	i := ipToInt(ip.IP)
	i.Add(i, bigInt[1])

	return xIP{intToIP(i)}
}

// nextN returns the next nth IP address of xIP.
func (ip xIP) nextN(n *big.Int) xIP {
	if n.Sign() == 0 {
		return ip
	}

	i := ipToInt(ip.IP)
	i.Add(i, n)

	return xIP{intToIP(i)}
}

// prev returns the previous IP address of xIP.
func (ip xIP) prev() xIP {
	i := ipToInt(ip.IP)
	i.Sub(i, bigInt[1])

	return xIP{intToIP(i)}
}

// cmp compares xIP ip and ip2 with the same IP version and returns:
//
//	-1: ip <  ip2
//	 0: ip == ip2
//	+1: ip >  ip2
func (ip xIP) cmp(ip2 xIP) int {
	nIP1 := normalizeIP(ip.IP)
	nIP2 := normalizeIP(ip2.IP)

	return ipToInt(nIP1).Cmp(ipToInt(nIP2))
}

// ipToInt converts net.IP to a big number.
func ipToInt(ip net.IP) *big.Int {
	return new(big.Int).SetBytes(ip)
}

// ipToInt converts big number to a net.IP.
func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

// normalizeIP normalizes net.IP by family:
//
//	IPv4:      4-byte form
//	IPv6:      16-byte form
//	not an IP: nil
func normalizeIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}

	return ip.To16()
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

// max returns the larger of x and y.
func maxN(x, y int) int {
	if x > y {
		return x
	}

	return y
}

// min returns the smaller of x and y.
func minN(x, y int) int {
	if x < y {
		return x
	}

	return y
}
