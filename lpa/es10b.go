package lpa

import (
	"errors"
	"net/url"

	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
)

func (c *Client) EUICCChallenge() ([]byte, error) {
	euiccChallenge, err := sgp22.InvokeAPDU(c.APDU, new(sgp22.GetEuiccChallengeRequest))
	if err != nil {
		return nil, err
	}
	return euiccChallenge.Challenge, nil
}

// EUICCInfo1 retrieves the eUICC information (version 1).
//
// See https://aka.pw/sgp22/v2.5#page=187 (Section 5.7.8, ES10b.GetEUICCInfo)
func (c *Client) EUICCInfo1() (*bertlv.TLV, error) {
	euiccInfo1, err := sgp22.InvokeAPDU(c.APDU, &sgp22.GetEuiccInfoRequest{Version: 1})
	if err != nil {
		return nil, err
	}
	return euiccInfo1.Response, nil
}

// EUICCInfo2 retrieves the eUICC information (version 2).
//
// See https://aka.pw/sgp22/v2.5#page=187 (Section 5.7.8, ES10b.GetEUICCInfo)
func (c *Client) EUICCInfo2() (*bertlv.TLV, error) {
	euiccInfo1, err := sgp22.InvokeAPDU(c.APDU, &sgp22.GetEuiccInfoRequest{Version: 2})
	if err != nil {
		return nil, err
	}
	return euiccInfo1.Response, nil
}

// AuthenticateClient authenticates the client to the eUICC.
//
// See https://aka.pw/sgp22/v2.5#page=195 (Section 5.7.13, ES10b.AuthenticateClient)
func (c *Client) AuthenticateClient(address *url.URL, request *sgp22.AuthenticateServerRequest) (*sgp22.ES9AuthenticateClientResponse, error) {
	authenticateClientRequest, err := sgp22.InvokeAPDU(c.APDU, request)
	if err != nil {
		return nil, err
	}
	return sgp22.InvokeHTTP(c.HTTP, address, authenticateClientRequest)
}

// PrepareDownload prepares the eUICC for a profile download.
//
// See https://aka.pw/sgp22/v2.5#page=184 (Section 5.7.13, ES10b.PrepareDownload)
func (c *Client) PrepareDownload(address *url.URL, request *sgp22.PrepareDownloadRequest) (*sgp22.ES9BoundProfilePackageResponse, error) {
	boundProfilePackageRequest, err := sgp22.InvokeAPDU(c.APDU, request)
	if err != nil {
		return nil, err
	}
	return sgp22.InvokeHTTP(c.HTTP, address, boundProfilePackageRequest)
}

// ListNotification retrieves a list of notifications from the eUICC.
//
// See https://aka.pw/sgp22/v2.5#page=191 (Section 5.7.9, ES10b.ListNotification)
func (c *Client) ListNotification(filters ...sgp22.NotificationEvent) ([]*sgp22.NotificationMetadata, error) {
	var request sgp22.ListNotificationRequest
	request.Filter = make(map[sgp22.NotificationEvent]bool)
	if len(filters) == 0 {
		filters = []sgp22.NotificationEvent{
			sgp22.NotificationEventInstall,
			sgp22.NotificationEventEnable,
			sgp22.NotificationEventDisable,
			sgp22.NotificationEventDelete,
		}
	}
	for _, event := range filters {
		request.Filter[event] = true
	}
	response, err := sgp22.InvokeAPDU(c.APDU, &request)
	if err != nil {
		return nil, err
	}
	return response.NotificationList, nil
}

// RetrieveNotificationList retrieves a list of notifications from the eUICC.
//
// Search Criteria:
// - [sgp22.SequenceNumber]: The sequence number of the notification.
// - [sgp22.NotificationEvent]: The event type of the notification.
func (c *Client) RetrieveNotificationList(searchCriteria any) ([]*sgp22.PendingNotification, error) {
	var request sgp22.RetrieveNotificationsListRequest
	switch v := searchCriteria.(type) {
	case sgp22.SequenceNumber:
		request.SearchCriteria, _ = bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(0), &v)
	case sgp22.NotificationEvent:
		request.SearchCriteria, _ = bertlv.MarshalValue(bertlv.ContextSpecific.Primitive(1), &v)
	default:
		return nil, errors.New("searchCriteria must be of type sgp22.SequenceNumber or sgp22.NotificationEvent")
	}
	response, err := sgp22.InvokeAPDU(c.APDU, &request)
	if err != nil {
		return nil, err
	}
	return response.NotificationList, nil
}

// RemoveNotificationFromList removes a notification from the eUICC's notification list.
//
// See https://aka.pw/sgp22/v2.5#page=193 (Section 5.7.11, ES10b.RemoveNotificationFromList)
func (c *Client) RemoveNotificationFromList(sequenceNumber sgp22.SequenceNumber) error {
	_, err := sgp22.InvokeAPDU(c.APDU, &sgp22.NotificationSentRequest{
		SequenceNumber: sequenceNumber,
	})
	return err
}
