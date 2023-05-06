/*
Package iprange parses IPv4/IPv6 addresses from strings in IP range
format and handles interval mathematics between multiple IP ranges.

The following IP range formats are supported:

	172.18.0.1              fd00::1
	172.18.0.0/24           fd00::/64
	172.18.0.1-10           fd00::1-a
	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a

It takes a set of IP range strings, and returns a list of start-end IP
address pairs, which can then be automatically extended and normalized,
for instance:

	v4Ranges, err := iprange.Parse("172.18.0.1", "172.18.0.0/24")  // √
	v6Ranges, err := iprange.Parse("fd00::1", "fd00::/64")         // √
	invalid, err := iprange.Parse("Invalid IP range string")       // ×
	dual, err := iprange.Parse("172.18.0.1", "fd00::/64")          // ×

When parsing an invalid IP range string, error errInvalidIPRangeFormat
will be returned, and dual-stack IP ranges are not allowed because this
approach is too complex and confusing. Use the following functions to
assert the errors:

	func IsInvalidIPRangeFormat(err error) bool
	func IsDualStackIPRanges(err error) bool

Use the interval methods of IPRanges to calculate the union, difference or
intersection of two IPRanges. They do not change the original parameters
(rr and rs), just calculate, and return the results.

	func (rr *IPRanges) Union(rs *IPRanges) *IPRanges
	func (rr *IPRanges) Diff(rs *IPRanges) *IPRanges
	func (rr *IPRanges) Intersect(rs *IPRanges) *IPRanges

However, do not attempt to perform calculations on two IPRanges with
different IP versions, it won't work:

	res := v4Ranges.Diff(v6Ranges)  // res will be equal to v4Ranges.

The IPRanges can be converted into individual net.IP through its own iterator.
Continuously call the method Next() to iterate through the IPRanges until
nil is returned:

	iter := ranges.Iterator()
	for {
		ip := iter.Next()
		if ip == nil {
			break
		}
		// Do someting.
	}

Finally, the inspiration for writing this package comes from

	CNI plugins: https://github.com/containernetworking/plugins
	malfunkt/iprange: https://github.com/malfunkt/iprange

both of which are great!
*/
package iprange
