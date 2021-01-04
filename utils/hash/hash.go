// Package hash ...
package hash

import (
	"unsafe"
)

// Bernstein ...
func Bernstein(data []byte) uint32 {
	hash := uint32(5381)
	for _, b := range data {
		hash = ((hash << 5) + hash) + uint32(b)
	}
	return hash
}

// Constants for FNV1A and derivatives
const (
	_Off32 = 2166136261
	_P32   = 16777619
	_Yp32  = 709607
)

// FNV1A ...
func FNV1A(data []byte) uint32 {
	var hash = uint32(_Off32)
	for _, c := range data {
		hash ^= uint32(c)
		hash *= _P32
	}
	return hash
}

// Constants for multiples of sizeof(WORD)
const (
	_wordSize    = 4              // 4
	_dwordSize   = _wordSize << 1 // 8
	_ddwordSize  = _wordSize << 2 // 16
	_dddwordSize = _wordSize << 3 // 32
)

// Jesteress derivative of FNV1A from [http://www.sanmayce.com/Fastest_Hash/]
func Jesteress(data []byte) uint32 {
	h32 := uint32(_Off32)
	i, dlen := 0, len(data)

	for ; dlen >= _ddwordSize; dlen -= _ddwordSize {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		k2 := *(*uint64)(unsafe.Pointer(&data[i+4]))
		h32 = uint32((uint64(h32) ^ ((k1<<5 | k1>>27) ^ k2)) * _Yp32)
		i += _ddwordSize
	}

	// Cases: 0,1,2,3,4,5,6,7
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		h32 = uint32(uint64(h32)^k1) * _Yp32
		i += _dwordSize
	}
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		h32 = (h32 ^ k1) * _Yp32
		i += _dwordSize
	}
	if (dlen & 1) > 0 {
		h32 = (h32 ^ uint32(data[i])) * _Yp32
	}
	return h32 ^ (h32 >> 16)
}

// Meiyan derivative of FNV1A from [http://www.sanmayce.com/Fastest_Hash/]
func Meiyan(data []byte) uint32 {
	h32 := uint32(_Off32)
	i, dlen := 0, len(data)

	for ; dlen >= _ddwordSize; dlen -= _ddwordSize {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		k2 := *(*uint64)(unsafe.Pointer(&data[i+4]))
		h32 = uint32((uint64(h32) ^ ((k1<<5 | k1>>27) ^ k2)) * _Yp32)
		i += _ddwordSize
	}

	// Cases: 0,1,2,3,4,5,6,7
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		h32 = uint32(uint64(h32)^k1) * _Yp32
		i += _dwordSize
		k1 = *(*uint64)(unsafe.Pointer(&data[i]))
		h32 = uint32(uint64(h32)^k1) * _Yp32
		i += _dwordSize
	}
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		h32 = (h32 ^ k1) * _Yp32
		i += _dwordSize
	}
	if (dlen & 1) > 0 {
		h32 = (h32 ^ uint32(data[i])) * _Yp32
	}
	return h32 ^ (h32 >> 16)
}

// Yorikke derivative of FNV1A from [http://www.sanmayce.com/Fastest_Hash/]
func Yorikke(data []byte) uint32 {
	h32 := uint32(_Off32)
	h32b := uint32(_Off32)
	i, dlen := 0, len(data)

	for ; dlen >= _dddwordSize; dlen -= _dddwordSize {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		k2 := *(*uint64)(unsafe.Pointer(&data[i+4]))
		h32 = uint32((uint64(h32) ^ ((k1<<5 | k1>>27) ^ k2)) * _Yp32)
		k1 = *(*uint64)(unsafe.Pointer(&data[i+8]))
		k2 = *(*uint64)(unsafe.Pointer(&data[i+12]))
		h32b = uint32((uint64(h32b) ^ ((k1<<5 | k1>>27) ^ k2)) * _Yp32)
		i += _dddwordSize
	}
	if (dlen & _ddwordSize) > 0 {
		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
		k2 := *(*uint64)(unsafe.Pointer(&data[i+4]))
		h32 = uint32((uint64(h32) ^ k1) * _Yp32)
		h32b = uint32((uint64(h32b) ^ k2) * _Yp32)
		i += _ddwordSize
	}
	// Cases: 0,1,2,3,4,5,6,7
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		k2 := *(*uint32)(unsafe.Pointer(&data[i+2]))
		h32 = (h32 ^ k1) * _Yp32
		h32b = (h32b ^ k2) * _Yp32
		i += _dwordSize
	}
	if (dlen & _dwordSize) > 0 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		h32 = (h32 ^ k1) * _Yp32
		i += _dwordSize
	}
	if (dlen & 1) > 0 {
		h32 = (h32 ^ uint32(data[i])) * _Yp32
	}
	h32 = (h32 ^ (h32b<<5 | h32b>>27)) * _Yp32
	return h32 ^ (h32 >> 16)
}

// Constants defined by the Murmur3 algorithm
const (
	_C1 = uint32(0xcc9e2d51)
	_C2 = uint32(0x1b873593)
	_F1 = uint32(0x85ebca6b)
	_F2 = uint32(0xc2b2ae35)
)

// M3Seed ...
const M3Seed = uint32(0x9747b28c)

// Murmur3 ...
func Murmur3(data []byte, seed uint32) uint32 {
	h1 := seed
	ldata := len(data)
	end := ldata - (ldata % 4)
	i := 0

	// Inner
	for ; i < end; i += 4 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		k1 *= _C1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= _C2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*5 + 0xe6546b64
	}

	// Tail
	var k1 uint32
	switch ldata - i {
	case 3:
		k1 |= uint32(data[i+2]) << 16
		fallthrough
	case 2:
		k1 |= uint32(data[i+1]) << 8
		fallthrough
	case 1:
		k1 |= uint32(data[i])
		k1 *= _C1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= _C2
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint32(ldata)
	h1 ^= (h1 >> 16)
	h1 *= _F1
	h1 ^= (h1 >> 13)
	h1 *= _F2
	h1 ^= (h1 >> 16)

	return h1
}
