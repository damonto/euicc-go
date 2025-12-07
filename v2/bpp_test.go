package sgp22

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
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
			bpp, err := LoadBoundProfilePackage(fixture.BPP)
			assert.NoError(t, err)
			expectedSegments, err := LoadSegmentedBoundProfilePackage(fixture.SBPP)
			assert.NoError(t, err)
			segments, err := SegmentedBoundProfilePackage(bpp)
			assert.NoError(t, err)
			var index int
			for index = 0; index < len(segments); index++ {
				assert.Equal(t, expectedSegments[index], segments[index], index)
			}
		})
	}
}

func LoadBoundProfilePackage(name string) (*bertlv.TLV, error) {
	fp, err := os.Open(filepath.Join("fixtures", name))
	if err != nil {
		return nil, err
	}
	bpp := new(bertlv.TLV)
	_, err = bpp.ReadFrom(base64.NewDecoder(base64.StdEncoding, fp))
	return bpp, err
}

func LoadSegmentedBoundProfilePackage(name string) ([][]byte, error) {
	fp, err := os.Open(filepath.Join("fixtures", name))
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)
	var block []byte
	var line int
	var text string
	var sbpp [][]byte
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
	return sbpp, nil
}
