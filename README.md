# eUICC Go

`euicc-go` is a Go library for working with GSMA RSP / eUICC profile
management. The code is organized around SGP.22 v2.x, with a high-level LPA
client, protocol message types, BER-TLV helpers, and several smart-card
transport drivers.

The project is currently a library and example program, not a complete
end-user CLI.

## Features

- LPA client for common ES10a / ES10b / ES10c, ES9+, and ES11 flows.
- Profile download flow with progress, confirmation, and confirmation-code
  callbacks.
- Profile operations: list, enable, disable, delete, set nickname, memory
  reset, and EID retrieval.
- Notification operations: list local notifications, retrieve pending
  notification payloads, send notification to SM-DP+, and remove sent
  notifications from the eUICC list.
- Discovery through SM-DS.
- SGP.22 v2 message models for APDU and HTTP request/response encoding.
- Bound Profile Package segmentation for ES10b `LoadBoundProfilePackage`.
- ASN.1 BER-TLV parser and builder used by the RSP protocol implementation.
- Smart-card channels over CCID / PCSC, AT serial, MBIM, Qualcomm QMI, and
  Qualcomm QRTR.
- HTTP client configured with bundled eUICC CI root certificates.

## Packages

| Package | Purpose |
| --- | --- |
| `lpa` | High-level Local Profile Assistant client. This is the main package most callers should use. |
| `v2` | SGP.22 v2.x APDU / HTTP message types, identifiers, profile types, notification types, and errors. |
| `driver` | Shared smart-card channel and APDU transmitter interfaces. |
| `driver/ccid` | CCID / PCSC reader channel. |
| `driver/at` | AT-command modem channel over a serial device. |
| `driver/mbim` | MBIM proxy modem channel. |
| `driver/qcom` | Qualcomm QMI and QRTR modem channels. |
| `driver/qmi` | Deprecated compatibility aliases for `driver/qcom`. |
| `http` | RSP JSON-over-HTTP client helpers. |
| `http/rootci` | Embedded eUICC CI root certificate bundle. |
| `bertlv` | BER-TLV read, write, selector, and primitive helpers. |

## Requirements

- Access to an eUICC through one of the supported channel drivers.
- For CCID / PCSC usage, a working smart-card reader stack on the host.
- For modem usage, access to the relevant modem device, slot, and permissions
  for AT, MBIM, QMI, or QRTR transport.

## Installation

```sh
go get github.com/damonto/euicc-go
```

## Minimal Usage

This example opens the first available CCID reader, creates an LPA client,
prints the EID, and lists profiles.

```go
package main

import (
	"fmt"

	"github.com/damonto/euicc-go/driver/ccid"
	"github.com/damonto/euicc-go/lpa"
)

func main() {
	ch, err := ccid.New()
	if err != nil {
		panic(err)
	}

	readers, err := ch.ListReaders()
	if err != nil {
		panic(err)
	}
	if len(readers) == 0 {
		panic("no CCID readers found")
	}
	if err := ch.SetReader(readers[0]); err != nil {
		panic(err)
	}

	client, err := lpa.New(&lpa.Options{Channel: ch})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	eid, err := client.EID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("EID: %X\n", eid)

	profiles, err := client.ListProfile(nil, nil)
	if err != nil {
		panic(err)
	}
	for _, profile := range profiles {
		fmt.Printf("Profile: %s, ICCID: %s\n", profile.ProfileName, profile.ICCID)
	}
}
```

The repository also contains a runnable example in `examples/main.go`.

```sh
cd examples
go run .
```

## Choosing A Channel Driver

All high-level LPA operations use a `driver.SmartCardChannel`. Select the
driver that matches how the eUICC is exposed on your system.

```go
// CCID / PCSC reader.
ch, err := ccid.NewWithReader("reader name")

// AT modem serial port.
ch, err := at.New("/dev/ttyUSB7")

// MBIM proxy device, slot numbering starts at 1.
ch, err := mbim.New("/dev/cdc-wdm0", 1)

// Qualcomm QMI proxy device, slot numbering starts at 1.
ch, err := qcom.NewQMI("/dev/cdc-wdm0", 1)

// Qualcomm QRTR UIM service, slot numbering starts at 1.
ch, err := qcom.NewQRTR(1)
```

Import the matching packages as needed:

```go
import (
	"github.com/damonto/euicc-go/driver/at"
	"github.com/damonto/euicc-go/driver/ccid"
	"github.com/damonto/euicc-go/driver/mbim"
	"github.com/damonto/euicc-go/driver/qcom"
)
```

## LPA Client

Create a client with `lpa.New`:

```go
client, err := lpa.New(&lpa.Options{
	Channel: ch,
	// Optional. Defaults are shown by behavior:
	// AID: GSMA ISD-R AID
	// MSS: 254
	// AdminProtocolVersion: "2.5.0"
	// Timeout: 30 * time.Second
	// Logger: slog.Default()
})
if err != nil {
	return err
}
defer client.Close()
```

The current `AdminProtocolVersion` validation accepts SGP.22 v2.x values. A
leading `v` is normalized, so values like `v2.5.0` are accepted.

## Common Operations

### eUICC Data

```go
eid, err := client.EID()
info1, err := client.EUICCInfo1()
info2, err := client.EUICCInfo2()
challenge, err := client.EUICCChallenge()
addresses, err := client.EUICCConfiguredAddresses()
err = client.SetDefaultDPAddress("smdp.example.com")
```

### Profile Management

```go
profiles, err := client.ListProfile(nil, nil)

iccid, err := sgp22.NewICCID("8944476500001224158")
if err != nil {
	return err
}

err = client.EnableProfile(iccid, true)
err = client.DisableProfile(iccid, true)
err = client.SetNickname(iccid, "travel")
err = client.DeleteProfile(iccid)
```

`ListProfile` accepts these search criteria:

- `nil` for all profiles.
- `sgp22.ICCID`.
- `sgp22.ISDPAID`.
- `sgp22.ProfileClass`.

Profile enable, disable, and delete accept either `sgp22.ICCID` or
`sgp22.ISDPAID` as the identifier.

`MemoryReset` is also implemented. It deletes operational profiles, deletes
field-loaded test profiles, and resets the default SM-DP+ address:

```go
err = client.MemoryReset()
```

Use destructive operations only when the target eUICC and profile state are
known.

### Download A Profile

```go
ac := &lpa.ActivationCode{
	SMDP:       &url.URL{Scheme: "https", Host: "smdp.example.com"},
	MatchingID: "matching-id",
	IMEI:       "356938035643809",
}

result, err := client.DownloadProfile(ctx, ac, &lpa.DownloadOptions{
	OnProgress: func(stage lpa.DownloadStage) {
		fmt.Println(stage)
	},
	OnConfirm: func(metadata *sgp22.ProfileInfo) bool {
		return true
	},
	OnEnterConfirmationCode: func() string {
		return ""
	},
})
if err != nil {
	return err
}
if result != nil {
	fmt.Println(result.ISDPAID(), result.Notification)
}
```

`ActivationCode` supports text marshal and unmarshal for `LPA:1$...` activation
codes. The download helper currently requires an IMEI in addition to the
SM-DP+ address.

### Notifications

```go
notifications, err := client.ListNotification()
pending, err := client.RetrieveNotificationList(nil)
pendingBySeq, err := client.RetrieveNotificationList(sgp22.SequenceNumber(1))
pendingByEvent, err := client.RetrieveNotificationList(sgp22.NotificationEventInstall)

if len(pending) > 0 {
	err = client.HandleNotification(pending[0])
	err = client.RemoveNotificationFromList(pending[0].Notification.SequenceNumber)
}
```

`RetrieveNotificationList` accepts `nil`, `sgp22.SequenceNumber`, or
`sgp22.NotificationEvent` as search criteria.

### Discovery

```go
imei, err := sgp22.NewIMEI("356938035643809")
if err != nil {
	return err
}

entries, err := client.Discovery(&url.URL{
	Scheme: "https",
	Host:   "lpa.ds.gsma.com",
}, imei)
```

Each returned `sgp22.EventEntry` contains the event ID and RSP server address.

## Lower-Level Protocol Use

Callers that need direct protocol access can use the lower-level helpers in
`v2`:

```go
response, err := sgp22.InvokeAPDU(client.APDU, &sgp22.GetEuiccDataRequest{})
remote, err := sgp22.InvokeHTTP(client.HTTP, smdpURL, request)
```

The `bertlv` package can be used independently for BER-TLV parsing and
building.

## Testing

Run all unit tests:

```sh
go test ./...
```

Run the example module separately:

```sh
cd examples
go test ./...
```

Most tests use fixtures and fake transports. Real profile operations require
hardware, carrier / SM-DP+ access, and the correct host permissions.

## References

- [SGP.22 v2.5](https://aka.pw/sgp22/v2.5)
- [Infineon LPA](https://github.com/CursedHardware/infineon-lpa-mirror/tree/4.0.3/messages/src/main/java/com/gsma/sgp/messages/rspdefinitions)
- [asn1bean](https://github.com/beanit/asn1bean)

## License

MIT
