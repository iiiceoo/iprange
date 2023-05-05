package iprange

import (
	"math/big"
	"net"
)

// xIP wraps net.IP in order to expand the method of net.IP.
type xIP struct {
	net.IP
}

// next returns the next IP address of xIP.
func (ip xIP) next() xIP {
	i := ipToInt(ip.IP)
	i.Add(i, big.NewInt(1))

	return xIP{intToIP(i)}
}

// prev returns the previous IP address of xIP.
func (ip xIP) prev() xIP {
	i := ipToInt(ip.IP)
	i.Sub(i, big.NewInt(1))

	return xIP{intToIP(i)}
}

// cmp compares xIP ip1 and ip2 and returns:
//
//	-1: ip1 <  ip2
//	 0: ip1 == ip2
//	+1: ip1 >  ip2
//	-2: ip1 and ip2 are not comparable
//
// Incomparable means that ip1 is IPv4 and ip2 is IPv6, and vice versa.
func (ip1 xIP) cmp(ip2 xIP) int {
	nIP1 := normalizeIP(ip1.IP)
	nIP2 := normalizeIP(ip2.IP)
	if len(nIP1) != 0 && len(nIP1) == len(nIP2) {
		return ipToInt(nIP1).Cmp(ipToInt(nIP2))
	}

	return -2
}

// ipToInt converts net.IP to a big number.
func ipToInt(ip net.IP) *big.Int {
	return big.NewInt(0).SetBytes(ip)
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
