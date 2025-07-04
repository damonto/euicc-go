package mbim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

// region Proxy Configuration

type ProxyConfigRequest struct {
	TxnID      uint32
	DevicePath string
	Timeout    uint32
	Response   *ProxyConfigResponse
}

func (r *ProxyConfigRequest) Message() *Message {
	utf16s := utf16.Encode([]rune(r.DevicePath))
	utf16s = append(utf16s, 0) // null terminator
	pb := new(bytes.Buffer)
	_ = binary.Write(pb, binary.LittleEndian, utf16s)
	devicePathUTF16 := pb.Bytes()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(12))
	binary.Write(buf, binary.LittleEndian, uint32(len(devicePathUTF16)))
	binary.Write(buf, binary.LittleEndian, r.Timeout)
	buf.Write(devicePathUTF16)

	r.Response = new(ProxyConfigResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMbimProxyControl,
			CID:             CIDProxyControlConfiguration,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        r.Response,
		},
	}
}

// ProxyConfigResponse is empty for now
type ProxyConfigResponse struct{}

func (r *ProxyConfigResponse) UnmarshalBinary(data []byte) error { return nil }

// endregion

// region Open Device Request

type OpenDeviceRequest struct {
	TxnID   uint32
	Payload *OpenDevicePayload
}

type OpenDevicePayload struct {
	MaxControlTransfer uint32
}

func (p *OpenDevicePayload) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, p.MaxControlTransfer)
	return buf, nil
}

func (p *OpenDevicePayload) UnmarshalBinary(data []byte) error {
	p.MaxControlTransfer = binary.LittleEndian.Uint32(data[0:4])
	return nil
}

func (r *OpenDeviceRequest) Message() *Message {
	r.Payload = &OpenDevicePayload{
		MaxControlTransfer: 4096,
	}
	return &Message{
		Type:          MessageTypeOpen,
		TransactionID: r.TxnID,
		Payload:       r.Payload,
	}
}

func (r *OpenDeviceRequest) UnmarshalBinary(data []byte) error {
	return r.Payload.UnmarshalBinary(data)
}

// endregion

// region Open Logical Channel

type OpenLogicalChannelRequest struct {
	TxnID       uint32
	AppId       []byte
	SelectP2Arg uint32
	Group       uint32
	Response    *OpenLogicalChannelResponse
}

func (r *OpenLogicalChannelRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(r.AppId)))
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, r.SelectP2Arg)
	binary.Write(buf, binary.LittleEndian, r.Group)
	buf.Write(r.AppId)
	r.Response = new(OpenLogicalChannelResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccOpenChannel,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        r.Response,
		},
	}
}

type OpenLogicalChannelResponse struct {
	Status   uint32
	Channel  uint32
	Response []byte
}

func (r *OpenLogicalChannelResponse) UnmarshalBinary(data []byte) error {
	r.Status = binary.LittleEndian.Uint32(data[0:4])
	r.Channel = binary.LittleEndian.Uint32(data[4:8])
	n := binary.LittleEndian.Uint32(data[8:12])
	if len(data) < int(16+n) {
		return errors.New("APDU response buffer too short")
	}
	r.Response = data[16 : 16+n]
	return nil
}

// endregion

// region Close Logical Channel

type CloseLogicalChannelRequest struct {
	Channel  uint32 // Channel to close
	Group    uint32 // Channel group to close
	TxnID    uint32
	Response *CloseLogicalChannelResponse
}

func (r *CloseLogicalChannelRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(r.Channel))
	binary.Write(buf, binary.LittleEndian, r.Group)
	r.Response = new(CloseLogicalChannelResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccCloseChannel,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        r.Response,
		},
	}
}

type CloseLogicalChannelResponse struct {
	Status uint32
}

func (r *CloseLogicalChannelResponse) UnmarshalBinary(data []byte) error {
	r.Status = binary.LittleEndian.Uint32(data[0:4])
	return nil
}

// endregion

// region Transmit APDU
type TransmitAPDURequest struct {
	TxnID           uint32
	Channel         uint32
	SecureMessaging uint32
	ClassByteType   uint32
	APDU            []byte
	Response        *TransmitAPDUResponse
}

func (r *TransmitAPDURequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Channel)
	binary.Write(buf, binary.LittleEndian, r.SecureMessaging)
	binary.Write(buf, binary.LittleEndian, r.ClassByteType)
	binary.Write(buf, binary.LittleEndian, uint32(len(r.APDU)))
	binary.Write(buf, binary.LittleEndian, uint32(20))
	buf.Write(r.APDU)
	r.Response = new(TransmitAPDUResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TxnID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsUiccLowLevelAccess,
			CID:             CIDUiccAPDU,
			CommandType:     CommandTypeSet,
			Data:            buf.Bytes(),
			Response:        r.Response,
		},
	}
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
	if len(data) < int(12+n) {
		return errors.New("APDU response buffer too short")
	}
	r.APDU = data[12 : 12+n]
	return nil
}

// endregion
