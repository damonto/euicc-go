package bertlv

type Class byte
type Form byte

const (
	Universal       Class = 0b00
	Application     Class = 0b01
	ContextSpecific Class = 0b10
	Private         Class = 0b11
	Primitive       Form  = 0b0
	Constructed     Form  = 0b1
)

func (c Class) Primitive(value uint64) Tag   { return NewTag(c, Primitive, value) }
func (c Class) Constructed(value uint64) Tag { return NewTag(c, Constructed, value) }

func (f Form) Universal(value uint64) Tag       { return NewTag(Universal, f, value) }
func (f Form) Application(value uint64) Tag     { return NewTag(Application, f, value) }
func (f Form) ContextSpecific(value uint64) Tag { return NewTag(ContextSpecific, f, value) }
func (f Form) Private(value uint64) Tag         { return NewTag(Private, f, value) }
