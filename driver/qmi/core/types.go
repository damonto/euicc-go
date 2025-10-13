package core

import "time"

type Request struct {
	ClientID      uint8
	TransactionID uint16
	ServiceType   ServiceType
	ReadTimeout   time.Duration
	MessageID     MessageID
	Value         TLVs
	Response      ResponseUnmarshaler
}

type ResponseUnmarshaler interface {
	UnmarshalResponse(TLVs *TLVs) error
}

type Transport interface {
	Transmit(request *Request) error
}
