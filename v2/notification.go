package sgp22

import (
	"errors"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
)

// region notification search criteria

// region sequence number

type SequenceNumber int64

func (n SequenceNumber) MarshalBinary() ([]byte, error) {
	return primitive.MarshalInt(n).MarshalBinary()
}

// endregion

// region profile management operation

type NotificationEvent byte

func (n *NotificationEvent) UnmarshalBinary(data []byte) error {
	var bits []bool
	if err := primitive.UnmarshalBitString(&bits).UnmarshalBinary(data); err != nil {
		return err
	}
	var found bool
	for index, bit := range bits {
		if !bit {
			continue
		}
		if NotificationEvent(index) > NotificationEventDelete {
			return errors.New("invalid notification event")
		}
		if found {
			return errors.New("notification event has multiple bits set")
		}
		*n = NotificationEvent(index)
		found = true
	}
	if !found {
		return errors.New("notification event has no bits set")
	}
	return nil
}

func (n *NotificationEvent) MarshalBinary() ([]byte, error) {
	if *n > NotificationEventDelete {
		return nil, errors.New("invalid notification event")
	}
	bits := make([]bool, 4)
	bits[*n] = true
	return primitive.MarshalBitString(bits).MarshalBinary()
}

const (
	NotificationEventInstall NotificationEvent = 0
	NotificationEventEnable  NotificationEvent = 1
	NotificationEventDisable NotificationEvent = 2
	NotificationEventDelete  NotificationEvent = 3
)

// endregion

// endregion

type NotificationMetadata struct {
	SequenceNumber             SequenceNumber
	ProfileManagementOperation NotificationEvent
	Address                    string
	ICCID                      ICCID
}

func (n *NotificationMetadata) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 47) {
		return ErrUnexpectedTag
	}
	*n = NotificationMetadata{
		Address: string(tlv.First(bertlv.Universal.Primitive(12)).Value),
	}
	if err := tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&n.SequenceNumber)); err != nil {
		return err
	}
	if iccid := tlv.First(bertlv.Application.Primitive(26)); iccid != nil {
		n.ICCID = ICCID(iccid.Value)
	}
	if err := tlv.First(bertlv.ContextSpecific.Primitive(1)).
		UnmarshalValue(&n.ProfileManagementOperation); err != nil {
		return err
	}
	return nil
}

type PendingNotification struct {
	PendingNotification *bertlv.TLV
	Notification        *NotificationMetadata
}

func (p *PendingNotification) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if len(tlv.Children) == 0 {
		return errors.New("notification does not exist")
	}
	var metadata *bertlv.TLV
	switch {
	case tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 55):
		metadata = tlv.Select(
			bertlv.ContextSpecific.Constructed(39),
			bertlv.ContextSpecific.Constructed(47),
		)
	case tlv.Tag.If(bertlv.Universal, bertlv.Constructed, 16):
		metadata = tlv.First(bertlv.ContextSpecific.Constructed(47))
	default:
		return ErrUnexpectedTag
	}
	*p = PendingNotification{PendingNotification: tlv}
	p.Notification = new(NotificationMetadata)
	return p.Notification.UnmarshalBERTLV(metadata)
}
