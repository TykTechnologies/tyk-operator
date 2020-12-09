package dashboard_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type Cert struct {
	*Client
}

func (c *Cert) Get(id string) (string, error) {
	res, err := c.Client.Get(c.Env.JoinURL(endpointCerts, id), nil)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	// TODO:(gernest) return certificate data?
	return "", nil
}

func (c *Cert) Delete(id string) error {
	res, err := c.Client.Delete(c.Env.JoinURL(endpointCerts, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return nil
}

func (c *Cert) Upload(key []byte, crt []byte) (id string, err error) {
	combined := make([]byte, 0)
	combined = append(combined, key...)
	combined = append(combined, crt...)
	fullPath := c.Env.JoinURL(endpointCerts)

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
	res, err := c.Client.Post(fullPath, body, universal_client.AddHeaders(
		map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
	))
	defer res.Body.Close()

	rBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		errStruct := CertErrorResponse{}
		json.Unmarshal(rBody, &errStruct)

		reg := regexp.MustCompile(`Could not create certificate: Certificate with (?P<ID>.*) id already exists`)

		matches := reg.FindStringSubmatch(errStruct.Message)

		if len(matches) != 2 {
			return "", fmt.Errorf("api returned error: %v", string(rBody))
		}

		return matches[1], nil
	}
	dbResp := CertResponse{}
	if err := json.Unmarshal(rBody, &dbResp); err != nil {
		return "", err
	}
	if strings.ToLower(dbResp.Status) != "ok" {
		return "", fmt.Errorf("non ok response message: %v", dbResp.Message)
	}

	return dbResp.Id, nil
}
