package sgp22

import (
	"crypto/sha256"
	"errors"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
)

// region Section 5.7.5, ES10b.PrepareDownload

// PrepareDownloadRequest is used to prepare a download.
// The confirmation code is required if the profile is not yet confirmed.
//
// See https://aka.pw/sgp22/v2.5#page=184 (Section 5.7.5, ES10b.PrepareDownload)
type PrepareDownloadRequest struct {
	TransactionID    []byte
	ProfileMetadata  *bertlv.TLV
	Signed2          *bertlv.TLV
	Signature2       *bertlv.TLV
	Certificate      *bertlv.TLV
	ConfirmationCode []byte
}

func (r *PrepareDownloadRequest) CardResponse() *ES9BoundProfilePackageRequest {
	return &ES9BoundProfilePackageRequest{TransactionID: r.TransactionID}
}

func (r *PrepareDownloadRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	hashedConfirmationCode := r.HashedConfirmationCode()
	if hashedConfirmationCode != nil && len(r.ConfirmationCode) == 0 {
		return nil, errors.New("confirmation code is required")
	}
	request := bertlv.NewChildrenIter(bertlv.ContextSpecific.Constructed(33), func(yield func(*bertlv.TLV) bool) {
		if !yield(r.Signed2) {
			return
		}
		if !yield(r.Signature2) {
			return
		}
		if hashedConfirmationCode != nil {
			if !yield(bertlv.NewValue(bertlv.Universal.Primitive(4), hashedConfirmationCode)) {
				return
			}
		}
		yield(r.Certificate)
	})
	return request, nil
}

func (r *PrepareDownloadRequest) HashedConfirmationCode() []byte {
	if !r.NeedConfirmationCode() {
		return nil
	}
	hashed := sha256.New()
	hashed.Write(r.ConfirmationCode)
	confirmationCode := hashed.Sum(nil)
	hashed.Reset()
	hashed.Write(confirmationCode)
	hashed.Write(r.TransactionID)
	return hashed.Sum(nil)
}

func (r *PrepareDownloadRequest) NeedConfirmationCode() bool {
	if r == nil || r.Signed2 == nil {
		return false
	}
	var confirmationCodeRequired bool
	_ = r.Signed2.First(bertlv.Universal.Primitive(1)).
		UnmarshalValue(primitive.UnmarshalBool(&confirmationCodeRequired))
	return confirmationCodeRequired
}

// endregion

// region Section 5.7.6, ES10b.LoadBoundProfilePackage

// LoadBoundProfilePackageRequest is used to load a bound profile package.
//
// See https://aka.pw/sgp22/v2.5#page=186 (Section 5.7.6, ES10b.LoadBoundProfilePackage)
//
// See https://aka.pw/sgp22/v2.5#page=35 (Section 2.5.6, ProfileInstallationResult)
type LoadBoundProfilePackageRequest struct {
	BoundProfilePackage *bertlv.TLV
}

func (r *LoadBoundProfilePackageRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	if !r.BoundProfilePackage.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 54) {
		return nil, errors.New("expected a BoundProfilePackage")
	}
	return r.BoundProfilePackage, nil
}

func (r *LoadBoundProfilePackageRequest) CardResponse() *LoadBoundProfilePackageResponse {
	return new(LoadBoundProfilePackageResponse)
}

type LoadBoundProfilePackageResponse struct {
	TransactionID []byte
	Notification  *NotificationMetadata
	FinalResult   *bertlv.TLV
}

func (r *LoadBoundProfilePackageResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 55) {
		return ErrUnexpectedTag
	}
	tlv = tlv.First(bertlv.ContextSpecific.Constructed(39))
	r.TransactionID = tlv.First(bertlv.ContextSpecific.Primitive(0)).Value
	r.FinalResult = tlv.First(bertlv.ContextSpecific.Constructed(2))
	r.Notification = new(NotificationMetadata)
	return r.Notification.UnmarshalBERTLV(tlv.First(bertlv.ContextSpecific.Constructed(47)))
}

func (r *LoadBoundProfilePackageResponse) ISDPAID() ISDPAID {
	if successResult := r.FinalResult.First(bertlv.ContextSpecific.Constructed(0)); successResult != nil {
		return successResult.First(bertlv.Application.Primitive(15)).Value
	}
	return nil
}

func (r *LoadBoundProfilePackageResponse) Valid() error {
	result := r.FinalResult.First(bertlv.ContextSpecific.Constructed(1))
	if result == nil {
		return nil
	}
	return &LoadBoundProfilePackageError{
		BPPCommandID: result.First(bertlv.ContextSpecific.Primitive(0)).Value[0],
		ErrorReason:  result.First(bertlv.ContextSpecific.Primitive(1)).Value[0],
	}
}

// endregion

// region Section 5.7.7, ES10b.GetEuiccChallenge

// GetEuiccChallengeRequest is used to get a challenge from the eUICC.
//
// See https://aka.pw/sgp22/v2.5#page=187 (Section 5.7.7, ES10b.GetEUICCChallenge)
type GetEuiccChallengeRequest struct{}

func (r *GetEuiccChallengeRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	return bertlv.NewChildren(bertlv.ContextSpecific.Constructed(46)), nil
}

func (r *GetEuiccChallengeRequest) CardResponse() *GetEuiccChallengeResponse {
	return new(GetEuiccChallengeResponse)
}

type GetEuiccChallengeResponse struct {
	Challenge []byte
}

func (r *GetEuiccChallengeResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 46) {
		return ErrUnexpectedTag
	}
	r.Challenge = tlv.At(0).Value
	return nil
}

func (r *GetEuiccChallengeResponse) Valid() error {
	return nil
}

// endregion

// region Section 5.7.8, ES10b.GetEuiccInfo

// GetEuiccInfoRequest is used to get eUICC information.
// The version specifies the format of the response.
// Version 1 is used for ES10b.EUICCInfo1 and version 2 is used for ES10b.EUICCInfo2.
//
// See https://aka.pw/sgp22/v2.5#page=187 (Section 5.7.8, ES10b.GetEUICCInfo)
type GetEuiccInfoRequest struct {
	Version int
}

func (r *GetEuiccInfoRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	switch r.Version {
	case 1:
		return bertlv.NewChildren(bertlv.ContextSpecific.Constructed(32)), nil
	case 2:
		return bertlv.NewChildren(bertlv.ContextSpecific.Constructed(34)), nil
	}
	return nil, errors.New("unsupported version")
}

func (r *GetEuiccInfoRequest) CardResponse() *GetEuiccInfoResponse {
	return &GetEuiccInfoResponse{Version: r.Version}
}

type GetEuiccInfoResponse struct {
	Version  int
	Response *bertlv.TLV
}

func (r *GetEuiccInfoResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !(tlv.Tag.ContextSpecific() && tlv.Tag.Constructed()) {
		return ErrUnexpectedTag
	}
	if value := tlv.Tag.Value(); value == 32 || value == 34 {
		r.Response = tlv
		return nil
	}
	return ErrUnexpectedTag
}

func (r *GetEuiccInfoResponse) Valid() error {
	return nil
}

// endregion

// region Section 5.7.9, ES10b.ListNotification

// ListNotificationRequest is used to list notifications.
//
// See https://aka.pw/sgp22/v2.5#page=191 (Section 5.7.9, ES10b.ListNotification)
type ListNotificationRequest struct {
	Filter map[NotificationEvent]bool
}

func (r *ListNotificationRequest) CardResponse() *ListNotificationResponse {
	return new(ListNotificationResponse)
}

func (r *ListNotificationRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	bits := []bool{
		r.Filter[NotificationEventInstall],
		r.Filter[NotificationEventEnable],
		r.Filter[NotificationEventDisable],
		r.Filter[NotificationEventDelete],
	}
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(40),
		mustMarshalValue(bertlv.MarshalValue(
			bertlv.ContextSpecific.Primitive(1),
			primitive.MarshalBitString(bits),
		)),
	)
	return request, nil
}

type ListNotificationResponse struct {
	NotificationList []*NotificationMetadata
}

func (r *ListNotificationResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	tlv = tlv.First(bertlv.ContextSpecific.Constructed(0))
	notifications := make([]*NotificationMetadata, 0, len(tlv.Children))
	var notification *NotificationMetadata
	for _, child := range tlv.Children {
		notification = new(NotificationMetadata)
		if err := notification.UnmarshalBERTLV(child); err != nil {
			return err
		}
		notifications = append(notifications, notification)
	}
	r.NotificationList = notifications
	return nil
}

func (r *ListNotificationResponse) Valid() error {
	return nil
}

// endregion

// region Section 5.7.10, ES10b.RetrieveNotificationsList

// RetrieveNotificationsListRequest is used to retrieve a list of notifications.
//
// See https://aka.pw/sgp22/v2.5#page=191 (Section 5.7.10, ES10b.RetrieveNotificationsList)
type RetrieveNotificationsListRequest struct {
	SearchCriteria *bertlv.TLV
}

func (r *RetrieveNotificationsListRequest) CardResponse() *RetrieveNotificationsListResponse {
	return new(RetrieveNotificationsListResponse)
}

func (r *RetrieveNotificationsListRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildrenIter(bertlv.ContextSpecific.Constructed(43), func(yield func(*bertlv.TLV) bool) {
		if r.SearchCriteria != nil {
			yield(bertlv.NewChildren(bertlv.ContextSpecific.Constructed(0), r.SearchCriteria))
		}
	})
	return request, nil
}

type RetrieveNotificationsListResponse struct {
	NotificationList []*PendingNotification
}

func (r *RetrieveNotificationsListResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 43) {
		return ErrUnexpectedTag
	}
	var notifications []*PendingNotification
	for _, child := range tlv.Children {
		notification := new(PendingNotification)
		if err := notification.UnmarshalBERTLV(child); err != nil {
			return err
		}
		notifications = append(notifications, notification)
	}
	r.NotificationList = notifications
	return nil
}

func (r *RetrieveNotificationsListResponse) Valid() error {
	if len(r.NotificationList) > 0 {
		return nil
	}
	return ErrUndefined
}

// endregion

// region Section 5.7.11, ES10b.RemoveNotificationFromList

// NotificationSentRequest is used to remove a notification from the list.
//
// See https://aka.pw/sgp22/v2.5#page=193 (Section 5.7.11, ES10b.RemoveNotificationFromList)
type NotificationSentRequest struct {
	SequenceNumber SequenceNumber
}

func (r *NotificationSentRequest) CardResponse() *NotificationSentResponse {
	return new(NotificationSentResponse)
}

func (r *NotificationSentRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(48),
		mustMarshalValue(bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(0), &r.SequenceNumber)),
	)
	return request, nil
}

type NotificationSentResponse struct {
	DeleteNotificationStatus int8
}

func (r *NotificationSentResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 48) {
		return ErrUnexpectedTag
	}
	return tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&r.DeleteNotificationStatus))
}

func (r *NotificationSentResponse) Valid() error {
	switch r.DeleteNotificationStatus {
	case 0:
		return nil
	case 1:
		return ErrNothingToDelete
	}
	return ErrUndefined
}

// endregion

// region Section 5.7.13, ES10b.AuthenticateServer

// AuthenticateServerRequest is used to authenticate the server.
//
// See https://aka.pw/sgp22/v2.5#page=195 (Section 5.7.13, ES10b.AuthenticateServer)
type AuthenticateServerRequest struct {
	TransactionID []byte
	Signed1       *bertlv.TLV
	Signature1    *bertlv.TLV
	UsedIssuer    *bertlv.TLV
	Certificate   *bertlv.TLV
	IMEI          IMEI
	MatchingID    []byte
}

func (r *AuthenticateServerRequest) CardResponse() *ES9AuthenticateClientRequest {
	return &ES9AuthenticateClientRequest{TransactionID: r.TransactionID}
}

func (r *AuthenticateServerRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	deviceInfo := bertlv.NewChildrenIter(bertlv.ContextSpecific.Constructed(1), func(yield func(*bertlv.TLV) bool) {
		if !yield(bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), r.IMEI[:4])) {
			return
		}
		if !yield(bertlv.NewChildren(bertlv.ContextSpecific.Constructed(1))) {
			return
		}
		yield(bertlv.NewValue(bertlv.ContextSpecific.Primitive(2), r.IMEI))
	})
	ctxParams1 := bertlv.NewChildrenIter(bertlv.ContextSpecific.Constructed(0), func(yield func(*bertlv.TLV) bool) {
		if !yield(bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), r.MatchingID)) {
			return
		}
		yield(deviceInfo)
	})
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(56),
		r.Signed1,
		r.Signature1,
		r.UsedIssuer,
		r.Certificate,
		ctxParams1,
	)
	return request, nil
}

// endregion

// region Section 5.7.14, ES10b.CancelSession

// CancelSessionRequest is used to cancel a session.
//
// See https://aka.pw/sgp22/v2.5#page=197 (Section 5.7.14, ES10b.CancelSession)

type CancelSessionReason byte

const (
	CancelSessionReasonEndUserRejection      CancelSessionReason = 0
	CancelSessionReasonPostponed             CancelSessionReason = 1
	CancelSessionReasonTimeout               CancelSessionReason = 2
	CancelSessionReasonPPRNotAllowed         CancelSessionReason = 3
	CancelSessionReasonMetadataMismatch      CancelSessionReason = 4
	CancelSessionReasonLoadBppExecutionError CancelSessionReason = 5
	CancelSessionReasonUndefined             CancelSessionReason = 127
)

type CancelSessionRequest struct {
	TransactionID []byte
	Reason        CancelSessionReason
}

func (r *CancelSessionRequest) CardResponse() *ES9CancelSessionRequest {
	return &ES9CancelSessionRequest{TransactionID: r.TransactionID}
}

func (r *CancelSessionRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(65),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), r.TransactionID),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), []byte{byte(r.Reason)}),
	)
	return request, nil
}

// endregion
