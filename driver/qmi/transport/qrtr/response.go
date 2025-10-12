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
	if len(data) < 5 {
		return fmt.Errorf("data too short: got %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	binary.Read(reader, binary.LittleEndian, &r.MessageType)
	binary.Read(reader, binary.LittleEndian, &r.TransactionID)
	binary.Read(reader, binary.LittleEndian, &r.MessageID)
	binary.Read(reader, binary.LittleEndian, &r.MessageLength)
	if r.MessageLength > 0 {
		_, err := r.Value.ReadFrom(io.LimitReader(reader, int64(r.MessageLength)))
		return err
	}
	return nil
}

// endregion
