package sgp22

import (
	"errors"
	"unicode/utf8"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
)

// region Section 5.7.3, ES10a.GetEuiccConfiguredAddresses

// EuiccConfiguredAddressesRequest is a request to get the default SM-DP+ address and the root SM-DS address.
//
// See https://aka.pw/sgp22/v2.5#page=183 (Section 5.7.3, ES10a.GetEuiccConfiguredAddresses)
type EuiccConfiguredAddressesRequest struct{}

func (r *EuiccConfiguredAddressesRequest) CardResponse() *EuiccConfiguredAddressesResponse {
	return new(EuiccConfiguredAddressesResponse)
}

func (r *EuiccConfiguredAddressesRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	return bertlv.NewChildren(bertlv.ContextSpecific.Constructed(60)), nil
}

type EuiccConfiguredAddressesResponse struct {
	DefaultSMDPAddress string
	RootSMDSAddress    string
}

func (r *EuiccConfiguredAddressesResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 60) {
		return ErrUnexpectedTag
	}
	var response EuiccConfiguredAddressesResponse
	var child *bertlv.TLV
	if child = tlv.First(bertlv.ContextSpecific.Primitive(0)); child != nil {
		response.DefaultSMDPAddress = string(child.Value)
	}
	if child = tlv.First(bertlv.ContextSpecific.Primitive(1)); child != nil {
		response.RootSMDSAddress = string(child.Value)
	}
	*r = response
	return nil
}

func (r *EuiccConfiguredAddressesResponse) Valid() error {
	return nil
}

// endregion

// region Section 5.7.4, ES10a.SetDefaultDpAddress

// SetDefaultDPAddressRequest is a request to set the default SM-DP+ address.
//
// See https://aka.pw/sgp22/v2.5#page=183 (Section 5.7.4, ES10a.SetDefaultDpAddress)
type SetDefaultDPAddressRequest struct {
	DefaultDPAddress string
}

func (r *SetDefaultDPAddressRequest) CardResponse() *SetDefaultDPAddressResponse {
	return new(SetDefaultDPAddressResponse)
}

func (r *SetDefaultDPAddressRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	if !utf8.Valid([]byte(r.DefaultDPAddress)) {
		return nil, errors.New("DefaultDPAddress is not a valid UTF-8 string")
	}
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(63),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(0), []byte(r.DefaultDPAddress)),
	)
	return request, nil
}

type SetDefaultDPAddressResponse struct {
	Result int8
}

func (r *SetDefaultDPAddressResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 63) {
		return ErrUnexpectedTag
	}
	return tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&r.Result))
}

func (r *SetDefaultDPAddressResponse) Valid() error {
	if r == nil || r.Result == 0 {
		return nil
	}
	return ErrUndefined
}

// endregion
