package uim

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

func TestGetSlotStatusParsesVariableLengthICCID(t *testing.T) {
	value := new(bytes.Buffer)
	value.WriteByte(1)
	mustWrite(t, value, PhysicalCardStatePresent)
	mustWrite(t, value, SlotStateActive)
	value.WriteByte(1)
	value.WriteByte(3)
	value.Write([]byte{0x89, 0x67, 0x45})

	var response GetSlotStatusResponse
	err := response.UnmarshalResponse(&protocol.TLVs{
		{Type: 0x10, Len: uint16(value.Len()), Value: value.Bytes()},
	})
	if err != nil {
		t.Fatalf("UnmarshalResponse failed: %v", err)
	}
	if response.ActivatedSlot != 1 {
		t.Fatalf("ActivatedSlot = %d, want 1", response.ActivatedSlot)
	}
	if got, want := response.Slots[0].ICCID, []byte{0x89, 0x67, 0x45}; !bytes.Equal(got, want) {
		t.Fatalf("ICCID = %X, want %X", got, want)
	}
}

func TestGetCardStatusParsesVariableLengthApplications(t *testing.T) {
	value := new(bytes.Buffer)
	mustWrite(t, value, uint16(0))
	mustWrite(t, value, uint16(0))
	mustWrite(t, value, uint16(0))
	mustWrite(t, value, uint16(0))
	value.WriteByte(1) // cards
	value.WriteByte(byte(CardStatePresent))
	value.Write([]byte{0x00, 0x00, 0x00, 0x00}) // UPIN state/retries and error.
	value.WriteByte(2)                          // applications

	writeCardApplication(t, value, CardApplicationTypeSIM, CardApplicationStateDetected, []byte{0xa0, 0x01})
	writeCardApplication(t, value, CardApplicationTypeUSIM, CardApplicationStateReady, []byte{0xa0, 0x00, 0x00})

	var response GetCardStatusResponse
	err := response.UnmarshalResponse(&protocol.TLVs{
		{Type: 0x10, Len: uint16(value.Len()), Value: value.Bytes()},
	})
	if err != nil {
		t.Fatalf("UnmarshalResponse failed: %v", err)
	}
	if !response.Ready() {
		t.Fatal("Ready() = false, want true")
	}
}

func writeCardApplication(t *testing.T, w *bytes.Buffer, typ CardApplicationType, state CardApplicationState, aid []byte) {
	t.Helper()

	w.WriteByte(byte(typ))
	w.WriteByte(byte(state))
	w.Write([]byte{0x00, 0x00, 0x00, 0x00}) // personalization state/feature/retries.
	w.WriteByte(byte(len(aid)))
	w.Write(aid)
	w.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) // UPIN/PIN status.
}

func mustWrite(t *testing.T, w *bytes.Buffer, value any) {
	t.Helper()
	if err := binary.Write(w, binary.LittleEndian, value); err != nil {
		t.Fatalf("binary.Write failed: %v", err)
	}
}
