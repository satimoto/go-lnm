package util

import "encoding/binary"

func ReverseBytes(bytes []byte) []byte {
	bytesLen := len(bytes)
	reversed := make([]byte, bytesLen)

	for i := 0; i < bytesLen; i++ {
		reversed[i] = bytes[bytesLen-1-i]
	}

	return reversed
}

func BytesToUint64(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

func Uint64ToBytes(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)

	return bytes
}
