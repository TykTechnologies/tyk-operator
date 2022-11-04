package cert_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/matryer/is"
)

func TestCalculateFingerPrint(t *testing.T) {
	is := is.New(t)

	testCert, _, _ := generateTestCertificate()

	testCases := map[string]struct {
		Data  []byte
		Error error
	}{
		"nil data": {
			Data:  nil,
			Error: cert.ErrInvalidCert,
		},
		"empty data": {
			Data:  []byte(""),
			Error: cert.ErrInvalidCert,
		},
		"invalid data": {
			Data:  []byte("random-value"),
			Error: cert.ErrInvalidCert,
		},
		"Valid data": {
			Data: testCert,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			_, err := cert.CalculateFingerPrint([]byte(tc.Data))

			is.Equal(err, tc.Error)
		})
	}
}

func generateTestCertificate() ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	// PEM encoding of private key
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		},
	)

	//Create certificate templet
	template := x509.Certificate{
		SerialNumber:       big.NewInt(0),
		Subject:            pkix.Name{CommonName: "localhost"},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	//Create certificate using templet
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err

	}
	//pem encoding of certificate
	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		},
	)

	return certPem, keyPEM, nil
}
