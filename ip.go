package iprange

import (
	"math/big"
	"net"
)

type xIP struct {
	net.IP
}

func (ip xIP) next() xIP {
	i := ip.toInt()
	i.Add(i, big.NewInt(1))

	return xIP{IP: net.IP(i.Bytes())}
}

func (ip xIP) prev() xIP {
	i := ip.toInt()
	i.Sub(i, big.NewInt(1))

	return xIP{net.IP(i.Bytes())}
}

func (ip1 xIP) cmp(ip2 xIP) int {
	nIP1 := normalizeIP(ip1.IP)
	nIP2 := normalizeIP(ip2.IP)
	if len(nIP1) != 0 && len(nIP1) == len(nIP2) {
		return ip1.toInt().Cmp(ip2.toInt())
	}

	return -2
}

func (ip xIP) toInt() *big.Int {
	return big.NewInt(0).SetBytes(ip.IP)
}

func normalizeIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}

	return ip.To16()
}
