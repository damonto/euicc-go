package sgp22

import (
	"bytes"
	"errors"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestProfileOperationResponseValidEnableProfileStateError(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: EnableProfile,
		Result:    ProfileOperationResultProfileNotInDisabledState,
	}

	err := response.Valid()
	var operationError *ProfileOperationError
	if !errors.As(err, &operationError) {
		t.Fatalf("Valid() error = %v, want *ProfileOperationError", err)
	}
	if operationError.Operation != EnableProfile || operationError.Result != ProfileOperationResultProfileNotInDisabledState {
		t.Errorf("operation error = {%v, %v}, want {%v, %v}", operationError.Operation, operationError.Result, EnableProfile, ProfileOperationResultProfileNotInDisabledState)
	}
	if got, want := operationError.Error(), "enableProfile,profileNotInDisabledState"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestProfileOperationResponseValidDisableProfileStateError(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DisableProfile,
		Result:    ProfileOperationResultProfileNotInEnabledState,
	}

	err := response.Valid()
	var operationError *ProfileOperationError
	if !errors.As(err, &operationError) {
		t.Fatalf("Valid() error = %v, want *ProfileOperationError", err)
	}
	if operationError.Operation != DisableProfile || operationError.Result != ProfileOperationResultProfileNotInEnabledState {
		t.Errorf("operation error = {%v, %v}, want {%v, %v}", operationError.Operation, operationError.Result, DisableProfile, ProfileOperationResultProfileNotInEnabledState)
	}
	if got, want := operationError.Error(), "disableProfile,profileNotInEnabledState"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestProfileOperationResponseValidCATBusy(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DisableProfile,
		Result:    ProfileOperationResultCATBusy,
	}

	err := response.Valid()

	if !errors.Is(err, ErrCatBusy) {
		t.Errorf("Valid() error = %v, want CAT busy", err)
	}
}

func TestProfileOperationResponseValidRejectsInvalidResultForOperation(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DeleteProfile,
		Result:    ProfileOperationResultCATBusy,
	}

	if err := response.Valid(); !errors.Is(err, ErrUndefined) {
		t.Errorf("Valid() error = %v, want undefined", err)
	}
}

func TestProfileOperationRequestWrapsIdentifierChoice(t *testing.T) {
	request, err := (&ProfileOperationRequest{
		Operation:  EnableProfile,
		Identifier: bertlv.NewValue(bertlv.Application.Primitive(15), []byte{0x01}),
		Refresh:    true,
	}).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x31, 0x08, 0xa0, 0x03, 0x4f, 0x01, 0x01, 0x81, 0x01, 0xff}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestEuiccMemoryResetRequestUsesContextSpecificResetOptions(t *testing.T) {
	request, err := (&EuiccMemoryResetRequest{
		DeleteOperationalProfiles:     true,
		DeleteFieldLoadedTestProfiles: true,
		ResetDefaultSMDPAddress:       false,
	}).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x34, 0x04, 0x82, 0x02, 0x05, 0xc0}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestGetEuiccDataResponseUnmarshal(t *testing.T) {
	eid := []byte{
		0x89, 0x10, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12,
		0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78,
	}
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(62),
		bertlv.NewValue(bertlv.Application.Primitive(26), eid),
	)
	response := new(GetEuiccDataResponse)

	if err := response.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if !bytes.Equal(response.EID, eid) {
		t.Errorf("EID = % X, want % X", response.EID, eid)
	}
}

func TestGetEuiccDataResponseRejectsUnexpectedTag(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(0),
		bertlv.NewValue(bertlv.Application.Primitive(26), []byte{0x89, 0x10}),
	)
	response := new(GetEuiccDataResponse)

	if err := response.UnmarshalBERTLV(tlv); !errors.Is(err, ErrUnexpectedTag) {
		t.Errorf("UnmarshalBERTLV() error = %v, want unexpected tag", err)
	}
}
