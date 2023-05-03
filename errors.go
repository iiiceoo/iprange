package iprange

import "errors"

var (
	errInvalidIPRangeFormat = errors.New("invalid IP range format")
	errDualStackIPRanges    = errors.New("dual-stack IP ranges")
)

func IsInvalidIPRangeFormat(err error) bool {
	return errors.Is(err, errInvalidIPRangeFormat)
}

func IsDualStackIPRanges(err error) bool {
	return errors.Is(err, errDualStackIPRanges)
}
