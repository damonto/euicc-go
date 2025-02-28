package sgp22

import (
	"net/url"

	"github.com/damonto/euicc-go/bertlv"
)

// region Section 5.6.1, ES9+.HandleNotification

type ES9InitiateAuthenticationRequest struct {
	Challenge []byte      `json:"euiccChallenge"`
	Info1     *bertlv.TLV `json:"euiccInfo1"`
	Address   string      `json:"smdpAddress"`
}

func (r *ES9InitiateAuthenticationRequest) URL(address *url.URL) *url.URL {
	return address.JoinPath("/gsma/rsp2/es9plus/initiateAuthentication")
}

func (r *ES9InitiateAuthenticationRequest) RemoteResponse() *ES9InitiateAuthenticationResponse {
	return new(ES9InitiateAuthenticationResponse)
}

type ES9InitiateAuthenticationResponse struct {
	Header        *Header     `json:"header"`
	TransactionID HexString   `json:"transactionId"`
	Signed1       *bertlv.TLV `json:"serverSigned1"`
	Signature1    *bertlv.TLV `json:"serverSignature1"`
	UsedIssuer    *bertlv.TLV `json:"euiccCiPKIdToBeUsed"`
	Certificate   *bertlv.TLV `json:"serverCertificate"`
}

func (r *ES9InitiateAuthenticationResponse) FunctionExecutionStatus() *ExecutionStatus {
	return r.Header.ExecutionStatus
}

func (r *ES9InitiateAuthenticationResponse) CardRequest() *AuthenticateServerRequest {
	return &AuthenticateServerRequest{
		TransactionID: r.TransactionID,
		Signed1:       r.Signed1,
		Signature1:    r.Signature1,
		UsedIssuer:    r.UsedIssuer,
		Certificate:   r.Certificate,
	}
}

// endregion

// region Section 5.6.2, ES9+.GetBoundProfilePackage

type ES9BoundProfilePackageRequest struct {
	TransactionID HexString   `json:"transactionId"`
	Response      *bertlv.TLV `json:"prepareDownloadResponse"`
}

func (r *ES9BoundProfilePackageRequest) URL(address *url.URL) *url.URL {
	return address.JoinPath("/gsma/rsp2/es9plus/getBoundProfilePackage")
}

func (r *ES9BoundProfilePackageRequest) RemoteResponse() *ES9BoundProfilePackageResponse {
	return new(ES9BoundProfilePackageResponse)
}

func (r *ES9BoundProfilePackageRequest) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 33) {
		return ErrUnexpectedTag
	}
	r.Response = tlv
	return nil
}

func (r *ES9BoundProfilePackageRequest) Valid() error {
	return nil
}

type ES9BoundProfilePackageResponse struct {
	Header              *Header     `json:"header"`
	TransactionID       HexString   `json:"transactionId"`
	BoundProfilePackage *bertlv.TLV `json:"boundProfilePackage"`
}

func (r *ES9BoundProfilePackageResponse) FunctionExecutionStatus() *ExecutionStatus {
	return r.Header.ExecutionStatus
}

func (r *ES9BoundProfilePackageResponse) CardRequest() *LoadBoundProfilePackageRequest {
	return &LoadBoundProfilePackageRequest{BoundProfilePackage: r.BoundProfilePackage}
}

// endregion

// region Section 5.6.3, ES9+.AuthenticateClient

type ES9AuthenticateClientRequest struct {
	TransactionID HexString   `json:"transactionId"`
	Response      *bertlv.TLV `json:"authenticateServerResponse"`
}

func (r *ES9AuthenticateClientRequest) URL(address *url.URL) *url.URL {
	return address.JoinPath("/gsma/rsp2/es9plus/authenticateClient")
}

func (r *ES9AuthenticateClientRequest) RemoteResponse() *ES9AuthenticateClientResponse {
	return new(ES9AuthenticateClientResponse)
}

func (r *ES9AuthenticateClientRequest) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 56) {
		return ErrUnexpectedTag
	}
	r.Response = tlv
	return nil
}

func (r *ES9AuthenticateClientRequest) Valid() error {
	return nil
}

type ES9AuthenticateClientResponse struct {
	Header          *Header     `json:"header"`
	TransactionID   HexString   `json:"transactionId"`
	ProfileMetadata *bertlv.TLV `json:"profileMetadata"`
	Signed2         *bertlv.TLV `json:"smdpSigned2"`
	Signature2      *bertlv.TLV `json:"smdpSignature2"`
	Certificate     *bertlv.TLV `json:"smdpCertificate"`
}

func (r *ES9AuthenticateClientResponse) FunctionExecutionStatus() *ExecutionStatus {
	return r.Header.ExecutionStatus
}

func (r *ES9AuthenticateClientResponse) CardRequest() *PrepareDownloadRequest {
	return &PrepareDownloadRequest{
		TransactionID:   r.TransactionID,
		ProfileMetadata: r.ProfileMetadata,
		Signed2:         r.Signed2,
		Signature2:      r.Signature2,
		Certificate:     r.Certificate,
	}
}

// endregion

// region Section 5.6.4, ES9+.HandleNotification

// ES9HandleNotificationRequest is used to handle a notification.
//
// See https://aka.pw/sgp22/v2.5#page=177 (Section 5.6.4, ES9+.HandleNotification)
type ES9HandleNotificationRequest struct {
	PendingNotification *bertlv.TLV `json:"pendingNotification"`
}

func (r *ES9HandleNotificationRequest) URL(address *url.URL) *url.URL {
	return address.JoinPath("/gsma/rsp2/es9plus/handleNotification")
}

func (r *ES9HandleNotificationRequest) RemoteResponse() *ES9HandleNotificationResponse {
	return new(ES9HandleNotificationResponse)
}

type ES9HandleNotificationResponse struct {
	Header *Header `json:"header"`
}

func (r *ES9HandleNotificationResponse) FunctionExecutionStatus() *ExecutionStatus {
	// HandleNotification does not return an ExecutionStatus.
	return &ExecutionStatus{Status: "Executed-Success"}
}

// endregion

// region Section 5.6.5, ES9+.CancelSession

// ES9CancelSessionRequest is used to cancel a session.
//
// See https://aka.pw/sgp22/v2.5#page=177 (Section 5.6.5, ES9+.CancelSession)
//
// See https://aka.pw/sgp22/v2.5#page=197 (Section 5.7.14, ES10b.CancelSession)
type ES9CancelSessionRequest struct {
	TransactionID HexString   `json:"transactionId"`
	Response      *bertlv.TLV `json:"cancelSessionResponse"`
}

func (r *ES9CancelSessionRequest) URL(address *url.URL) *url.URL {
	return address.JoinPath("/gsma/rsp2/es9plus/cancelSession")
}

func (r *ES9CancelSessionRequest) RemoteResponse() *ES9CancelSessionResponse {
	return new(ES9CancelSessionResponse)
}

func (r *ES9CancelSessionRequest) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 65) {
		return ErrUnexpectedTag
	}
	r.Response = tlv
	return nil
}

func (r *ES9CancelSessionRequest) Valid() error {
	return nil
}

type ES9CancelSessionResponse struct {
	Header *Header `json:"header"`
}

func (r *ES9CancelSessionResponse) FunctionExecutionStatus() *ExecutionStatus {
	return r.Header.ExecutionStatus
}

// endregion
