package qmi

import "fmt"

// region Internal Open Request

type InternalOpenRequest struct {
	TransactionID uint16
	DevicePath    []byte
	Response      *InternalOpenResponse
}

func (r *InternalOpenRequest) Request() *Request {
	r.Response = new(InternalOpenResponse)
	request := Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlInternalProxyOpen,
		ServiceType:   QMIServiceControl,
		TLVs: []TLV{
			{Type: 0x01, Len: uint16(len(r.DevicePath)), Value: r.DevicePath},
		},
		Response: r.Response,
	}
	return &request
}

type InternalOpenResponse struct{}

func (r *InternalOpenResponse) UnmarshalResponse(TLVs map[uint8]TLV) error { return nil }

// endregion

// region Allocate Client ID Requests

type AllocateClientIDRequest struct {
	TransactionID uint16
	Response      *AllocateClientIDResponse
}

func (r *AllocateClientIDRequest) Request() *Request {
	r.Response = new(AllocateClientIDResponse)
	request := Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdAllocateClientID,
		ServiceType:   QMIServiceControl,
		TLVs: []TLV{
			{Type: 0x01, Len: 1, Value: []byte{byte(QMIServiceUIM)}},
		},
		Response: r.Response,
	}
	return &request
}

type AllocateClientIDResponse struct {
	ClientID uint8
}

func (r *AllocateClientIDResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x01]; ok && len(value.Value) >= 2 {
		r.ClientID = value.Value[1]
		return nil
	}
	return fmt.Errorf("could not find allocated client ID in response")
}

// endregion

// region Release Client ID Request

type ReleaseClientIDRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *ReleaseClientIDResponse
}

func (r *ReleaseClientIDRequest) Request() *Request {
	r.Response = new(ReleaseClientIDResponse)
	request := Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdReleaseClientID,
		ServiceType:   QMIServiceControl,
		TLVs: []TLV{
			{Type: 0x01, Len: 2, Value: []byte{byte(QMIServiceUIM), r.ClientID}},
		},
		Response: r.Response,
	}
	return &request
}

type ReleaseClientIDResponse struct {
	ClientID uint8
}

func (r *ReleaseClientIDResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x01]; ok && len(value.Value) >= 2 {
		r.ClientID = value.Value[1]
		return nil
	}
	return fmt.Errorf("could not find released client ID in response")
}

// endregion

// region Open Logical Channel Request

type OpenLogicalChannelRequest struct {
	ClientID      uint8
	TransactionID uint16
	Slot          byte
	AID           []byte
	Response      *OpenLogicalChannelResponse
}

func (r *OpenLogicalChannelRequest) Request() *Request {
	value := append([]byte{byte(len(r.AID))}, r.AID...)
	r.Response = new(OpenLogicalChannelResponse)
	request := Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMOpenLogicalChannel,
		ServiceType:   QMIServiceUIM,
		TLVs: []TLV{
			{Type: 0x10, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
		Response: r.Response,
	}
	return &request
}

type OpenLogicalChannelResponse struct {
	Channel byte
}

func (r *OpenLogicalChannelResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x10]; ok && len(value.Value) >= 1 {
		r.Channel = value.Value[0]
		return nil
	}
	return fmt.Errorf("could not find logical channel in response")
}

// endregion

// region Close Logical Channel Request

type CloseLogicalChannelRequest struct {
	ClientID      uint8
	TransactionID uint16
	Slot          byte
	Channel       byte
	Response      *CloseLogicalChannelResponse
}

func (r *CloseLogicalChannelRequest) Request() *Request {
	r.Response = new(CloseLogicalChannelResponse)
	request := Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMCloseLogicalChannel,
		ServiceType:   QMIServiceUIM,
		TLVs: []TLV{
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
			{Type: 0x11, Len: 1, Value: []byte{r.Channel}},
			{Type: 0x13, Len: 1, Value: []byte{0x01}},
		},
		Response: r.Response,
	}
	return &request
}

type CloseLogicalChannelResponse struct {
	Slot    byte
	Channel byte
}

func (r *CloseLogicalChannelResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x01]; ok && len(value.Value) >= 1 {
		r.Slot = value.Value[0]
	} else {
		return fmt.Errorf("could not find slot in response")
	}
	if value, ok := TLVs[0x11]; ok && len(value.Value) >= 1 {
		r.Channel = value.Value[0]
		return nil
	}
	return fmt.Errorf("could not find channel in response")
}

// endregion

// region Transmit APDU Request

type TransmitAPDURequest struct {
	ClientID      uint8
	TransactionID uint16
	Slot          byte
	Channel       byte
	Command       []byte
	Response      *TransmitAPDUResponse
}

func (r *TransmitAPDURequest) Request() *Request {
	length := len(r.Command)
	value := append([]byte{byte(length), byte(length >> 8)}, r.Command...)
	r.Response = new(TransmitAPDUResponse)
	request := Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMSendAPDU,
		ServiceType:   QMIServiceUIM,
		TLVs: []TLV{
			{Type: 0x10, Len: 1, Value: []byte{r.Channel}},
			{Type: 0x02, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
		Response: r.Response,
	}
	return &request
}

type TransmitAPDUResponse struct {
	Response []byte
}

func (r *TransmitAPDUResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x10]; ok && len(value.Value) >= 2 {
		n := int(value.Value[0]) | (int(value.Value[1]) << 8)
		if len(value.Value) >= 2+n {
			r.Response = value.Value[2 : 2+n]
			return nil
		}
	}
	return fmt.Errorf("could not find APDU response in message")
}

// endregion
