package sgp22

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
