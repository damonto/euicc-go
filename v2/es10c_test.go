package sgp22

import (
	"errors"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileOperationResponseValidEnableProfileStateError(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: EnableProfile,
		Result:    ProfileOperationResultProfileNotInDisabledState,
	}

	err := response.Valid()
	var operationError *ProfileOperationError
	if assert.True(t, errors.As(err, &operationError)) {
		assert.Equal(t, EnableProfile, operationError.Operation)
		assert.Equal(t, ProfileOperationResultProfileNotInDisabledState, operationError.Result)
		assert.Equal(t, "enableProfile,profileNotInDisabledState", operationError.Error())
	}
}

func TestProfileOperationResponseValidDisableProfileStateError(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DisableProfile,
		Result:    ProfileOperationResultProfileNotInEnabledState,
	}

	err := response.Valid()
	var operationError *ProfileOperationError
	if assert.True(t, errors.As(err, &operationError)) {
		assert.Equal(t, DisableProfile, operationError.Operation)
		assert.Equal(t, ProfileOperationResultProfileNotInEnabledState, operationError.Result)
		assert.Equal(t, "disableProfile,profileNotInEnabledState", operationError.Error())
	}
}

func TestProfileOperationResponseValidCATBusy(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DisableProfile,
		Result:    ProfileOperationResultCATBusy,
	}

	err := response.Valid()

	assert.ErrorIs(t, err, ErrCatBusy)
}

func TestProfileOperationResponseValidRejectsInvalidResultForOperation(t *testing.T) {
	response := ProfileOperationResponse{
		Operation: DeleteProfile,
		Result:    ProfileOperationResultCATBusy,
	}

	assert.ErrorIs(t, response.Valid(), ErrUndefined)
}

func TestProfileOperationRequestWrapsIdentifierChoice(t *testing.T) {
	request, err := (&ProfileOperationRequest{
		Operation:  EnableProfile,
		Identifier: bertlv.NewValue(bertlv.Application.Primitive(15), []byte{0x01}),
		Refresh:    true,
	}).MarshalBERTLV()

	require.NoError(t, err)
	assert.Equal(t, []byte{0xbf, 0x31, 0x08, 0xa0, 0x03, 0x4f, 0x01, 0x01, 0x81, 0x01, 0xff}, request.Bytes())
}

func TestEuiccMemoryResetRequestUsesContextSpecificResetOptions(t *testing.T) {
	request, err := (&EuiccMemoryResetRequest{
		DeleteOperationalProfiles:     true,
		DeleteFieldLoadedTestProfiles: true,
		ResetDefaultSMDPAddress:       false,
	}).MarshalBERTLV()

	require.NoError(t, err)
	assert.Equal(t, []byte{0xbf, 0x34, 0x04, 0x82, 0x02, 0x05, 0xc0}, request.Bytes())
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

	require.NoError(t, response.UnmarshalBERTLV(tlv))

	assert.Equal(t, eid, response.EID)
}

func TestGetEuiccDataResponseRejectsUnexpectedTag(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(0),
		bertlv.NewValue(bertlv.Application.Primitive(26), []byte{0x89, 0x10}),
	)
	response := new(GetEuiccDataResponse)

	assert.ErrorIs(t, response.UnmarshalBERTLV(tlv), ErrUnexpectedTag)
}
