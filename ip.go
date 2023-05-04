package iprange

import (
	"math/big"
	"net"
)

type xIP struct {
	net.IP
}

func (ip xIP) next() xIP {
	i := ipToInt(ip.IP)
	i.Add(i, big.NewInt(1))

	return xIP{intToIP(i)}
}

func (ip xIP) prev() xIP {
	i := ipToInt(ip.IP)
	i.Sub(i, big.NewInt(1))

	return xIP{intToIP(i)}
}

func (ip1 xIP) cmp(ip2 xIP) int {
	nIP1 := normalizeIP(ip1.IP)
	nIP2 := normalizeIP(ip2.IP)
	if len(nIP1) != 0 && len(nIP1) == len(nIP2) {
		return ipToInt(nIP1).Cmp(ipToInt(nIP2))
	}

	return -2
}

func ipToInt(ip net.IP) *big.Int {
	return big.NewInt(0).SetBytes(ip)
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func normalizeIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}

	return ip.To16()
}
