package sgp22

import (
	"bytes"
	"encoding/hex"
)

type HexString []byte

func (h *HexString) MarshalText() ([]byte, error) {
	text := make([]byte, hex.EncodedLen(len(*h)))
	hex.Encode(text, *h)
	text = bytes.ToUpper(text)
	return text, nil
}

func (h *HexString) UnmarshalText(text []byte) error {
	dst := make([]byte, hex.DecodedLen(len(text)))
	n, err := hex.Decode(dst, text)
	*h = dst[:n]
	return err
}
