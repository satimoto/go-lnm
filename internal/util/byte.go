package util

func ReverseBytes(bytes []byte) []byte {
	bytesLen := len(bytes)
	reversed := make([]byte, bytesLen)

	for i := 0; i < bytesLen; i++ {
		reversed[i] = bytes[bytesLen-1-i]
	}

	return reversed
}
