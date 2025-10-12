package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type TLV struct {
	Type  uint8
	Len   uint16
	Value []byte
}

func (t *TLV) Error() error {
	if len(t.Value) < 4 {
		return fmt.Errorf("result TLV too short, expected 4 bytes, got %d", len(t.Value))
	}
	if binary.LittleEndian.Uint16(t.Value[0:2]) == uint16(QMIResultSuccess) {
		return nil
	}
	return QMIError(binary.LittleEndian.Uint16(t.Value[2:4]))
}

type TLVs []TLV

func (ts *TLVs) ReadFrom(r io.Reader) (int64, error) {
	var len uint16
	for {
		var t uint8
		if err := binary.Read(r, binary.LittleEndian, &t); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return int64(len), fmt.Errorf("read TLV type: %w", err)
		}

		var n uint16
		binary.Read(r, binary.LittleEndian, &n)
		v := make([]byte, n)
		if _, err := io.ReadFull(r, v); err != nil {
			return int64(len), err
		}
		*ts = append(*ts, TLV{Type: t, Len: n, Value: v})
		len += n
	}
	return int64(len), ts.Error()
}

func (ts TLVs) WriteTo(w io.Writer) (int64, error) {
	var len int64
	for _, tlv := range ts {
		binary.Write(w, binary.LittleEndian, tlv.Type)
		binary.Write(w, binary.LittleEndian, tlv.Len)
		w.Write(tlv.Value)
		len += int64(tlv.Len)
	}
	return len, nil
}

func (ts TLVs) Find(t uint8) (TLV, bool) {
	for _, tlv := range ts {
		if tlv.Type == t {
			return tlv, true
		}
	}
	return TLV{}, false
}

func (ts TLVs) Error() error {
	tlv, ok := ts.Find(0x02)
	if !ok {
		return errors.New("no result TLV found")
	}
	return tlv.Error()
}
