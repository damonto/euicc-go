package bertlv

type Reflective interface {
	Tag() Tag
}

type Unmarshaler interface {
	UnmarshalBERTLV(*TLV) error
}

type Marshaler interface {
	MarshalBERTLV() (*TLV, error)
}
