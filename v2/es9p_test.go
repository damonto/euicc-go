package sgp22

import (
	"bytes"
	"errors"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestES9AuthenticateClientRequestUnmarshalAuthenticateResponseError(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x09,
		0xA1, 0x07,
		0x80, 0x02, 0x01, 0x02,
		0x81, 0x01, byte(AuthenticateErrorCodeInvalidSignature),
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var request ES9AuthenticateClientRequest

	if err := request.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}

	err := request.Valid()
	var authenticateError *AuthenticateResponseError
	if !errors.As(err, &authenticateError) {
		t.Fatalf("Valid() error = %v, want *AuthenticateResponseError", err)
	}
	if want := (HexString{0x01, 0x02}); !bytes.Equal(authenticateError.TransactionID, want) {
		t.Errorf("TransactionID = % X, want % X", authenticateError.TransactionID, want)
	}
	if authenticateError.ErrorCode != AuthenticateErrorCodeInvalidSignature {
		t.Errorf("ErrorCode = %d, want %d", authenticateError.ErrorCode, AuthenticateErrorCodeInvalidSignature)
	}
}

func TestES9AuthenticateClientRequestUnmarshalAuthenticateResponseOk(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x02,
		0xA0, 0x00,
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var request ES9AuthenticateClientRequest

	if err := request.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if err := request.Valid(); err != nil {
		t.Errorf("Valid() error = %v", err)
	}
	if request.Response != &tlv {
		t.Errorf("Response = %p, want %p", request.Response, &tlv)
	}
}

func TestES9AuthenticateClientRequestUnmarshalMalformedAuthenticateResponseError(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x04,
		0xA1, 0x02,
		0x80, 0x00,
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var request ES9AuthenticateClientRequest

	if err := request.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if err := request.Valid(); !errors.Is(err, ErrUnexpectedTag) {
		t.Errorf("Valid() error = %v, want unexpected tag", err)
	}
}

func TestES9AuthenticateClientRequestUnmarshalUnexpectedTag(t *testing.T) {
	tlv := bertlv.NewChildren(bertlv.ContextSpecific.Constructed(55))
	var request ES9AuthenticateClientRequest

	if err := request.UnmarshalBERTLV(tlv); !errors.Is(err, ErrUnexpectedTag) {
		t.Errorf("UnmarshalBERTLV() error = %v, want unexpected tag", err)
	}
}
