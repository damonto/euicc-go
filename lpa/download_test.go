package lpa

import (
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestConfirmationCodeRequiredUsesBooleanTag(t *testing.T) {
	client := new(Client)
	signed2 := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(0),
		bertlv.NewValue(bertlv.Universal.Primitive(1), []byte{0xff}),
	)

	got, err := client.confirmationCodeRequired(signed2)
	if err != nil {
		t.Fatalf("confirmationCodeRequired() error = %v", err)
	}
	if !got {
		t.Error("confirmationCodeRequired() = false, want true")
	}
}
