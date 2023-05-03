package iprange

import (
	"fmt"
	"net"
	"strings"
)

type IPRange struct {
	start xIP
	end   xIP
}

func Parse(r string) (IPRange, error) {
	if len(r) == 0 {
		return IPRange{}, fmt.Errorf(`%w: ""`, errInvalidIPRangeFormat)
	}

	fmtErr := fmt.Errorf("%w: %s", errInvalidIPRangeFormat, r)
	if strings.Contains(r, "/") {
		ip, ipNet, err := net.ParseCIDR(r)
		if err != nil {
			return IPRange{}, fmtErr
		}

		n := len(ipNet.IP)
		lastIP := make(net.IP, 0, n)
		for i := 0; i < n; i++ {
			lastIP = append(lastIP, ipNet.IP[i]|^ipNet.Mask[i])
		}

		return IPRange{
			start: xIP{normalizeIP(ip)},
			end:   xIP{normalizeIP(lastIP)},
		}, nil
	}

	before, after, found := strings.Cut(r, "-")
	if found {
		startIP := net.ParseIP(before)
		if startIP == nil {
			return IPRange{}, fmtErr
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
				return IPRange{}, fmtErr
			}

			return IPRange{
				start: xIP{normalizeIP(startIP)},
				end:   xIP{normalizeIP(endIP)},
			}, nil
		}

		ns := normalizeIP(startIP)
		ne := normalizeIP(endIP)
		if len(ns) != len(ne) {
			return IPRange{}, fmtErr
		}

		return IPRange{
			start: xIP{ns},
			end:   xIP{ne},
		}, nil
	}

	ip := net.ParseIP(r)
	if ip == nil {
		return IPRange{}, fmtErr
	}
	nIP := normalizeIP(ip)

	return IPRange{
		start: xIP{nIP},
		end:   xIP{nIP},
	}, nil
}

func (r IPRange) Contains(ip net.IP) bool {
	return r.contains(xIP{ip})
}

func (r IPRange) contains(ip xIP) bool {
	switch r.start.cmp(ip) {
	case 0:
		return true
	case 1:
		return false
	case -2:
		return false
	default:
		return r.end.cmp(ip) >= 0
	}
}

func (r1 IPRange) Equal(r2 IPRange) bool {
	return r1.start.Equal(r2.start.IP) && r1.end.Equal(r2.end.IP)
}
