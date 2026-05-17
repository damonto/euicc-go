package qrtr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

func (c *Conn) findService(serviceType protocol.ServiceType) (*service, error) {
	if err := c.sendControlPacket(serviceType); err != nil {
		return nil, err
	}
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) {
		buf, _, err := c.recvPacketWithDeadline(timeout, c.currentDeadlineSeq())
		if err != nil {
			if errors.Is(err, errReadDeadlineChanged) {
				continue
			}
			if errors.Is(err, os.ErrDeadlineExceeded) {
				break
			}
			return nil, err
		}
		if len(buf) < int(unsafe.Sizeof(controlPacket{})) {
			continue
		}
		if packetType(binary.LittleEndian.Uint32(buf[:4])) != packetTypeNewServer {
			continue
		}
		var service service
		if err := binary.Read(bytes.NewReader(buf[4:]), binary.LittleEndian, &service); err != nil {
			return nil, fmt.Errorf("read QRTR service announcement: %w", err)
		}
		if protocol.ServiceType(service.Service) == serviceType {
			return &service, nil
		}
	}
	return nil, fmt.Errorf("service %d not found", serviceType)
}

func (c *Conn) sendControlPacket(serviceType protocol.ServiceType) error {
	pkt := &controlPacket{
		Command: packetTypeNewLookup,
		Service: service{
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
	addr, err := c.localAddr()
	if err != nil {
		return err
	}
	addr.Port = portControl
	_, err = c.sendTo(addr, buf.Bytes())
	return err
}
