package bertlv

type TLV struct {
	Tag      Tag
	Value    []byte
	Children []*TLV
}
