package sgp22

import (
	"errors"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
)

func TestES9AuthenticateClientRequestUnmarshalAuthenticateResponseError(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x09,
		0xA1, 0x07,
		0x80, 0x02, 0x01, 0x02,
		0x81, 0x01, byte(AuthenticateErrorCodeInvalidSignature),
	}))
	var request ES9AuthenticateClientRequest

	assert.NoError(t, request.UnmarshalBERTLV(&tlv))

	err := request.Valid()
	var authenticateError *AuthenticateResponseError
	if assert.True(t, errors.As(err, &authenticateError)) {
		assert.Equal(t, HexString{0x01, 0x02}, authenticateError.TransactionID)
		assert.Equal(t, AuthenticateErrorCodeInvalidSignature, authenticateError.ErrorCode)
	}
}

func TestES9AuthenticateClientRequestUnmarshalAuthenticateResponseOk(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x02,
		0xA0, 0x00,
	}))
	var request ES9AuthenticateClientRequest

	assert.NoError(t, request.UnmarshalBERTLV(&tlv))
	assert.NoError(t, request.Valid())
	assert.Same(t, &tlv, request.Response)
}

func TestES9AuthenticateClientRequestUnmarshalMalformedAuthenticateResponseError(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x38, 0x04,
		0xA1, 0x02,
		0x80, 0x00,
	}))
	var request ES9AuthenticateClientRequest

	assert.NoError(t, request.UnmarshalBERTLV(&tlv))
	assert.ErrorIs(t, request.Valid(), ErrUnexpectedTag)
}

func TestES9AuthenticateClientRequestUnmarshalUnexpectedTag(t *testing.T) {
	tlv := bertlv.NewChildren(bertlv.ContextSpecific.Constructed(55))
	var request ES9AuthenticateClientRequest

	assert.ErrorIs(t, request.UnmarshalBERTLV(tlv), ErrUnexpectedTag)
}
