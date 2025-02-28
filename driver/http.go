package driver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/damonto/euicc-go/http/rootci"
)

type LoggingRoundTripper struct {
	transport *http.Transport
}

func NewLoggingRoundTripper(rootci *x509.CertPool) *LoggingRoundTripper {
	return &LoggingRoundTripper{
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
	request.Body = io.NopCloser(bytes.NewBuffer(body))
	slog.Debug("[HTTP] sending request to", "url", request.URL.String(), "body", string(body))

	response, err := l.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	slog.Debug("[HTTP] received response from", "url", request.URL.String(), "body", string(responseBody))
	return response, nil
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: NewLoggingRoundTripper(
			rootci.TrustedRootCIs(),
		),
	}
}
