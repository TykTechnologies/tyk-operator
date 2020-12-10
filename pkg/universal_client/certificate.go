package universal_client

type UniversalCertificate interface {
	All() ([]string, error)
	Upload(key []byte, crt []byte) (id string, err error)
	Delete(id string) error
	// Exists returns true if a certificate with id exists
	Exists(id string) bool
}
