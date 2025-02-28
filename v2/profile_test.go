package sgp22

import "testing"

func TestOperatorId_UnmarshalBERTLV(t *testing.T) {
	var operator OperatorId
	operator.UnmarshalBERTLV(nil)
}
