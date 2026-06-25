package lpa

import (
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
)

func TestConfirmationCodeRequiredUsesBooleanTag(t *testing.T) {
	client := new(Client)
	signed2 := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(0),
		bertlv.NewValue(bertlv.Universal.Primitive(1), []byte{0xff}),
	)

	assert.True(t, client.confirmationCodeRequired(signed2))
}
