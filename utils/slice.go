// Package utils ...
package utils

// InSliceInt64 ...
func InSliceInt64(v int64, arr []int64) bool {
	if IndexSliceInt64(v, arr) < 0 {
		return false
	}
	return true
}

// IndexSliceInt64 ...
func IndexSliceInt64(v int64, arr []int64) int {
	for i := range arr {
		if arr[i] == v {
			return i
		}
	}
	return -1
}

// InSliceInt ...
func InSliceInt(v int, arr []int) bool {
	if IndexSliceInt(v, arr) < 0 {
		return false
	}
	return true
}

// IndexSliceInt ...
func IndexSliceInt(v int, arr []int) int {
	for i := range arr {
		if arr[i] == v {
			return i
		}
	}
	return -1
}
