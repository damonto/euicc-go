package driver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/damonto/euicc-go/http/rootci"
)

type LoggingRoundTripper struct {
	transport *http.Transport
	logger    *slog.Logger
}

func NewLoggingRoundTripper(rootci *x509.CertPool, logger *slog.Logger) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		logger: logger,
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootci,
			},
		},
	}
}

func (l *LoggingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	// workaround: Orange PL notification address contains space in the host.
	request.URL.Host = strings.ReplaceAll(request.URL.Host, " ", "")
	request.Body = io.NopCloser(bytes.NewBuffer(body))
	l.logger.Debug("[HTTP] sending request to", "url", request.URL.String(), "body", string(body))
	response, err := l.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	rb, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(bytes.NewBuffer(rb))
	l.logger.Debug("[HTTP] received response from", "url", request.URL.String(), "body", string(rb))
	return response, nil
}

func NewHTTPClient(logger *slog.Logger, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: NewLoggingRoundTripper(
			rootci.TrustedRootCIs(),
			logger,
		),
	}
}
