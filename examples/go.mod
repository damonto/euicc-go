module github.com/damonto/euicc-go/examples

go 1.24.0

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/mbim => ../driver/mbim
)

require (
	github.com/damonto/euicc-go v0.0.0-20250304064401-cce9769bc8f1
	github.com/damonto/euicc-go/driver/mbim v0.0.0-20250304064401-cce9769bc8f1
)
