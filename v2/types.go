package sgp22

import (
	"errors"
	"net/url"

	"github.com/damonto/euicc-go/bertlv"
)

type Transmitter interface {
	Transmit(bertlv.Marshaler, bertlv.Unmarshaler) error
	TransmitRaw([]byte) ([]byte, error)
}

type CardRequest[R CardResponse] interface {
	bertlv.Marshaler
	CardResponse() R
}

type CardResponse interface {
	bertlv.Unmarshaler
	Valid() error
}

func InvokeAPDU[I CardRequest[O], O CardResponse](transmitter Transmitter, request I) (O, error) {
	response := request.CardResponse()
	err := transmitter.Transmit(request, response)
	if err == nil {
		err = response.Valid()
	}
	return response, err
}

func InvokeRawAPDU(transmitter Transmitter, command []byte) ([]byte, error) {
	return transmitter.TransmitRaw(command)
}

type HTTPClient interface {
	SendRequest(url *url.URL, request, response any) error
}

type HTTPRequest[R HTTPResponse] interface {
	URL(*url.URL) *url.URL
	RemoteResponse() R
}

type HTTPResponse interface {
	FunctionExecutionStatus() *ExecutionStatus
}

func InvokeHTTP[I HTTPRequest[O], O HTTPResponse](client HTTPClient, address *url.URL, request I) (O, error) {
	response := request.RemoteResponse()
	if err := client.SendRequest(request.URL(address), request, response); err != nil {
		return response, err
	}
	if !response.FunctionExecutionStatus().ExecutedSuccess() {
		return response, errors.New(response.FunctionExecutionStatus().StatusCodeData.Error())
	}
	return response, nil
}
