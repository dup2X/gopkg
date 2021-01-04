// Package utils ...
package utils

import (
	"encoding/binary"
)

var byteOrder = binary.BigEndian

// ByteReadUint64 Read bytes to uint64
func ByteReadUint64(data []byte) uint64 {
	return byteOrder.Uint64(data)
}

// ByteReadUint32 Read bytes to uint32
func ByteReadUint32(data []byte) uint32 {
	return byteOrder.Uint32(data)
}

// ByteReadUint16 Read bytes to uint16
func ByteReadUint16(data []byte) uint16 {
	return byteOrder.Uint16(data)
}

// ByteWriteUint64 Write uint64 to bytes
func ByteWriteUint64(data []byte, val uint64) {
	byteOrder.PutUint64(data, val)
}

// ByteWriteUint32 Write uint32 to bytes
func ByteWriteUint32(data []byte, val uint32) {
	byteOrder.PutUint32(data, val)
}

// ByteWriteUint16 Write uint16 to bytes
func ByteWriteUint16(data []byte, val uint16) {
	byteOrder.PutUint16(data, val)
}
