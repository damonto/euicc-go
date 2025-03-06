package sgp22

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestICCID_String(t *testing.T) {
	iccid := ICCID{0x98, 0x44, 0x74, 0x68, 0x00, 0x00, 0x54, 0x37, 0x21, 0xF8}
	assert.Equal(t, "8944478600004573128", iccid.String())
	parsed, err := NewICCID(iccid.String())
	assert.NoError(t, err)
	assert.Equal(t, iccid, parsed)
}
