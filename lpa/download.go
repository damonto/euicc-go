package lpa

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

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

func (c *Client) DownloadProfile(ctx context.Context, ac *ActivationCode, handler DownloadHandler) (*sgp22.LoadBoundProfilePackageResponse, error) {
	handler.Progress(DownloadProgressAuthenticateClient)
	clientResponse, metadata, ccRequired, err := c.authenticateClient(ac)
	if err != nil {
		if clientResponse.Header.ExecutionStatus.ExecutedSuccess() {
			return nil, c.raiseError(ac, clientResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
		}
		return nil, err
	}

	if c.isCanceled(ctx) || !<-handler.Confirm(metadata) {
		_, err := c.cancelSession(ac, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}

	if ccRequired && ac.ConfirmationCode == "" {
		ac.ConfirmationCode = <-handler.ConfirmationCode()
		if ac.ConfirmationCode == "" {
			return nil, c.raiseError(
				ac,
				clientResponse.TransactionID,
				errors.New("confirmation code is required"),
				sgp22.CancelSessionReasonEndUserRejection,
			)
		}
	}

	handler.Progress(DownloadProgressAuthenticateServer)
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(ac, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	serverResponse, err := c.authenticateServer(ac, clientResponse)
	if err != nil {
		return nil, c.raiseError(ac, serverResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
	}

	handler.Progress(DownloadProgressLoadBPP)
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(ac, serverResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	result, err := c.install(serverResponse)
	if err != nil {
		return result, c.raiseError(ac, serverResponse.TransactionID, err, sgp22.CancelSessionReasonLoadBppExecutionError)
	}
	return result, nil
}

func (c *Client) install(bppResponse *sgp22.ES9BoundProfilePackageResponse) (*sgp22.LoadBoundProfilePackageResponse, error) {
	segments, err := sgp22.SegmentedBoundProfilePackage(bppResponse.BoundProfilePackage)
	if err != nil {
		return nil, err
	}
	var r []byte
	for _, command := range segments {
		r, err = sgp22.InvokeRawAPDU(c.APDU, command)
		if err != nil {
			return nil, err
		}
		if len(r) > 0 {
			break
		}
	}
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary(r); err != nil {
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

func (c *Client) authenticateServer(ac *ActivationCode, clientResponse *sgp22.ES9AuthenticateClientResponse) (*sgp22.ES9BoundProfilePackageResponse, error) {
	return c.PrepareDownload(ac.SMDP, &sgp22.PrepareDownloadRequest{
		TransactionID:    clientResponse.TransactionID,
		ProfileMetadata:  clientResponse.ProfileMetadata,
		Signed2:          clientResponse.Signed2,
		Signature2:       clientResponse.Signature2,
		Certificate:      clientResponse.Certificate,
		ConfirmationCode: []byte(ac.ConfirmationCode),
	})
}

func (c *Client) authenticateClient(ac *ActivationCode) (*sgp22.ES9AuthenticateClientResponse, *sgp22.ProfileInfo, bool, error) {
	initiateAuthenticationResponse, err := c.InitiateAuthentication(ac.SMDP)
	if err != nil {
		return nil, nil, false, err
	}
	imei, err := sgp22.NewIMEI(ac.IMEI)
	if err != nil {
		return nil, nil, false, err
	}
	response, err := c.AuthenticateClient(ac.SMDP, &sgp22.AuthenticateServerRequest{
		TransactionID: initiateAuthenticationResponse.TransactionID,
		Signed1:       initiateAuthenticationResponse.Signed1,
		Signature1:    initiateAuthenticationResponse.Signature1,
		UsedIssuer:    initiateAuthenticationResponse.UsedIssuer,
		Certificate:   initiateAuthenticationResponse.Certificate,
		IMEI:          imei,
		MatchingID:    []byte(ac.MatchingID),
	})
	if err != nil {
		return response, nil, false, err
	}
	metadata, err := c.profileMetadata(response.ProfileMetadata)
	if err != nil {
		return response, nil, false, err
	}
	return response, metadata, c.confirmationCodeRequired(response.Signed2), nil
}

func (c *Client) confirmationCodeRequired(tlv *bertlv.TLV) bool {
	var required bool
	tlv.First(bertlv.Universal.Primitive(1)).UnmarshalValue(primitive.UnmarshalBool(&required))
	return required
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

func (c *Client) raiseError(ac *ActivationCode, transactionID []byte, err error, cancelReason sgp22.CancelSessionReason) error {
	_, cancelErr := c.cancelSession(ac, transactionID, cancelReason)
	if cancelErr != nil {
		return errors.Join(err, fmt.Errorf("cancel session error: %w", cancelErr))
	}
	return err
}

func (c *Client) cancelSession(ac *ActivationCode, transactionID []byte, reason sgp22.CancelSessionReason) (*sgp22.ES9CancelSessionResponse, error) {
	cancelSessionRequest, err := sgp22.InvokeAPDU(c.APDU, &sgp22.CancelSessionRequest{
		TransactionID: transactionID,
		Reason:        reason,
	})
	if err != nil {
		return nil, err
	}
	return sgp22.InvokeHTTP(c.HTTP, ac.SMDP, cancelSessionRequest)
}

func ParseActivationCode(acString string, imei string) (*ActivationCode, error) {
	acString = strings.TrimPrefix(acString, "LPA:")

	parts := strings.Split(acString, "$")

	if len(parts) < 3 {
		return nil, fmt.Errorf("activation code is invalid")
	}

	acFormat := parts[0]

	if acFormat != "1" {
		return nil, fmt.Errorf("invalid activation code format: %s", acFormat)
	}

	// Add https:// scheme if not present
	smdpAddress := parts[1]
	if !strings.Contains(smdpAddress, "://") {
		smdpAddress = "https://" + smdpAddress
	}
	
	smdp, err := url.Parse(smdpAddress)
	if err != nil {
		return nil, fmt.Errorf("smdp+ address is invalid : %w", err)
	}
	
	// Validate hostname format
	if !strings.Contains(smdp.Host, ".") || smdp.Path != "" || smdp.RawQuery != "" || smdp.Fragment != "" {
		return nil, fmt.Errorf("smdp+ address is not a valid hostname")
	}

	acToken := parts[2]

	oid := ""
	if len(parts) > 3 {
		oid = parts[3]
	}

	confirmationCode := ""
	if len(parts) > 4 {
		confirmationCode = parts[4]
	}

	return &ActivationCode{
		SMDP:             smdp,
		MatchingID:       acToken,
		IMEI:             imei,
		OID:              oid,
		ConfirmationCode: confirmationCode,
	}, nil
}
