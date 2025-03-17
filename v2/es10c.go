package sgp22

import (
	"errors"
	"slices"
	"unicode/utf8"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
)

// region Section 5.7.15, ES10c.GetProfilesInfo

// ProfileInfoListRequest is a request to get a list of profiles.
//
// See https://aka.pw/sgp22/v2.5#page=199 (Section 5.7.15, ES10c.GetProfilesInfo)
type ProfileInfoListRequest struct {
	SearchCriteria *bertlv.TLV
	Tags           []bertlv.Tag
}

func (r *ProfileInfoListRequest) CardResponse() *ProfileInfoListResponse {
	return new(ProfileInfoListResponse)
}

func (r *ProfileInfoListRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildrenIter(bertlv.ContextSpecific.Constructed(45), func(yield func(*bertlv.TLV) bool) {
		if r.SearchCriteria != nil {
			if !yield(bertlv.NewChildren(bertlv.ContextSpecific.Constructed(0), r.SearchCriteria)) {
				return
			}
		}
		if tags := slices.Concat(r.Tags...); len(tags) > 0 {
			yield(bertlv.NewValue(bertlv.Application.Primitive(28), tags))
		}
	})
	return request, nil
}

type ProfileInfoListResponse struct {
	ProfileList []*ProfileInfo
	error       *bertlv.TLV
}

func (r *ProfileInfoListResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if r.error = tlv.First(bertlv.ContextSpecific.Primitive(1)); r.error != nil {
		return r.Valid()
	}
	tlv = tlv.First(bertlv.ContextSpecific.Constructed(0))
	var profile *ProfileInfo
	profiles := make([]*ProfileInfo, 0, len(tlv.Children))
	for _, child := range tlv.Children {
		profile = new(ProfileInfo)
		if err := profile.UnmarshalBERTLV(child); err != nil {
			return err
		}
		profiles = append(profiles, profile)
	}
	*r = ProfileInfoListResponse{ProfileList: profiles}
	return nil
}

func (r *ProfileInfoListResponse) Valid() error {
	if r.error == nil {
		return nil
	}
	switch r.error.Value[0] {
	case 1:
		return errors.New("incorrect input values")
	}
	return ErrUndefined
}

// endregion

// region Section 5.7.{16,17,18}, ES10c.{Enable,Disable,Delete}ProfileInfo

type (
	EnableProfileRequest struct {
		Identifier *bertlv.TLV
		Refresh    bool
	}
	EnableProfileResponse  struct{ Result int8 }
	DisableProfileRequest  EnableProfileRequest
	DisableProfileResponse EnableProfileResponse
	DeleteProfileRequest   EnableProfileRequest
	DeleteProfileResponse  EnableProfileResponse
)

type ProfileOperation byte

const (
	EnableProfile ProfileOperation = iota + 49
	DisableProfile
	DeleteProfile
)

// ProfileOperationRequest is a request to enable, disable, or delete a profile.
//
// See https://aka.pw/sgp22/v2.5#page=201 (Section 5.7.16, ES10c.EnableProfile)
//
// See https://aka.pw/sgp22/v2.5#page=204 (Section 5.7.17, ES10c.DisableProfile)
//
// See https://aka.pw/sgp22/v2.5#page=206 (Section 5.7.18, ES10c.DeleteProfile)
type ProfileOperationRequest struct {
	Operation  ProfileOperation
	Identifier *bertlv.TLV
	Refresh    bool
}

func (r *ProfileOperationRequest) CardResponse() *ProfileOperationResponse {
	return &ProfileOperationResponse{
		Operation: r.Operation,
	}
}

type ProfileOperationResponse struct {
	Operation ProfileOperation
	Result    int8
}

func (r *ProfileOperationRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildrenIter(
		bertlv.ContextSpecific.Constructed(uint64(r.Operation)),
		func(yield func(*bertlv.TLV) bool) {
			// DeleteProfile does not require the refresh flag.
			if r.Operation == DeleteProfile {
				yield(r.Identifier)
				return
			}
			// Refresh flag is optional for EnableProfile and DisableProfile.
			if !yield(bertlv.NewChildren(bertlv.ContextSpecific.Constructed(0), r.Identifier)) {
				return
			}
			yield(mustMarshalValue(bertlv.MarshalValue(
				bertlv.ContextSpecific.Primitive(1),
				primitive.MarshalBool(r.Refresh),
			)))
		},
	)
	return request, nil
}

func (r *ProfileOperationResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, uint64(r.Operation)) {
		return ErrUnexpectedTag
	}
	return tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&r.Result))
}

func (r *ProfileOperationResponse) Valid() error {
	switch r.Result {
	case 0:
		return nil
	case 1:
		return errors.New("iccid or aid not found")
	case 2:
		if r.Operation == EnableProfile {
			return errors.New("profile not in disabled state")
		}
		return errors.New("profile not in enabled state")
	case 3:
		return errors.New("disallowed by policy")
	case 4:
		return errors.New("wrong profile re-enabling")
	case 5:
		return ErrCatBusy
	}
	return ErrUndefined
}

// endregion

// region Section 5.7.19, ES10c.eUICCMemoryReset

// EuiccMemoryResetRequest is a request to reset the eUICC memory.
//
// See https://aka.pw/sgp22/v2.5#page=207 (Section 5.7.19, ES10c.eUICCMemoryReset)
type EuiccMemoryResetRequest struct {
	DeleteOperationalProfiles     bool
	DeleteFieldLoadedTestProfiles bool
	ResetDefaultSMDPAddress       bool
}

func (r *EuiccMemoryResetRequest) CardResponse() *EuiccMemoryResetResponse {
	return new(EuiccMemoryResetResponse)
}

func (r *EuiccMemoryResetRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(52),
		mustMarshalValue(bertlv.MarshalValue(
			bertlv.Application.Primitive(2),
			primitive.MarshalBitString([]bool{
				r.DeleteOperationalProfiles,
				r.DeleteFieldLoadedTestProfiles,
				r.ResetDefaultSMDPAddress,
			}),
		)),
	)
	return request, nil
}

type EuiccMemoryResetResponse struct {
	Result int8
}

func (r *EuiccMemoryResetResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 52) {
		return ErrUnexpectedTag
	}
	return tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&r.Result))
}

func (r *EuiccMemoryResetResponse) Valid() error {
	switch r.Result {
	case 0:
		return nil
	case 1:
		return ErrNothingToDelete
	case 5:
		return ErrCatBusy
	}
	return ErrUndefined
}

// endregion

// region Section 5.7.20, ES10c.GetEID

// GetEuiccDataRequest is a request to get the eUICC data.
//
// See https://aka.pw/sgp22/v2.5#page=209 (Section 5.7.20, ES10c.GetEID)
type GetEuiccDataRequest struct{}

func (r *GetEuiccDataRequest) CardResponse() *GetEuiccDataResponse {
	return new(GetEuiccDataResponse)
}

func (r *GetEuiccDataRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(62),
		bertlv.NewValue(
			bertlv.Application.Primitive(28),
			bertlv.Application.Primitive(26),
		),
	)
	return request, nil
}

type GetEuiccDataResponse struct {
	EID []byte
}

func (r *GetEuiccDataResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if tlv = tlv.First(bertlv.Application.Primitive(26)); tlv == nil {
		return errors.New("no eid found")
	}
	r.EID = tlv.Value
	return nil
}

func (r *GetEuiccDataResponse) Valid() error {
	return nil
}

// endregion

// region Section 5.7.21, ES10c.SetNickname

// SetNicknameRequest is a request to set the nickname of a profile.
//
// See https://aka.pw/sgp22/v2.5#page=209 (Section 5.7.21, ES10c.SetNickname)
type SetNicknameRequest struct {
	ICCID    ICCID
	Nickname []byte
}

func (r *SetNicknameRequest) CardResponse() *SetNicknameResponse {
	return new(SetNicknameResponse)
}

func (r *SetNicknameRequest) Valid() error {
	if !utf8.Valid(r.Nickname) {
		return errors.New("the nickname invalid utf-8 string")
	} else if len(r.Nickname) > 64 {
		return errors.New("the nickname too long")
	}
	return nil
}

func (r *SetNicknameRequest) MarshalBERTLV() (*bertlv.TLV, error) {
	if err := r.Valid(); err != nil {
		return nil, err
	}
	request := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(41),
		bertlv.NewValue(bertlv.Application.Primitive(26), r.ICCID),
		bertlv.NewValue(bertlv.ContextSpecific.Primitive(16), r.Nickname),
	)
	return request, nil
}

type SetNicknameResponse struct {
	Result int8
}

func (r *SetNicknameResponse) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 41) {
		return ErrUnexpectedTag
	}
	return tlv.First(bertlv.ContextSpecific.Primitive(0)).
		UnmarshalValue(primitive.UnmarshalInt(&r.Result))
}

func (r *SetNicknameResponse) Valid() error {
	switch r.Result {
	case 0:
		return nil
	case 1:
		return ErrICCIDNotFound
	}
	return ErrUndefined
}

// endregion
