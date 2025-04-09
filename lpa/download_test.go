package lpa

import (
	"testing"
)

func TestParseActivationCode(t *testing.T) {
	tests := []struct {
		name          string
		acString      string
		imei          string
		expectedSMDP  string
		expectedToken string
		expectedOID   string
		expectedCC    string
		expectedError bool
	}{
		{
			name:          "Basic activation code without OID and CC",
			acString:      "1$SMDP.GSMA.COM$04386-AGYFT-A74Y8-3F815",
			imei:          "123456789012345",
			expectedSMDP:  "https://SMDP.GSMA.COM",
			expectedToken: "04386-AGYFT-A74Y8-3F815",
			expectedOID:   "",
			expectedCC:    "",
			expectedError: false,
		},
		{
			name:          "Activation code with CC but without OID",
			acString:      "1$SMDP.GSMA.COM$04386-AGYFT-A74Y8-3F815$$1",
			imei:          "123456789012345",
			expectedSMDP:  "https://SMDP.GSMA.COM",
			expectedToken: "04386-AGYFT-A74Y8-3F815",
			expectedOID:   "",
			expectedCC:    "1",
			expectedError: false,
		},
		{
			name:          "Activation code with both OID and CC",
			acString:      "1$SMDP.GSMA.COM$04386-AGYFT-A74Y8-3F815$1.3.6.1.4.1.31746$1",
			imei:          "123456789012345",
			expectedSMDP:  "https://SMDP.GSMA.COM",
			expectedToken: "04386-AGYFT-A74Y8-3F815",
			expectedOID:   "1.3.6.1.4.1.31746",
			expectedCC:    "1",
			expectedError: false,
		},
		{
			name:          "Activation code with OID but without CC",
			acString:      "1$SMDP.GSMA.COM$04386-AGYFT-A74Y8-3F815$1.3.6.1.4.1.31746",
			imei:          "123456789012345",
			expectedSMDP:  "https://SMDP.GSMA.COM",
			expectedToken: "04386-AGYFT-A74Y8-3F815",
			expectedOID:   "1.3.6.1.4.1.31746",
			expectedCC:    "",
			expectedError: false,
		},
		{
			name:          "Activation code with OID and empty token",
			acString:      "1$SMDP.GSMA.COM$$1.3.6.1.4.1.31746",
			imei:          "123456789012345",
			expectedSMDP:  "https://SMDP.GSMA.COM",
			expectedToken: "",
			expectedOID:   "1.3.6.1.4.1.31746",
			expectedCC:    "",
			expectedError: false,
		},
		{
			name:          "Invalid activation code format",
			acString:      "2$SMDP.GSMA.COM$04386-AGYFT-A74Y8-3F815",
			imei:          "123456789012345",
			expectedError: true,
		},
		{
			name:          "Invalid SMDP address",
			acString:      "1$invalid-url$04386-AGYFT-A74Y8-3F815",
			imei:          "123456789012345",
			expectedError: true,
		},
		{
			name:          "Too few parts",
			acString:      "1$SMDP.GSMA.COM",
			imei:          "123456789012345",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac, err := ParseActivationCode(tt.acString, tt.imei)
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if ac.SMDP.String() != tt.expectedSMDP {
				t.Errorf("Expected SMDP %s, got %s", tt.expectedSMDP, ac.SMDP.String())
			}

			if ac.MatchingID != tt.expectedToken {
				t.Errorf("Expected token %s, got %s", tt.expectedToken, ac.MatchingID)
			}

			if ac.OID != tt.expectedOID {
				t.Errorf("Expected OID %s, got %s", tt.expectedOID, ac.OID)
			}

			if ac.ConfirmationCode != tt.expectedCC {
				t.Errorf("Expected confirmation code %s, got %s", tt.expectedCC, ac.ConfirmationCode)
			}

			if ac.IMEI != tt.imei {
				t.Errorf("Expected IMEI %s, got %s", tt.imei, ac.IMEI)
			}
		})
	}
}
