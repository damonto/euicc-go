package driver

import (
	"encoding/hex"
	"io"
	"log/slog"
	"strings"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
)

// GSMAISDRApplicationAID is the AID of the GSMA SGP.02 ISD-R application.
// See https://www.gsma.com/solutions-and-impact/technologies/esim/wp-content/uploads/2020/07/SGP.02-v4.2.pdf#page=26 (Section 2.2.3 Identification of Security Domains: AID and TAR)
var GSMAISDRApplicationAID = []byte{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x00, 0x01, 0x00}

type Transmitter interface {
	sgp22.Transmitter
	Close() error
}

type transmitter struct {
	card io.ReadWriteCloser
}

func NewTransmitter(channel apdu.SmartCardChannel, AID []byte, MSS int) (Transmitter, error) {
	t, err := apdu.NewTransmitter(channel, AID, MSS)
	if err != nil {
		return nil, err
	}
	return &transmitter{card: t}, nil
}

func (t *transmitter) Transmit(request bertlv.Marshaler, response bertlv.Unmarshaler) error {
	req, err := request.MarshalBERTLV()
	if err != nil {
		return err
	}
	bs, err := t.TransmitRaw(req.Bytes())
	if err != nil {
		return err
	}
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary(bs); err != nil {
		return err
	}
	return response.UnmarshalBERTLV(&tlv)
}

func (t *transmitter) TransmitRaw(command []byte) ([]byte, error) {
	slog.Debug("[APDU] sending", "command", strings.ToUpper(hex.EncodeToString(command)))
	if _, err := t.card.Write(command); err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(t.card)
	if err != nil {
		return nil, err
	}
	slog.Debug("[APDU] received", "response", strings.ToUpper(hex.EncodeToString(bs)))
	return bs, err
}

func (t *transmitter) Close() error {
	return t.card.Close()
}
