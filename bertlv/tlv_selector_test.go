package bertlv

import (
	"bytes"
	"testing"
)

func TestTLV_At(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	for index, want := range map[int][]byte{
		0: {0x01}, -3: {0x01},
		1: {0x02}, -2: {0x02},
		2: {0x03}, -1: {0x03},
	} {
		if got := tree.At(index).Value; !bytes.Equal(got, want) {
			t.Errorf("TLV.At(%d).Value = % X, want % X", index, got, want)
		}
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("TLV.At(4) did not panic")
			}
		}()
		tree.At(4)
	}()
}

func TestTLV_Find(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	for i := range 3 {
		got := tree.Find(Primitive.ContextSpecific(uint64(i + 1)))
		if len(got) != 1 || got[0] != tree.Children[i] {
			t.Errorf("TLV.Find(%d) = %#v, want child %d", i+1, got, i)
		}
	}
	if got := tree.Find(Primitive.ContextSpecific(4)); got != nil {
		t.Errorf("TLV.Find(4) = %#v, want nil", got)
	}
}

func TestTLV_First(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	for i := range 3 {
		if got := tree.First(Primitive.ContextSpecific(uint64(i + 1))); got != tree.Children[i] {
			t.Errorf("TLV.First(%d) = %p, want %p", i+1, got, tree.Children[i])
		}
	}
}

func TestTLV_Select(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewChildren(
			Constructed.ContextSpecific(0),
			NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		),
	)
	want := tree.Children[0].Children[0]
	if got := tree.Select(
		Constructed.ContextSpecific(0),
		Primitive.ContextSpecific(1),
	); got != want {
		t.Errorf("TLV.Select() = %p, want %p", got, want)
	}
	if got := tree.Select(
		Constructed.ContextSpecific(0),
		Constructed.ContextSpecific(1),
		Primitive.ContextSpecific(2),
	); got != nil {
		t.Errorf("TLV.Select() = %p, want nil", got)
	}
}
