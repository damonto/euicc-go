package qmi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

// Response represents a complete parsed QMI message
type Response struct {
	QMUXHeader
	TransactionID uint16
	MessageID     core.MessageID
	MessageType   core.MessageType
	MessageLength uint16
	Value         core.TLVs
}

// UnmarshalBinary parses binary data into a Response
func (r *Response) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data too short: got %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.LittleEndian, &r.QMUXHeader); err != nil {
		return fmt.Errorf("read QMUX header: %w", err)
	}
	if r.QMUXHeader.IfType != core.QMUXHeaderIfType {
		return fmt.Errorf("unexpected QMUX marker 0x%02X", r.QMUXHeader.IfType)
	}
	if got, want := len(data), int(r.QMUXHeader.Length)+1; got != want {
		return fmt.Errorf("QMUX length mismatch: got %d bytes, header declares %d", got, want)
	}
	switch r.QMUXHeader.ServiceType {
	case core.QMIServiceControl:
		var header Header[uint8]
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			return fmt.Errorf("read control QMI header: %w", err)
		}
		r.MessageType = header.MessageType
		r.TransactionID = uint16(header.TransactionID)
		r.MessageID = header.MessageID
		r.MessageLength = header.MessageLength
	default:
		var header Header[uint16]
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			return fmt.Errorf("read service QMI header: %w", err)
		}
		r.MessageType = header.MessageType
		r.TransactionID = header.TransactionID
		r.MessageID = header.MessageID
		r.MessageLength = header.MessageLength
	}

	if got, want := reader.Len(), int(r.MessageLength); got != want {
		return fmt.Errorf("QMI TLV length mismatch: got %d bytes, header declares %d", got, want)
	}
	if r.MessageLength > 0 {
		n, err := r.Value.ReadFrom(io.LimitReader(reader, int64(r.MessageLength)))
		if err != nil {
			return err
		}
		if n != int64(r.MessageLength) {
			return fmt.Errorf("QMI TLV length mismatch: parsed %d bytes, header declares %d", n, r.MessageLength)
		}
	}
	return nil
}

// endregion
