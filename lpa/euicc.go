package lpa

import (
	"github.com/damonto/euicc-go/http"
	"github.com/damonto/euicc-go/v2"
)

type Client struct {
	HTTP *http.Client
	APDU sgp22.Transmitter
}
