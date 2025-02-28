//go:generate curl -L -o bundle.pem https://euicc-manual.osmocom.org/docs/pki/ci/bundle.pem

package rootci

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
)

//go:embed bundle.pem
var bundle []byte

var store = x509.NewCertPool()

func init() {
	store.AppendCertsFromPEM(bundle)
}

func TrustedRootCIs() *x509.CertPool {
	return store.Clone()
}

func TrustedTLSConfig() *tls.Config {
	return &tls.Config{RootCAs: TrustedRootCIs()}
}

func UntrustedTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
