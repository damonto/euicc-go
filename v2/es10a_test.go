package sgp22

import (
	"bytes"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func marshalRequest(request bertlv.Marshaler) (*bertlv.TLV, error) {
	return request.MarshalBERTLV()
}

func TestEUICCConfiguredAddressesRequest(t *testing.T) {
	request, err := marshalRequest(new(EuiccConfiguredAddressesRequest))
	if err != nil {
		t.Fatalf("marshalRequest() error = %v", err)
	}
	if request == nil {
		t.Fatal("marshalRequest() = nil")
	}
	expected := []byte{0xBF, 0x3C, 0x00}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, expected)
	}
}

func TestEUICCConfiguredAddressesResponse(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x3C, 0x17,
		0x81, 0x15, 0x74, 0x65, 0x73, 0x74, 0x72, 0x6F, 0x6F, 0x74, 0x73,
		0x6D, 0x64, 0x73, 0x2E, 0x67, 0x73, 0x6D, 0x61, 0x2E, 0x63, 0x6F, 0x6D,
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var response EuiccConfiguredAddressesResponse
	if err := response.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if response.DefaultSMDPAddress != "" {
		t.Errorf("DefaultSMDPAddress = %q, want empty", response.DefaultSMDPAddress)
	}
	if got, want := response.RootSMDSAddress, "testrootsmds.gsma.com"; got != want {
		t.Errorf("RootSMDSAddress = %q, want %q", got, want)
	}
}

func TestEUICCConfiguredAddressesResponse2(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x3C, 0x24,
		0x80, 0x0B, 0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x63, 0x6F, 0x6D,
		0x81, 0x15, 0x74, 0x65, 0x73, 0x74, 0x72, 0x6F, 0x6F, 0x74, 0x73, 0x6D, 0x64,
		0x73, 0x2E, 0x67, 0x73, 0x6D, 0x61, 0x2E, 0x63, 0x6F, 0x6D,
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var response EuiccConfiguredAddressesResponse
	if err := response.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if got, want := response.DefaultSMDPAddress, "example.com"; got != want {
		t.Errorf("DefaultSMDPAddress = %q, want %q", got, want)
	}
	if got, want := response.RootSMDSAddress, "testrootsmds.gsma.com"; got != want {
		t.Errorf("RootSMDSAddress = %q, want %q", got, want)
	}
}

func TestSetDefaultDPAddressRequest(t *testing.T) {
	request, err := marshalRequest(&SetDefaultDPAddressRequest{
		DefaultDPAddress: "example.com",
	})
	if err != nil {
		t.Fatalf("marshalRequest() error = %v", err)
	}
	if request == nil {
		t.Fatal("marshalRequest() = nil")
	}
	expected := []byte{
		0xBF, 0x3F,
		0x0D,
		0x80, 0x0B, 0x65, 0x78,
		0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x63, 0x6F, 0x6D,
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, expected)
	}
}

func TestSetDefaultDPAddressResponse(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary([]byte{
		0xBF, 0x3F, 0x03, 0x80, 0x01, 0x00,
	}); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	var response SetDefaultDPAddressResponse
	if err := response.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if err := response.Valid(); err != nil {
		t.Errorf("Valid() error = %v", err)
	}
}
