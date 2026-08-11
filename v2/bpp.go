package sgp22

import (
	"errors"
	"fmt"
	"slices"

	"github.com/damonto/euicc-go/bertlv"
)

func SegmentedBoundProfilePackage(bpp *bertlv.TLV) ([][]byte, error) {
	if err := ValidBoundProfilePackage(bpp); err != nil {
		return nil, err
	}
	marshalHeader := func(tlv *bertlv.TLV) ([]byte, error) {
		var n int
		for _, child := range tlv.Children {
			if child == nil {
				continue
			}
			n += child.Len()
		}
		if n >= 16777216 {
			return nil, fmt.Errorf("TLV length %d exceeds 3-byte limit", n)
		}
		header := slices.Clone(tlv.Tag)
		switch {
		case n < 128:
			header = append(header, byte(n))
		case n < 256:
			header = append(header, 0x81, byte(n))
		case n < 65536:
			header = append(header, 0x82, byte(n>>8), byte(n))
		case n < 16777216:
			header = append(header, 0x83, byte(n>>16), byte(n>>8), byte(n))
		}
		return header, nil
	}
	appendSegmentedSequence := func(segments [][]byte, sequence *bertlv.TLV) ([][]byte, error) {
		header, err := marshalHeader(sequence)
		if err != nil {
			return nil, err
		}
		headerWritten := false
		for _, child := range sequence.Children {
			if child == nil {
				continue
			}
			encoded, err := child.Bytes()
			if err != nil {
				return nil, fmt.Errorf("marshal segmented TLV: %w", err)
			}
			if !headerWritten {
				segments = append(segments, slices.Concat(header, encoded))
				headerWritten = true
				continue
			}
			segments = append(segments, encoded)
		}
		if !headerWritten {
			segments = append(segments, header)
		}
		return segments, nil
	}
	var (
		initialiseSecureChannelRequest = bpp.First(bertlv.Constructed.ContextSpecific(35))
		firstSequenceOf87              = bpp.First(bertlv.Constructed.ContextSpecific(0))
		sequenceOf88                   = bpp.First(bertlv.Constructed.ContextSpecific(1))
		secondSequenceOf87             = bpp.First(bertlv.Constructed.ContextSpecific(2))
		sequenceOf86                   = bpp.First(bertlv.Constructed.ContextSpecific(3))
	)
	var segments [][]byte
	bppHeader, err := marshalHeader(bpp)
	if err != nil {
		return nil, err
	}
	initialiseSecureChannel, err := initialiseSecureChannelRequest.Bytes()
	if err != nil {
		return nil, fmt.Errorf("marshal initialise secure channel request: %w", err)
	}
	// Tag and length fields of the BoundProfilePackage TLV plus the initialiseSecureChannelRequest TLV
	segments = append(segments, slices.Concat(
		// Tag and length fields of the BoundProfilePackage TLV
		bppHeader,
		// initialiseSecureChannelRequest TLV
		initialiseSecureChannel,
	))
	// Tag and length fields of the firstSequenceOf87 TLV plus the first '87' TLV,
	// followed by the remaining '87' TLVs.
	segments, err = appendSegmentedSequence(segments, firstSequenceOf87)
	if err != nil {
		return nil, err
	}
	// Tag and length fields of the sequenceOf88 TLV
	header, err := marshalHeader(sequenceOf88)
	if err != nil {
		return nil, err
	}
	segments = append(segments, header)
	// Each of the '88' TLVs
	for _, child := range sequenceOf88.Children {
		if child == nil {
			continue
		}
		encoded, err := child.Bytes()
		if err != nil {
			return nil, fmt.Errorf("marshal sequenceOf88 child: %w", err)
		}
		segments = append(segments, encoded)
	}
	// Tag and length fields of the secondSequenceOf87 TLV plus the first '87' TLV,
	// followed by the remaining '87' TLVs.
	if secondSequenceOf87 != nil {
		segments, err = appendSegmentedSequence(segments, secondSequenceOf87)
		if err != nil {
			return nil, err
		}
	}
	// Tag and length fields of the sequenceOf86 TLV
	header, err = marshalHeader(sequenceOf86)
	if err != nil {
		return nil, err
	}
	segments = append(segments, header)
	// Each of the '86' TLVs
	for _, child := range sequenceOf86.Children {
		if child == nil {
			continue
		}
		encoded, err := child.Bytes()
		if err != nil {
			return nil, fmt.Errorf("marshal sequenceOf86 child: %w", err)
		}
		segments = append(segments, encoded)
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
	}
	if !bpp.Tag.Equal(bertlv.Constructed.ContextSpecific(54)) {
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
