package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/driver/ccid"
	sgp22http "github.com/damonto/euicc-go/http"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
)

type DownloadHandler struct{}

// HandleConfirm implements lpa.Handler.
func (d *DownloadHandler) Confirm(metadata *sgp22.ProfileInfo) chan bool {
	fmt.Println(metadata)
	bool := make(chan bool, 1)
	bool <- true
	return bool
}

// HandleProgress implements lpa.Handler.
func (d *DownloadHandler) Progress(process lpa.DownloadProgress) {
	fmt.Println(process)
}

func (d *DownloadHandler) ConfirmationCode() chan string {
	code := make(chan string, 1)
	code <- "0000"
	return code
}

func NewDownloadHandler() lpa.DownloadHandler {
	return &DownloadHandler{}
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ch, err := ccid.New()
	if err != nil {
		panic(err)
	}
	readers, err := ch.ListReaders()
	if err != nil {
		panic(err)
	}
	ch.SetReader(readers[0])
	transmitter, err := driver.NewTransmitter(ch, driver.ISDRAID, 240)
	if err != nil {
		panic(err)
	}
	defer transmitter.Close()
	client := &lpa.Client{
		HTTP: &sgp22http.Client{
			Client:        driver.NewHTTPClient(30 * time.Second),
			AdminProtocol: "gsma/rsp/v2.2.0",
		},
		APDU: transmitter,
	}

	// id, _ := sgp22.NewICCID("8944840003001326791")
	// fmt.Println(client.EnableProfile(id))

	// pn, err := client.RetrieveNotificationList(sgp22.SequenceNumber(89))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// if err := client.HandleNotification(pn[0]); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	eid, _ := client.EID()
	fmt.Println(eid)
	// for _, child := range tlv.First(bertlv.ContextSpecific.Constructed(10)).Children {
	// 	fmt.Println(hex.EncodeToString(child.Value))
	// }

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// installResult, err := client.DownloadProfile(ctx, &lpa.ActivationCode{
	// 	SMDP:       &url.URL{Scheme: "https", Host: "abc.smdp.com"},
	// 	MatchingID: "123131313131",
	// 	IMEI:       "356938035643809",
	// }, NewDownloadHandler())
	// if err != nil {
	// 	panic(err)
	// }
	// if installResult != nil {
	// 	fmt.Println(installResult.ISDPAID(), installResult.Notification)
	// }
}
