module github.com/damonto/euicc-go/examples

go 1.24.0

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/ccid => ../driver/ccid
)

require (
	github.com/damonto/euicc-go v0.0.6
	github.com/damonto/euicc-go/driver/ccid v0.0.0-00010101000000-000000000000
)

require (
	github.com/ElMostafaIdrassi/goscard v1.0.0 // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	golang.org/x/sys v0.31.0 // indirect
)
