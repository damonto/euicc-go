package sgp22

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/damonto/euicc-go/bertlv"
)

func SegmentedBoundProfilePackage(bpp *bertlv.TLV) ([][]byte, error) {
	if err := ValidBoundProfilePackage(bpp); err != nil {
		return nil, err
	}
	marshalHeader := func(tlv *bertlv.TLV) []byte {
		var n int
		for _, child := range tlv.Children {
			n += child.Len()
		}
		var buf bytes.Buffer
		buf.Write(tlv.Tag)
		switch {
		case n < 128:
			buf.WriteByte(byte(n))
		case n < 256:
			buf.Write([]byte{0x81, byte(n)})
		case n < 65536:
			buf.Write([]byte{0x82, byte(n >> 8), byte(n)})
		case n < 16777216:
			buf.Write([]byte{0x83, byte(n >> 16), byte(n >> 8), byte(n)})
		default:
			panic(fmt.Sprintf("TLV too large: %d exceeds 3-byte length limit (3 bytes max)", n))
		}
		return buf.Bytes()
	}
	var (
		initialiseSecureChannelRequest = bpp.First(bertlv.Constructed.ContextSpecific(35))
		firstSequenceOf87              = bpp.First(bertlv.Constructed.ContextSpecific(0))
		sequenceOf88                   = bpp.First(bertlv.Constructed.ContextSpecific(1))
		secondSequenceOf87             = bpp.First(bertlv.Constructed.ContextSpecific(2))
		sequenceOf86                   = bpp.First(bertlv.Constructed.ContextSpecific(3))
	)
	var segments [][]byte
	// Tag and length fields of the BoundProfilePackage TLV plus the initialiseSecureChannelRequest TLV
	segments = append(segments, slices.Concat(
		// Tag and length fields of the BoundProfilePackage TLV
		marshalHeader(bpp),
		// initialiseSecureChannelRequest TLV
		initialiseSecureChannelRequest.Bytes(),
	))
	// Tag and length fields of the first firstSequenceOf87 TLV plus the first '87' TLV
	segments = append(segments, firstSequenceOf87.Bytes())
	// Tag and length fields of the sequenceOf88 TLV
	segments = append(segments, marshalHeader(sequenceOf88))
	// Each of the '88' TLVs
	for _, child := range sequenceOf88.Children {
		segments = append(segments, child.Bytes())
	}
	// Tag and length fields of the firstSequenceOf87 TLV plus the first '87' TLV
	if secondSequenceOf87 != nil {
		segments = append(segments, secondSequenceOf87.Bytes())
	}
	// Tag and length fields of the sequenceOf86 TLV
	segments = append(segments, marshalHeader(sequenceOf86))
	// Each of the '86' TLVs
	for _, child := range sequenceOf86.Children {
		segments = append(segments, child.Bytes())
	}
	return segments, nil
}

func ValidBoundProfilePackage(bpp *bertlv.TLV) error {
	type Item struct {
		Name string
		Tag  bertlv.Tag
	}
	var fields []error
	if bpp == nil {
		return errors.New("missing boundProfilePackage")
	} else if !bpp.Tag.Equal(bertlv.Constructed.ContextSpecific(54)) {
		return errors.New("invalid boundProfilePackage tag")
	}
	items := []*Item{
		{"initialiseSecureChannelRequest", bertlv.Constructed.ContextSpecific(35)},
		{"firstSequenceOf87", bertlv.Constructed.ContextSpecific(0)},
		{"sequenceOf88", bertlv.Constructed.ContextSpecific(1)},
		{"sequenceOf86", bertlv.Constructed.ContextSpecific(3)},
	}
	for _, item := range items {
		if bpp.First(item.Tag) == nil {
			fields = append(fields, fmt.Errorf("missing %s", item.Name))
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return errors.Join(fields...)
}
