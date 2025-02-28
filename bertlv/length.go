package bertlv

import (
	"errors"
	"fmt"
	"io"
)

func marshalLength(n uint16) []byte {
	switch {
	case n < 128:
		return []byte{byte(n)}
	case n < 256:
		return []byte{0x81, byte(n)}
	}
	return []byte{0x82, byte(n >> 8), byte(n)}
}

func readLength(r io.Reader) (value uint16, err error) {
	var n int
	length := make([]byte, 1)
	switch n, err = io.ReadAtLeast(r, length, 1); length[0] {
	case 0x81:
		n, err = io.ReadAtLeast(r, length, 1)
		value = uint16(length[0])
	case 0x82:
		length = make([]byte, 2)
		n, err = io.ReadAtLeast(r, length, 2)
		value = uint16(length[1]) | uint16(length[0])<<8
	default:
		if length[0] >= 0x80 {
			err = errors.New("if length is greater than 127, first byte must indicate encoding of length")
		}
		value = uint16(length[0])
	}
	if len(length) != n {
		err = fmt.Errorf("expected %d bytes, got %d", len(length), n)
	}
	if err != nil {
		err = fmt.Errorf("failed to read length: %w", err)
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
