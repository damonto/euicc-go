//go:generate curl -L -o bundle.pem https://euicc-manual.osmocom.org/docs/pki/ci/bundle.pem
//go:generate curl -L -o bundle-tests.pem https://euicc-manual.osmocom.org/docs/pki/ci/bundle-tests.pem

package rootci

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
)

//go:embed bundle.pem
var bundle []byte

//go:embed bundle-tests.pem
var bundleTests []byte

var store, storeErr = newTrustedRootCAs(bundle, bundleTests)

func newTrustedRootCAs(production, test []byte) (*x509.CertPool, error) {
	store := x509.NewCertPool()
	var errs []error
	if !store.AppendCertsFromPEM(production) {
		errs = append(errs, errors.New("parse production root CI bundle"))
	}
	if !store.AppendCertsFromPEM(test) {
		errs = append(errs, errors.New("parse test root CI bundle"))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return store, nil
}

func TrustedRootCAs() (*x509.CertPool, error) {
	if storeErr != nil {
		return nil, storeErr
	}
	return store.Clone(), nil
}

func TrustedTLSConfig() (*tls.Config, error) {
	rootCAs, err := TrustedRootCAs()
	if err != nil {
		return nil, err
	}
	return &tls.Config{RootCAs: rootCAs}, nil
}

func UntrustedTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
