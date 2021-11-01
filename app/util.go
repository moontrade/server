package app

import "encoding/binary"

func appendUvarint(dst []byte, x uint64) []byte {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], x)
	return append(dst, buf[:n]...)
}
