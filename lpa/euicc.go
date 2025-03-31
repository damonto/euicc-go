package lpa

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/http"
	sgp22 "github.com/damonto/euicc-go/v2"
)

// GSMAISDRApplicationAID is the AID of the GSMA SGP.02 ISD-R application.
// See https://www.gsma.com/solutions-and-impact/technologies/esim/wp-content/uploads/2020/07/SGP.02-v4.2.pdf#page=26 (Section 2.2.3 Identification of Security Domains: AID and TAR)
var GSMAISDRApplicationAID = []byte{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x00, 0x01, 0x00}

// Client is the main structure for the LPA client.
type Client struct {
	HTTP *http.Client
	APDU sgp22.Transmitter

	transmitter driver.Transmitter
}

// Option is the configuration for the LPA client.
// It includes the channel for APDU communication, logger, AID, maximum APDU size (MSS), admin protocol version, and timeout.
type Option struct {
	// Channel is the channel for APDU communication. It is required for APDU communication.
	Channel apdu.SmartCardChannel
	// AID is the application identifier for the GSMA ISD-R application. It defaults to GSMA ISD-R Application AID.
	AID []byte
	// MSS is the maximum APDU size. It defaults to 254.
	MSS int
	// AdminProtocolVersion is the version of the admin protocol. It defaults to "v2.2.2".
	AdminProtocolVersion string
	// Logger is the logger for the LPA client. It defaults to slog.Default().
	Logger *slog.Logger
	// Timeout is the timeout for the HTTP client. It defaults to 30 seconds.
	Timeout time.Duration
}

func (opt *Option) validateAdminProtocolVersion() error {
	if opt.AdminProtocolVersion == "" {
		opt.AdminProtocolVersion = "2.2.2"
	}
	// If the version starts with "v", remove it
	if opt.AdminProtocolVersion[0] == 'v' {
		opt.AdminProtocolVersion = opt.AdminProtocolVersion[1:]
	}
	// Currently only v2.x.x is supported
	if opt.AdminProtocolVersion[0] != '2' {
		return fmt.Errorf("unsupported admin protocol version: %s", opt.AdminProtocolVersion)
	}
	return nil
}

func (opt *Option) validateMSS() error {
	if opt.MSS == 0 {
		opt.MSS = 254
	}
	if opt.MSS < 0 || opt.MSS > 254 {
		return fmt.Errorf("invalid maximum APDU size: %d", opt.MSS)
	}
	return nil
}

func (opt *Option) Validate() error {
	if opt.AID == nil {
		opt.AID = GSMAISDRApplicationAID
	}
	if err := opt.validateMSS(); err != nil {
		return err
	}
	if err := opt.validateAdminProtocolVersion(); err != nil {
		return err
	}
	if opt.Channel == nil {
		return errors.New("channel is required for APDU communication")
	}
	if opt.Logger == nil {
		opt.Logger = slog.Default()
	}
	if opt.Timeout == 0 {
		opt.Timeout = 30 * time.Second
	}
	return nil
}

// New creates a new LPA client with the given options.
func New(opt *Option) (*Client, error) {
	var c Client
	var err error
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	if c.transmitter, err = driver.NewTransmitter(opt.Logger, opt.Channel, opt.AID, opt.MSS); err != nil {
		return nil, err
	}
	c.APDU = c.transmitter
	c.HTTP = &http.Client{
		Client:        driver.NewHTTPClient(opt.Logger, opt.Timeout),
		AdminProtocol: fmt.Sprintf("gsma/rsp/v%s", opt.AdminProtocolVersion),
	}
	return &c, nil
}

// Close closes the LPA client and the underlying APDU transmitter.
// You should call this method when you are done using the client to release resources.
func (c *Client) Close() error {
	return c.transmitter.Close()
}
