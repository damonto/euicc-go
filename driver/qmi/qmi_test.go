package qmi

import (
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/core"
	"golang.org/x/sys/unix"
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

func TestNewRejectsInvalidInputsBeforeDial(t *testing.T) {
	if _, err := New("/dev/cdc-wdm1", 0); err == nil {
		t.Fatal("New error = nil, want invalid slot error")
	}

	if _, err := New(strings.Repeat("x", 0x10000), 1); err == nil {
		t.Fatal("New error = nil, want oversized device path error")
	}
}

func TestNewQRTRRejectsInvalidSlotBeforeSocket(t *testing.T) {
	if _, err := NewQRTR(0); err == nil {
		t.Fatal("NewQRTR error = nil, want invalid slot error")
	}
}

func TestQRTRConnBoundarySemantics(t *testing.T) {
	conn := &qrtrConn{}

	if n, err := conn.Read(nil); n != 0 || err != nil {
		t.Fatalf("Read(nil) = %d, %v; want 0, nil", n, err)
	}
	if n, err := conn.Write([]byte{0x01}); n != 0 || err == nil {
		t.Fatalf("Write without service = %d, %v; want 0, error", n, err)
	}
	if _, _, err := conn.recv([]byte{0x01}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("recv on zero conn error = %v, want net.ErrClosed", err)
	}
}

func TestQRTRConnCloseIsStateful(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[1])

	conn := &qrtrConn{fd: fds[0], fdValid: true}
	if err := conn.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	if err := conn.Close(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("second Close error = %v, want net.ErrClosed", err)
	}
}
