package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
	"unsafe"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/driver/qmi/core"
	transport "github.com/damonto/euicc-go/driver/qmi/transport/qrtr"
	"golang.org/x/sys/unix"
)

var _ net.Conn = (*QRTRConn)(nil)

const (
	// QRTR node and port constants
	QRTRNodeBroadcast = 0xffffffff
	QRTRPortControl   = 0xfffffffe
)

type QRTRPacketType uint32

const (
	QRTRPacketTypeData QRTRPacketType = iota + 1
	QRTRPacketTypeHello
	QRTRPacketTypeBye
	QRTRPacketTypeNewServer
	QRTRPacketTypeDelServer
	QRTRPacketTypeDelClient
	QRTRPacketTypeResumeTx
	QRTRPacketTypeExit
	QRTRPacketTypePing
	QRTRPacketTypeNewLookup
	QRTRPacketTypeDelLookup
)

// SockAddr represents a QRTR socket address
type SockAddr struct {
	Family uint16
	Node   uint32
	Port   uint32
}

func (s SockAddr) Network() string {
	return "qrtr"
}

func (s SockAddr) String() string {
	return fmt.Sprintf("qrtr://%d:%d/%d", unix.AF_QIPCRTR, s.Node, s.Port)
}

// ControlPacket represents a QRTR control packet
type ControlPacket struct {
	Command QRTRPacketType
	Service Service
}

// Service represents a QRTR service
type Service struct {
	Service  uint32
	Instance uint32
	Node     uint32
	Port     uint32
}

type QRTRConn struct {
	fd          int
	Service     *Service
	readTimeout time.Duration
}

// QRTR implements the apdu.SmartCardChannel interface using QRTR protocol
type QRTR struct {
	conn *QRTRConn
	core.QMIClient
}

// NewQRTR creates a new QRTR connection to the UIM service
func NewQRTR(slot uint8) (apdu.SmartCardChannel, error) {
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
	q.conn.Service, err = q.findService(core.QMIServiceUIM)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return q, nil
}

func (c *QRTR) findService(serviceType core.ServiceType) (*Service, error) {
	if err := c.sendControlPacket(serviceType); err != nil {
		return nil, err
	}
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) {
		buf := make([]byte, 1024)
		n, _, err := c.conn.Recv(buf)
		if err != nil {
			return nil, err
		}
		if QRTRPacketType(binary.LittleEndian.Uint32(buf[:4])) != QRTRPacketTypeNewServer {
			continue
		}
		var service Service
		binary.Read(bytes.NewReader(buf[4:n]), binary.LittleEndian, &service)
		if core.ServiceType(service.Service) == serviceType {
			return &service, nil
		}
	}
	return nil, fmt.Errorf("service %d not found", serviceType)
}

func (c *QRTR) sendControlPacket(serviceType core.ServiceType) error {
	pkt := &ControlPacket{
		Command: QRTRPacketTypeNewLookup,
		Service: Service{
			Service:  uint32(serviceType),
			Instance: 0,
			Node:     0,
			Port:     0,
		},
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pkt)
	_, err := c.conn.Sendto(&SockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   QRTRNodeBroadcast,
		Port:   QRTRPortControl,
	}, buf.Bytes())
	return err
}

func (c *QRTR) Disconnect() error {
	return c.conn.Close()
}

func newQRTRConn() (*QRTRConn, error) {
	fd, err := unix.Socket(unix.AF_QIPCRTR, unix.SOCK_DGRAM, 0)
	if err != nil {
		return nil, fmt.Errorf("create QRTR socket: %w", err)
	}
	return &QRTRConn{fd: fd, readTimeout: 30 * time.Second}, nil
}

func (c *QRTRConn) Sendto(dest *SockAddr, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("data is empty")
	}
	n, _, errno := unix.Syscall6(unix.SYS_SENDTO,
		uintptr(c.fd),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		0,
		uintptr(unsafe.Pointer(dest)),
		uintptr(unsafe.Sizeof(*dest)))
	if errno != 0 {
		return 0, fmt.Errorf("send data: %w", errno)
	}
	return int(n), nil
}

func (c *QRTRConn) Recvfrom(buf []byte) (int, *SockAddr, error) {
	var addr SockAddr
	addrLen := uintptr(unsafe.Sizeof(addr))
	n, _, errno := unix.Syscall6(unix.SYS_RECVFROM,
		uintptr(c.fd),
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

func (c *QRTRConn) Recv(b []byte) (int, *SockAddr, error) {
	tv := unix.NsecToTimeval((1 * time.Second).Nanoseconds())
	if err := unix.SetsockoptTimeval(c.fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv); err != nil {
		return 0, nil, err
	}

	timeout := time.Now().Add(c.readTimeout)
	for time.Now().Before(timeout) {
		n, from, err := c.Recvfrom(b)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return 0, nil, err
		}
		return n, from, nil
	}
	return 0, nil, os.ErrDeadlineExceeded
}

func (c *QRTRConn) Read(b []byte) (int, error) {
	for {
		n, from, err := c.Recv(b)
		if err != nil {
			return 0, err
		}
		if from.Port == QRTRPortControl {
			continue
		}
		if c.Service != nil && (from.Node != c.Service.Node || from.Port != c.Service.Port) {
			continue
		}
		return n, nil
	}
}

func (c *QRTRConn) Write(b []byte) (int, error) {
	return c.Sendto(&SockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   c.Service.Node,
		Port:   c.Service.Port,
	}, b)
}

func (c *QRTRConn) Close() error {
	return unix.Close(c.fd)
}

func (c *QRTRConn) LocalAddr() net.Addr {
	return &SockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   c.Service.Node,
		Port:   c.Service.Port,
	}
}

func (c *QRTRConn) RemoteAddr() net.Addr {
	return &SockAddr{
		Family: unix.AF_QIPCRTR,
		Node:   c.Service.Node,
		Port:   c.Service.Port,
	}
}

func (c *QRTRConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *QRTRConn) SetReadDeadline(t time.Time) error {
	c.readTimeout = c.toTimeDuration(t)
	return nil
}

func (c *QRTRConn) SetWriteDeadline(t time.Time) error {
	tv := unix.NsecToTimeval(c.toTimeDuration(t).Nanoseconds())
	return unix.SetsockoptTimeval(c.fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, &tv)
}

func (c *QRTRConn) toTimeDuration(t time.Time) time.Duration {
	if t.IsZero() {
		return 0
	}
	d := time.Until(t)
	if d <= 0 {
		return time.Microsecond
	}
	return d
}
