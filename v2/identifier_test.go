package sgp22

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestICCID_String(t *testing.T) {
	var iccid ICCID
	iccid = ICCID{0x98, 0x44, 0x74, 0x68, 0x00, 0x00, 0x54, 0x37, 0x21, 0xF8}
	assert.Equal(t, "8944478600004573128", iccid.String())
	parsed, err := NewICCID(iccid.String())
	assert.NoError(t, err)
	assert.Equal(t, iccid, parsed)
}
