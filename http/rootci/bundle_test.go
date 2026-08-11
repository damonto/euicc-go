package rootci

import (
	"crypto/x509"
	"testing"
)

func TestTrustedRootCAs(t *testing.T) {
	store, err := TrustedRootCAs()
	if err != nil {
		t.Fatalf("TrustedRootCAs() error = %v", err)
	}
	if store.Equal(x509.NewCertPool()) {
		t.Fatal("TrustedRootCAs() returned an empty certificate pool")
	}
}

func TestNewTrustedRootCAsRejectsInvalidBundle(t *testing.T) {
	if _, err := newTrustedRootCAs(bundle, []byte("invalid")); err == nil {
		t.Error("newTrustedRootCAs() error = nil for invalid bundle")
	}
}
