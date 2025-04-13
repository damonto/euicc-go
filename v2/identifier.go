package sgp22

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/damonto/euicc-go/bertlv/primitive"
)

// region ICCID

// ICCID represents the Integrated Circuit Card Identifier.
// The format is GSM-BCD
//
// See https://en.wikipedia.org/wiki/Binary-coded_decimal
type ICCID []byte

func NewICCID(iccid string) (ICCID, error) {
	return binaryCodedDecimalEncode[ICCID](iccid)
}

func (id ICCID) String() string {
	return binaryCodedDecimalDecode(id)
}

// endregion

// region IMEI

// IMEI represents the International Mobile Equipment Identity.
// The format is GSM-BCD
//
// See https://en.wikipedia.org/wiki/Binary-coded_decimal
type IMEI []byte

func NewIMEI(imei string) (IMEI, error) {
	return binaryCodedDecimalEncode[IMEI](imei)
}

func (imei IMEI) String() string {
	return binaryCodedDecimalDecode(imei)
}

// endregion

func binaryCodedDecimalEncode[T ~[]byte](value string) (T, error) {
	for _, r := range value {
		if (r < '0' || r > '9') && !(r == 'f' || r == 'F') {
			return nil, errors.New("invalid value")
		}
	}
	if len(value)%2 != 0 {
		value += "F"
	}
	id, _ := hex.DecodeString(value)
	for index := range id {
		id[index] = id[index]>>4 | id[index]<<4
	}
	return id, nil
}

func binaryCodedDecimalDecode(value []byte) string {
	id := make([]byte, len(value))
	var index int
	for index = range value {
		id[index] = value[index]>>4 | value[index]<<4
	}
	points := hex.EncodeToString(id)
	if index = strings.IndexByte(points, 'f'); index != -1 {
		points = points[:index]
	}
	return points
}

// region ISD-P Application Identifier

// ISDPAID represents the ISD-P Application Identifier.
type ISDPAID []byte

func (id ISDPAID) String() string {
	return strings.ToUpper(hex.EncodeToString(id))
}

// endregion

type ProfileClass int8

const (
	ProfileClassTest         ProfileClass = 0x00
	ProfileClassProvisioning ProfileClass = 0x01
	ProfileClassOperational  ProfileClass = 0x02
)

func (p ProfileClass) MarshalBinary() ([]byte, error) {
	return primitive.MarshalInt(p).MarshalBinary()
}

func (p ProfileClass) String() string {
	switch p {
	case ProfileClassTest:
		return "test"
	case ProfileClassProvisioning:
		return "provisioning"
	case ProfileClassOperational:
		return "operational"
	}
	return "unknown"
}
