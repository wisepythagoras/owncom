package core

import (
	"encoding/binary"
)

// bytesChecksum computes the checksum of a byte array. Based on the following implementations:
// https://go-review.googlesource.com/c/net/+/112817/2/ipv4/header.go#131
// https://tools.ietf.org/html/rfc1071
func bytesChecksum(bt []byte) uint16 {
	var sum uint32 = 0

	if len(bt)%2 == 1 {
		bt = append(bt, 0)
	}

	for i := 0; i < len(bt); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(bt[i : i+2]))
	}

	carry := uint16(sum >> 16)
	return ^(carry + uint16(sum&0x0ffff))
}
