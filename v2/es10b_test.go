package sgp22

import (
	"bytes"
	"errors"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestListNotificationResponseErrorChoice(t *testing.T) {
	response := new(ListNotificationResponse)
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(40),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), []byte{0x7f}),
	)

	err := response.UnmarshalBERTLV(tlv)

	if !errors.Is(err, ErrUndefined) {
		t.Errorf("UnmarshalBERTLV() error = %v, want undefined", err)
	}
}

func TestListNotificationRequestOmitsEmptyFilter(t *testing.T) {
	request, err := new(ListNotificationRequest).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x28, 0x00}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestListNotificationRequestEncodesFilterBits(t *testing.T) {
	request, err := (&ListNotificationRequest{
		Filter: map[NotificationEvent]bool{
			NotificationEventInstall: true,
			NotificationEventDelete:  true,
		},
	}).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x28, 0x04, 0x81, 0x02, 0x04, 0x90}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestRetrieveNotificationsListRequestEncodesSearchCriteria(t *testing.T) {
	criteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(0), SequenceNumber(1))
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}

	request, err := (&RetrieveNotificationsListRequest{SearchCriteria: criteria}).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x2b, 0x05, 0xa0, 0x03, 0x80, 0x01, 0x01}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestRetrieveNotificationsListRequestOmitsSearchCriteria(t *testing.T) {
	request, err := new(RetrieveNotificationsListRequest).MarshalBERTLV()

	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}
	encoded, err := request.Bytes()
	if err != nil {
		t.Fatalf("request.Bytes() error = %v", err)
	}
	if want := []byte{0xbf, 0x2b, 0x00}; !bytes.Equal(encoded, want) {
		t.Errorf("request.Bytes() = % X, want % X", encoded, want)
	}
}

func TestPrepareDownloadRequestNeedConfirmationCodeUsesBooleanTag(t *testing.T) {
	request := &PrepareDownloadRequest{
		Signed2: bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(0),
			bertlv.NewValue(bertlv.Universal.Primitive(1), []byte{0xff}),
		),
	}

	got, err := request.NeedConfirmationCode()
	if err != nil {
		t.Fatalf("NeedConfirmationCode() error = %v", err)
	}
	if !got {
		t.Error("NeedConfirmationCode() = false, want true")
	}
}

func TestPrepareDownloadRequestNeedConfirmationCodeRejectsInvalidEncoding(t *testing.T) {
	request := &PrepareDownloadRequest{Signed2: bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(0),
		bertlv.NewValue(bertlv.Universal.Primitive(1), []byte{0xff, 0x00}),
	)}
	if _, err := request.NeedConfirmationCode(); err == nil {
		t.Error("NeedConfirmationCode() error = nil for invalid boolean encoding")
	}
}

func TestPrepareDownloadRequestEncodesHashCcAsOctetString(t *testing.T) {
	request := &PrepareDownloadRequest{
		TransactionID: []byte{0x01, 0x02},
		Signed2: bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(0),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), []byte{0x01, 0x02}),
			bertlv.NewValue(bertlv.Universal.Primitive(1), []byte{0xff}),
		),
		Signature2:       bertlv.NewValue(bertlv.Application.Primitive(55), []byte{0x03}),
		Certificate:      bertlv.NewChildren(bertlv.ContextSpecific.Constructed(3)),
		ConfirmationCode: []byte("1234"),
	}

	tlv, err := request.MarshalBERTLV()
	if err != nil {
		t.Fatalf("MarshalBERTLV() error = %v", err)
	}

	expected := []byte{
		0xbf, 0x21, 0x31,
		0xa0, 0x07, 0x80, 0x02, 0x01, 0x02, 0x01, 0x01, 0xff,
		0x5f, 0x37, 0x01, 0x03,
		0x04, 0x20,
	}
	hashedConfirmationCode, err := request.HashedConfirmationCode()
	if err != nil {
		t.Fatalf("HashedConfirmationCode() error = %v", err)
	}
	expected = append(expected, hashedConfirmationCode...)
	expected = append(expected, 0xa3, 0x00)
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, expected)
	}
}

func TestRetrieveNotificationsListResponseAllowsEmptyList(t *testing.T) {
	response := new(RetrieveNotificationsListResponse)
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(43),
		bertlv.NewChildren(bertlv.ContextSpecific.Constructed(0)),
	)

	if err := response.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if len(response.NotificationList) != 0 {
		t.Errorf("NotificationList = %#v, want empty", response.NotificationList)
	}
	if err := response.Valid(); err != nil {
		t.Errorf("Valid() error = %v", err)
	}
}

func TestRetrieveNotificationsListResponseErrorChoice(t *testing.T) {
	response := new(RetrieveNotificationsListResponse)
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(43),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), []byte{0x7f}),
	)

	err := response.UnmarshalBERTLV(tlv)

	if !errors.Is(err, ErrUndefined) {
		t.Errorf("UnmarshalBERTLV() error = %v, want undefined", err)
	}
}

func TestNotificationEventRejectsInvalidBitCount(t *testing.T) {
	var event NotificationEvent

	for _, input := range [][]byte{{0x04, 0x00}, {0x04, 0xc0}, {0x03, 0x08}} {
		if err := event.UnmarshalBinary(input); err == nil {
			t.Errorf("NotificationEvent.UnmarshalBinary(% X) error = nil", input)
		}
	}
}

func TestNotificationMetadataUsesUTF8StringAddressTag(t *testing.T) {
	metadata := new(NotificationMetadata)

	if err := metadata.UnmarshalBERTLV(notificationMetadataTLV()); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if metadata.SequenceNumber != 1 || metadata.ProfileManagementOperation != NotificationEventInstall || metadata.Address != "example.com" {
		t.Errorf("metadata = %#v, want sequence 1, install, example.com", metadata)
	}
}

func TestPendingNotificationUnmarshalProfileInstallationResult(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(55),
		bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(39),
			notificationMetadataTLV(),
			bertlv.NewChildren(bertlv.ContextSpecific.Constructed(2)),
		),
		bertlv.NewValue(bertlv.Application.Primitive(55), []byte{0x01}),
	)
	notification := new(PendingNotification)

	if err := notification.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if notification.PendingNotification != tlv {
		t.Errorf("PendingNotification = %p, want %p", notification.PendingNotification, tlv)
	}
	if got, want := notification.Notification.Address, "example.com"; got != want {
		t.Errorf("notification address = %q, want %q", got, want)
	}
}

func TestPendingNotificationUnmarshalOtherSignedNotification(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.Universal.Constructed(16),
		notificationMetadataTLV(),
		bertlv.NewValue(bertlv.Application.Primitive(55), []byte{0x01}),
		bertlv.NewChildren(bertlv.ContextSpecific.Constructed(2)),
		bertlv.NewChildren(bertlv.ContextSpecific.Constructed(3)),
	)
	notification := new(PendingNotification)

	if err := notification.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if notification.PendingNotification != tlv {
		t.Errorf("PendingNotification = %p, want %p", notification.PendingNotification, tlv)
	}
	if got, want := notification.Notification.Address, "example.com"; got != want {
		t.Errorf("notification address = %q, want %q", got, want)
	}
}

func notificationMetadataTLV() *bertlv.TLV {
	return bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(47),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), []byte{0x01}),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), []byte{0x04, 0x80}),
		bertlv.NewValue(bertlv.Universal.Primitive(12), []byte("example.com")),
	)
}

func TestAuthenticateServerRequestRejectsInvalidIMEI(t *testing.T) {
	request := &AuthenticateServerRequest{IMEI: []byte{0x12, 0x34, 0x56}}

	_, err := request.MarshalBERTLV()

	if err == nil {
		t.Error("MarshalBERTLV() error = nil for invalid IMEI")
	}
}
