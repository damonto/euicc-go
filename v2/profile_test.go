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
}

func TestProfileInfoUnmarshalOptionalProfileClassUsesPrimitiveInt(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.Private.Constructed(3),
		bertlv.NewValue(TagProfileClass, []byte{0x00, 0x01}),
	)
	profile := new(ProfileInfo)

	err := profile.UnmarshalBERTLV(tlv)

	assert.Error(t, err)
}

func TestOperatorIdUnmarshalAllowsMissingPLMN(t *testing.T) {
	tlv := bertlv.NewChildren(bertlv.ContextSpecific.Constructed(23))
	operator := new(OperatorId)

	require.NoError(t, operator.UnmarshalBERTLV(tlv))

	assert.Empty(t, operator.MCC())
	assert.Empty(t, operator.MNC())
}

func TestOperatorIdShortPLMNDoesNotPanic(t *testing.T) {
	operator := OperatorId{PLMN: []byte{0x13}}

	assert.NotPanics(t, func() {
		assert.Empty(t, operator.MCC())
		assert.Empty(t, operator.MNC())
	})
}

func TestNotificationConfigurationInfoUnmarshalAllowsMissingOptionalFields(t *testing.T) {
	tlv := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(22),
		bertlv.NewChildren(bertlv.Universal.Constructed(16)),
	)
	info := new(NotificationConfigurationInfo)

	require.NoError(t, info.UnmarshalBERTLV(tlv))

	require.Len(t, *info, 1)
	assert.Equal(t, NotificationEventInstall, (*info)[0].ProfileManagementOperation)
	assert.Empty(t, (*info)[0].Address)
}
