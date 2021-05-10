package universal_client

import "context"

type UniversalCertificate interface {
	All(ctx context.Context) ([]string, error)
	Upload(ctx context.Context, key []byte, crt []byte) (id string, err error)
	Delete(ctx context.Context, id string) error
	// Exists returns true if a certificate with id exists
	Exists(ctx context.Context, id string) bool
}
