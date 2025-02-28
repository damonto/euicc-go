package bertlv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTLV_At(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	assert.Equal(t, []byte{0x01}, tree.At(0).Value)
	assert.Equal(t, []byte{0x01}, tree.At(-3).Value)
	assert.Equal(t, []byte{0x02}, tree.At(1).Value)
	assert.Equal(t, []byte{0x02}, tree.At(-2).Value)
	assert.Equal(t, []byte{0x03}, tree.At(2).Value)
	assert.Equal(t, []byte{0x03}, tree.At(-1).Value)
	assert.Panics(t, func() { tree.At(4) })
}

func TestTLV_Find(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	assert.Equal(t, []*TLV{tree.Children[0]}, tree.Find(Primitive.ContextSpecific(1)))
	assert.Equal(t, []*TLV{tree.Children[1]}, tree.Find(Primitive.ContextSpecific(2)))
	assert.Equal(t, []*TLV{tree.Children[2]}, tree.Find(Primitive.ContextSpecific(3)))
	assert.Nil(t, tree.Find(Primitive.ContextSpecific(4)))
}

func TestTLV_First(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		NewValue(Primitive.ContextSpecific(2), []byte{0x02}),
		NewValue(Primitive.ContextSpecific(3), []byte{0x03}),
	)
	assert.Equal(t, tree.Children[0], tree.First(Primitive.ContextSpecific(1)))
	assert.Equal(t, tree.Children[1], tree.First(Primitive.ContextSpecific(2)))
	assert.Equal(t, tree.Children[2], tree.First(Primitive.ContextSpecific(3)))
}

func TestTLV_Select(t *testing.T) {
	tree := NewChildren(
		Constructed.ContextSpecific(0),
		NewChildren(
			Constructed.ContextSpecific(0),
			NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		),
	)
	assert.Equal(t,
		tree.Children[0].Children[0],
		tree.Select(
			Constructed.ContextSpecific(0),
			Primitive.ContextSpecific(1),
		),
	)
	assert.Nil(t,
		tree.Select(
			Constructed.ContextSpecific(0),
			Constructed.ContextSpecific(1),
			Primitive.ContextSpecific(2),
		),
	)
}
