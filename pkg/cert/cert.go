package cert

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"errors"
)

var ErrInvalidCert = errors.New("certificate is invalid")

func HexSHA256(cert []byte) string {
	certSHA := sha256.Sum256(cert)
	return hex.EncodeToString(certSHA[:])
}

func CalculateFingerPrint(data []byte) (string, error) {
	publicKey, _ := pem.Decode(data)
	if publicKey != nil {
		return HexSHA256(publicKey.Bytes), nil
	}

	return "", ErrInvalidCert
}
