package driver

import (
	"bytes"
	"crypto/x509"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type fakeHTTPTransport struct {
	request     *http.Request
	requestBody []byte
	response    *http.Response
	err         error
	called      bool
	idleClosed  bool
}

func (f *fakeHTTPTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	f.called = true
	f.request = request
	if request.Body != nil {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		f.requestBody = body
		_ = request.Body.Close()
	}
	if f.err != nil {
		return nil, f.err
	}
	if f.response == nil {
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	}
	return f.response, nil
}

func (f *fakeHTTPTransport) CloseIdleConnections() {
	f.idleClosed = true
}

type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

type failingBody struct {
	err    error
	closed bool
}

func (b *failingBody) Read([]byte) (int, error) { return 0, b.err }

func (b *failingBody) Close() error {
	b.closed = true
	return nil
}

func debugLogger(output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestLoggingRoundTripperHandlesNilBodyWithoutMutatingRequest(t *testing.T) {
	transport := new(fakeHTTPTransport)
	roundTripper := NewLoggingRoundTripper(x509.NewCertPool(), discardLogger())
	roundTripper.transport = transport
	originalURL := &url.URL{Scheme: "https", Host: "exa mple.com", Path: "/notification"}
	request := &http.Request{Method: http.MethodGet, URL: originalURL}

	response, err := roundTripper.RoundTrip(request)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	_ = response.Body.Close()
	if request.URL.Host != "exa mple.com" {
		t.Fatalf("RoundTrip() mutated original host to %q", request.URL.Host)
	}
	if transport.request.URL.Host != "example.com" {
		t.Fatalf("transport host = %q, want example.com", transport.request.URL.Host)
	}
	if request.Body != nil {
		t.Fatal("RoundTrip() replaced nil body on original request")
	}
}

func TestLoggingRoundTripperLogsAndRestoresRawBodies(t *testing.T) {
	var logs bytes.Buffer
	requestBody := &trackingBody{Reader: strings.NewReader("raw request")}
	transport := &fakeHTTPTransport{response: &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("raw response")),
	}}
	roundTripper := NewLoggingRoundTripper(x509.NewCertPool(), debugLogger(&logs))
	roundTripper.transport = transport
	request, err := http.NewRequest(http.MethodPost, "https://example.com", requestBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	response, err := roundTripper.RoundTrip(request)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll(response.Body) error = %v", err)
	}
	_ = response.Body.Close()
	if !requestBody.closed {
		t.Fatal("RoundTrip() did not close original request body")
	}
	if got := string(transport.requestBody); got != "raw request" {
		t.Fatalf("transport request body = %q, want raw request", got)
	}
	if got := string(responseBody); got != "raw response" {
		t.Fatalf("response body = %q, want raw response", got)
	}
	if output := logs.String(); !strings.Contains(output, "raw request") || !strings.Contains(output, "raw response") {
		t.Fatalf("debug logs do not contain raw bodies: %s", output)
	}
}

func TestLoggingRoundTripperClosesBodiesAfterReadError(t *testing.T) {
	readErr := errors.New("read")
	requestBody := &failingBody{err: readErr}
	transport := new(fakeHTTPTransport)
	roundTripper := NewLoggingRoundTripper(x509.NewCertPool(), debugLogger(io.Discard))
	roundTripper.transport = transport
	request, err := http.NewRequest(http.MethodPost, "https://example.com", requestBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	if _, err := roundTripper.RoundTrip(request); !errors.Is(err, readErr) {
		t.Fatalf("RoundTrip() error = %v, want request read error", err)
	}
	if !requestBody.closed {
		t.Fatal("RoundTrip() did not close request body after read error")
	}
	if transport.called {
		t.Fatal("RoundTrip() called transport after request body read error")
	}

	responseBody := &failingBody{err: readErr}
	transport = &fakeHTTPTransport{response: &http.Response{StatusCode: http.StatusOK, Body: responseBody}}
	roundTripper.transport = transport
	request, err = http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if _, err := roundTripper.RoundTrip(request); !errors.Is(err, readErr) {
		t.Fatalf("RoundTrip() error = %v, want response read error", err)
	}
	if !responseBody.closed {
		t.Fatal("RoundTrip() did not close response body after read error")
	}
}

func TestLoggingRoundTripperUsesDefaultTransportAndClosesIdleConnections(t *testing.T) {
	roundTripper := NewLoggingRoundTripper(x509.NewCertPool(), discardLogger())
	transport, ok := roundTripper.transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", roundTripper.transport)
	}
	if transport.Proxy == nil || transport.IdleConnTimeout == 0 || transport.TLSHandshakeTimeout == 0 {
		t.Fatal("transport did not retain http.DefaultTransport production defaults")
	}

	fake := new(fakeHTTPTransport)
	roundTripper.transport = fake
	roundTripper.CloseIdleConnections()
	if !fake.idleClosed {
		t.Fatal("CloseIdleConnections() was not forwarded")
	}
}
