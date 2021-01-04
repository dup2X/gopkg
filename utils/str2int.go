// Package utils ...
package utils

import (
	"strconv"
)

// StrToUint64Must ...
func StrToUint64Must(src string) uint64 {
	res, err := strconv.ParseUint(src, 10, 64)
	if err != nil {
		return 0
	}
	return res
}
