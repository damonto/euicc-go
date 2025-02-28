package sgp22

import (
	"net/url"
)

// region Section 5.8.2, ES11.AuthenticateClient

type ES11AuthenticateClientRequest struct{ *ES9AuthenticateClientRequest }

func (r *ES11AuthenticateClientRequest) RemoteResponse() *ES11AuthenticateClientResponse {
	return new(ES11AuthenticateClientResponse)
}

type ES11AuthenticateClientResponse struct {
	Header        *Header       `json:"header"`
	TransactionID HexString     `json:"transactionId"`
	EventEntries  []*EventEntry `json:"eventEntries"`
}

func (r *ES11AuthenticateClientResponse) FunctionExecutionStatus() *ExecutionStatus {
	return r.Header.ExecutionStatus
}

type EventEntry struct {
	EventID string `json:"eventId"`
	Address string `json:"rspServerAddress"`
}

func (e *EventEntry) URL() *url.URL {
	return &url.URL{Scheme: "https:", Host: e.Address}
}

// endregion
