package core

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

// endregion

// region Allocate Client ID Requests

type AllocateClientIDRequest struct {
	TransactionID uint16
	Response      *AllocateClientIDResponse
}

func (r *AllocateClientIDRequest) Request() *Request {
	r.Response = new(AllocateClientIDResponse)
	return &Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdAllocateClientID,
		ServiceType:   QMIServiceControl,
		Value: TLVs{
			{Type: 0x01, Len: 1, Value: []byte{byte(QMIServiceUIM)}},
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

// endregion

// region Release Client ID Request

type ReleaseClientIDRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *ReleaseClientIDResponse
}

func (r *ReleaseClientIDRequest) Request() *Request {
	r.Response = new(ReleaseClientIDResponse)
	return &Request{
		TransactionID: r.TransactionID,
		MessageID:     QMICtlCmdReleaseClientID,
		ServiceType:   QMIServiceControl,
		Value: TLVs{
			{Type: 0x01, Len: 2, Value: []byte{byte(QMIServiceUIM), r.ClientID}},
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
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMSwitchSlot,
		ServiceType:   QMIServiceUIM,
		Value: TLVs{
			{Type: 0x01, Len: 1, Value: []byte{r.LogicalSlot}},
			{Type: 0x02, Len: 4, Value: buf.Bytes()},
		},
		Response: r.Response,
	}
}

type SwitchSlotResponse struct{}

func (r *SwitchSlotResponse) UnmarshalResponse(TLVs *TLVs) error { return nil }

// endregion

// region Get Slot Status Request

type GetSlotStatusRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *GetSlotStatusResponse
}

func (r *GetSlotStatusRequest) Request() *Request {
	r.Response = new(GetSlotStatusResponse)
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMGetSlotStatus,
		ServiceType:   QMIServiceUIM,
		ReadTimeout:   1 * time.Second,
		Response:      r.Response,
	}
}

type GetSlotStatusResponse struct {
	Slots         []Slot
	ActivatedSlot uint8
}

type Slot struct {
	CardState   UIMPhysicalCardState
	SlotState   UIMSlotState
	LogicalSlot uint8
	ICCID       [10]byte
}

func (r *GetSlotStatusResponse) UnmarshalResponse(TLVs *TLVs) error {
	value, ok := TLVs.Find(0x10)
	if !ok {
		return errors.New("could not find slot status in response")
	}
	var slotCount uint8
	buf := bytes.NewBuffer(value.Value)
	binary.Read(buf, binary.LittleEndian, &slotCount)
	r.Slots = make([]Slot, 0, slotCount)
	for i := range slotCount {
		var slot Slot
		binary.Read(buf, binary.LittleEndian, &slot.CardState)
		binary.Read(buf, binary.LittleEndian, &slot.SlotState)
		binary.Read(buf, binary.LittleEndian, &slot.LogicalSlot)
		var iccidLen uint8
		binary.Read(buf, binary.LittleEndian, &iccidLen)
		if iccidLen > 0 {
			binary.Read(buf, binary.LittleEndian, &slot.ICCID)
		}
		if slot.SlotState == UIMSlotStateActive {
			r.ActivatedSlot = uint8(i + 1)
		}
		r.Slots = append(r.Slots, slot)
	}
	return nil
}

// endregion

// region Get Card Status Request

type GetCardStatusRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *GetCardStatusResponse
}

func (r *GetCardStatusRequest) Request() *Request {
	r.Response = new(GetCardStatusResponse)
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMGetCardStatus,
		ServiceType:   QMIServiceUIM,
		Response:      r.Response,
	}
}

type GetCardStatusResponse struct {
	IndexGWPrimary   uint16
	Index1XPrimary   uint16
	IndexGWSecondary uint16
	Index1XSecondary uint16
	Cards            []Card
}

type Card struct {
	State        UIMCardState
	Applications []Application
}

type Application struct {
	Type  UIMCardApplicationType
	State UIMCardApplicationState
}

func (r *GetCardStatusResponse) UnmarshalResponse(TLVs *TLVs) error {
	value, ok := TLVs.Find(0x10)
	if !ok {
		return errors.New("could not find card status in response")
	}

	buf := bytes.NewBuffer(value.Value)
	binary.Read(buf, binary.LittleEndian, &r.IndexGWPrimary)
	binary.Read(buf, binary.LittleEndian, &r.Index1XPrimary)
	binary.Read(buf, binary.LittleEndian, &r.IndexGWSecondary)
	binary.Read(buf, binary.LittleEndian, &r.Index1XSecondary)

	var cardLen uint8
	binary.Read(buf, binary.LittleEndian, &cardLen)
	r.Cards = make([]Card, 0, cardLen)
	for range cardLen {
		var card Card
		binary.Read(buf, binary.LittleEndian, &card.State)

		buf.Next(4)

		var appLen uint8
		binary.Read(buf, binary.LittleEndian, &appLen)
		card.Applications = make([]Application, 0, appLen)
		for range appLen {
			var app Application
			binary.Read(buf, binary.LittleEndian, &app.Type)
			binary.Read(buf, binary.LittleEndian, &app.State)
			card.Applications = append(card.Applications, app)

			buf.Next(28)
		}
		r.Cards = append(r.Cards, card)
	}
	return nil
}

func (r *GetCardStatusResponse) Ready() bool {
	for _, card := range r.Cards {
		if card.State == UIMCardStatusPresent {
			for _, app := range card.Applications {
				if app.Type == UIMCardApplicationTypeUSIM && app.State == UIMCardApplicationStateReady {
					return true
				}
			}
		}
	}
	return false
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
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMOpenLogicalChannel,
		ServiceType:   QMIServiceUIM,
		Value: TLVs{
			{Type: 0x10, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
		Response: r.Response,
	}
}

type OpenLogicalChannelResponse struct {
	Channel byte
}

func (r *OpenLogicalChannelResponse) UnmarshalResponse(TLVs *TLVs) error {
	if value, ok := TLVs.Find(0x10); ok && len(value.Value) >= 1 {
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
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMCloseLogicalChannel,
		ServiceType:   QMIServiceUIM,
		Value: TLVs{
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
			{Type: 0x11, Len: 1, Value: []byte{r.Channel}},
			{Type: 0x13, Len: 1, Value: []byte{0x01}},
		},
		Response: r.Response,
	}
}

type CloseLogicalChannelResponse struct{}

func (r *CloseLogicalChannelResponse) UnmarshalResponse(TLVs *TLVs) error { return nil }

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
	return &Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     QMIUIMSendAPDU,
		ServiceType:   QMIServiceUIM,
		Value: TLVs{
			{Type: 0x10, Len: 1, Value: []byte{r.Channel}},
			{Type: 0x02, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
		Response: r.Response,
	}
}

type TransmitAPDUResponse struct {
	Response []byte
}

func (r *TransmitAPDUResponse) UnmarshalResponse(TLVs *TLVs) error {
	if value, ok := TLVs.Find(0x10); ok && len(value.Value) >= 2 {
		n := int(value.Value[0]) | (int(value.Value[1]) << 8)
		if len(value.Value) >= 2+n {
			r.Response = value.Value[2 : 2+n]
			return nil
		}
	}
	return errors.New("could not find APDU response in message")
}

// endregion
