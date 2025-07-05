package mbim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
	"unicode/utf16"
)

// region Proxy Configuration

type ProxyConfigRequest struct {
	TransactionID uint32
	DevicePath    string
	Timeout       uint32
	Response      *ProxyConfigResponse
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
		TransactionID: r.TransactionID,
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
	TransactionID uint32
	Payload       *OpenDevicePayload
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
		TransactionID: r.TransactionID,
		Payload:       r.Payload,
	}
}

func (r *OpenDeviceRequest) UnmarshalBinary(data []byte) error {
	return r.Payload.UnmarshalBinary(data)
}

// endregion

// region Device Slot Mappings

type DeviceSlotMappingsRequest struct {
	TransactionID uint32
	MapCount      uint32
	SlotMappings  []SlotMapping
	Response      *DeviceSlotMappingsResponse
}

type SlotMapping struct {
	Slot uint32
}

func (r *DeviceSlotMappingsRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.MapCount)

	if r.MapCount > 0 {
		dataOffset := 4 + (r.MapCount * 8) // 4 bytes for MapCount + offset table
		for i := uint32(0); i < r.MapCount; i++ {
			binary.Write(buf, binary.LittleEndian, dataOffset+i*4) // offset to slot data
			binary.Write(buf, binary.LittleEndian, uint32(4))      // size of slot data
		}
		for _, mapping := range r.SlotMappings {
			binary.Write(buf, binary.LittleEndian, mapping.Slot)
		}
	}

	r.Response = new(DeviceSlotMappingsResponse)
	commandType := CommandTypeQuery
	if r.MapCount > 0 {
		commandType = CommandTypeSet
	}
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TransactionID,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceMsBasicConnectExtensions,
			CID:             CIDDeviceSlotMappings,
			CommandType:     uint32(commandType),
			Data:            buf.Bytes(),
			Response:        r.Response,
		},
	}
}

type DeviceSlotMappingsResponse struct {
	MapCount     uint32
	SlotMappings []SlotMapping
}

func (r *DeviceSlotMappingsResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("device slot mappings response data too short")
	}
	r.MapCount = binary.LittleEndian.Uint32(data[0:4])
	r.SlotMappings = make([]SlotMapping, r.MapCount)

	if r.MapCount == 0 {
		return nil
	}
	dataOffset := 4 + r.MapCount*8
	if len(data) < int(dataOffset) {
		return errors.New("device slot mappings response buffer too short")
	}
	for i := uint32(0); i < r.MapCount; i++ {
		slotDataOffset := dataOffset + i*4
		if len(data) < int(slotDataOffset+4) {
			return errors.New("device slot mappings response slot data too short")
		}
		r.SlotMappings[i].Slot = binary.LittleEndian.Uint32(data[slotDataOffset : slotDataOffset+4])
	}
	return nil
}

// endregion

// region Subscriber Ready Status

type SubscriberReadyStatusRequest struct {
	TransactionID uint32
	Response      *SubscriberReadyStatusResponse
}

func (r *SubscriberReadyStatusRequest) Message() *Message {
	r.Response = new(SubscriberReadyStatusResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TransactionID,
		ReadTimeout:   1 * time.Second,
		Payload: &Command{
			FragmentTotal:   1,
			FragmentCurrent: 0,
			Service:         ServiceBasicConnect,
			CID:             CIDSubscriberReadyStatus,
			CommandType:     CommandTypeQuery,
			Data:            []byte{},
			Response:        r.Response,
		},
	}
}

type SubscriberReadyStatusResponse struct {
	ReadyState uint32
}

func (r *SubscriberReadyStatusResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("subscriber ready status response data too short")
	}
	r.ReadyState = binary.LittleEndian.Uint32(data[0:4])
	return nil
}

// endregion

// region Open Logical Channel

type OpenLogicalChannelRequest struct {
	TransactionID uint32
	AppId         []byte
	SelectP2Arg   uint32
	Group         uint32
	Response      *OpenLogicalChannelResponse
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
		TransactionID: r.TransactionID,
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
	Channel       uint32 // Channel to close
	Group         uint32 // Channel group to close
	TransactionID uint32
	Response      *CloseLogicalChannelResponse
}

func (r *CloseLogicalChannelRequest) Message() *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(r.Channel))
	binary.Write(buf, binary.LittleEndian, r.Group)
	r.Response = new(CloseLogicalChannelResponse)
	return &Message{
		Type:          MessageTypeCommand,
		TransactionID: r.TransactionID,
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
	TransactionID   uint32
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
		TransactionID: r.TransactionID,
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
	Status   uint32
	Response []byte
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
	r.Response = data[12 : 12+n]
	return nil
}

// endregion
