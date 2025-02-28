package primitive

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBigInt(t *testing.T) {
	var n big.Int
	assert.NoError(t, UnmarshalBigInt(&n).UnmarshalBinary([]byte{0x7f}))
	assert.Equal(t, int64(127), n.Int64())
	data, err := MarshalBigInt(&n).MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x7f}, data)
}
