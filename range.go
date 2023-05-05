package iprange

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

type ipRange struct {
	start xIP
	end   xIP
}

func parse(r string) (ipRange, error) {
	if len(r) == 0 {
		return ipRange{}, fmt.Errorf(`%w: ""`, errInvalidIPRangeFormat)
	}

	fmtErr := fmt.Errorf("%w: %s", errInvalidIPRangeFormat, r)
	if strings.Contains(r, "/") {
		ip, ipNet, err := net.ParseCIDR(r)
		if err != nil {
			return ipRange{}, fmtErr
		}

		n := len(ipNet.IP)
		lastIP := make(net.IP, 0, n)
		for i := 0; i < n; i++ {
			lastIP = append(lastIP, ipNet.IP[i]|^ipNet.Mask[i])
		}

		return ipRange{
			start: xIP{normalizeIP(ip)},
			end:   xIP{normalizeIP(lastIP)},
		}, nil
	}

	before, after, found := strings.Cut(r, "-")
	if found {
		startIP := net.ParseIP(before)
		if startIP == nil {
			return ipRange{}, fmtErr
		}

		endIP := net.ParseIP(after)
		if endIP == nil {
			index := strings.LastIndex(before, ".")
			if index == -1 {
				index = strings.LastIndex(before, ":")
			}
			after = before[:index+1] + after
			endIP = net.ParseIP(after)
			if endIP == nil {
				return ipRange{}, fmtErr
			}

			start := xIP{normalizeIP(startIP)}
			end := xIP{normalizeIP(endIP)}
			if end.cmp(start) < 0 {
				return ipRange{}, fmtErr
			}

			return ipRange{
				start: start,
				end:   end,
			}, nil
		}

		start := xIP{normalizeIP(startIP)}
		end := xIP{normalizeIP(endIP)}
		if end.cmp(start) < 0 {
			return ipRange{}, fmtErr
		}

		return ipRange{
			start: start,
			end:   end,
		}, nil
	}

	ip := net.ParseIP(r)
	if ip == nil {
		return ipRange{}, fmtErr
	}
	nIP := normalizeIP(ip)

	return ipRange{
		start: xIP{nIP},
		end:   xIP{nIP},
	}, nil
}

func (r ipRange) contains(ip net.IP) bool {
	w := xIP{ip}
	switch r.start.cmp(w) {
	case 0:
		return true
	case 1:
		return false
	case -2:
		return false
	default:
		return r.end.cmp(w) >= 0
	}
}

func (r1 ipRange) equal(r2 ipRange) bool {
	return r1.start.Equal(r2.start.IP) && r1.end.Equal(r2.end.IP)
}

func (r ipRange) size() *big.Int {
	n := big.NewInt(1)
	n.Add(n, ipToInt(r.end.IP))

	return n.Sub(n, ipToInt(r.start.IP))
}

func (r ipRange) String() string {
	return r.start.String() + "-" + r.end.String()
}
