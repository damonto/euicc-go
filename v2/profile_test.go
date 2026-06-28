package sgp22

import (
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileInfoUnmarshalAllowsMissingOptionalFields(t *testing.T) {
	tlv := bertlv.NewChildren(bertlv.Private.Constructed(3))
	profile := new(ProfileInfo)

	require.NoError(t, profile.UnmarshalBERTLV(tlv))

	assert.Nil(t, profile.ICCID)
	assert.Nil(t, profile.ISDPAID)
	assert.Equal(t, ProfileDisabled, profile.ProfileState)
	assert.Empty(t, profile.ProfileNickname)
	assert.Empty(t, profile.ServiceProviderName)
	assert.Empty(t, profile.ProfileName)
	assert.Nil(t, profile.Icon)
	assert.Equal(t, ProfileClassProvisioning, profile.ProfileClass)
	assert.Nil(t, profile.ProfileOwner.PLMN)
	assert.Nil(t, profile.NotificationConfigurationInfo)
	assert.Equal(t, ProfilePolicyRules{}, profile.ProfilePolicyRules)
}

func TestProfileInfoUnmarshalAuthenticateClientProfileMetadata(t *testing.T) {
	var tlv bertlv.TLV
	require.NoError(t, tlv.UnmarshalText([]byte("vyWBjVoKmFgyJCBCSCZpZJEGQ01MSU5LkgdDTUlfR0RTthowGIACBHCBEmNvbnN1bWVyLnJzcC53b3JsZLcdgANU9CGBCv////////////+CCv////////////+/djLiMOEiwSA6yVumdHCV8I0+WJoSqtB4vuLOqh4/PnGVvchLJYeB2OMK2wgAAAAAAAAAAQ==")))
	profile := new(ProfileInfo)

	require.NoError(t, profile.UnmarshalBERTLV(&tlv))

	assert.Equal(t, "89852342022484629646", profile.ICCID.String())
	assert.Equal(t, "CMI_GDS", profile.ProfileName)
	assert.Equal(t, "CMLINK", profile.ServiceProviderName)
	require.Len(t, profile.NotificationConfigurationInfo, 1)
	assert.Equal(t, []NotificationEvent{
		NotificationEventEnable,
		NotificationEventDisable,
		NotificationEventDelete,
	}, profile.NotificationConfigurationInfo[0].ProfileManagementOperations)
	assert.Equal(t, "consumer.rsp.world", profile.NotificationConfigurationInfo[0].Address)
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

	require.NoError(t, profile.UnmarshalBERTLV(tlv))

	assert.Equal(t, ProfileIconTypePNG, profile.IconType)
	assert.Equal(t, ProfilePolicyRules{
		DisablingNotAllowed: true,
		DeletionNotAllowed:  true,
	}, profile.ProfilePolicyRules)
	assert.NotNil(t, profile.SMDPProprietaryData)
	assert.NotNil(t, profile.ServiceSpecificData)
}

func TestProfileInfoUnmarshalOptionalProfileClassAllowsSignExtendedInt(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.Private.Constructed(3),
		bertlv.NewValue(TagProfileClass, []byte{0x00, 0x01}),
	)
	profile := new(ProfileInfo)

	require.NoError(t, profile.UnmarshalBERTLV(tlv))

	assert.Equal(t, ProfileClassProvisioning, profile.ProfileClass)
}

func TestOperatorIdShortPLMNDoesNotPanic(t *testing.T) {
	operator := OperatorId{PLMN: []byte{0x13}}

	assert.NotPanics(t, func() {
		assert.Empty(t, operator.MCC())
		assert.Empty(t, operator.MNC())
	})
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

	require.NoError(t, info.UnmarshalBERTLV(tlv))

	require.Len(t, *info, 1)
	assert.Equal(t, []NotificationEvent{NotificationEventInstall}, (*info)[0].ProfileManagementOperations)
	assert.Equal(t, "example.com", (*info)[0].Address)
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

	require.ErrorIs(t, info.UnmarshalBERTLV(tlv), ErrUnexpectedTag)
}
