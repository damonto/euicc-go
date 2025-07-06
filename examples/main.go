package main

import (
	"fmt"
	"log/slog"

	"github.com/damonto/euicc-go/driver/qmi"
	"github.com/damonto/euicc-go/lpa"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// ch, err := mbim.New("/dev/cdc-wdm0", 1)
	// if err != nil {
	// 	panic(err)
	// }
	ch, err := qmi.New("/dev/cdc-wdm1", 1)
	if err != nil {
		panic(err)
	}
	// ch, err := at.New("/dev/ttyUSB7")
	// if err != nil {
	// 	panic(err)
	// }
	client, err := lpa.New(&lpa.Options{
		Channel: ch,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	// ns, _ := client.ListNotification()
	// for _, n := range ns {
	// 	fmt.Println(n.SequenceNumber, n.ICCID, n.ProfileManagementOperation)
	// }

	// id, _ := sgp22.NewICCID("8944476500001224158")
	// fmt.Println(client.DeleteProfile(id))
	// fmt.Println(client.EnableProfile(id))

	// pn, err := client.RetrieveNotificationList(sgp22.SequenceNumber(202))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// if err := client.HandleNotification(pn[0]); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// ps, err := client.ListProfile(nil, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// for _, p := range ps {
	// 	fmt.Println(p.ProfileName, p.ICCID)
	// }

	eid, _ := client.EID()
	fmt.Println(eid)
	// for _, child := range tlv.First(bertlv.ContextSpecific.Constructed(10)).Children {
	// 	fmt.Println(hex.EncodeToString(child.Value))
	// }

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// installResult, err := client.DownloadProfile(ctx, &lpa.ActivationCode{
	// 	SMDP:       &url.URL{Scheme: "https", Host: "smdp.io"},
	// 	MatchingID: "QR-G-5C-1LS-1W1Z9P7",
	// 	IMEI:       "356938035643809",
	// }, &lpa.DownloadOptions{
	// 	OnProgress: func(stage lpa.DownloadStage) {
	// 		fmt.Println(stage)
	// 	},
	// 	OnConfirm: func(metadata *sgp22.ProfileInfo) bool {
	// 		fmt.Printf("Confirm download of profile %s with ICCID %s\n", metadata.ProfileName, metadata.ICCID)
	// 		return true // Return true to confirm the download
	// 	},
	// 	OnEnterConfirmationCode: func() string { return "" },
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// if installResult != nil {
	// 	fmt.Println(installResult.ISDPAID(), installResult.Notification)
	// }
}
