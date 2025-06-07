module github.com/damonto/euicc-go/examples

go 1.24.1

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/at => ../driver/at
)

require (
	github.com/damonto/euicc-go v0.0.13
	github.com/damonto/euicc-go/driver/at v0.0.5
)

require golang.org/x/sys v0.33.0 // indirect
