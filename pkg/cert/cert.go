package cert

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
)

func HexSHA256(cert []byte) string {
	certSHA := sha256.Sum256(cert)
	return hex.EncodeToString(certSHA[:])
}

func CalculateFingerPrint(data []byte) string {
	// TODO: publicKey might be returned as nil. if that's the case, the program panics.
	publicKey, _ := pem.Decode(data)
	return HexSHA256(publicKey.Bytes)
}
