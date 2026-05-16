package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/driver/qmi/core"
	transport "github.com/damonto/euicc-go/driver/qmi/transport/qrtr"
	"golang.org/x/sys/unix"
)

const (
	qrtrPortControl = 0xfffffffe
)

type qrtrPacketType uint32

const (
	qrtrPacketTypeData qrtrPacketType = iota + 1
	qrtrPacketTypeHello
	qrtrPacketTypeBye
	qrtrPacketTypeNewServer
	qrtrPacketTypeDelServer
	qrtrPacketTypeDelClient
	qrtrPacketTypeResumeTx
	qrtrPacketTypeExit
	qrtrPacketTypePing
	qrtrPacketTypeNewLookup
	qrtrPacketTypeDelLookup
)

type qrtrSockAddr struct {
	Family uint16
	Node   uint32
	Port   uint32
}

type qrtrControlPacket struct {
	Command qrtrPacketType
	Service qrtrService
}

type qrtrService struct {
	Service  uint32
	Instance uint32
	Node     uint32
	Port     uint32
}

type qrtrConn struct {
	mu          sync.Mutex
	fd          int
	fdValid     bool
	service     *qrtrService
	readTimeout time.Duration
}

// QRTR implements the apdu.SmartCardChannel interface using QRTR protocol
type QRTR struct {
	conn *qrtrConn
	core.QMIClient
}

// NewQRTR creates a new QRTR connection to the UIM service
func NewQRTR(slot uint8) (apdu.SmartCardChannel, error) {
	if slot == 0 {
		return nil, errors.New("slot must be >= 1")
	}

	conn, err := newQRTRConn()
	if err != nil {
		return nil, err
	}
	q := &QRTR{
		conn: conn,
		QMIClient: core.QMIClient{
			Transport: transport.New(conn),
			Slot:      slot,
		},
	}
	q.conn.service, err = q.findService(core.QMIServiceUIM)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return q, nil
}

func (c *QRTR) findService(serviceType core.ServiceType) (*qrtrService, error) {
	if err := c.sendControlPacket(serviceType); err != nil {
		return nil, err
	}
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) {
		buf, _, err := c.conn.recvPacketWithTimeout(time.Until(timeout))
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				break
			}
			return nil, err
		}
		if len(buf) < int(unsafe.Sizeof(qrtrControlPacket{})) {
			continue
		}
		if qrtrPacketType(binary.LittleEndian.Uint32(buf[:4])) != qrtrPacketTypeNewServer {
			continue
		}
		var service qrtrService
		if err := binary.Read(bytes.NewReader(buf[4:]), binary.LittleEndian, &service); err != nil {
			return nil, fmt.Errorf("read QRTR service announcement: %w", err)
		}
		if core.ServiceType(service.Service) == serviceType {
			return &service, nil
		}
	}
	return nil, fmt.Errorf("service %d not found", serviceType)
}

func (c *QRTR) sendControlPacket(serviceType core.ServiceType) error {
	pkt := &qrtrControlPacket{
		Command: qrtrPacketTypeNewLookup,
		Service: qrtrService{
			Service:  uint32(serviceType),
			Instance: 0,
			Node:     0,
			Port:     0,
		},
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, pkt); err != nil {
		return fmt.Errorf("write QRTR control packet: %w", err)
	}
	addr, err := c.conn.localAddr()
	if err != nil {
		return err
	}
	addr.Port = qrtrPortControl
	_, err = c.conn.sendTo(addr, buf.Bytes())
	return err
}

func (c *QRTR) Disconnect() error {
	return c.conn.Close()
}

func newQRTRConn() (*qrtrConn, error) {
	fd, err := unix.Socket(unix.AF_QIPCRTR, unix.SOCK_DGRAM, 0)
	if err != nil {
		return nil, fmt.Errorf("create QRTR socket: %w", err)
	}
	return &qrtrConn{fd: fd, fdValid: true, readTimeout: 30 * time.Second}, nil
}

func (c *qrtrConn) sendTo(dest *qrtrSockAddr, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("data is empty")
	}
	fd, err := c.currentFD()
	if err != nil {
		return 0, err
	}
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

func (c *qrtrConn) recvFrom(buf []byte) (int, *qrtrSockAddr, error) {
	if len(buf) == 0 {
		return 0, nil, nil
	}

	fd, err := c.currentFD()
	if err != nil {
		return 0, nil, err
	}
	var addr qrtrSockAddr
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

func (c *qrtrConn) recv(b []byte) (int, *qrtrSockAddr, error) {
	return c.recvWithTimeout(b, c.readTimeout)
}

func (c *qrtrConn) recvWithTimeout(b []byte, timeout time.Duration) (int, *qrtrSockAddr, error) {
	if len(b) == 0 {
		return 0, nil, nil
	}

	packet, from, err := c.recvPacketWithTimeout(timeout)
	if err != nil {
		return 0, nil, err
	}
	if len(packet) > len(b) {
		return 0, nil, fmt.Errorf("QRTR message size %d exceeds read buffer %d", len(packet), len(b))
	}
	return copy(b, packet), from, nil
}

func (c *qrtrConn) recvPacketWithTimeout(timeout time.Duration) ([]byte, *qrtrSockAddr, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return nil, nil, os.ErrDeadlineExceeded
		}

		fd, err := c.currentFD()
		if err != nil {
			return nil, nil, err
		}
		pollTimeout := -1
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return nil, nil, os.ErrDeadlineExceeded
			}
			if remaining > time.Second {
				remaining = time.Second
			}
			pollTimeout = durationMillis(remaining)
		}
		if err := waitReadable(fd, pollTimeout); err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return nil, nil, err
		}

		size, err := nextDatagramSize(fd)
		if err != nil {
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
		n, from, err := c.recvFrom(packet)
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

func waitReadable(fd int, timeoutMillis int) error {
	pollFDs := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	n, err := unix.Poll(pollFDs, timeoutMillis)
	if err != nil {
		return err
	}
	if n == 0 {
		return os.ErrDeadlineExceeded
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

func (c *qrtrConn) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	for {
		n, from, err := c.recv(b)
		if err != nil {
			return 0, err
		}
		if from.Port == qrtrPortControl {
			continue
		}
		if c.service != nil && (from.Node != c.service.Node || from.Port != c.service.Port) {
			continue
		}
		return n, nil
	}
}

func (c *qrtrConn) Write(b []byte) (int, error) {
	if c.service == nil {
		return 0, errors.New("QRTR service is not set")
	}
	return c.sendTo(&qrtrSockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   c.service.Node,
		Port:   c.service.Port,
	}, b)
}

func (c *qrtrConn) Close() error {
	c.mu.Lock()
	if !c.fdValid {
		c.mu.Unlock()
		return net.ErrClosed
	}
	fd := c.fd
	c.fd = -1
	c.fdValid = false
	c.mu.Unlock()

	return unix.Close(fd)
}

func (c *qrtrConn) localAddr() (*qrtrSockAddr, error) {
	fd, err := c.currentFD()
	if err != nil {
		return nil, err
	}
	addr := &qrtrSockAddr{}
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

func (c *qrtrConn) currentFD() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fdValid {
		return -1, net.ErrClosed
	}
	return c.fd, nil
}
