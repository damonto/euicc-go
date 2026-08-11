package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/driver/qcom"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "euicc example: %v\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	// ch, err := mbim.New(mbim.WithProxy("/dev/cdc-wdm0"), mbim.WithSlot(1), mbim.WithTimeout(30*time.Second))
	// if err != nil {
	// 	return fmt.Errorf("open MBIM channel: %w", err)
	// }
	ch, err := qcom.NewQMI(qcom.WithProxy("/dev/cdc-wdm0"), qcom.WithSlot(1), qcom.WithTimeout(30*time.Second))
	// ch, err := qcom.NewQRTR(qcom.WithSlot(1))
	if err != nil {
		return fmt.Errorf("open QMI channel: %w", err)
	}
	// ch, err := at.New("/dev/ttyUSB7")
	// if err != nil {
	// 	return fmt.Errorf("open AT channel: %w", err)
	// }
	// ch := ccid.New()
	// reader, err := ch.ListReaders()
	// if err != nil {
	// 	return fmt.Errorf("list CCID readers: %w", err)
	// }
	// if len(reader) == 0 {
	// 	return errors.New("no CCID readers found")
	// }
	// fmt.Printf("Using reader: %s\n", reader[0])
	// if err := ch.SetReader(reader[0]); err != nil {
	// 	return fmt.Errorf("select CCID reader: %w", err)
	// }

	client, err := lpa.New(&lpa.Options{
		Channel: ch,
	})
	if err != nil {
		return fmt.Errorf("create LPA client: %w", err)
	}
	defer func() {
		err = errors.Join(err, client.Close())
	}()

	if err := testEID(client); err != nil {
		return err
	}

	// if err := testDownload(client); err != nil {
	// 	return err
	// }

	if err := testListProfiles(client); err != nil {
		return err
	}

	// return testDiscovery(client)
	return nil
}

func testEID(client *lpa.Client) error {
	eid, err := client.EID()
	if err != nil {
		return fmt.Errorf("read EID: %w", err)
	}
	fmt.Printf("EID: %X\n", eid)
	return nil
}

func testDownload(client *lpa.Client) error {
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
		return fmt.Errorf("download profile: %w", err)
	}
	if installResult != nil {
		fmt.Println("install Result", installResult.ISDPAID(), installResult.Notification)
	}
	return nil
}

func testListProfiles(client *lpa.Client) error {
	profiles, err := client.ListProfile(nil, []bertlv.Tag{sgp22.TagNotificationConfigurationInfo})
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}
	for _, profile := range profiles {
		// fmt.Printf("Profile: %s, ICCID: %s\n", profile.ProfileName, profile.ICCID)
		for _, n := range profile.NotificationConfigurationInfo {
			fmt.Printf("Address: %s, Operations %+v \n", n.Address, n.ProfileManagementOperations)
		}
	}
	return nil
}

func testListNotifications(client *lpa.Client) error {
	notifications, err := client.ListNotification()
	if err != nil {
		return fmt.Errorf("list notifications: %w", err)
	}
	for _, notification := range notifications {
		fmt.Printf("Sequence: %d, ICCID: %s, Operation: %d\n",
			notification.SequenceNumber, notification.ICCID, notification.ProfileManagementOperation)
	}
	return nil
}

func testEnableProfile(client *lpa.Client) error {
	id, err := sgp22.NewICCID("8944476500001224158")
	if err != nil {
		return fmt.Errorf("parse ICCID: %w", err)
	}
	if err := client.EnableProfile(id, true); err != nil {
		return fmt.Errorf("enable profile: %w", err)
	}
	fmt.Println("Profile enabled successfully")
	return nil
}

func testDisableProfile(client *lpa.Client) error {
	id, err := sgp22.NewICCID("8944476500001224158")
	if err != nil {
		return fmt.Errorf("parse ICCID: %w", err)
	}
	if err := client.DisableProfile(id, true); err != nil {
		return fmt.Errorf("disable profile: %w", err)
	}
	fmt.Println("Profile disabled successfully")
	return nil
}

func testSendNotification(client *lpa.Client, sequenceNumber sgp22.SequenceNumber) error {
	notifications, err := client.RetrieveNotificationList(sequenceNumber)
	if err != nil {
		return fmt.Errorf("retrieve notifications: %w", err)
	}
	if len(notifications) == 0 {
		fmt.Println("No notifications found")
		return nil
	}
	if err := client.HandleNotification(notifications[0]); err != nil {
		return fmt.Errorf("handle notification: %w", err)
	}
	fmt.Println("Notification handled successfully")
	return nil
}

func testDiscovery(client *lpa.Client) error {
	addresses := []url.URL{
		{Scheme: "https", Host: "lpa.ds.gsma.com"},
		{Scheme: "https", Host: "lpa.live.esimdiscovery.com"},
	}
	imei, err := sgp22.NewIMEI("356938035643809")
	if err != nil {
		return fmt.Errorf("parse IMEI: %w", err)
	}

	var errs []error
	for _, address := range addresses {
		fmt.Printf("Discovering profiles at %s...\n", address.Host)
		entries, err := client.Discovery(&address, imei)
		if err != nil {
			errs = append(errs, fmt.Errorf("discover profiles at %s: %w", address.Host, err))
			continue
		}
		for _, entry := range entries {
			fmt.Printf("Discovered profile: %s, URL: %s\n", entry.EventID, entry.Address)
		}
	}
	return errors.Join(errs...)
}
