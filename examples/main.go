package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/damonto/euicc-go/driver/qmi"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// ch, err := mbim.New("/dev/cdc-wdm0", 1)
	// if err != nil {
	// 	panic(err)
	// }
	ch, err := qmi.New("/dev/cdc-wdm0", 1)
	if err != nil {
		panic(err)
	}
	// ch, err := at.New("/dev/ttyUSB7")
	// if err != nil {
	// 	panic(err)
	// }
	// ch, err := ccid.New()
	// if err != nil {
	// 	panic(err)
	// }
	// reader, err := ch.ListReaders()
	// if err != nil {
	// 	panic(err)
	// }
	// if len(reader) == 0 {
	// 	panic("No readers found")
	// }
	// fmt.Printf("Using reader: %s\n", reader[0])
	// ch.SetReader(reader[0])

	client, err := lpa.New(&lpa.Options{
		Channel: ch,
	})
	if err != nil {
		fmt.Printf("Failed to create LPA client: %v\n", err)
		return
	}
	defer client.Close()

	testEID(client)

	testListProfiles(client)

	// testDiscovery(client)
}

func testEID(client *lpa.Client) {
	eid, err := client.EID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("EID: %X\n", eid)
}

func testDownload(client *lpa.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	installResult, err := client.DownloadProfile(ctx, &lpa.ActivationCode{
		SMDP:       &url.URL{Scheme: "https", Host: "smdp.io"},
		MatchingID: "QR-G-5C-1LS-1W1Z9P7",
		IMEI:       "356938035643809",
	}, &lpa.DownloadOptions{
		OnProgress: func(stage lpa.DownloadStage) {
			fmt.Println(stage)
		},
		OnConfirm: func(metadata *sgp22.ProfileInfo) bool {
			fmt.Printf("Confirm download of profile %s with ICCID %s\n", metadata.ProfileName, metadata.ICCID)
			return true // Return true to confirm the download
		},
		OnEnterConfirmationCode: func() string { return "" },
	})
	if err != nil {
		panic(err)
	}
	if installResult != nil {
		fmt.Println(installResult.ISDPAID(), installResult.Notification)
	}
}

func testListProfiles(client *lpa.Client) {
	profiles, err := client.ListProfile(nil, nil)
	if err != nil {
		panic(err)
	}
	for _, profile := range profiles {
		fmt.Printf("Profile: %s, ICCID: %s\n", profile.ProfileName, profile.ICCID)
	}
}

func testListNotifications(client *lpa.Client) {
	notifications, err := client.ListNotification()
	if err != nil {
		panic(err)
	}
	for _, notification := range notifications {
		fmt.Printf("Sequence: %d, ICCID: %s, Operation: %d\n",
			notification.SequenceNumber, notification.ICCID, notification.ProfileManagementOperation)
	}
}

func testEnableProfile(client *lpa.Client) {
	id, _ := sgp22.NewICCID("8944476500001224158")
	if err := client.EnableProfile(id, true); err != nil {
		fmt.Printf("Failed to enable profile: %v\n", err)
	} else {
		fmt.Println("Profile enabled successfully")
	}
}

func testDisableProfile(client *lpa.Client) {
	id, _ := sgp22.NewICCID("8944476500001224158")
	if err := client.DisableProfile(id, true); err != nil {
		fmt.Printf("Failed to disable profile: %v\n", err)
	} else {
		fmt.Println("Profile disabled successfully")
	}
}

func testSendNotification(client *lpa.Client, sequenceNumber sgp22.SequenceNumber) {
	notifications, err := client.RetrieveNotificationList(sequenceNumber)
	if err != nil {
		fmt.Printf("Failed to retrieve notifications: %v\n", err)
		return
	}
	if len(notifications) == 0 {
		fmt.Println("No notifications found")
		return
	}
	if err := client.HandleNotification(notifications[0]); err != nil {
		fmt.Printf("Failed to handle notification: %v\n", err)
	} else {
		fmt.Println("Notification handled successfully")
	}
}

func testDiscovery(client *lpa.Client) {
	addresses := []url.URL{
		{Scheme: "https", Host: "lpa.ds.gsma.com"},
		{Scheme: "https", Host: "lpa.live.esimdiscovery.com"},
	}

	for _, address := range addresses {
		fmt.Printf("Discovering profiles at %s...\n", address.Host)
		imei, _ := sgp22.NewIMEI("356938035643809") // Example IMEI, replace with actual if needed
		entries, err := client.Discovery(&address, imei)
		if err != nil {
			fmt.Printf("Failed to discover profiles: %v\n", err)
			continue
		}
		for _, entry := range entries {
			fmt.Printf("Discovered profile: %s, URL: %s\n", entry.EventID, entry.Address)
		}
	}
}
