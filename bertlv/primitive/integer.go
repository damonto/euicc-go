package primitive

import (
	"encoding"
	"errors"
	"fmt"
	"unsafe"
)

type signedInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func UnmarshalInt[Int signedInt](value *Int) encoding.BinaryUnmarshaler {
	size := int(unsafe.Sizeof(*value))
	return Unmarshaler(func(data []byte) error {
		if len(data) == 0 {
			return errors.New("invalid integer length")
		}
		if data[0] == 0x00 || data[0] == 0xff {
			sign := data[0]
			start := 0
			for start < len(data)-1 &&
				data[start] == sign &&
				data[start+1]>>7 == sign>>7 {
				start++
			}
			data = data[start:]
		}
		if len(data) > size {
			return fmt.Errorf("the value is too large, expected at most %d bytes, got %d", size, len(data))
		}
		var n Int
		var index int
		if data[0] > 0x7f {
			n = -1
		}
		for index = range data {
			n = Int(int64(n) << 8)
			n ^= Int(data[index])
		}
		*value = n
		return nil
	})
}

func MarshalInt[Int signedInt](value Int) encoding.BinaryMarshaler {
	return Marshaler(func() ([]byte, error) {
		var buf [8]byte
		size := len(buf)
		for i := size - 1; i >= 0; i-- {
			buf[i] = byte(value)
			value = Int(int64(value) >> 8)
		}
		start := 0
		for start < size-1 &&
			((buf[start] == 0x00 && buf[start+1]&0x80 == 0x00) ||
				(buf[start] == 0xFF && buf[start+1]&0x80 == 0x80)) {
			start++
		}
		return buf[start:], nil
	})
}
