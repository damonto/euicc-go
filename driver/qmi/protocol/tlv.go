package protocol

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
	var read int64
	for {
		var t uint8
		if err := binary.Read(r, binary.LittleEndian, &t); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return read, fmt.Errorf("read TLV type: %w", err)
		}
		read++

		var n uint16
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return read, fmt.Errorf("read TLV length: %w", err)
		}
		read += 2

		v := make([]byte, n)
		nr, err := io.ReadFull(r, v)
		read += int64(nr)
		if err != nil {
			return read, fmt.Errorf("read TLV value: %w", err)
		}
		*ts = append(*ts, TLV{Type: t, Len: n, Value: v})
	}
	return read, ts.Error()
}

func (ts TLVs) WriteTo(w io.Writer) (int64, error) {
	var written int64
	for _, tlv := range ts {
		if int(tlv.Len) != len(tlv.Value) {
			return written, fmt.Errorf("TLV type 0x%02X length mismatch: header %d, value %d", tlv.Type, tlv.Len, len(tlv.Value))
		}
		if err := binary.Write(w, binary.LittleEndian, tlv.Type); err != nil {
			return written, fmt.Errorf("write TLV type: %w", err)
		}
		written++
		if err := binary.Write(w, binary.LittleEndian, tlv.Len); err != nil {
			return written, fmt.Errorf("write TLV length: %w", err)
		}
		written += 2
		n, err := w.Write(tlv.Value)
		written += int64(n)
		if err != nil {
			return written, fmt.Errorf("write TLV value: %w", err)
		}
		if n != len(tlv.Value) {
			return written, io.ErrShortWrite
		}
	}
	return written, nil
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
