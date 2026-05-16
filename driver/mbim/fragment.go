package mbim

import (
	"encoding/binary"
	"errors"
	"fmt"
)

func isFragmentMessage(messageType MessageType) bool {
	return messageType == MessageTypeCommand ||
		messageType == MessageTypeCommandDone ||
		messageType == MessageTypeIndicateStatus
}

type fragmentedMessage struct {
	data         []byte
	maxFrameSize int
}

func (m fragmentedMessage) Frames() ([][]byte, error) {
	if len(m.data) <= m.maxFrameSize {
		return [][]byte{m.data}, nil
	}
	if len(m.data) < 20 {
		return nil, fmt.Errorf("MBIM message too short to fragment: %d", len(m.data))
	}
	messageType := MessageType(binary.LittleEndian.Uint32(m.data[0:4]))
	if !isFragmentMessage(messageType) {
		return nil, fmt.Errorf("MBIM message type %#x does not support fragments", messageType)
	}
	if binary.LittleEndian.Uint32(m.data[4:8]) != uint32(len(m.data)) {
		return nil, fmt.Errorf("MBIM message length mismatch: header=%d actual=%d", binary.LittleEndian.Uint32(m.data[4:8]), len(m.data))
	}
	if m.maxFrameSize <= 20 {
		return nil, fmt.Errorf("MBIM max fragment size %d is too small", m.maxFrameSize)
	}

	transactionID := binary.LittleEndian.Uint32(m.data[8:12])
	payload := m.data[20:]
	maxPayloadSize := m.maxFrameSize - 20
	total := (len(payload) + maxPayloadSize - 1) / maxPayloadSize
	if total == 0 {
		total = 1
	}

	frames := make([][]byte, 0, total)
	for current, offset := 0, 0; offset < len(payload); current++ {
		end := min(offset+maxPayloadSize, len(payload))
		fragmentLen := 20 + end - offset
		fragment := make([]byte, fragmentLen)
		binary.LittleEndian.PutUint32(fragment[0:4], uint32(messageType))
		binary.LittleEndian.PutUint32(fragment[4:8], uint32(fragmentLen))
		binary.LittleEndian.PutUint32(fragment[8:12], transactionID)
		binary.LittleEndian.PutUint32(fragment[12:16], uint32(total))
		binary.LittleEndian.PutUint32(fragment[16:20], uint32(current))
		copy(fragment[20:], payload[offset:end])
		frames = append(frames, fragment)
		offset = end
	}
	return frames, nil
}

type fragmentCollector struct {
	messageType   MessageType
	transactionID uint32
	total         uint32
	current       uint32
	payload       []byte
}

func newFragmentCollector(frame []byte) (*fragmentCollector, error) {
	var f fragment
	if err := f.UnmarshalBinary(frame); err != nil {
		return nil, err
	}
	if f.current != 0 {
		return nil, fmt.Errorf("expecting MBIM fragment 0/%d, got %d/%d", f.total, f.current, f.total)
	}
	return &fragmentCollector{
		messageType:   f.messageType,
		transactionID: f.transactionID,
		total:         f.total,
		current:       f.current,
		payload:       append([]byte(nil), f.payload...),
	}, nil
}

func (c *fragmentCollector) add(frame []byte) error {
	var f fragment
	if err := f.UnmarshalBinary(frame); err != nil {
		return err
	}
	if f.messageType != c.messageType {
		return fmt.Errorf("MBIM fragment message type mismatch: got %#x want %#x", f.messageType, c.messageType)
	}
	if f.transactionID != c.transactionID {
		return fmt.Errorf("MBIM fragment transaction ID mismatch: got %d want %d", f.transactionID, c.transactionID)
	}
	if f.current != c.current+1 {
		return fmt.Errorf("expecting MBIM fragment %d/%d, got %d/%d", c.current+1, c.total, f.current, f.total)
	}
	if f.total != c.total {
		return fmt.Errorf("MBIM fragment total mismatch: got %d want %d", f.total, c.total)
	}
	c.current = f.current
	c.payload = append(c.payload, f.payload...)
	return nil
}

func (c *fragmentCollector) complete() bool {
	return c.current == c.total-1
}

func (c *fragmentCollector) MarshalBinary() ([]byte, error) {
	if !c.complete() {
		return nil, fmt.Errorf("incomplete MBIM fragments: got %d/%d", c.current, c.total)
	}
	length := uint32(20 + len(c.payload))
	buf := make([]byte, length)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(c.messageType))
	binary.LittleEndian.PutUint32(buf[4:8], length)
	binary.LittleEndian.PutUint32(buf[8:12], c.transactionID)
	binary.LittleEndian.PutUint32(buf[12:16], 1)
	binary.LittleEndian.PutUint32(buf[16:20], 0)
	copy(buf[20:], c.payload)
	return buf, nil
}

type fragment struct {
	messageType   MessageType
	transactionID uint32
	total         uint32
	current       uint32
	payload       []byte
}

func (f *fragment) UnmarshalBinary(frame []byte) error {
	if len(frame) < 20 {
		return fmt.Errorf("MBIM fragment too short: %d", len(frame))
	}
	messageType := MessageType(binary.LittleEndian.Uint32(frame[0:4]))
	if !isFragmentMessage(messageType) {
		return fmt.Errorf("MBIM message type %#x does not support fragments", messageType)
	}
	messageLength := binary.LittleEndian.Uint32(frame[4:8])
	if messageLength != uint32(len(frame)) {
		return fmt.Errorf("MBIM fragment length mismatch: header=%d actual=%d", messageLength, len(frame))
	}
	transactionID := binary.LittleEndian.Uint32(frame[8:12])
	total := binary.LittleEndian.Uint32(frame[12:16])
	current := binary.LittleEndian.Uint32(frame[16:20])
	if total == 0 {
		return errors.New("invalid MBIM fragment total 0")
	}
	if current >= total {
		return fmt.Errorf("invalid MBIM fragment %d/%d", current, total)
	}
	*f = fragment{
		messageType:   messageType,
		transactionID: transactionID,
		total:         total,
		current:       current,
		payload:       frame[20:],
	}
	return nil
}
