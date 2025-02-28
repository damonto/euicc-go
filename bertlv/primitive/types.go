package primitive

type Unmarshaler func([]byte) error

func (u Unmarshaler) UnmarshalBinary(data []byte) error {
	return u(data)
}

type Marshaler func() ([]byte, error)

func (m Marshaler) MarshalBinary() ([]byte, error) {
	return m()
}
