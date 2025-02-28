package sgp22

import (
	"bytes"
	"encoding/hex"
)

type HexString []byte

func (h *HexString) MarshalText() (text []byte, err error) {
	text = make([]byte, hex.EncodedLen(len(*h)))
	hex.Encode(text, *h)
	text = bytes.ToUpper(text)
	return
}

func (h *HexString) UnmarshalText(text []byte) (err error) {
	dst := make([]byte, hex.DecodedLen(len(text)))
	n, err := hex.Decode(dst, text)
	*h = dst[:n]
	return
}
