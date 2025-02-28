package lpa

import (
	"errors"

	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
)

// ListProfile returns a list of profiles that match the search criteria.
// If the search criteria is empty, all profiles are returned.
//
// Search Criteria:
// - [sgp22.ICCID]: The ICCID of the profile.
// - [sgp22.ISDPAID]: The ISD-P AID of the profile.
// - [sgp22.ProfileClass]: The profile class of the profile.
//
// See https://aka.pw/sgp22/v2.5#page=199 (Section 5.7.15, ES10c.GetProfilesInfo)
func (c *Client) ListProfile(searchCriteria any) ([]*sgp22.ProfileInfo, error) {
	var request sgp22.ProfileInfoListRequest
	switch v := searchCriteria.(type) {
	case nil:
		break
	case sgp22.ICCID:
		request.SearchCriteria = bertlv.NewValue(bertlv.Application.Primitive(26), v)
	case sgp22.ISDPAID:
		request.SearchCriteria = bertlv.NewValue(bertlv.Application.Primitive(15), v)
	case sgp22.ProfileClass:
		request.SearchCriteria, _ = bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(21), v)
	}
	response, err := sgp22.InvokeAPDU(c.APDU, &request)
	if err != nil {
		return nil, err
	}
	return response.ProfileList, nil
}

// EnableProfile enables a profile.
// The profile is identified by the ICCID or ISD-P AID.
//
// ProfileInfo Identifier:
// - [sgp22.ICCID]: The ICCID of the profile.
// - [sgp22.ISDPAID]: The ISD-P AID of the profile.
//
// See https://aka.pw/sgp22/v2.5#page=201 (Section 5.7.16, ES10c.EnableProfile)
func (c *Client) EnableProfile(identifier any) error {
	return c.setProfile(sgp22.EnableProfile, identifier)
}

// DisableProfile disables a profile.
// The profile is identified by the ICCID or ISD-P AID.
//
// ProfileInfo Identifier:
// - [sgp22.ICCID]: The ICCID of the profile.
// - [sgp22.ISDPAID]: The ISD-P AID of the profile.
//
// See https://aka.pw/sgp22/v2.5#page=204 (Section 5.7.17, ES10c.DisableProfile)
func (c *Client) DisableProfile(identifier any) error {
	return c.setProfile(sgp22.DisableProfile, identifier)
}

// DeleteProfile deletes a profile.
// The profile is identified by the ICCID or ISD-P AID.
//
// ProfileInfo Identifier:
// - [sgp22.ICCID]: The ICCID of the profile.
// - [sgp22.ISDPAID]: The ISD-P AID of the profile.
//
// See https://aka.pw/sgp22/v2.5#page=206 (Section 5.7.18, ES10c.DeleteProfile)
func (c *Client) DeleteProfile(identifier any) error {
	return c.setProfile(sgp22.DeleteProfile, identifier)
}

func (c *Client) setProfile(operation sgp22.ProfileOperation, identifier any) (err error) {
	var request sgp22.ProfileOperationRequest
	request.Operation = operation
	switch v := identifier.(type) {
	case sgp22.ICCID:
		request.Identifier = bertlv.NewValue(bertlv.Application.Primitive(26), v)
	case sgp22.ISDPAID:
		request.Identifier = bertlv.NewValue(bertlv.Application.Primitive(15), v)
	default:
		return errors.New("invalid profile identifier")
	}
	request.Refresh = true
	_, err = sgp22.InvokeAPDU(c.APDU, &request)
	return
}

// MemoryReset resets the eUICC memory.
// This operation deletes all operational profiles, field-loaded test profiles,
// and resets the default SM-DP+ address.
//
// See https://aka.pw/sgp22/v2.5#page=207 (Section 5.7.19, ES10c.eUICCMemoryReset)
func (c *Client) MemoryReset() error {
	_, err := sgp22.InvokeAPDU(c.APDU, &sgp22.EuiccMemoryResetRequest{
		DeleteOperationalProfiles:     true,
		DeleteFieldLoadedTestProfiles: true,
		ResetDefaultSMDPAddress:       true,
	})
	return err
}

// EID returns the EID of the eUICC.
// The EID is a unique identifier of the eUICC.
//
// See https://aka.pw/sgp22/v2.5#page=209 (Section 5.7.20, ES10c.GetEID)
func (c *Client) EID() ([]byte, error) {
	response, err := sgp22.InvokeAPDU(c.APDU, new(sgp22.GetEuiccDataRequest))
	if err != nil {
		return nil, err
	}
	return response.EID, nil
}

// SetNickname sets the nickname of the profile.
//
// See https://aka.pw/sgp22/v2.5#page=209 (Section 5.7.21, ES10c.SetNickname)
func (c *Client) SetNickname(iccid sgp22.ICCID, nickname string) error {
	_, err := sgp22.InvokeAPDU(c.APDU, &sgp22.SetNicknameRequest{
		ICCID:    iccid,
		Nickname: []byte(nickname),
	})
	return err
}
