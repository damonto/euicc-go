module github.com/damonto/euicc-go/examples

go 1.24.0

replace github.com/damonto/euicc-go => ..

require (
	github.com/damonto/euicc-go v0.0.0-00010101000000-000000000000
	github.com/damonto/libeuicc-go/driver/qmi v0.0.0-20250109023509-4f4c20bbcbc8
)

require github.com/damonto/libeuicc-go v0.0.0-20241014073658-6122c976bf1b // indirect
