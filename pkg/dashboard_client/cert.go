package dashboard_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type Cert struct {
	*Client
}

type CertificateList struct {
	CertIDs []string `json:"certs"`
	Pages   int      `json:"pages"`
}

// All returns a list of all certificates ID's
func (c *Cert) All(ctx context.Context) ([]string, error) {
	res, err := c.Client.Get(ctx, toURL(endpointCerts), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, universal_client.Error(res)
	}
	var o CertificateList
	err = universal_client.JSON(res, &o)
	if err != nil {
		return nil, err
	}
	return o.CertIDs, nil
}

func (c *Cert) Exists(ctx context.Context, id string) bool {
	res, err := c.Client.Get(ctx, toURL(endpointCerts, id), nil)
	if err != nil {
		c.Log.Error(err, "failed to get certificate")
		return false
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		c.Log.Error(universal_client.Error(res), "Unexepcted status")
		return false
	}
	return true
}

func (c *Cert) Delete(ctx context.Context, id string) error {
	res, err := c.Client.Delete(ctx, toURL(endpointCerts, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return nil
}

func (c *Cert) Upload(ctx context.Context, key []byte, crt []byte) (id string, err error) {
	combined := make([]byte, 0)
	combined = append(combined, key...)
	combined = append(combined, crt...)
	fullPath := toURL(endpointCerts)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("cert", "cert.pem")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, ioutil.NopCloser(bytes.NewReader(combined)))

	err = writer.Close()
	if err != nil {
		return "", err
	}
	res, err := c.Client.Post(ctx, fullPath, body, universal_client.SetHeaders(
		map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
	))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", universal_client.Error(res)
	}
	dbResp := CertResponse{}
	if err := universal_client.JSON(res, &dbResp); err != nil {
		return "", err
	}
	return dbResp.Id, nil
}
