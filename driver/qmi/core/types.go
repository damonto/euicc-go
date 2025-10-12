package core

import (
	"time"
)

type QMUXHeader struct {
	IfType       uint8
	Length       uint16
	ControlFlags uint8
	ServiceType  ServiceType
	ClientID     uint8
}

type Header[T uint8 | uint16] struct {
	MessageType   MessageType
	TransactionID T
	MessageID     MessageID
	MessageLength uint16
}

type ResponseUnmarshaler interface {
	UnmarshalResponse(TLVs *TLVs) error
}

type Request struct {
	ClientID      uint8
	TransactionID uint16
	ServiceType   ServiceType
	ReadTimeout   time.Duration
	MessageID     MessageID
	Value         TLVs
	Response      ResponseUnmarshaler
}

type Transport interface {
	Transmit(request *Request) error
}

func Transmit(transport Transport, request *Request) error {
	return transport.Transmit(request)
}
