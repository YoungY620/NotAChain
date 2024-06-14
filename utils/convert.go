package utils

import (
	"encoding/binary"
)

func UintToBytes(num uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, num)
	return buf
}

func BytesToInt(buf []byte) uint64 {
	localBuf := make([]byte, 8)
	copy(localBuf, buf)
	return binary.BigEndian.Uint64(localBuf)
}

func LongBytesToInt(buf []byte) []uint64 {
	ints := make([]uint64, 0)
	for i := 0; i < len(buf); i += 8 {
		ints = append(ints, binary.BigEndian.Uint64(buf[i:i+8]))
	}
	return ints
}
