package mbim

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// region Open Device Request

type OpenDeviceRequest struct {
	message *Message
	TxnID   uint32
}

type OpenDeviceResponse struct{}

func (p *OpenDeviceResponse) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 4096)
	return buf, nil
}

func (p *OpenDeviceResponse) UnmarshalBinary(data []byte) error {
	if err := MBIMStatusError(binary.LittleEndian.Uint32(data)); err != MBIMStatusErrorNone {
		return MBIMError{Code: err}
	}
	return nil
}

func (r *OpenDeviceRequest) Message() *Message {
	r.message = &Message{
		Type:          MessageTypeOpen,
		TransactionID: r.TxnID,
		Payload:       new(OpenDeviceResponse),
	}
	return r.message
}

func (r *OpenDeviceRequest) UnmarshalBinary(data []byte) error {
	return r.message.UnmarshalBinary(data)
}

// endregion

// region Open Logical Channel

type OpenLogicalChannelRequest struct {
	message      *Message
	TxnID        uint32
	AppId        []byte
	SelectP2Arg  uint32
	ChannelGroup uint32
}

func (r *OpenLogicalChannelRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(r.AppId)))
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, r.SelectP2Arg)
	binary.Write(buf, binary.LittleEndian, r.ChannelGroup)
	buf.Write(r.AppId)
	r.message = &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccOpenChannel,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        new(OpenLogicalChannelResponse),
		},
	}
	return r.message
}

func (r *OpenLogicalChannelRequest) Response() *OpenLogicalChannelResponse {
	if cmd, ok := r.message.Payload.(*Command); ok {
		if response, ok := cmd.Response.(*OpenLogicalChannelResponse); ok {
			return response
		}
	}
	return nil
}

type OpenLogicalChannelResponse struct {
	Status  uint32
	Channel uint32
}

func (r *OpenLogicalChannelResponse) UnmarshalBinary(data []byte) error {
	offset := 8
	r.Status = binary.LittleEndian.Uint32(data[offset : offset+4])
	r.Channel = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
	return nil
}

// endregion

// region Close Logical Channel

type CloseLogicalChannelRequest struct {
	message *Message
	Channel uint32 // Channel to close
	Group   uint32 // Channel group to close
	TxnID   uint32
}

func (r *CloseLogicalChannelRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(r.Channel))
	binary.Write(buf, binary.LittleEndian, r.Group)
	r.message = &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccCloseChannel,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        new(CloseLogicalChannelResponse),
		},
	}
	return r.message
}

func (r *CloseLogicalChannelRequest) Response() *CloseLogicalChannelResponse {
	if cmd, ok := r.message.Payload.(*Command); ok {
		if response, ok := cmd.Response.(*CloseLogicalChannelResponse); ok {
			return response
		}
	}
	return nil
}

type CloseLogicalChannelResponse struct {
	Status uint32
}

func (r *CloseLogicalChannelResponse) UnmarshalBinary(data []byte) error {
	return nil
}

// endregion

// region Transmit APDU
type TransmitAPDURequest struct {
	message         *Message
	TxnID           uint32
	Channel         uint32
	SecureMessaging uint32
	ClassByteType   uint32
	APDU            []byte
}

func (r *TransmitAPDURequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Channel)
	binary.Write(buf, binary.LittleEndian, r.SecureMessaging)
	binary.Write(buf, binary.LittleEndian, r.ClassByteType)
	binary.Write(buf, binary.LittleEndian, uint32(len(r.APDU)))
	binary.Write(buf, binary.LittleEndian, uint32(20))
	buf.Write(r.APDU)

	r.message = &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccAPDU,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        new(TransmitAPDUResponse),
		},
	}
	return r.message
}

func (r *TransmitAPDURequest) Response() *TransmitAPDUResponse {
	if cmd, ok := r.message.Payload.(*Command); ok {
		if response, ok := cmd.Response.(*TransmitAPDUResponse); ok {
			return response
		}
	}
	return nil
}

type TransmitAPDUResponse struct {
	Status uint32
	APDU   []byte
}

func (r *TransmitAPDUResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("APDU response data too short")
	}
	r.Status = binary.LittleEndian.Uint32(data[0:4])
	n := binary.LittleEndian.Uint32(data[4:8])
	if len(data) < int(8+n) {
		return errors.New("APDU response buffer too short")
	}
	r.APDU = data[8 : 8+n]
	return nil
}

// endregion
