package universal_client

type UniversalCertificate interface {
	Upload(key []byte, crt []byte) (id string, err error)
	Delete(id string) error
	Get(id string) (string, error)
}

func UploadCertificate(c UniversalClient, key []byte, crt []byte) (string, error) {
	return c.Certificate().Upload(key, crt)
}

func GetCertificate(c UniversalClient, id string) (string, error) {
	return c.Certificate().Get(id)
}

func DeleteCertificate(c UniversalClient, id string) error {
	return c.Certificate().Delete(id)
}
