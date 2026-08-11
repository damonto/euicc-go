package sgp22

import (
	"errors"
	"reflect"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestProfileInfoUnmarshalAllowsMissingOptionalFields(t *testing.T) {
	tlv := bertlv.NewChildren(bertlv.Private.Constructed(3))
	profile := new(ProfileInfo)

	if err := profile.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if profile.ICCID != nil || profile.ISDPAID != nil || profile.Icon != nil || profile.ProfileOwner.PLMN != nil || profile.NotificationConfigurationInfo != nil {
		t.Errorf("optional slice fields were populated: %#v", profile)
	}
	if profile.ProfileState != ProfileDisabled || profile.ProfileClass != ProfileClassProvisioning {
		t.Errorf("profile defaults = state:%v class:%v", profile.ProfileState, profile.ProfileClass)
	}
	if profile.ProfileNickname != "" || profile.ServiceProviderName != "" || profile.ProfileName != "" {
		t.Errorf("optional names were populated: %#v", profile)
	}
	if profile.ProfilePolicyRules != (ProfilePolicyRules{}) {
		t.Errorf("ProfilePolicyRules = %#v, want zero value", profile.ProfilePolicyRules)
	}
}

func TestProfileInfoUnmarshalAuthenticateClientProfileMetadata(t *testing.T) {
	var tlv bertlv.TLV
	if err := tlv.UnmarshalText([]byte("vyWBjVoKmFgyJCBCSCZpZJEGQ01MSU5LkgdDTUlfR0RTthowGIACBHCBEmNvbnN1bWVyLnJzcC53b3JsZLcdgANU9CGBCv////////////+CCv////////////+/djLiMOEiwSA6yVumdHCV8I0+WJoSqtB4vuLOqh4/PnGVvchLJYeB2OMK2wgAAAAAAAAAAQ==")); err != nil {
		t.Fatalf("TLV.UnmarshalText() error = %v", err)
	}
	profile := new(ProfileInfo)

	if err := profile.UnmarshalBERTLV(&tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if got, want := profile.ICCID.String(), "89852342022484629646"; got != want {
		t.Errorf("ICCID = %q, want %q", got, want)
	}
	if profile.ProfileName != "CMI_GDS" || profile.ServiceProviderName != "CMLINK" {
		t.Errorf("profile names = %q, %q", profile.ProfileName, profile.ServiceProviderName)
	}
	if len(profile.NotificationConfigurationInfo) != 1 {
		t.Fatalf("notification configuration count = %d, want 1", len(profile.NotificationConfigurationInfo))
	}
	wantEvents := []NotificationEvent{
		NotificationEventEnable,
		NotificationEventDisable,
		NotificationEventDelete,
	}
	if got := profile.NotificationConfigurationInfo[0].ProfileManagementOperations; !reflect.DeepEqual(got, wantEvents) {
		t.Errorf("profile operations = %v, want %v", got, wantEvents)
	}
	if got, want := profile.NotificationConfigurationInfo[0].Address, "consumer.rsp.world"; got != want {
		t.Errorf("notification address = %q, want %q", got, want)
	}
}

func TestProfileInfoUnmarshalAdditionalOptionalFields(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.Private.Constructed(3),
		bertlv.NewValue(TagProfileIconType, []byte{0x01}),
		bertlv.NewValue(TagProfilePolicyRules, []byte{0x05, 0x60}),
		bertlv.NewChildren(TagSMDPProprietaryData),
		bertlv.NewChildren(TagServiceSpecificData),
	)
	profile := new(ProfileInfo)

	if err := profile.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	wantRules := ProfilePolicyRules{
		DisablingNotAllowed: true,
		DeletionNotAllowed:  true,
	}
	if profile.IconType != ProfileIconTypePNG || profile.ProfilePolicyRules != wantRules {
		t.Errorf("optional fields = icon:%v rules:%#v", profile.IconType, profile.ProfilePolicyRules)
	}
	if profile.SMDPProprietaryData == nil || profile.ServiceSpecificData == nil {
		t.Error("optional proprietary data fields are nil")
	}
}

func TestProfileInfoUnmarshalOptionalProfileClassAllowsSignExtendedInt(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.Private.Constructed(3),
		bertlv.NewValue(TagProfileClass, []byte{0x00, 0x01}),
	)
	profile := new(ProfileInfo)

	if err := profile.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if profile.ProfileClass != ProfileClassProvisioning {
		t.Errorf("ProfileClass = %v, want provisioning", profile.ProfileClass)
	}
}

func TestOperatorIdShortPLMNDoesNotPanic(t *testing.T) {
	operator := OperatorId{PLMN: []byte{0x13}}

	if got := operator.MCC(); got != "" {
		t.Errorf("MCC() = %q, want empty", got)
	}
	if got := operator.MNC(); got != "" {
		t.Errorf("MNC() = %q, want empty", got)
	}
}

func TestNotificationConfigurationInfoUnmarshal(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(22),
		bertlv.NewChildren(
			bertlv.Universal.Constructed(16),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), []byte{0x04, 0x80}),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), []byte("example.com")),
		),
	)
	info := new(NotificationConfigurationInfo)

	if err := info.UnmarshalBERTLV(tlv); err != nil {
		t.Fatalf("UnmarshalBERTLV() error = %v", err)
	}
	if len(*info) != 1 {
		t.Fatalf("notification configuration count = %d, want 1", len(*info))
	}
	if got, want := (*info)[0].ProfileManagementOperations, []NotificationEvent{NotificationEventInstall}; !reflect.DeepEqual(got, want) {
		t.Errorf("profile operations = %v, want %v", got, want)
	}
	if got, want := (*info)[0].Address, "example.com"; got != want {
		t.Errorf("address = %q, want %q", got, want)
	}
}

func TestNotificationConfigurationInfoUnmarshalMissingFieldDoesNotPanic(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(22),
		bertlv.NewChildren(
			bertlv.Universal.Constructed(16),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), []byte{0x04, 0x80}),
		),
	)
	info := new(NotificationConfigurationInfo)

	if err := info.UnmarshalBERTLV(tlv); !errors.Is(err, ErrUnexpectedTag) {
		t.Errorf("UnmarshalBERTLV() error = %v, want unexpected tag", err)
	}
}
