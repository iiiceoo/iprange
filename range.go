package iprange

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

// The core abstraction of the IP range concept, which uses the starting
// and ending IP addresses to represent any IP range of any size.
type ipRange struct {
	start xIP
	end   xIP
}

// parse parses the IP range format string as ipRange that records the
// starting and ending IP addresses. The error errInvalidIPRangeFormat
// wiil be returned when r is invalid.
func parse(r string) (*ipRange, error) {
	if r == "" {
		return nil, fmt.Errorf(`%w: ""`, errInvalidIPRangeFormat)
	}

	fmtErr := fmt.Errorf("%w: %s", errInvalidIPRangeFormat, r)
	// 172.18.0.0/24
	// fd00::/64
	if strings.Contains(r, "/") {
		ip, ipNet, err := net.ParseCIDR(r)
		if err != nil {
			return nil, fmtErr
		}

		n := len(ipNet.IP)
		lastIP := make(net.IP, 0, n)
		for i := 0; i < n; i++ {
			lastIP = append(lastIP, ipNet.IP[i]|^ipNet.Mask[i])
		}

		return &ipRange{
			start: xIP{normalizeIP(ip)},
			end:   xIP{normalizeIP(lastIP)},
		}, nil
	}

	before, after, found := strings.Cut(r, "-")
	if found {
		startIP := net.ParseIP(before)
		if startIP == nil {
			return nil, fmtErr
		}

		endIP := net.ParseIP(after)
		if endIP == nil {
			// 172.18.0.1-10
			// fd00::1-a
			index := strings.LastIndex(before, ".")
			if index == -1 {
				index = strings.LastIndex(before, ":")
			}
			after = before[:index+1] + after
			endIP = net.ParseIP(after)
			if endIP == nil {
				return nil, fmtErr
			}

			start := xIP{normalizeIP(startIP)}
			end := xIP{normalizeIP(endIP)}
			if end.cmp(start) < 0 {
				return nil, fmtErr
			}

			return &ipRange{
				start: start,
				end:   end,
			}, nil
		}

		// 172.18.0.1-172.18.1.10
		// fd00::1-fd00::1:a
		start := xIP{normalizeIP(startIP)}
		end := xIP{normalizeIP(endIP)}
		if end.cmp(start) < 0 {
			return nil, fmtErr
		}

		return &ipRange{
			start: start,
			end:   end,
		}, nil
	}

	// 172.18.0.1
	// fd00::1
	ip := net.ParseIP(r)
	if ip == nil {
		return nil, fmtErr
	}
	nIP := normalizeIP(ip)

	return &ipRange{
		start: xIP{nIP},
		end:   xIP{nIP},
	}, nil
}

// contains reports whether ipRange r contains net.IP ip.
func (r *ipRange) contains(ip net.IP) bool {
	w := xIP{ip}
	switch r.start.cmp(w) {
	case 0:
		return true
	case 1:
		return false
	default:
		return r.end.cmp(w) >= 0
	}
}

// equal reports whether ipRange r is equal to r2.
func (r *ipRange) equal(r2 *ipRange) bool {
	return r.start.Equal(r2.start.IP) && r.end.Equal(r2.end.IP)
}

// size calculates the total number of IP addresses that pertain to ipRange r.
func (r *ipRange) size() *big.Int {
	n := big.NewInt(1)
	n.Add(n, ipToInt(r.end.IP))

	return n.Sub(n, ipToInt(r.start.IP))
}

// String implements fmt.Stringer.
func (r *ipRange) String() string {
	inc := r.size()
	dv := new(big.Int).Sub(inc, bigInt[1])
	bl := dv.BitLen()
	if bl == 0 {
		return r.start.String()
	}

	if inc.And(inc, dv).Sign() == 0 {
		bits := 32
		if r.start.version() == IPv6 {
			bits = 128
		}

		ip := r.start.IP
		mask := net.CIDRMask(bits-bl, bits)
		if ip.Mask(mask).Equal(ip) {
			ipNet := net.IPNet{
				IP:   ip,
				Mask: mask,
			}
			return ipNet.String()
		}
	}

	return r.start.String() + "-" + r.end.String()
}
