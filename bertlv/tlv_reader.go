package bertlv

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
)

func (tlv *TLV) ReadFrom(r io.Reader) (int64, error) {
	var n int64
	r = &countReader{Reader: r, Length: &n}
	var t TLV
	if _, err := t.Tag.ReadFrom(r); err != nil {
		return 0, err
	}
	var length uint32
	var err error
	if length, err = readLength(r); err != nil {
		return n, fmt.Errorf("tag %02X: invalid length encoding\n%w", t.Tag, err)
	}
	if t.Tag.Constructed() {
		limited := &io.LimitedReader{R: r, N: int64(length)}
		var child *TLV
		for limited.N > 0 {
			remaining := limited.N
			child = new(TLV)
			if _, err = child.ReadFrom(limited); err != nil {
				return n, fmt.Errorf("tag %02X: invalid child object\n%w", t.Tag, err)
			}
			if limited.N == remaining {
				return n, fmt.Errorf("tag %02X: invalid child object\nno bytes consumed", t.Tag)
			}
			t.Children = append(t.Children, child)
		}
	} else if length > 0 {
		t.Value = make([]byte, length)
		if _, err := io.ReadAtLeast(r, t.Value, len(t.Value)); err != nil {
			return n, fmt.Errorf("tag %02X: invalid length encoding\n%w", t.Tag, err)
		}
	}
	*tlv = t
	return n, nil
}

func (tlv *TLV) UnmarshalText(text []byte) error {
	var t TLV
	if _, err := t.ReadFrom(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(text))); err == nil {
		*tlv = t
		return nil
	}
	if _, err := t.ReadFrom(base64.NewDecoder(base64.RawStdEncoding, bytes.NewReader(text))); err != nil {
		return err
	}
	*tlv = t
	return nil
}

func (tlv *TLV) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)
	if _, err := tlv.ReadFrom(reader); err != nil {
		return err
	}
	if reader.Len() != 0 {
		return fmt.Errorf("trailing data after TLV: %d bytes", reader.Len())
	}
	return nil
}

func (tlv *TLV) UnmarshalBERTLV(cloned *TLV) error {
	*tlv = *cloned.Clone()
	return nil
}
