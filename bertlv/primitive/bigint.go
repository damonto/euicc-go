package primitive

import (
	"encoding"
	"math/big"
)

func UnmarshalBigInt(value *big.Int) encoding.BinaryUnmarshaler {
	return Unmarshaler(func(data []byte) error {
		value.SetBytes(data)
		return nil
	})
}

func MarshalBigInt(value *big.Int) encoding.BinaryMarshaler {
	return Marshaler(func() ([]byte, error) {
		return value.Bytes(), nil
	})
}
