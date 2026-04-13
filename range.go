package iprange

import (
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"strings"
)

var bigInt = [...]*big.Int{
	big.NewInt(0),
	big.NewInt(1),
}

// The core abstraction of the IP range concept, which uses the starting and
// ending IP addresses to represent any IP range of any size.
type ipRange struct {
	start ip
	end   ip
}

// parse parses the IP range format string as ipRange that records the starting
// and ending IP addresses. The error errInvalidIPRangeFormat will be returned
// when format is invalid.
func parse(format string) (ipRange, error) {
	fmtErr := fmt.Errorf("%w %s", errInvalidIPRangeFormat, format)
	if format == "" {
		return ipRange{}, fmtErr
	}

	// 172.18.0.0/24 fd00::/64
	if strings.Contains(format, "/") {
		prefix, err := netip.ParsePrefix(format)
		if err != nil {
			return ipRange{}, errors.Join(fmtErr, err)
		}

		bytes := prefix.Masked().Addr().AsSlice()
		mask := net.CIDRMask(prefix.Bits(), len(bytes)*8)
		for i := 0; i < len(bytes); i++ {
			bytes[i] |= ^mask[i]
		}
		lastAddr, _ := netip.AddrFromSlice(bytes)

		return ipRange{
			start: ip{addr: prefix.Addr().Unmap()},
			end:   ip{addr: lastAddr.Unmap()},
		}, nil
	}

	before, after, found := strings.Cut(format, "-")
	if !found {
		// 172.18.0.1 fd00::1
		addr, err := netip.ParseAddr(format)
		if err != nil {
			return ipRange{}, errors.Join(fmtErr, err)
		}

		addr = addr.Unmap()
		return ipRange{
			start: ip{addr: addr},
			end:   ip{addr: addr},
		}, nil
	}

	start, err := netip.ParseAddr(before)
	if err != nil {
		return ipRange{}, errors.Join(fmtErr, err)
	}

	_, err = netip.ParseAddr(after)
	if err != nil {
		// 172.18.0.1-10 fd00::1-a
		index := strings.LastIndex(before, ".")
		if index == -1 {
			index = strings.LastIndex(before, ":")
		}
		after = before[:index+1] + after
	}

	// 172.18.0.1-172.18.1.10 fd00::1-fd00::1:a
	end, err := netip.ParseAddr(after)
	if err != nil {
		return ipRange{}, errors.Join(fmtErr, err)
	}

	start = start.Unmap()
	end = end.Unmap()
	if start.Is4() != end.Is4() {
		return ipRange{}, fmtErr
	}
	if end.Compare(start) < 0 {
		return ipRange{}, fmtErr
	}

	return ipRange{
		start: ip{addr: start},
		end:   ip{addr: end},
	}, nil
}

// Version returns the IP version of ipRange.
func (r ipRange) version() family {
	return r.start.version()
}

// contains reports whether ipRange r contains netip.Addr addr.
func (r ipRange) contains(addr netip.Addr) bool {
	switch r.start.addr.Compare(addr) {
	case 0:
		return true
	case 1:
		return false
	default:
		return r.end.addr.Compare(addr) >= 0
	}
}

// equal reports whether ipRange r is equal to other.
func (r ipRange) equal(other ipRange) bool {
	return r.start.cmp(other.start) == 0 && r.end.cmp(other.end) == 0
}

// size calculates the total number of IP addresses that pertain to ipRange r.
// buf is an optional big.Int used to avoid memory allocations.
func (r *ipRange) size(buf *big.Int) *big.Int {
	if buf == nil {
		buf = new(big.Int)
	}

	var start big.Int
	start.SetBytes(r.start.addr.AsSlice())
	buf.SetBytes(r.end.addr.AsSlice())

	buf.Sub(buf, &start)
	buf.Add(buf, bigInt[1])

	return buf
}

// String implements fmt.Stringer.
func (r ipRange) String() string {
	if r.start.addr == r.end.addr {
		return r.start.addr.String()
	}

	inc := r.size(nil)
	dv := new(big.Int).Sub(inc, bigInt[1])
	if inc.And(inc, dv).Sign() == 0 {
		addr := r.start.addr
		prefix, _ := addr.Prefix(addr.BitLen() - dv.BitLen())
		if prefix.Masked().Addr().Compare(addr) == 0 {
			return prefix.String()
		}
	}

	return r.start.addr.String() + "-" + r.end.addr.String()
}
