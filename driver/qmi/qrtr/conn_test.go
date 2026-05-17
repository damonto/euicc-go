package qrtr

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestConnBoundarySemantics(t *testing.T) {
	conn := &Conn{}

	if n, err := conn.Read(nil); n != 0 || err != nil {
		t.Fatalf("Read(nil) = %d, %v; want 0, nil", n, err)
	}
	if n, err := conn.Write([]byte{0x01}); n != 0 || err == nil {
		t.Fatalf("Write without service = %d, %v; want 0, error", n, err)
	}
	if _, err := conn.Read([]byte{0x01}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Read on zero conn error = %v, want net.ErrClosed", err)
	}
	if err := conn.SetReadDeadline(time.Now()); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("SetReadDeadline on zero conn error = %v, want net.ErrClosed", err)
	}
}

func TestConnCloseIsStateful(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}

	conn := newConnWithFDs(fds[0], wakeFD)
	if err := conn.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	if err := conn.Close(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("second Close error = %v, want net.ErrClosed", err)
	}
}

func TestConnSetReadDeadline(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}
	defer unix.Close(wakeFD)

	conn := newConnWithFDs(fds[0], wakeFD)
	deadline := time.Now().Add(time.Second)
	if err := conn.SetReadDeadline(deadline); err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}
	if got, _ := conn.currentReadDeadline(); !got.Equal(deadline) {
		t.Fatalf("read deadline = %s, want %s", got, deadline)
	}
}

func TestWaitReadableWakesOnDeadlineChange(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}
	defer unix.Close(wakeFD)

	done := make(chan error, 1)
	go func() {
		done <- waitReadable(fds[0], wakeFD, true, -1)
	}()

	time.Sleep(10 * time.Millisecond)
	if _, err := unix.Write(wakeFD, []byte{1, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		t.Fatalf("write eventfd: %v", err)
	}

	select {
	case err := <-done:
		if !errors.Is(err, errReadDeadlineChanged) {
			t.Fatalf("waitReadable error = %v, want errReadDeadlineChanged", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waitReadable did not wake")
	}
}

func TestReadDeadlineCanWakeBlockedRead(t *testing.T) {
	conn, err := newConn()
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("QRTR socket unavailable: %v", err)
		}
		t.Skipf("QRTR socket unavailable: %v", err)
	}
	defer conn.Close()

	done := make(chan error, 1)
	go func() {
		_, err := conn.Read(make([]byte, 1))
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	if err := conn.SetReadDeadline(time.Now()); err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}

	select {
	case err := <-done:
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("Read error = %v, want deadline exceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Read did not wake after deadline change")
	}
}

func TestCloseWakesBlockedRead(t *testing.T) {
	conn, err := newConn()
	if err != nil {
		t.Skipf("QRTR socket unavailable: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := conn.Read(make([]byte, 1))
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	if err := conn.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	select {
	case err := <-done:
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("Read error = %v, want net.ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Read did not wake after Close")
	}
}

func TestCloseWaitsForActiveOperation(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}

	conn := newConnWithFDs(fds[0], wakeFD)
	conn.activeOps = 1

	done := make(chan error, 1)
	go func() {
		done <- conn.Close()
	}()

	select {
	case <-done:
		t.Fatal("Close returned before active operation was released")
	case <-time.After(10 * time.Millisecond):
	}

	conn.releaseFD()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Close error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not return after active operation was released")
	}
}

func TestSetReadDeadlineDoesNotWakeWithoutPollWaiter(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}
	defer unix.Close(wakeFD)

	conn := newConnWithFDs(fds[0], wakeFD)
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}

	var buf [8]byte
	if _, err := unix.Read(wakeFD, buf[:]); !errors.Is(err, unix.EAGAIN) {
		t.Fatalf("wakeFD read error = %v, want EAGAIN", err)
	}
}

func TestExpiredPacketDeadlineReleasesActiveOperation(t *testing.T) {
	fds := make([]int, 2)
	if err := unix.Pipe(fds); err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer unix.Close(fds[1])

	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		t.Fatalf("create eventfd: %v", err)
	}

	conn := newConnWithFDs(fds[0], wakeFD)
	_, _, err = conn.recvPacketWithDeadline(time.Now().Add(-time.Millisecond), conn.currentDeadlineSeq())
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("recvPacketWithDeadline error = %v, want deadline exceeded", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- conn.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Close error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not return after expired read")
	}
}
