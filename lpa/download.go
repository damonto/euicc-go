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

// ActivationCode represents the activation code for downloading a profile.
//
// See https://aka.pw/sgp22/v2.5#page=113 (Section 4.1 Activation Code)
type ActivationCode struct {
	SMDP             *url.URL
	MatchingID       string
	IMEI             string
	OID              string
	ConfirmationCode string
}

func (ac *ActivationCode) MarshalText() ([]byte, error) {
	if ac.SMDP == nil {
		return nil, errors.New("SM-DP+ is required")
	}
	b := []byte("LPA:1$")
	b = append(append(b, ac.SMDP.Host...), '$')
	if ac.MatchingID != "" {
		b = append(b, ac.MatchingID...)
	}
	if ac.OID != "" {
		b = append(append(b, '$'), ac.OID...)
	}
	if ac.ConfirmationCode != "" {
		if ac.OID == "" {
			b = append(b, '$')
		}
		b = append(b, '$', '1')
	}
	return b, nil
}

func (ac *ActivationCode) UnmarshalText(text []byte) error {
	if text == nil {
		return errors.New("activation code is required")
	}
	code := string(text)
	if !strings.HasPrefix(code, "LPA:1") {
		return errors.New("invalid activation code format")
	}
	parts := strings.Split(code, "$")
	if len(parts) < 2 {
		return errors.New("invalid activation code format")
	}
	var err error
	if ac.SMDP, err = url.Parse("https://" + parts[1]); err != nil {
		return err
	}
	if len(parts) > 2 {
		ac.MatchingID = parts[2]
	}
	if len(parts) > 3 {
		ac.OID = parts[3]
	}
	return nil
}

func (ac *ActivationCode) validate() error {
	if ac.SMDP == nil || ac.SMDP.Host == "" {
		return errors.New("SM-DP+ is required")
	}
	if ac.IMEI == "" {
		return errors.New("IMEI is required")
	}
	return nil
}

type DownloadStage uint8

const (
	DownloadStageAuthenticateClient DownloadStage = iota
	DownloadStageAuthenticateServer
	DownloadStageInstall
)

// String returns a string representation of the DownloadStage.
func (s DownloadStage) String() string {
	switch s {
	case DownloadStageAuthenticateClient:
		return "Authenticating Client"
	case DownloadStageAuthenticateServer:
		return "Authenticating Server"
	case DownloadStageInstall:
		return "Installing"
	default:
		return fmt.Sprintf("Unknown Stage (%d)", s)
	}
}

// DownloadOptions provides user interaction callbacks during profile download.
type DownloadOptions struct {
	OnProgress              func(stage DownloadStage)
	OnConfirm               func(metadata *sgp22.ProfileInfo) bool
	OnEnterConfirmationCode func() string
}

// DownloadProfile downloads a profile using the provided activation code and options.
func (c *Client) DownloadProfile(ctx context.Context, ac *ActivationCode, opts *DownloadOptions) (*sgp22.LoadBoundProfilePackageResponse, error) {
	if err := ac.validate(); err != nil {
		return nil, err
	}

	if opts != nil && opts.OnProgress != nil {
		opts.OnProgress(DownloadStageAuthenticateClient)
	}

	clientResponse, metadata, ccRequired, err := c.authenticateClient(ac)
	if err != nil {
		if clientResponse != nil && clientResponse.FunctionExecutionStatus().ExecutedSuccess() {
			return nil, c.abort(ac, clientResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
		}
		return nil, err
	}

	if c.isCanceled(ctx) || (opts != nil && opts.OnConfirm != nil && !opts.OnConfirm(metadata)) {
		_, err := c.cancelSession(ac, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}

	if ccRequired && ac.ConfirmationCode == "" {
		if opts != nil && opts.OnEnterConfirmationCode != nil {
			ac.ConfirmationCode = opts.OnEnterConfirmationCode()
		}
		if ac.ConfirmationCode == "" {
			return nil, c.abort(
				ac,
				clientResponse.TransactionID,
				errors.New("confirmation code is required"),
				sgp22.CancelSessionReasonEndUserRejection,
			)
		}
	}

	if opts != nil && opts.OnProgress != nil {
		opts.OnProgress(DownloadStageAuthenticateServer)
	}
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(ac, clientResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	serverResponse, err := c.authenticateServer(ac, clientResponse)
	if err != nil {
		return nil, c.abort(ac, clientResponse.TransactionID, err, sgp22.CancelSessionReasonEndUserRejection)
	}

	if opts != nil && opts.OnProgress != nil {
		opts.OnProgress(DownloadStageInstall)
	}
	if c.isCanceled(ctx) {
		_, err := c.cancelSession(ac, serverResponse.TransactionID, sgp22.CancelSessionReasonEndUserRejection)
		return nil, err
	}
	result, err := c.install(serverResponse)
	if err != nil {
		return result, c.abort(ac, serverResponse.TransactionID, err, sgp22.CancelSessionReasonLoadBppExecutionError)
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
	return &response, response.Valid()
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
	profileInfo := new(sgp22.ProfileInfo)
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

func (c *Client) abort(ac *ActivationCode, transactionID []byte, err error, cancelReason sgp22.CancelSessionReason) error {
	_, cancelErr := c.cancelSession(ac, transactionID, cancelReason)
	if cancelErr != nil {
		return fmt.Errorf("%w (cancel session error: %v)", err, cancelErr)
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
