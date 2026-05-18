package uim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

// region Switch Slot Request

type SwitchSlotRequest struct {
	ClientID      uint8
	TransactionID uint16
	LogicalSlot   uint8
	PhysicalSlot  uint32
	Response      *SwitchSlotResponse
}

func (r *SwitchSlotRequest) Request() *protocol.Request {
	r.Response = new(SwitchSlotResponse)
	physicalSlot := make([]byte, 4)
	binary.LittleEndian.PutUint32(physicalSlot, r.PhysicalSlot)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMSwitchSlot,
		ServiceType:   protocol.QMIServiceUIM,
		Value: protocol.TLVs{
			{Type: 0x01, Len: 1, Value: []byte{r.LogicalSlot}},
			{Type: 0x02, Len: uint16(len(physicalSlot)), Value: physicalSlot},
		},
		Response: r.Response,
	}
}

type SwitchSlotResponse struct{}

func (r *SwitchSlotResponse) UnmarshalResponse(TLVs *protocol.TLVs) error { return nil }

// endregion

// region Get Slot Status Request

type GetSlotStatusRequest struct {
	ClientID      uint8
	TransactionID uint16
	Response      *GetSlotStatusResponse
}

func (r *GetSlotStatusRequest) Request() *protocol.Request {
	r.Response = new(GetSlotStatusResponse)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMGetSlotStatus,
		ServiceType:   protocol.QMIServiceUIM,
		ReadTimeout:   1 * time.Second,
		Response:      r.Response,
	}
}

type GetSlotStatusResponse struct {
	Slots         []Slot
	ActivatedSlot uint8
}

type Slot struct {
	CardState   PhysicalCardState
	SlotState   SlotState
	LogicalSlot uint8
	ICCID       []byte
}

func (r *GetSlotStatusResponse) UnmarshalResponse(TLVs *protocol.TLVs) error {
	value, ok := TLVs.Find(0x10)
	if !ok {
		return errors.New("could not find slot status in response")
	}

	buf := bytes.NewReader(value.Value)
	slotCount, err := readUint8(buf, "slot count")
	if err != nil {
		return err
	}
	r.Slots = make([]Slot, 0, slotCount)
	for i := range slotCount {
		var slot Slot
		if err := readValue(buf, &slot.CardState, "physical card status"); err != nil {
			return err
		}
		if err := readValue(buf, &slot.SlotState, "physical slot status"); err != nil {
			return err
		}
		slot.LogicalSlot, err = readUint8(buf, "logical slot")
		if err != nil {
			return err
		}
		slot.ICCID, err = readUint8Array(buf, "ICCID")
		if err != nil {
			return err
		}
		if slot.SlotState == SlotStateActive {
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

func (r *GetCardStatusRequest) Request() *protocol.Request {
	r.Response = new(GetCardStatusResponse)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMGetCardStatus,
		ServiceType:   protocol.QMIServiceUIM,
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
	State        CardState
	Applications []Application
}

type Application struct {
	Type  CardApplicationType
	State CardApplicationState
}

func (r *GetCardStatusResponse) UnmarshalResponse(TLVs *protocol.TLVs) error {
	value, ok := TLVs.Find(0x10)
	if !ok {
		return errors.New("could not find card status in response")
	}

	buf := bytes.NewReader(value.Value)
	if err := readValue(buf, &r.IndexGWPrimary, "GW primary index"); err != nil {
		return err
	}
	if err := readValue(buf, &r.Index1XPrimary, "1X primary index"); err != nil {
		return err
	}
	if err := readValue(buf, &r.IndexGWSecondary, "GW secondary index"); err != nil {
		return err
	}
	if err := readValue(buf, &r.Index1XSecondary, "1X secondary index"); err != nil {
		return err
	}

	cardLen, err := readUint8(buf, "card count")
	if err != nil {
		return err
	}
	r.Cards = make([]Card, 0, cardLen)
	for range cardLen {
		var card Card
		if err := readValue(buf, &card.State, "card state"); err != nil {
			return err
		}
		if err := skipBytes(buf, 4, "card PIN/error status"); err != nil {
			return err
		}

		appLen, err := readUint8(buf, "application count")
		if err != nil {
			return err
		}
		card.Applications = make([]Application, 0, appLen)
		for range appLen {
			var app Application
			if err := readValue(buf, &app.Type, "application type"); err != nil {
				return err
			}
			if err := readValue(buf, &app.State, "application state"); err != nil {
				return err
			}
			card.Applications = append(card.Applications, app)

			if err := skipBytes(buf, 4, "application personalization status"); err != nil {
				return err
			}
			if _, err := readUint8Array(buf, "application identifier"); err != nil {
				return err
			}
			if err := skipBytes(buf, 7, "application PIN status"); err != nil {
				return err
			}
		}
		r.Cards = append(r.Cards, card)
	}
	return nil
}

func (r *GetCardStatusResponse) Ready() bool {
	for _, card := range r.Cards {
		if card.State == CardStatePresent {
			for _, app := range card.Applications {
				if app.Type == CardApplicationTypeUSIM && app.State == CardApplicationStateReady {
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

func (r *OpenLogicalChannelRequest) Request() *protocol.Request {
	value := append([]byte{byte(len(r.AID))}, r.AID...)
	r.Response = new(OpenLogicalChannelResponse)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMOpenLogicalChannel,
		ServiceType:   protocol.QMIServiceUIM,
		Value: protocol.TLVs{
			{Type: 0x10, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
		Response: r.Response,
	}
}

type OpenLogicalChannelResponse struct {
	Channel byte
}

func (r *OpenLogicalChannelResponse) UnmarshalResponse(TLVs *protocol.TLVs) error {
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

func (r *CloseLogicalChannelRequest) Request() *protocol.Request {
	r.Response = new(CloseLogicalChannelResponse)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMCloseLogicalChannel,
		ServiceType:   protocol.QMIServiceUIM,
		Value: protocol.TLVs{
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
			{Type: 0x11, Len: 1, Value: []byte{r.Channel}},
			// Optional TLV 0x13: Terminate Application. Value 0x01 asks the
			// modem to terminate the selected application while closing the
			// logical channel. libqmi/qmicli do not send it by default.
			// {Type: 0x13, Len: 1, Value: []byte{0x01}},
		},
		Response: r.Response,
	}
}

type CloseLogicalChannelResponse struct{}

func (r *CloseLogicalChannelResponse) UnmarshalResponse(TLVs *protocol.TLVs) error { return nil }

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

func (r *TransmitAPDURequest) Request() *protocol.Request {
	length := len(r.Command)
	value := append([]byte{byte(length), byte(length >> 8)}, r.Command...)
	r.Response = new(TransmitAPDUResponse)
	return &protocol.Request{
		ClientID:      r.ClientID,
		TransactionID: r.TransactionID,
		MessageID:     protocol.QMIUIMSendAPDU,
		ServiceType:   protocol.QMIServiceUIM,
		Value: protocol.TLVs{
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

func (r *TransmitAPDUResponse) UnmarshalResponse(TLVs *protocol.TLVs) error {
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

func readValue(r io.Reader, out any, field string) error {
	if err := binary.Read(r, binary.LittleEndian, out); err != nil {
		return fmt.Errorf("read %s: %w", field, err)
	}
	return nil
}

func readUint8(r io.Reader, field string) (uint8, error) {
	var v uint8
	if err := readValue(r, &v, field); err != nil {
		return 0, err
	}
	return v, nil
}

func readUint8Array(r *bytes.Reader, field string) ([]byte, error) {
	n, err := readUint8(r, field+" length")
	if err != nil {
		return nil, err
	}
	value := make([]byte, n)
	if _, err := io.ReadFull(r, value); err != nil {
		return nil, fmt.Errorf("read %s: %w", field, err)
	}
	return value, nil
}

func skipBytes(r *bytes.Reader, n int, field string) error {
	if r.Len() < n {
		return fmt.Errorf("read %s: %w", field, io.ErrUnexpectedEOF)
	}
	if _, err := r.Seek(int64(n), io.SeekCurrent); err != nil {
		return fmt.Errorf("skip %s: %w", field, err)
	}
	return nil
}
