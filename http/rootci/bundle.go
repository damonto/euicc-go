//go:generate curl -L -o bundle.pem https://euicc-manual.osmocom.org/docs/pki/ci/bundle.pem
//go:generate curl -L -o bundle-tests.pem https://euicc-manual.osmocom.org/docs/pki/ci/bundle-tests.pem

package rootci

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
)

//go:embed bundle.pem
var bundle []byte

//go:embed bundle-tests.pem
var bundleTests []byte

var store = x509.NewCertPool()

func init() {
	store.AppendCertsFromPEM(bundle)
	store.AppendCertsFromPEM(bundleTests)
}

func TrustedRootCAs() *x509.CertPool {
	return store.Clone()
}

func TrustedTLSConfig() *tls.Config {
	return &tls.Config{RootCAs: TrustedRootCAs()}
}

func UntrustedTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
