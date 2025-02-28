package lpa

import (
	"net/url"

	"github.com/damonto/euicc-go/v2"
)

// InitiateAuthentication initiates the authentication process.
//
// See https://aka.pw/sgp22/v2.5#page=170 (Section 5.6.1, ES9p.InitiateAuthentication)
func (c *Client) InitiateAuthentication(address *url.URL) (*sgp22.ES9InitiateAuthenticationResponse, error) {
	var err error
	request := sgp22.ES9InitiateAuthenticationRequest{Address: address.Host}
	if request.Challenge, err = c.EUICCChallenge(); err != nil {
		return nil, err
	}
	if request.Info1, err = c.EUICCInfo1(); err != nil {
		return nil, err
	}
	return sgp22.InvokeHTTP(c.HTTP, address, &request)
}

// HandleNotification handles the pending notification.
//
// See https://aka.pw/sgp22/v2.5#page=177 (Section 5.6.4, ES9p.HandleNotification)
func (c *Client) HandleNotification(pendingNotification *sgp22.PendingNotification) error {
	request := sgp22.ES9HandleNotificationRequest{
		PendingNotification: pendingNotification.PendingNotification,
	}
	_, err := sgp22.InvokeHTTP(c.HTTP, &url.URL{
		Scheme: "https",
		Host:   pendingNotification.Notification.Address,
	}, &request)
	return err
}
