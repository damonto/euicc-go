package sgp22

import "fmt"

type Header struct {
	ExecutionStatus *ExecutionStatus `json:"functionExecutionStatus,omitempty"`
	RequesterID     string           `json:"functionRequesterIdentifier,omitempty"`
	CallID          string           `json:"functionCallIdentifier,omitempty"`
}

func (h Header) Error() error {
	if h.ExecutionStatus.ExecutedSuccess() {
		return nil
	}
	return h.ExecutionStatus.StatusCodeData
}

// HeaderExecutionStatus returns the execution status of an ES9+ response,
// synthesising a failed status when the header is missing. An SM-DP+ that
// replies with an empty or invalid body leaves Header (and thus
// ExecutionStatus) nil; callers dereference the returned status on the failure
// path, so returning nil here would panic. Always return a non-nil status with
// a non-nil StatusCodeData instead.
func HeaderExecutionStatus(h *Header) *ExecutionStatus {
	if h == nil || h.ExecutionStatus == nil {
		return &ExecutionStatus{
			Status: "Failed",
			StatusCodeData: &StatusCodeData{
				Message: "SM-DP+ returned an empty or invalid response (missing functionExecutionStatus)",
			},
		}
	}
	return h.ExecutionStatus
}

type ExecutionStatus struct {
	Status         string          `json:"status,omitempty"`
	StatusCodeData *StatusCodeData `json:"statusCodeData,omitempty"`
}

func (s *ExecutionStatus) ExecutedSuccess() bool {
	return s != nil && s.Status == "Executed-Success"
}

func (s *ExecutionStatus) ExecutedWithWarning() bool {
	return s != nil && s.Status == "Executed-WithWarning"
}

func (s *ExecutionStatus) Failed() bool {
	return s != nil && s.Status == "Failed"
}

func (s *ExecutionStatus) Expired() bool {
	return s != nil && s.Status == "Expired"
}

type StatusCodeData struct {
	SubjectCode string `json:"subjectCode,omitempty"`
	ReasonCode  string `json:"reasonCode,omitempty"`
	Message     string `json:"message,omitempty"`
}

func (s StatusCodeData) Error() string {
	if len(s.Message) > 0 {
		return s.Message
	}
	for _, err := range rspErrors {
		if err.ReasonCode == s.ReasonCode && err.SubjectCode == s.SubjectCode {
			return err.Message
		}
	}
	return fmt.Sprintf("SubjectCode: %s, ReasonCode: %s", s.SubjectCode, s.ReasonCode)
}
