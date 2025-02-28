package bertlv

import (
	"crypto/subtle"
	"fmt"
	"io"
)

type Tag []byte

func NewTag(class Class, form Form, value uint64) Tag {
	mask := byte(class)<<6 | byte(form)<<5
	if value < 0x1f {
		return Tag{mask | byte(value)}
	}
	tag := Tag{mask | 0x1f}
	for value > 0 {
		b := byte(value & 0x7f)
		value >>= 7
		if len(tag) > 1 || value > 0 {
			b |= 0x80
		}
		tag = append(tag, b)
	}
	return tag
}

func (t *Tag) ReadFrom(r io.Reader) (n int64, err error) {
	var tag [11]byte
	if _, err = io.ReadAtLeast(r, tag[0:1], 1); err != nil {
		return n, fmt.Errorf("tag encoding with less than one byte\n%w", err)
	}
	if tag[0]&0x1f != 0x1f {
		*t = tag[0:1]
		return
	}
	for n = 1; ; n++ {
		if _, err = io.ReadAtLeast(r, tag[n:n+1], 1); err != nil {
			return n, fmt.Errorf("tag encoding with more than %d bytes\n%w", n+1, err)
		}
		if tag[n]>>7 == 0b0 {
			*t = tag[0 : n+1]
			return
		}
	}
}

func (t *Tag) String() string {
	switch t.Class() {
	case Universal:
		return fmt.Sprintf("[UNIVERSAL %d]", t.Value())
	case Application:
		return fmt.Sprintf("[APPLICATION %d]", t.Value())
	case Private:
		return fmt.Sprintf("[PRIVATE %d]", t.Value())
	}
	return fmt.Sprintf("[%d]", t.Value())
}

func (t *Tag) Value() uint64 {
	tag := *t
	if value := uint64(tag[0] & 0x1f); value != 0x1f {
		return value
	}
	value := uint64(0)
	for index := 1; index < len(tag); index++ {
		value <<= 7
		value |= uint64(tag[index] & 0x7f)
		if tag[index]>>7 == 0 {
			break
		}
	}
	return value
}

func (t *Tag) If(class Class, form Form, value uint64) bool {
	return t.Class() == class && t.Form() == form && t.Value() == value
}

func (t *Tag) Equal(tag Tag) bool {
	return subtle.ConstantTimeCompare(*t, tag) == 1
}

func (t *Tag) Form() Form        { return Form((*t)[0] >> 5 & 0b1) }
func (t *Tag) Primitive() bool   { return t.Form() == Primitive }
func (t *Tag) Constructed() bool { return t.Form() == Constructed }

func (t *Tag) Class() Class          { return Class((*t)[0] >> 6) }
func (t *Tag) Universal() bool       { return t.Class() == Universal }
func (t *Tag) Application() bool     { return t.Class() == Application }
func (t *Tag) ContextSpecific() bool { return t.Class() == ContextSpecific }
func (t *Tag) Private() bool         { return t.Class() == Private }
