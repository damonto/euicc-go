package qrtr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

// Response represents a complete parsed QMI message
type Response struct {
	TransactionID uint16
	MessageID     core.MessageID
	MessageType   core.MessageType
	MessageLength uint16
	Value         core.TLVs
}

// UnmarshalBinary parses binary data into a Response
func (r *Response) UnmarshalBinary(data []byte) error {
	const headerLen = 7
	if len(data) < headerLen {
		return fmt.Errorf("data too short: got %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	var header Header
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("read QRTR QMI header: %w", err)
	}
	r.MessageType = header.MessageType
	r.TransactionID = header.TransactionID
	r.MessageID = header.MessageID
	r.MessageLength = header.MessageLength

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
