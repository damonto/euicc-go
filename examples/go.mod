module github.com/damonto/euicc-go/examples

go 1.24.1

replace (
	github.com/damonto/euicc-go => ../
	github.com/damonto/euicc-go/driver/at => ../driver/at
)

require (
	github.com/damonto/euicc-go v0.0.12
	github.com/damonto/euicc-go/driver/at v0.0.3
)

require golang.org/x/sys v0.33.0 // indirect
