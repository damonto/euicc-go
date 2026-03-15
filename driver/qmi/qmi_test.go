package qmi

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

type fakeTransport struct {
	called bool
	err    error
}

func (f *fakeTransport) Transmit(*core.Request) error {
	f.called = true
	return f.err
}

type fakeConn struct {
	closed   bool
	closeErr error
}

func (c *fakeConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (c *fakeConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *fakeConn) Close() error                     { c.closed = true; return c.closeErr }
func (c *fakeConn) LocalAddr() net.Addr              { return nil }
func (c *fakeConn) RemoteAddr() net.Addr             { return nil }
func (c *fakeConn) SetDeadline(time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(time.Time) error { return nil }

func TestDisconnectClosesConnectionWhenReleaseFails(t *testing.T) {
	releaseErr := errors.New("release client id")
	closeErr := errors.New("close conn")
	transport := &fakeTransport{err: releaseErr}
	conn := &fakeConn{closeErr: closeErr}
	q := &QMI{
		conn: conn,
		QMIClient: core.QMIClient{
			Transport: transport,
			ClientID:  7,
		},
	}

	err := q.Disconnect()
	if !transport.called {
		t.Fatal("releaseClientID was not called")
	}
	if !conn.closed {
		t.Fatal("connection was not closed")
	}
	if !errors.Is(err, releaseErr) {
		t.Fatalf("disconnect error %v does not include release error", err)
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("disconnect error %v does not include close error", err)
	}
}
