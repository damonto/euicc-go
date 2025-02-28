package apdu

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
)

type Response []byte

func (r Response) Data() []byte   { return r[0 : len(r)-2] }
func (r Response) SW() uint16     { return binary.BigEndian.Uint16(r[len(r)-2:]) }
func (r Response) SW1() byte      { return r[len(r)-2] }
func (r Response) SW2() byte      { return r[len(r)-1] }
func (r Response) OK() bool       { return r.SW() == 0x9000 }
func (r Response) HasMore() bool  { return r.SW1() == 0x61 }
func (r Response) String() string { return strings.ToUpper(hex.EncodeToString(r)) }
