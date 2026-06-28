package sgp22

import (
	"bytes"
	"encoding/base64"
	"errors"

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
	IconType                      ProfileIconType
	Icon                          ProfileIcon
	ProfileClass                  ProfileClass
	ProfileOwner                  OperatorId
	NotificationConfigurationInfo NotificationConfigurationInfo
	SMDPProprietaryData           *bertlv.TLV
	ProfilePolicyRules            ProfilePolicyRules
	ServiceSpecificData           *bertlv.TLV
}

func (p *ProfileInfo) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.Private, bertlv.Constructed, 3) && !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 37) {
		return ErrUnexpectedTag
	}
	*p = ProfileInfo{
		ProfileClass: ProfileClassProvisioning,
	}
	if err := optional(tlv, TagICCID, &p.ICCID, ICCID(nil)); err != nil {
		return err
	}
	if err := optional(tlv, TagISDPAID, &p.ISDPAID, ISDPAID(nil)); err != nil {
		return err
	}
	if err := optional(tlv, TagProfileState, &p.ProfileState, ProfileDisabled); err != nil {
		return err
	}
	if err := optional(tlv, TagNickname, &p.ProfileNickname, ""); err != nil {
		return err
	}
	if err := optional(tlv, TagServiceProviderName, &p.ServiceProviderName, ""); err != nil {
		return err
	}
	if err := optional(tlv, TagProfileName, &p.ProfileName, ""); err != nil {
		return err
	}
	if err := optional(tlv, TagProfileIconType, &p.IconType, ProfileIconTypeJPG); err != nil {
		return err
	}
	if err := optional(tlv, TagProfileIcon, &p.Icon, ProfileIcon(nil)); err != nil {
		return err
	}
	if err := optional(tlv, TagProfileClass, &p.ProfileClass, ProfileClassProvisioning); err != nil {
		return err
	}
	if notification := tlv.First(bertlv.ContextSpecific.Constructed(22)); notification != nil {
		if err := p.NotificationConfigurationInfo.UnmarshalBERTLV(notification); err != nil {
			return err
		}
	}
	if owner := tlv.First(bertlv.ContextSpecific.Constructed(23)); owner != nil {
		if err := p.ProfileOwner.UnmarshalBERTLV(owner); err != nil {
			return err
		}
	}
	p.SMDPProprietaryData = tlv.First(TagSMDPProprietaryData)
	if err := optional(tlv, TagProfilePolicyRules, &p.ProfilePolicyRules, ProfilePolicyRules{}); err != nil {
		return err
	}
	p.ServiceSpecificData = tlv.First(TagServiceSpecificData)
	return nil
}

func optional[T any](tlv *bertlv.TLV, tag bertlv.Tag, dst *T, def T) error {
	*dst = def
	field := tlv.First(tag)
	if field == nil {
		return nil
	}
	switch v := any(dst).(type) {
	case *string:
		*v = string(field.Value)
	case *[]byte:
		*v = field.Value
	case *ICCID:
		*v = ICCID(field.Value)
	case *ISDPAID:
		*v = ISDPAID(field.Value)
	case *ProfileIcon:
		*v = ProfileIcon(field.Value)
	case *ProfileIconType:
		return field.UnmarshalValue(primitive.UnmarshalInt(v))
	case *ProfileState:
		return field.UnmarshalValue(primitive.UnmarshalInt(v))
	case *ProfileClass:
		return field.UnmarshalValue(primitive.UnmarshalInt(v))
	case *NotificationEvent:
		return field.UnmarshalValue(v)
	case *ProfilePolicyRules:
		return field.UnmarshalValue(v)
	case *[]bool:
		return field.UnmarshalValue(primitive.UnmarshalBitString(v))
	default:
		return errors.New("unsupported optional field")
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
		return "enabled"
	}
	return "unknown"
}

type ProfileIcon []byte

func (p ProfileIcon) Valid() bool    { return len(p.FileType()) > 0 }
func (p ProfileIcon) String() string { return base64.StdEncoding.EncodeToString(p) }

func (p ProfileIcon) FileType() string {
	switch {
	case bytes.HasPrefix(p, []byte("\xFF\xD8\xFF\xDB")):
		return "image/jpeg"
	case bytes.HasPrefix(p, []byte("\x89PNG")):
		return "image/png"
	}
	return ""
}

type ProfileIconType int8

const (
	ProfileIconTypeJPG ProfileIconType = 0
	ProfileIconTypePNG ProfileIconType = 1
)

type OperatorId struct {
	PLMN, GID1, GID2 []byte
}

func (id *OperatorId) MCC() string {
	if len(id.PLMN) < 2 {
		return ""
	}
	return string([]byte{
		'0' + id.PLMN[0]&0x0f,
		'0' + id.PLMN[0]>>4,
		'0' + id.PLMN[1]&0x0f,
	})
}

func (id *OperatorId) MNC() string {
	if len(id.PLMN) < 3 {
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
	ProfileManagementOperations []NotificationEvent
	Address                     string
}

type NotificationConfigurationInfo []*NotificationConfiguration

func (n *NotificationConfigurationInfo) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 22) {
		return ErrUnexpectedTag
	}
	configs := make(NotificationConfigurationInfo, 0, len(tlv.Children))
	for _, child := range tlv.Children {
		if child == nil || !child.Tag.If(bertlv.Universal, bertlv.Constructed, 16) {
			return ErrUnexpectedTag
		}
		operation := child.First(bertlv.ContextSpecific.Primitive(0))
		address := child.First(bertlv.ContextSpecific.Primitive(1))
		if operation == nil || address == nil {
			return ErrUnexpectedTag
		}
		c := NotificationConfiguration{
			Address: string(address.Value),
		}
		events, err := notificationEvents(operation.Value)
		if err != nil {
			return err
		}
		c.ProfileManagementOperations = events
		configs = append(configs, &c)
	}
	*n = configs
	return nil
}

func notificationEvents(data []byte) ([]NotificationEvent, error) {
	var bits []bool
	if err := primitive.UnmarshalBitString(&bits).UnmarshalBinary(data); err != nil {
		return nil, err
	}
	events := make([]NotificationEvent, 0, len(bits))
	for index, bit := range bits {
		if !bit {
			continue
		}
		event := NotificationEvent(index)
		if event > NotificationEventDelete {
			return nil, errors.New("invalid notification event")
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil, errors.New("notification event has no bits set")
	}
	return events, nil
}

type ProfilePolicyRules struct {
	UpdateControl       bool
	DisablingNotAllowed bool
	DeletionNotAllowed  bool
}

func (p *ProfilePolicyRules) UnmarshalBinary(data []byte) error {
	var bits []bool
	if err := primitive.UnmarshalBitString(&bits).UnmarshalBinary(data); err != nil {
		return err
	}

	*p = ProfilePolicyRules{}
	if len(bits) > 0 {
		p.UpdateControl = bits[0]
	}
	if len(bits) > 1 {
		p.DisablingNotAllowed = bits[1]
	}
	if len(bits) > 2 {
		p.DeletionNotAllowed = bits[2]
	}
	return nil
}
