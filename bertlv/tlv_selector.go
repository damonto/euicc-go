package bertlv

func (tlv *TLV) At(index int) *TLV {
	switch {
	case index >= 0 && index < len(tlv.Children):
		return tlv.Children[index]
	case index < 0 && index >= -len(tlv.Children):
		return tlv.Children[len(tlv.Children)+index]
	}
	panic("tlv: index out of bounds")
}

func (tlv *TLV) First(tag Tag) *TLV {
	for _, child := range tlv.Children {
		if child != nil && child.Tag.Equal(tag) {
			return child
		}
	}
	return nil
}

func (tlv *TLV) Find(tag Tag) (matches []*TLV) {
	for _, child := range tlv.Children {
		if child != nil && child.Tag.Equal(tag) {
			matches = append(matches, child)
		}
	}
	return matches
}

func (tlv *TLV) Select(tags ...Tag) *TLV {
	next := tlv
	for index := 0; index < len(tags); index++ {
		if next = next.First(tags[index]); next == nil {
			return nil
		}
	}
	return next
}
