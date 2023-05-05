/*
Package iprange parses IPv4/IPv6 addresses from strings in IP range
format and handle interval mathematics between multiple IP ranges.

The following IP range formats are supported:

	172.18.0.1              fd00::1
	172.18.0.0/24           fd00::/64
	172.18.0.1-10           fd00::1-a
	172.18.0.1-172.18.1.10  fd00::1-fd00::1:a

It takes a set of IP range strings, and returns a list of start-end IP
address pairs, which can then be automatically extended and normalized,
for instance:

	ranges, err := iprange.Parse("172.18.0.1", "172.18.0.0/24")  // √
	ranges, err = iprange.Parse("fd00::1", "fd00::/64")          // √
	ranges, err = iprange.Parse("Invalid IP range string")       // ×
	ranges, err = iprange.Parse("172.18.0.1", "fd00::/64")       // ×

When parsing an invalid IP range string, error errInvalidIPRangeFormat
will be returned, and dual-stack IP ranges are not allowed because this
approach is too complex and confusing. Use the following functions to
assert the errors:

	func IsInvalidIPRangeFormat(err error) bool
	func IsDualStackIPRanges(err error) bool

Never never nerver use Go's append to combine two IPRanges (essentially,
it is a Go's slice). Instead, its interval operation method should be used:

	func (rr IPRanges) Union(rs IPRanges) IPRanges
	func (rr IPRanges) Diff(rs IPRanges) IPRanges
	func (rr IPRanges) Intersect(rs IPRanges) IPRanges

Similarly, do not attempt to calculate the intersection, union or difference
of an IPv4 IPRanges and an IPv6 IPRanges:

	res := v4Ranges.Diff(v6Ranges)  // Don't do this.

This behavior is not prohibited because the method literal does not return
an error (for convenience), but the calculated result must be meaningless
and incorrect.

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
