package sgp22

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
)

func TestSegmentedBoundProfilePackage(t *testing.T) {
	type Fixture struct {
		BPP  string
		SBPP string
		Name string
	}
	fixtures := []Fixture{
		{"bpp@1.txt", "sbpp@1.txt", "Infineon"},
		{"bpp@2.txt", "sbpp@2.txt", "Redtea Mobile"},
		{"bpp@3.txt", "sbpp@3.txt", "Tigo"},
		{"bpp@4.txt", "sbpp@4.txt", "Tele2"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			bpp, err := loadBoundProfilePackage(fixture.BPP)
			if err != nil {
				t.Fatalf("loadBoundProfilePackage(%q) error = %v", fixture.BPP, err)
			}
			expectedSegments, err := loadSegmentedBoundProfilePackage(fixture.SBPP)
			if err != nil {
				t.Fatalf("loadSegmentedBoundProfilePackage(%q) error = %v", fixture.SBPP, err)
			}
			segments, err := SegmentedBoundProfilePackage(bpp)
			if err != nil {
				t.Fatalf("SegmentedBoundProfilePackage() error = %v", err)
			}
			if len(segments) != len(expectedSegments) {
				t.Fatalf("segment count = %d, want %d", len(segments), len(expectedSegments))
			}
			for index := range segments {
				if !bytes.Equal(segments[index], expectedSegments[index]) {
					t.Errorf("segment %d = % X, want % X", index, segments[index], expectedSegments[index])
				}
			}
		})
	}
}

func TestSegmentedBoundProfilePackageSplitsSequenceOf87(t *testing.T) {
	bpp := bertlv.NewChildren(
		bertlv.ContextSpecific.Constructed(54),
		bertlv.NewChildren(bertlv.ContextSpecific.Constructed(35)),
		bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(0),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(7), []byte{0x01}),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(7), []byte{0x02}),
		),
		bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(1),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(8), []byte{0x03}),
		),
		bertlv.NewChildren(
			bertlv.ContextSpecific.Constructed(3),
			bertlv.NewValue(bertlv.ContextSpecific.Primitive(6), []byte{0x04}),
		),
	)

	segments, err := SegmentedBoundProfilePackage(bpp)

	if err != nil {
		t.Fatalf("SegmentedBoundProfilePackage() error = %v", err)
	}
	want := [][]byte{
		{0xbf, 0x36, 0x15, 0xbf, 0x23, 0x00},
		{0xa0, 0x06, 0x87, 0x01, 0x01},
		{0x87, 0x01, 0x02},
		{0xa1, 0x03},
		{0x88, 0x01, 0x03},
		{0xa3, 0x03},
		{0x86, 0x01, 0x04},
	}
	if !reflect.DeepEqual(segments, want) {
		t.Errorf("SegmentedBoundProfilePackage() = % X, want % X", segments, want)
	}
}

func loadBoundProfilePackage(name string) (bpp *bertlv.TLV, err error) {
	fp, err := os.Open(filepath.Join("fixtures", name))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, fp.Close())
	}()
	bpp = new(bertlv.TLV)
	_, err = bpp.ReadFrom(base64.NewDecoder(base64.StdEncoding, fp))
	return bpp, err
}

func loadSegmentedBoundProfilePackage(name string) (sbpp [][]byte, err error) {
	fp, err := os.Open(filepath.Join("fixtures", name))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, fp.Close())
	}()
	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)
	var block []byte
	var line int
	var text string
	for scanner.Scan() {
		line++
		text = scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		if block, err = hex.DecodeString(text); err != nil {
			return nil, fmt.Errorf("line %d: %w", line+1, err)
		}
		sbpp = append(sbpp, block)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sbpp, nil
}
