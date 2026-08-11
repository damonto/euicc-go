package bertlv

import (
	"bytes"
	"testing"
)

func TestLength(t *testing.T) {
	fixtures := map[uint32][]byte{
		0x00:    {0x00},
		0x01:    {0x01},
		0x7f:    {0x7f},
		0x80:    {0x81, 0x80},
		0xff:    {0x81, 0xff},
		0x100:   {0x82, 0x01, 0x00},
		0xffff:  {0x82, 0xff, 0xff},
		0x10000: {0x83, 0x01, 0x00, 0x00},
	}
	for length, expected := range fixtures {
		got, err := marshalLength(length)
		if err != nil {
			t.Errorf("marshalLength(%d) error = %v", length, err)
			continue
		}
		if !bytes.Equal(got, expected) {
			t.Errorf("marshalLength(%d) = % X, want % X", length, got, expected)
		}
		value, err := readLength(bytes.NewReader(expected))
		if err != nil {
			t.Errorf("readLength(% X) error = %v", expected, err)
		}
		if value != length {
			t.Errorf("readLength(% X) = %d, want %d", expected, value, length)
		}
	}
	if _, err := marshalLength(0x1000000); err == nil {
		t.Error("marshalLength(0x1000000) error = nil")
	}
}

func TestLength_Error(t *testing.T) {
	type Fixture struct {
		Length []byte
		Error  string
	}
	fixtures := []*Fixture{
		{[]byte{}, "read length: expected 1 bytes, got 0"},
		{[]byte{0x80}, "read length: unsupported length encoding"},
		{[]byte{0x81}, "read length: expected 1 bytes, got 0"},
		{[]byte{0x82}, "read length: expected 2 bytes, got 0"},
	}
	var err error
	for _, fixture := range fixtures {
		_, err = readLength(bytes.NewReader(fixture.Length))
		if err == nil || err.Error() != fixture.Error {
			t.Errorf("readLength(% X) error = %v, want %q", fixture.Length, err, fixture.Error)
		}
	}
}
