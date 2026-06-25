package driver

import (
	"encoding/hex"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
)

func TestTransmitterBuildsRecordedES10APDUs(t *testing.T) {
	iccid, err := sgp22.NewICCID("00000000000000000000")
	if err != nil {
		t.Fatalf("NewICCID() error = %v", err)
	}
	seqNumber := sgp22.SequenceNumber(1 << 30)
	searchCriteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(0), seqNumber)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}
	installEvent := sgp22.NotificationEventInstall
	enableEvent := sgp22.NotificationEventEnable
	disableEvent := sgp22.NotificationEventDisable
	deleteEvent := sgp22.NotificationEventDelete
	installCriteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(1), &installEvent)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}
	enableCriteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(1), &enableEvent)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}
	disableCriteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(1), &disableEvent)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}
	deleteCriteria, err := bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(1), &deleteEvent)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}

	tests := []struct {
		name    string
		request bertlv.Marshaler
		apdu    string
	}{
		{
			name:    "EuiccConfiguredAddresses",
			request: new(sgp22.EuiccConfiguredAddressesRequest),
			apdu:    "81E2910003BF3C00",
		},
		{
			name:    "SetDefaultDPAddress",
			request: &sgp22.SetDefaultDPAddressRequest{DefaultDPAddress: "example.com"},
			apdu:    "81E2910010BF3F0D800B6578616D706C652E636F6D",
		},
		{
			name:    "GetEuiccData",
			request: new(sgp22.GetEuiccDataRequest),
			apdu:    "81E2910006BF3E035C015A",
		},
		{
			name:    "GetEuiccChallenge",
			request: new(sgp22.GetEuiccChallengeRequest),
			apdu:    "81E2910003BF2E00",
		},
		{
			name:    "GetEuiccInfo1",
			request: &sgp22.GetEuiccInfoRequest{Version: sgp22.EuiccInfoVersion1},
			apdu:    "81E2910003BF2000",
		},
		{
			name:    "GetEuiccInfo2",
			request: &sgp22.GetEuiccInfoRequest{Version: sgp22.EuiccInfoVersion2},
			apdu:    "81E2910003BF2200",
		},
		{
			name:    "ListNotificationAll",
			request: new(sgp22.ListNotificationRequest),
			apdu:    "81E2910003BF2800",
		},
		{
			name: "ListNotificationInstall",
			request: &sgp22.ListNotificationRequest{
				Filter: map[sgp22.NotificationEvent]bool{sgp22.NotificationEventInstall: true},
			},
			apdu: "81E2910007BF280481020480",
		},
		{
			name: "ListNotificationEnable",
			request: &sgp22.ListNotificationRequest{
				Filter: map[sgp22.NotificationEvent]bool{sgp22.NotificationEventEnable: true},
			},
			apdu: "81E2910007BF280481020440",
		},
		{
			name: "ListNotificationDisable",
			request: &sgp22.ListNotificationRequest{
				Filter: map[sgp22.NotificationEvent]bool{sgp22.NotificationEventDisable: true},
			},
			apdu: "81E2910007BF280481020420",
		},
		{
			name: "ListNotificationDelete",
			request: &sgp22.ListNotificationRequest{
				Filter: map[sgp22.NotificationEvent]bool{sgp22.NotificationEventDelete: true},
			},
			apdu: "81E2910007BF280481020410",
		},
		{
			name:    "RetrieveNotificationsListAll",
			request: new(sgp22.RetrieveNotificationsListRequest),
			apdu:    "81E2910003BF2B00",
		},
		{
			name: "RetrieveNotificationsListInstall",
			request: &sgp22.RetrieveNotificationsListRequest{
				SearchCriteria: installCriteria,
			},
			apdu: "81E2910009BF2B06A00481020480",
		},
		{
			name: "RetrieveNotificationsListEnable",
			request: &sgp22.RetrieveNotificationsListRequest{
				SearchCriteria: enableCriteria,
			},
			apdu: "81E2910009BF2B06A00481020440",
		},
		{
			name: "RetrieveNotificationsListDisable",
			request: &sgp22.RetrieveNotificationsListRequest{
				SearchCriteria: disableCriteria,
			},
			apdu: "81E2910009BF2B06A00481020420",
		},
		{
			name: "RetrieveNotificationsListDelete",
			request: &sgp22.RetrieveNotificationsListRequest{
				SearchCriteria: deleteCriteria,
			},
			apdu: "81E2910009BF2B06A00481020410",
		},
		{
			name:    "ProfileInfoListAll",
			request: new(sgp22.ProfileInfoListRequest),
			apdu:    "81E2910003BF2D00",
		},
		{
			name: "EuiccMemoryReset",
			request: &sgp22.EuiccMemoryResetRequest{
				DeleteOperationalProfiles:     true,
				DeleteFieldLoadedTestProfiles: true,
				ResetDefaultSMDPAddress:       false,
			},
			apdu: "81E2910007BF3404820205C0",
		},
		{
			name: "SetNickname",
			request: &sgp22.SetNicknameRequest{
				ICCID:    iccid,
				Nickname: []byte("euicc-go-test"),
			},
			apdu: "81E291001EBF291B5A0A00000000000000000000900D65756963632D676F2D74657374",
		},
		{
			name: "EnableProfile",
			request: &sgp22.ProfileOperationRequest{
				Operation:  sgp22.EnableProfile,
				Identifier: bertlv.NewValue(bertlv.Application.Primitive(26), iccid),
				Refresh:    false,
			},
			apdu: "81E2910014BF3111A00C5A0A00000000000000000000810100",
		},
		{
			name: "DisableProfile",
			request: &sgp22.ProfileOperationRequest{
				Operation:  sgp22.DisableProfile,
				Identifier: bertlv.NewValue(bertlv.Application.Primitive(26), iccid),
				Refresh:    false,
			},
			apdu: "81E2910014BF3211A00C5A0A00000000000000000000810100",
		},
		{
			name: "DeleteProfile",
			request: &sgp22.ProfileOperationRequest{
				Operation:  sgp22.DeleteProfile,
				Identifier: bertlv.NewValue(bertlv.Application.Primitive(26), iccid),
			},
			apdu: "81E291000FBF330C5A0A00000000000000000000",
		},
		{
			name: "RetrieveNotificationsListBySequence",
			request: &sgp22.RetrieveNotificationsListRequest{
				SearchCriteria: searchCriteria,
			},
			apdu: "81E291000BBF2B08A006800440000000",
		},
		{
			name: "NotificationSent",
			request: &sgp22.NotificationSentRequest{
				SequenceNumber: seqNumber,
			},
			apdu: "81E2910009BF3006800440000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := &fakeSmartCardChannel{
				logicalChannel: 1,
				responses:      [][]byte{{0x90, 0x00}},
			}
			tx, err := NewTransmitter(discardLogger(), channel, []byte{0xA0}, 254)
			if err != nil {
				t.Fatalf("NewTransmitter() error = %v", err)
			}
			tlv, err := tt.request.MarshalBERTLV()
			if err != nil {
				t.Fatalf("MarshalBERTLV() error = %v", err)
			}
			if _, err := tx.TransmitRaw(tlv.Bytes()); err != nil {
				t.Fatalf("TransmitRaw() error = %v", err)
			}
			if got := hex.EncodeToString(channel.requests[0]); got != lowerHex(tt.apdu) {
				t.Fatalf("APDU = %s, want %s", got, lowerHex(tt.apdu))
			}
		})
	}
}

func lowerHex(s string) string {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(decoded)
}
