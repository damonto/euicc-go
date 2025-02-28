package primitive

import "encoding"

func UnmarshalBool(value *bool) encoding.BinaryUnmarshaler {
	return Unmarshaler(func(data []byte) error {
		*value = len(data) == 1 && data[0] == 0xff
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
