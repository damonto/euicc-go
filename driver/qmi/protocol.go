package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

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

// region Switch Slot Request

type SwitchSlotRequest struct {
	ClientID      uint8
	TransactionID uint16
	LogicalSlot   uint8
	PhysicalSlot  uint32
	Response      *SwitchSlotResponse
}

func (r *SwitchSlotRequest) Request() *Request {
	r.Response = new(SwitchSlotResponse)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.PhysicalSlot)
	request := Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMSwitchSlot,
		ServiceType:   QMIServiceUIM,
		TLVs: []TLV{
			{Type: 0x01, Len: 1, Value: []byte{r.LogicalSlot}},
			{Type: 0x02, Len: 4, Value: buf.Bytes()},
		},
		Response: r.Response,
	}
	return &request
}

type SwitchSlotResponse struct{}

func (r *SwitchSlotResponse) UnmarshalResponse(TLVs map[uint8]TLV) error { return nil }

// endregion

// region Get Slot Status Request

type GetSlotStatusRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *GetSlotStatusResponse
}

func (r *GetSlotStatusRequest) Request() *Request {
	r.Response = new(GetSlotStatusResponse)
	request := Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMGetSlotStatus,
		ServiceType:   QMIServiceUIM,
		ReadTimeout:   1 * time.Second,
		Response:      r.Response,
	}
	return &request
}

type Slot struct {
	CardStatus  UIMPhysicalCardState
	SlotStatus  UIMSlotState
	LogicalSlot uint8
	ICCID       [10]byte
}

type GetSlotStatusResponse struct {
	Slots         []Slot
	ActivatedSlot uint8
}

func (r *GetSlotStatusResponse) UnmarshalResponse(TLVs map[uint8]TLV) error {
	if value, ok := TLVs[0x10]; ok {
		var slotCount uint8
		buf := bytes.NewBuffer(value.Value)
		binary.Read(buf, binary.LittleEndian, &slotCount)
		r.Slots = make([]Slot, 0, slotCount)
		for i := range slotCount {
			var slot Slot
			binary.Read(buf, binary.LittleEndian, &slot.CardStatus)
			binary.Read(buf, binary.LittleEndian, &slot.SlotStatus)
			binary.Read(buf, binary.LittleEndian, &slot.LogicalSlot)
			var iccidLen uint8
			binary.Read(buf, binary.LittleEndian, &iccidLen)
			if iccidLen > 0 {
				binary.Read(buf, binary.LittleEndian, &slot.ICCID)
			}
			if slot.SlotStatus == UIMSlotStateActive {
				r.ActivatedSlot = uint8(i + 1)
			}
			r.Slots = append(r.Slots, slot)
		}
		return nil
	}
	return errors.New("could not find slot status in response")
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
	return errors.New("could not find logical channel in response")
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
		return errors.New("could not find slot in response")
	}
	if value, ok := TLVs[0x11]; ok && len(value.Value) >= 1 {
		r.Channel = value.Value[0]
		return nil
	}
	return errors.New("could not find channel in response")
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
	return errors.New("could not find APDU response in message")
}

// endregion
