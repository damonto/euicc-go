module github.com/damonto/euicc-go/examples

go 1.24.0

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/qmi => ../driver/qmi
)

require (
	github.com/damonto/euicc-go v0.0.6
	github.com/damonto/euicc-go/driver/qmi v0.0.0-00010101000000-000000000000
)
