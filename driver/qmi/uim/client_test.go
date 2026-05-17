package uim

import (
	"testing"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

type stubTransport struct {
	called bool
}

func (s *stubTransport) Transmit(*protocol.Request) error {
	s.called = true
	return nil
}

func TestOpenLogicalChannelRejectsOversizedAID(t *testing.T) {
	transport := &stubTransport{}
	client := &Client{Transport: transport, Slot: 1}

	_, err := client.OpenLogicalChannel(make([]byte, maxAIDLength+1))
	if err == nil {
		t.Fatal("OpenLogicalChannel error = nil, want oversized AID error")
	}
	if transport.called {
		t.Fatal("transport should not be called for invalid AID")
	}
}

func TestTransmitRejectsOversizedAPDU(t *testing.T) {
	transport := &stubTransport{}
	client := &Client{Transport: transport, Slot: 1}

	_, err := client.Transmit(make([]byte, maxTransmitAPDUCommandLength+1))
	if err == nil {
		t.Fatal("Transmit error = nil, want oversized APDU error")
	}
	if transport.called {
		t.Fatal("transport should not be called for invalid APDU")
	}
}
