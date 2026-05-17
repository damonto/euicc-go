package protocol

import "fmt"

type InternalOpenRequest struct {
	TransactionID uint16
	DevicePath    []byte
	Response      *InternalOpenResponse
}

func (r *InternalOpenRequest) Request() *Request {
	r.Response = new(InternalOpenResponse)
	return &Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlInternalProxyOpen,
		ServiceType:   QMIServiceControl,
		Value: TLVs{
			{Type: 0x01, Len: uint16(len(r.DevicePath)), Value: r.DevicePath},
		},
		Response: r.Response,
	}
}

type InternalOpenResponse struct{}

func (r *InternalOpenResponse) UnmarshalResponse(TLVs *TLVs) error { return nil }

type AllocateClientIDRequest struct {
	TransactionID uint16
	ServiceType   ServiceType
	Response      *AllocateClientIDResponse
}

func (r *AllocateClientIDRequest) Request() *Request {
	r.Response = new(AllocateClientIDResponse)
	return &Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdAllocateClientID,
		ServiceType:   QMIServiceControl,
		Value: TLVs{
			{Type: 0x01, Len: 1, Value: []byte{byte(r.ServiceType)}},
		},
		Response: r.Response,
	}
}

type AllocateClientIDResponse struct {
	ClientID uint8
}

func (r *AllocateClientIDResponse) UnmarshalResponse(TLVs *TLVs) error {
	if value, ok := TLVs.Find(0x01); ok && len(value.Value) >= 2 {
		r.ClientID = value.Value[1]
		return nil
	}
	return fmt.Errorf("could not find allocated client ID in response")
}

type ReleaseClientIDRequest struct {
	ClientID      uint8
	TransactionID uint16
	ServiceType   ServiceType
	Response      *ReleaseClientIDResponse
}

func (r *ReleaseClientIDRequest) Request() *Request {
	r.Response = new(ReleaseClientIDResponse)
	return &Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdReleaseClientID,
		ServiceType:   QMIServiceControl,
		Value: TLVs{
			{Type: 0x01, Len: 2, Value: []byte{byte(r.ServiceType), r.ClientID}},
		},
		Response: r.Response,
	}
}

type ReleaseClientIDResponse struct {
	ClientID uint8
}

func (r *ReleaseClientIDResponse) UnmarshalResponse(TLVs *TLVs) error {
	if value, ok := TLVs.Find(0x01); ok && len(value.Value) >= 2 {
		r.ClientID = value.Value[1]
		return nil
	}
	return fmt.Errorf("could not find released client ID in response")
}
