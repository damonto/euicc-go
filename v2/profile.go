package sgp22

import (
	"bytes"
	"encoding/base64"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
)

type ProfileInfo struct {
	ICCID                         ICCID
	ISDPAID                       ISDPAID
	ProfileState                  ProfileState
	ProfileNickname               string
	ServiceProviderName           string
	ProfileName                   string
	Icon                          ProfileIcon
	ProfileClass                  ProfileClass
	ProfileOwner                  OperatorId
	NotificationConfigurationInfo NotificationConfigurationInfo
}

func (p *ProfileInfo) UnmarshalBERTLV(tlv *bertlv.TLV) (err error) {
	if !tlv.Tag.If(bertlv.Private, bertlv.Constructed, 3) && !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 37) {
		return ErrUnexpectedTag
	}
	*p = ProfileInfo{
		ICCID:               tlv.First(bertlv.Application.Primitive(26)).Value,
		ServiceProviderName: string(tlv.First(bertlv.ContextSpecific.Primitive(17)).Value),
		ProfileName:         string(tlv.First(bertlv.ContextSpecific.Primitive(18)).Value),
		ProfileClass:        ProfileClassProvisioning,
	}
	if profileClass := tlv.First(bertlv.ContextSpecific.Primitive(21)); profileClass != nil {
		p.ProfileClass = ProfileClass(profileClass.Value[0])
	}
	if id := tlv.First(bertlv.Application.Primitive(15)); id != nil {
		p.ISDPAID = id.Value
	}
	if nickname := tlv.First(bertlv.ContextSpecific.Primitive(16)); nickname != nil {
		p.ProfileNickname = string(nickname.Value)
	}
	if icon := tlv.First(bertlv.ContextSpecific.Primitive(20)); icon != nil {
		p.Icon = icon.Value
	}
	if tlv.Tag.If(bertlv.Private, bertlv.Constructed, 3) {
		if err = tlv.First(bertlv.ContextSpecific.Primitive(112)).UnmarshalValue(primitive.UnmarshalInt(&p.ProfileState)); err != nil {
			return err
		}
	}
	if notification := tlv.First(bertlv.ContextSpecific.Constructed(22)); notification != nil {
		if err = p.NotificationConfigurationInfo.UnmarshalBERTLV(notification); err != nil {
			return err
		}
	}
	if owner := tlv.First(bertlv.ContextSpecific.Constructed(23)); owner != nil {
		if err = p.ProfileOwner.UnmarshalBERTLV(owner); err != nil {
			return err
		}
	}
	return nil
}

type ProfileState int8

const (
	ProfileDisabled ProfileState = 0
	ProfileEnabled  ProfileState = 1
)

func (state ProfileState) String() string {
	switch state {
	case ProfileDisabled:
		return "disabled"
	case ProfileEnabled:
		return "enable"
	}
	return "unknown"
}

type ProfileIcon []byte

func (p ProfileIcon) Valid() bool    { return len(p.FileType()) > 0 }
func (p ProfileIcon) String() string { return base64.URLEncoding.EncodeToString(p) }

func (p ProfileIcon) FileType() string {
	switch {
	case bytes.HasPrefix(p, []byte("\xFF\xD8\xFF\xDB")):
		return "image/jpeg"
	case bytes.HasPrefix(p, []byte("\x89PNG")):
		return "image/png"
	}
	return ""
}

type OperatorId struct {
	PLMN, GID1, GID2 []byte
}

func (id *OperatorId) MCC() string {
	if len(id.PLMN) == 0 {
		return ""
	}
	return string([]byte{
		'0' + id.PLMN[0]&0x0f,
		'0' + id.PLMN[0]>>4,
		'0' + id.PLMN[1]&0x0f,
	})
}

func (id *OperatorId) MNC() string {
	if len(id.PLMN) < 2 {
		return ""
	}
	mnc := []byte{'0' + id.PLMN[2]&0x0f, '0' + id.PLMN[2]>>4}
	if last := id.PLMN[1] >> 4; last != 0xf {
		mnc = append(mnc, '0'+last)
	}
	return string(mnc)
}

func (id *OperatorId) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 23) {
		return ErrUnexpectedTag
	}
	*id = OperatorId{
		PLMN: tlv.First(bertlv.ContextSpecific.Primitive(0)).Value,
	}
	if gid1 := tlv.First(bertlv.ContextSpecific.Primitive(1)); gid1 != nil {
		id.GID1 = gid1.Value
	}
	if gid2 := tlv.First(bertlv.ContextSpecific.Primitive(2)); gid2 != nil {
		id.GID2 = gid2.Value
	}
	return nil
}

type NotificationConfiguration struct {
	ProfileManagementOperation NotificationEvent
	Address                    string
}

type NotificationConfigurationInfo []*NotificationConfiguration

func (n *NotificationConfigurationInfo) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 22) {
		return ErrUnexpectedTag
	}
	configs := make(NotificationConfigurationInfo, 0, len(tlv.Children))
	for _, child := range tlv.Children {
		c := NotificationConfiguration{
			Address: string(child.First(bertlv.ContextSpecific.Primitive(1)).Value),
		}
		c.ProfileManagementOperation.UnmarshalBinary(child.First(bertlv.ContextSpecific.Primitive(0)).Value)
		configs = append(configs, &c)
	}
	*n = configs
	return nil
}
