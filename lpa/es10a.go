package lpa

import (
	"github.com/damonto/euicc-go/v2"
)

type EUICCConfiguredAddresses struct {
	DefaultSMDPAddress string
	RootSMDSAddress    string
}

// EUICCConfiguredAddresses returns the default SM-DP+ address and the root SM-DS address.
//
// See https://aka.pw/sgp22/v2.5#page=183 (Section 5.7.3, ES10a.GetEuiccConfiguredAddresses)
func (c *Client) EUICCConfiguredAddresses() (*EUICCConfiguredAddresses, error) {
	response, err := sgp22.InvokeAPDU(c.APDU, new(sgp22.EuiccConfiguredAddressesRequest))
	if err != nil {
		return nil, err
	}
	addresses := EUICCConfiguredAddresses{
		DefaultSMDPAddress: response.DefaultSMDPAddress,
		RootSMDSAddress:    response.RootSMDSAddress,
	}
	return &addresses, nil
}

// SetDefaultDPAddress sets the default SM-DP+ address.
//
// See https://aka.pw/sgp22/v2.5#page=183 (Section 5.7.4, ES10a.SetDefaultDpAddress)
func (c *Client) SetDefaultDPAddress(address string) error {
	_, err := sgp22.InvokeAPDU(c.APDU, &sgp22.SetDefaultDPAddressRequest{
		DefaultDPAddress: address,
	})
	return err
}
