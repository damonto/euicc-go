package bertlv

import (
	"errors"
	"fmt"
	"io"
)

func marshalLength(n uint32) []byte {
	switch {
	case n < 128:
		return []byte{byte(n)}
	case n < 256:
		return []byte{0x81, byte(n)}
	case n < 65536:
		return []byte{0x82, byte(n >> 8), byte(n)}
	case n < 16777216:
		return []byte{0x83, byte(n >> 16), byte(n >> 8), byte(n)}
	}
	panic(fmt.Sprintf("TLV too large: %d exceeds 3-byte length limit (3 bytes max)", n))
}

func readLength(r io.Reader) (value uint32, err error) {
	var n int
	length := make([]byte, 1)
	switch n, err = io.ReadAtLeast(r, length, 1); length[0] {
	case 0x81:
		n, err = io.ReadAtLeast(r, length, 1)
		value = uint32(length[0])
	case 0x82:
		length = make([]byte, 2)
		n, err = io.ReadAtLeast(r, length, 2)
		value = uint32(length[0])<<8 | uint32(length[1])
	case 0x83:
		length = make([]byte, 3)
		n, err = io.ReadAtLeast(r, length, 3)
		value = uint32(length[0])<<16 | uint32(length[1])<<8 | uint32(length[2])
	default:
		if length[0] >= 0x80 {
			err = errors.New("unsupported length encoding")
		}
		value = uint32(length[0])
	}
	if len(length) != n {
		err = fmt.Errorf("expected %d bytes, got %d", len(length), n)
	}
	if err != nil {
		err = fmt.Errorf("read length: %w", err)
	}
	return
}

func contentLength(tlv *TLV) int {
	if tlv.Tag.Primitive() {
		return len(tlv.Value)
	}
	var n int
	for _, child := range tlv.Children {
		if child == nil {
			continue
		}
		n += child.Len()
	}
	return n
}
