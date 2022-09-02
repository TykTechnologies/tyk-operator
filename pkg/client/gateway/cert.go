package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

type Cert struct{}

type CertificateList struct {
	CertIDs []string `json:"certs"`
}

// All returns a list of all certificates ID's
func (c Cert) All(ctx context.Context) ([]string, error) {
	res, err := client.Get(ctx, endpointCerts, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, client.Error(res)
	}

	var o CertificateList

	err = client.JSON(res, &o)
	if err != nil {
		return nil, err
	}

	return o.CertIDs, nil
}

func (c Cert) Exists(ctx context.Context, id string) bool {
	res, err := client.Get(ctx, client.Join(endpointCerts, id), nil)
	if err != nil {
		client.LError(ctx, err, "failed to get certificate")
		return false
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		client.LError(ctx, client.Error(res), "Unexpected status")
		return false
	}

	return true
}

func (c Cert) Delete(ctx context.Context, id string) error {
	res, err := client.Delete(ctx, client.Join(endpointCerts, id), nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	return nil
}

func (c Cert) Upload(ctx context.Context, key, crt []byte) (id string, err error) {
	combined := make([]byte, 0)
	combined = append(combined, key...)
	combined = append(combined, crt...)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("cert", "cert.pem")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, io.NopCloser(bytes.NewReader(combined)))

	err = writer.Close()
	if err != nil {
		return "", err
	}

	res, err := client.Post(ctx, endpointCerts, body, client.SetHeaders(
		map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
	))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", client.Error(res)
	}

	dbResp := CertResponse{}

	if err := client.JSON(res, &dbResp); err != nil {
		return "", err
	}

	return dbResp.Id, nil
}
