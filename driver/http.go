package driver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/damonto/euicc-go/http/rootci"
)

// LoggingRoundTripper logs complete HTTP request and response bodies at debug
// level. It preserves the concurrency guarantees of the underlying transport.
type LoggingRoundTripper struct {
	transport http.RoundTripper
	logger    *slog.Logger
}

// NewLoggingRoundTripper returns a transport that trusts rootCAs and logs raw
// HTTP bodies when debug logging is enabled. Logger must not be nil.
func NewLoggingRoundTripper(rootCAs *x509.CertPool, logger *slog.Logger) *LoggingRoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: rootCAs}
	return &LoggingRoundTripper{
		logger:    logger,
		transport: transport,
	}
}

// RoundTrip implements http.RoundTripper.
func (l *LoggingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	transportRequest := request.Clone(request.Context())
	// workaround: Orange PL notification address contains space in the host.
	transportRequest.URL.Host = strings.ReplaceAll(transportRequest.URL.Host, " ", "")

	debug := l.logger.Enabled(request.Context(), slog.LevelDebug)
	if debug {
		body, err := readAndClose(request.Body)
		if err != nil {
			return nil, fmt.Errorf("read HTTP request body: %w", err)
		}
		if request.Body != nil {
			transportRequest.Body = io.NopCloser(bytes.NewReader(body))
		}
		l.logger.DebugContext(request.Context(), "[HTTP] sending request to", "url", transportRequest.URL.String(), "body", string(body))
	}

	response, err := l.transport.RoundTrip(transportRequest)
	if err != nil {
		return nil, err
	}
	if !debug {
		return response, nil
	}

	rb, err := readAndClose(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read HTTP response body: %w", err)
	}
	response.Body = io.NopCloser(bytes.NewReader(rb))
	l.logger.DebugContext(request.Context(), "[HTTP] received response from", "url", transportRequest.URL.String(), "body", string(rb))
	return response, nil
}

func readAndClose(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	data, err := io.ReadAll(body)
	return data, errors.Join(err, body.Close())
}

// CloseIdleConnections closes idle connections held by the underlying
// transport.
func (l *LoggingRoundTripper) CloseIdleConnections() {
	if transport, ok := l.transport.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}

// NewHTTPClient creates an HTTP client configured with the trusted eSIM root
// certificates and raw debug logging. Logger must not be nil.
func NewHTTPClient(logger *slog.Logger, timeout time.Duration) (*http.Client, error) {
	rootCAs, err := rootci.TrustedRootCAs()
	if err != nil {
		return nil, fmt.Errorf("load trusted root CAs: %w", err)
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: NewLoggingRoundTripper(rootCAs, logger),
	}, nil
}
