package iprange

import "errors"

var (
	// The string is not a valid IP range format. It occurs when parsing an
	// invalid IP range string.
	errInvalidIPRangeFormat = errors.New("invalid IP range format")

	// Dual-stack IP ranges are not allowed. It occurs when parsing a set of
	// IP range strings, where there are both IPv4 and IPv6 addresses.
	errDualStackIPRanges = errors.New("dual-stack IP ranges")
)

// IsInvalidIPRangeFormat asserts whether the err is errInvalidIPRangeFormat.
func IsInvalidIPRangeFormat(err error) bool {
	return errors.Is(err, errInvalidIPRangeFormat)
}

// IsDualStackIPRanges asserts whether the err is errDualStackIPRanges.
func IsDualStackIPRanges(err error) bool {
	return errors.Is(err, errDualStackIPRanges)
}
