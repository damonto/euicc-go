module github.com/damonto/euicc-go/examples

go 1.24.1

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/at => ../driver/at
)

require (
	github.com/damonto/euicc-go v0.0.10
	github.com/damonto/euicc-go/driver/at v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.32.0 // indirect
