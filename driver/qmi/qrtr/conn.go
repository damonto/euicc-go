package qrtr

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
	"golang.org/x/sys/unix"
)

type Conn struct {
	mu           sync.Mutex
	cond         *sync.Cond
	activeOps    int
	fd           int
	fdValid      bool
	wakeFD       int
	wakeFDValid  bool
	service      *service
	readDeadline time.Time
	deadlineSeq  uint64
	pollWaiters  int
}

var errReadDeadlineChanged = errors.New("read deadline changed")

func Open(serviceType protocol.ServiceType) (*Conn, error) {
	conn, err := newConn()
	if err != nil {
		return nil, err
	}
	conn.service, err = conn.findService(serviceType)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func newConn() (*Conn, error) {
	fd, err := unix.Socket(unix.AF_QIPCRTR, unix.SOCK_DGRAM, 0)
	if err != nil {
		return nil, fmt.Errorf("create QRTR socket: %w", err)
	}
	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("create QRTR wake eventfd: %w", err)
	}
	return newConnWithFDs(fd, wakeFD), nil
}

func newConnWithFDs(fd int, wakeFD int) *Conn {
	conn := &Conn{fd: fd, fdValid: true, wakeFD: wakeFD, wakeFDValid: true}
	conn.cond = sync.NewCond(&conn.mu)
	return conn
}

func (c *Conn) sendTo(dest *sockAddr, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("data is empty")
	}
	fd, release, err := c.acquireFD()
	if err != nil {
		return 0, err
	}
	defer release()

	n, _, errno := unix.Syscall6(unix.SYS_SENDTO,
		uintptr(fd),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		0,
		uintptr(unsafe.Pointer(dest)),
		uintptr(unsafe.Sizeof(*dest)))
	if errno != 0 {
		return 0, fmt.Errorf("send data: %w", errno)
	}
	if int(n) != len(data) {
		return int(n), io.ErrShortWrite
	}
	return int(n), nil
}

func recvFrom(fd int, buf []byte) (int, *sockAddr, error) {
	if len(buf) == 0 {
		return 0, nil, nil
	}

	var addr sockAddr
	addrLen := uintptr(unsafe.Sizeof(addr))
	n, _, errno := unix.Syscall6(unix.SYS_RECVFROM,
		uintptr(fd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0,
		uintptr(unsafe.Pointer(&addr)),
		uintptr(unsafe.Pointer(&addrLen)))
	if errno != 0 {
		return 0, nil, fmt.Errorf("receive data: %w", errno)
	}
	return int(n), &addr, nil
}

func (c *Conn) recvWithDeadline(b []byte, deadline time.Time, deadlineSeq uint64) (int, *sockAddr, error) {
	if len(b) == 0 {
		return 0, nil, nil
	}

	packet, from, err := c.recvPacketWithDeadline(deadline, deadlineSeq)
	if err != nil {
		return 0, nil, err
	}
	if len(packet) > len(b) {
		return 0, nil, fmt.Errorf("QRTR message size %d exceeds read buffer %d", len(packet), len(b))
	}
	return copy(b, packet), from, nil
}

func (c *Conn) recvPacketWithDeadline(deadline time.Time, deadlineSeq uint64) ([]byte, *sockAddr, error) {
	for {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return nil, nil, os.ErrDeadlineExceeded
		}

		fd, wakeFD, wakeFDValid, release, err := c.acquireFDs()
		if err != nil {
			return nil, nil, err
		}
		pollTimeout := -1
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				release()
				return nil, nil, os.ErrDeadlineExceeded
			}
			if remaining > time.Second {
				remaining = time.Second
			}
			pollTimeout = durationMillis(remaining)
		}
		unregisterPoll, err := c.registerPoll(deadlineSeq)
		if err != nil {
			release()
			return nil, nil, err
		}
		if err := waitReadable(fd, wakeFD, wakeFDValid, pollTimeout); err != nil {
			unregisterPoll()
			release()
			if errors.Is(err, errReadDeadlineChanged) {
				return nil, nil, err
			}
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return nil, nil, err
		}
		unregisterPoll()

		size, err := nextDatagramSize(fd)
		if err != nil {
			release()
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				continue
			}
			return nil, nil, err
		}
		readSize := size
		if readSize == 0 {
			readSize = 1
		}

		packet := make([]byte, readSize)
		n, from, err := recvFrom(fd, packet)
		release()
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EINTR) {
				continue
			}
			return nil, nil, err
		}
		if n != size {
			return nil, nil, fmt.Errorf("unexpected QRTR message size: got %d bytes, expected %d", n, size)
		}
		return packet[:n], from, nil
	}
}

func nextDatagramSize(fd int) (int, error) {
	size, err := unix.IoctlGetInt(fd, unix.TIOCINQ)
	if err != nil {
		return 0, fmt.Errorf("get QRTR datagram size: %w", err)
	}
	return size, nil
}

func waitReadable(fd int, wakeFD int, wakeFDValid bool, timeoutMillis int) error {
	pollFDs := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	if wakeFDValid {
		pollFDs = append(pollFDs, unix.PollFd{Fd: int32(wakeFD), Events: unix.POLLIN})
	}
	n, err := unix.Poll(pollFDs, timeoutMillis)
	if err != nil {
		return err
	}
	if n == 0 {
		return os.ErrDeadlineExceeded
	}

	if wakeFDValid && pollFDs[1].Revents&unix.POLLIN != 0 {
		drainWakeFD(wakeFD)
		return errReadDeadlineChanged
	}

	revents := pollFDs[0].Revents
	if revents&unix.POLLNVAL != 0 {
		return net.ErrClosed
	}
	if revents&(unix.POLLERR|unix.POLLHUP) != 0 {
		return fmt.Errorf("QRTR socket poll failed: revents=0x%X", revents)
	}
	if revents&unix.POLLIN == 0 {
		return os.ErrDeadlineExceeded
	}
	return nil
}

func durationMillis(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	ms := d / time.Millisecond
	if d%time.Millisecond != 0 {
		ms++
	}
	if ms < 1 {
		return 1
	}
	const maxInt32 time.Duration = 1<<31 - 1
	if ms > maxInt32 {
		return int(maxInt32)
	}
	return int(ms)
}

func (c *Conn) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	for {
		deadline, deadlineSeq := c.currentReadDeadline()
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return 0, os.ErrDeadlineExceeded
		}

		n, from, err := c.recvWithDeadline(b, deadline, deadlineSeq)
		if err != nil {
			if errors.Is(err, errReadDeadlineChanged) {
				continue
			}
			return 0, err
		}
		if from.Port == portControl {
			continue
		}
		if c.service != nil && (from.Node != c.service.Node || from.Port != c.service.Port) {
			continue
		}
		return n, nil
	}
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fdValid {
		return net.ErrClosed
	}
	c.readDeadline = t
	c.deadlineSeq++
	if c.pollWaiters > 0 {
		return notifyWakeFD(c.wakeFD, c.wakeFDValid)
	}
	return nil
}

func (c *Conn) Write(b []byte) (int, error) {
	if c.service == nil {
		return 0, errors.New("QRTR service is not set")
	}
	return c.sendTo(&sockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   c.service.Node,
		Port:   c.service.Port,
	}, b)
}

func (c *Conn) Close() error {
	c.mu.Lock()
	if !c.fdValid {
		c.mu.Unlock()
		return net.ErrClosed
	}
	fd := c.fd
	wakeFD := c.wakeFD
	wakeFDValid := c.wakeFDValid
	if c.pollWaiters > 0 {
		if err := notifyWakeFD(wakeFD, wakeFDValid); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("wake QRTR reader: %w", err)
		}
	}
	c.fd = -1
	c.fdValid = false
	c.wakeFD = -1
	c.wakeFDValid = false
	c.mu.Unlock()

	c.waitIdle()

	var errs []error
	if err := unix.Close(fd); err != nil {
		errs = append(errs, err)
	}
	if wakeFDValid {
		if err := unix.Close(wakeFD); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *Conn) localAddr() (*sockAddr, error) {
	fd, release, err := c.acquireFD()
	if err != nil {
		return nil, err
	}
	defer release()

	addr := &sockAddr{}
	addrLen := uintptr(unsafe.Sizeof(*addr))
	_, _, errno := unix.Syscall6(unix.SYS_GETSOCKNAME,
		uintptr(fd),
		uintptr(unsafe.Pointer(addr)),
		uintptr(unsafe.Pointer(&addrLen)),
		0,
		0,
		0)
	if errno != 0 {
		return nil, fmt.Errorf("get QRTR socket name: %w", errno)
	}
	if addrLen != uintptr(unsafe.Sizeof(*addr)) || addr.Family != unix.AF_QIPCRTR {
		return nil, fmt.Errorf("unexpected QRTR socket address family %d length %d", addr.Family, addrLen)
	}
	return addr, nil
}

func (c *Conn) acquireFD() (int, func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fdValid {
		return -1, nil, net.ErrClosed
	}
	c.activeOps++
	return c.fd, c.releaseFD, nil
}

func (c *Conn) acquireFDs() (int, int, bool, func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fdValid {
		return -1, -1, false, nil, net.ErrClosed
	}
	c.activeOps++
	return c.fd, c.wakeFD, c.wakeFDValid, c.releaseFD, nil
}

func (c *Conn) currentReadDeadline() (time.Time, uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readDeadline, c.deadlineSeq
}

func (c *Conn) currentDeadlineSeq() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deadlineSeq
}

func (c *Conn) releaseFD() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activeOps--
	if c.activeOps == 0 && c.cond != nil {
		c.cond.Broadcast()
	}
}

func (c *Conn) waitIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.activeOps > 0 {
		c.cond.Wait()
	}
}

func (c *Conn) registerPoll(deadlineSeq uint64) (func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fdValid {
		return nil, net.ErrClosed
	}
	if c.deadlineSeq != deadlineSeq {
		return nil, errReadDeadlineChanged
	}
	c.pollWaiters++
	return c.unregisterPoll, nil
}

func (c *Conn) unregisterPoll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pollWaiters--
}

func notifyWakeFD(fd int, valid bool) error {
	if !valid {
		return nil
	}
	for {
		_, err := unix.Write(fd, []byte{1, 0, 0, 0, 0, 0, 0, 0})
		if err == nil || errors.Is(err, unix.EAGAIN) {
			return nil
		}
		if !errors.Is(err, unix.EINTR) {
			return err
		}
	}
}

func drainWakeFD(fd int) {
	var buf [8]byte
	for {
		_, err := unix.Read(fd, buf[:])
		if err == nil {
			continue
		}
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
			return
		}
		return
	}
}
