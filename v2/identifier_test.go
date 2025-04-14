package sgp22

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestICCID_String(t *testing.T) {
	var iccid ICCID
	var parsed ICCID
	var err error

	// Standard ICCID
	iccid = ICCID{0x98, 0x44, 0x74, 0x68, 0x00, 0x00, 0x54, 0x37, 0x21, 0xF8}
	assert.Equal(t, "8944478600004573128", iccid.String())
	parsed, err = NewICCID(iccid.String())
	assert.NoError(t, err)
	assert.Equal(t, iccid, parsed)

	// Non-standard ICCID
	// 89860110F9900160570
	iccid = ICCID{0x98, 0x68, 0x10, 0x01, 0x9F, 0x09, 0x10, 0x06, 0x75, 0xF0}
	assert.Equal(t, "89860110f9900160570", iccid.String())
	parsed, err = NewICCID(iccid.String())
	assert.NoError(t, err)
	assert.Equal(t, iccid, parsed)
}
