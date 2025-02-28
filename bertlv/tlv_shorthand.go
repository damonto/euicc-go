package bertlv

import (
	"encoding"
	"errors"
	"fmt"
	"iter"
)

func NewValue(tag Tag, value []byte) *TLV {
	if tag.Constructed() {
		panic("tlv: constructed tag cannot have value")
	}
	return &TLV{Tag: tag, Value: value}
}

func NewChildren(tag Tag, children ...*TLV) *TLV {
	if tag.Primitive() {
		panic("tlv: primitive tag cannot have children")
	}
	return &TLV{Tag: tag, Children: children}
}

func NewChildrenIter(tag Tag, children iter.Seq[*TLV]) *TLV {
	var elements []*TLV
	for child := range children {
		elements = append(elements, child)
	}
	return NewChildren(tag, elements...)
}

func MarshalValue(tag Tag, marshaler encoding.BinaryMarshaler) (tlv *TLV, err error) {
	tlv = NewValue(tag, nil)
	err = tlv.MarshalValue(marshaler)
	return
}

func (tlv *TLV) String() string {
	if tlv.Tag.Primitive() {
		return fmt.Sprintf("%s (%d byte)", tlv.Tag.String(), len(tlv.Value))
	}
	return fmt.Sprintf("%s (%d elem)", tlv.Tag.String(), len(tlv.Children))
}

func (tlv *TLV) MarshalValue(marshaler encoding.BinaryMarshaler) (err error) {
	if !tlv.Tag.Primitive() {
		return errors.New("cannot marshal value on constructed")
	}
	tlv.Value, err = marshaler.MarshalBinary()
	return
}

func (tlv *TLV) UnmarshalValue(unmarshaler encoding.BinaryUnmarshaler) error {
	if !tlv.Tag.Primitive() {
		return errors.New("cannot unmarshal value on constructed")
	}
	return unmarshaler.UnmarshalBinary(tlv.Value)
}
