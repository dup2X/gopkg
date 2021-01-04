// Package hash ...
package hash

// LuaHash32 ...
func LuaHash32(data []byte) uint32 {
	l := len(data)
	h := l
	step := (l >> 5) + 1
	for i := l; i >= step; i -= step {
		h = h ^ ((h << 5) + (h >> 2) + int(data[i-1]))
	}
	if h == 0 {
		return uint32(1)
	}
	return uint32(h)
}
