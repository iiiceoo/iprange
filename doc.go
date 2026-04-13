/*
Package iprange parses IPv4/IPv6 addresses from strings in IP range format.

The following IP range formats are supported:

	172.18.0.1              fd00::1
	172.18.0.0/24           fd00::/64
	172.18.0.1-10           fd00::1-a
	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a

It takes a set of IP range strings, and returns a list of start-end IP address
pairs, which can then be automatically extended and normalized, for instance:

	v4Ranges, err := iprange.Parse("172.18.0.1", "172.18.0.0/24")  // valid
	v6Ranges, err := iprange.Parse("fd00::1", "fd00::/64")         // valid
	invalid, err := iprange.Parse("Invalid IP range string")       // invalid
	dual, err := iprange.Parse("172.18.0.1", "fd00::/64")          // invalid

When parsing an invalid IP range string, error errInvalidIPRangeFormat will be
returned, and dual-stack IP ranges are not allowed. Use the following functions
to assert the errors:

	func IsInvalidIPRangeFormat(err error) bool
	func IsDualStackIPRanges(err error) bool

Use the interval methods of IPRanges to calculate the union, difference or
intersection of two IPRanges. They do not change the original parameters (rs and
other), just calculate, and return the results.

	func (rs *IPRanges) Union(other *IPRanges) *IPRanges
	func (rs *IPRanges) Diff(other *IPRanges) *IPRanges
	func (rs *IPRanges) Intersect(other *IPRanges) *IPRanges

The IPRanges can be converted into multiple netip.Addr through iterator.
Continuously call the method Next() until an zero value is returned:

	ipIter := ranges.IPIterator()
	for {
	    addr := ipIter.Next()
	    if !addr.IsValid() {
	        break
	    }
	    // TODO
	}

Finally, the inspiration for writing this package comes from

	CNI plugins:      https://github.com/containernetworking/plugins
	malfunkt/iprange: https://github.com/malfunkt/iprange
	netaddr/netaddr:  https://github.com/netaddr/netaddr

both of which are great!
*/
package iprange
