package sgp22

import (
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
)

func MarshalRequest(request bertlv.Marshaler) (*bertlv.TLV, error) {
	return request.MarshalBERTLV()
}

func TestEUICCConfiguredAddressesRequest(t *testing.T) {
	request, err := MarshalRequest(new(EuiccConfiguredAddressesRequest))
	assert.NoError(t, err)
	assert.NotNil(t, request)
	expected := []byte{0xBF, 0x3C, 0x00}
	assert.Equal(t, expected, request.Bytes())
}

func TestEUICCConfiguredAddressesResponse(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x3C, 0x17,
		0x81, 0x15, 0x74, 0x65, 0x73, 0x74, 0x72, 0x6F, 0x6F, 0x74, 0x73,
		0x6D, 0x64, 0x73, 0x2E, 0x67, 0x73, 0x6D, 0x61, 0x2E, 0x63, 0x6F, 0x6D,
	}))
	var response EuiccConfiguredAddressesResponse
	assert.NoError(t, response.UnmarshalBERTLV(&tlv))
	assert.Empty(t, response.DefaultSMDPAddress)
	assert.Equal(t, "testrootsmds.gsma.com", response.RootSMDSAddress)
}

func TestEUICCConfiguredAddressesResponse2(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x3C, 0x24,
		0x80, 0x0B, 0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x63, 0x6F, 0x6D,
		0x81, 0x15, 0x74, 0x65, 0x73, 0x74, 0x72, 0x6F, 0x6F, 0x74, 0x73, 0x6D, 0x64,
		0x73, 0x2E, 0x67, 0x73, 0x6D, 0x61, 0x2E, 0x63, 0x6F, 0x6D,
	}))
	var response EuiccConfiguredAddressesResponse
	assert.NoError(t, response.UnmarshalBERTLV(&tlv))
	assert.Equal(t, "example.com", response.DefaultSMDPAddress)
	assert.Equal(t, "testrootsmds.gsma.com", response.RootSMDSAddress)
}

func TestSetDefaultDPAddressRequest(t *testing.T) {
	request, err := MarshalRequest(&SetDefaultDPAddressRequest{
		DefaultDPAddress: "example.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, request)
	expected := []byte{
		0xBF, 0x3F,
		0x0D,
		0x80, 0x0B, 0x65, 0x78,
		0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x63, 0x6F, 0x6D,
	}
	assert.Equal(t, expected, request.Bytes())
}

func TestSetDefaultDPAddressResponse(t *testing.T) {
	var tlv bertlv.TLV
	assert.NoError(t, tlv.UnmarshalBinary([]byte{
		0xBF, 0x3F, 0x03, 0x80, 0x01, 0x00,
	}))
	var response SetDefaultDPAddressResponse
	assert.NoError(t, response.UnmarshalBERTLV(&tlv))
	assert.NoError(t, response.Valid())
}
