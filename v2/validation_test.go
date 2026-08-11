package sgp22

import "testing"

func TestLoadBoundProfilePackageRequestRejectsNilPackage(t *testing.T) {
	if _, err := new(LoadBoundProfilePackageRequest).MarshalBERTLV(); err == nil {
		t.Error("MarshalBERTLV() error = nil for missing bound profile package")
	}
}

func TestHTTPResponsesHandleMissingExecutionStatus(t *testing.T) {
	responses := map[string]HTTPResponse{
		"initiate authentication": new(ES9InitiateAuthenticationResponse),
		"bound profile package":   new(ES9BoundProfilePackageResponse),
		"authenticate client":     new(ES9AuthenticateClientResponse),
		"cancel session":          new(ES9CancelSessionResponse),
		"discovery":               new(ES11AuthenticateClientResponse),
	}
	for name, response := range responses {
		t.Run(name, func(t *testing.T) {
			status := response.FunctionExecutionStatus()
			if status == nil || !status.Failed() || status.StatusCodeData == nil {
				t.Errorf("FunctionExecutionStatus() = %#v, want failed status with details", status)
			}
		})
	}
}

func TestHeaderErrorHandlesMissingExecutionStatus(t *testing.T) {
	if err := (Header{}).Error(); err == nil {
		t.Error("Header.Error() = nil for missing execution status")
	}
}
