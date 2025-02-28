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
