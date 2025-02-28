package apdu

import (
	"bytes"
	"encoding/hex"
	"io"
	"strings"
)

type Request struct {
	CLA  byte
	INS  byte
	P1   byte
	P2   byte
	Data []byte
	Le   *byte
}

func (r *Request) APDU() []byte {
	var apdu bytes.Buffer
	_, _ = r.WriteTo(&apdu)
	return apdu.Bytes()
}

func (r *Request) WriteTo(w io.Writer) (n int64, err error) {
	var buf bytes.Buffer
	buf.WriteByte(r.CLA)
	buf.WriteByte(r.INS)
	buf.WriteByte(r.P1)
	buf.WriteByte(r.P2)
	if len(r.Data) > 0 {
		buf.WriteByte(byte(len(r.Data)))
		buf.Write(r.Data)
	}
	if r.Le != nil {
		buf.WriteByte(*r.Le)
	}
	return buf.WriteTo(w)
}

func (r *Request) String() string {
	return strings.ToUpper(hex.EncodeToString(r.APDU()))
}
