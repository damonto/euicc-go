package lpa

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
	sgp22 "github.com/damonto/euicc-go/v2"
)

type ActivationCode struct {
	SMDP             *url.URL
	MatchingID       string
	IMEI             string
	OID              string
	ConfirmationCode string
}

type DownloadProgress uint8

const (
	DownloadProgressAuthenticateClient DownloadProgress = iota
	DownloadProgressAuthenticateServer
	DownloadProgressLoadBPP
)

type DownloadHandler interface {
	Progress(process DownloadProgress)
	Confirm(metadata *sgp22.ProfileInfo) chan bool
	ConfirmationCode() chan string
}

func (c *Client) DownloadProfile(ctx context.Context, activationCode *ActivationCode, handler DownloadHandler) (*sgp22.LoadBoundProfilePackageResponse, error) {
	handler.Progress(DownloadProgressAuthenticateClient)
	clientResponse, metadata, ccRequired, err := c.authenticateClient(activationCode)
	if err != nil {
		if clientResponse.Header.ExecutionStatus.ExecutedSuccess() {
			return nil, c.handleDownloadError(activationCode, clientResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
		}
		return nil, err
	}
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(activationCode.SMDP, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}

	if !<-handler.Confirm(metadata) {
		_, err := c.cancelSession(activationCode.SMDP, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	if ccRequired {
		activationCode.ConfirmationCode = <-handler.ConfirmationCode()
		if activationCode.ConfirmationCode == "" {
			return nil, errors.New("confirmation code is required")
		}
	}

	handler.Progress(DownloadProgressAuthenticateServer)
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(activationCode.SMDP, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	serverResponse, err := c.authenticateServer(activationCode, clientResponse)
	if err != nil {
		return nil, c.handleDownloadError(activationCode, serverResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
	}

	handler.Progress(DownloadProgressLoadBPP)
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(activationCode.SMDP, serverResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	result, err := c.install(serverResponse)
	if err != nil {
		return nil, c.handleDownloadError(activationCode, serverResponse.TransactionID, err, sgp22.CancelSessionReasonLoadBppExecutionError)
	}
	return result, nil
}

func (c *Client) install(bppResponse *sgp22.ES9BoundProfilePackageResponse) (*sgp22.LoadBoundProfilePackageResponse, error) {
	segments, err := sgp22.SegmentedBoundProfilePackage(bppResponse.BoundProfilePackage)
	if err != nil {
		return nil, err
	}
	var sw []byte
	for _, command := range segments {
		sw, err = sgp22.InvokeRawAPDU(c.APDU, command)
		if err != nil {
			return nil, err
		}
	}
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary(sw); err != nil {
		return nil, err
	}
	var response sgp22.LoadBoundProfilePackageResponse
	if err := response.UnmarshalBERTLV(&tlv); err != nil {
		return nil, err
	}
	if valid := response.Valid(); valid != nil {
		return nil, errors.New(valid.Error())
	}
	return &response, nil
}

func (c *Client) authenticateServer(activationCode *ActivationCode, clientResponse *sgp22.ES9AuthenticateClientResponse) (*sgp22.ES9BoundProfilePackageResponse, error) {
	response, err := c.PrepareDownload(activationCode.SMDP, &sgp22.PrepareDownloadRequest{
		TransactionID:    clientResponse.TransactionID,
		ProfileMetadata:  clientResponse.ProfileMetadata,
		Signed2:          clientResponse.Signed2,
		Signature2:       clientResponse.Signature2,
		Certificate:      clientResponse.Certificate,
		ConfirmationCode: []byte(activationCode.ConfirmationCode),
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) authenticateClient(activationCode *ActivationCode) (*sgp22.ES9AuthenticateClientResponse, *sgp22.ProfileInfo, bool, error) {
	initiateAuthenticationResponse, err := c.InitiateAuthentication(activationCode.SMDP)
	if err != nil {
		return nil, nil, false, err
	}
	imei, err := sgp22.NewIMEI(activationCode.IMEI)
	if err != nil {
		return nil, nil, false, err
	}
	response, err := c.AuthenticateClient(activationCode.SMDP, &sgp22.AuthenticateServerRequest{
		TransactionID: initiateAuthenticationResponse.TransactionID,
		Signed1:       initiateAuthenticationResponse.Signed1,
		Signature1:    initiateAuthenticationResponse.Signature1,
		UsedIssuer:    initiateAuthenticationResponse.UsedIssuer,
		Certificate:   initiateAuthenticationResponse.Certificate,
		IMEI:          imei,
		MatchingID:    []byte(activationCode.MatchingID),
	})
	if err != nil {
		return response, nil, false, err
	}
	metadata, err := c.profileMetadata(response.ProfileMetadata)
	if err != nil {
		return response, nil, false, err
	}
	return response, metadata, c.needConfirmationCode(response.Signed2), nil
}

func (c *Client) profileMetadata(tlv *bertlv.TLV) (*sgp22.ProfileInfo, error) {
	var profileInfo = new(sgp22.ProfileInfo)
	if err := profileInfo.UnmarshalBERTLV(tlv); err != nil {
		return nil, err
	}
	return profileInfo, nil
}

func (c *Client) isCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (c *Client) handleDownloadError(activationCode *ActivationCode, transactionID []byte, err error, cancelReason sgp22.CancelSessionReason) error {
	_, cancelErr := c.cancelSession(activationCode.SMDP, transactionID, cancelReason)
	if cancelErr != nil {
		return errors.Join(err, fmt.Errorf("cancel session error: %w", cancelErr))
	}
	return err
}

func (c *Client) cancelSession(address *url.URL, transactionID []byte, reason sgp22.CancelSessionReason) (*sgp22.ES9CancelSessionResponse, error) {
	cancelSessionRequest, err := sgp22.InvokeAPDU(c.APDU, &sgp22.CancelSessionRequest{
		TransactionID: transactionID,
		Reason:        reason,
	})
	if err != nil {
		return nil, err
	}
	return sgp22.InvokeHTTP(c.HTTP, address, cancelSessionRequest)
}

func (c *Client) needConfirmationCode(tlv *bertlv.TLV) bool {
	var required bool
	_ = tlv.First(bertlv.Universal.Primitive(1)).
		UnmarshalValue(primitive.UnmarshalBool(&required))
	return required
}
