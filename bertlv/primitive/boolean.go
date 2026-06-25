package primitive

import (
	"encoding"
	"errors"
)

func UnmarshalBool(value *bool) encoding.BinaryUnmarshaler {
	return Unmarshaler(func(data []byte) error {
		if len(data) != 1 {
			return errors.New("invalid boolean length")
		}
		*value = data[0] != 0x00
		return nil
	})
}

func MarshalBool(value bool) encoding.BinaryMarshaler {
	return Marshaler(func() ([]byte, error) {
		if value {
			return []byte{0xff}, nil
		}
		return []byte{0x00}, nil
	})
}
