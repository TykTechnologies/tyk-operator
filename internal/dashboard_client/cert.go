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

	"github.com/levigross/grequests"
)

type Cert struct {
	*Client
}

func (c *Cert) Get(id string) (string, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointCerts, id)

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return "", nil
}

func (c *Cert) Delete(id string) error {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointCerts, id)

	_, err := sess.Delete(fullPath, nil)
	if err != nil {
		return err
	}

	//if res.StatusCode != http.StatusOK {
	//	return fmt.Errorf("expected 200 OK, got %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	//}
	return nil
}

func (c *Cert) Upload(key []byte, crt []byte) (id string, err error) {
	combined := make([]byte, 0)
	combined = append(combined, key...)
	combined = append(combined, crt...)
	fullPath := JoinUrl(c.url, endpointCerts)

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

	//sess := grequests.NewSession(c.opts)
	//sess.Post(fullPath, &grequests.RequestOptions{
	//	Files: []grequests.FileUpload{
	//		{
	//			FileName:     "",
	//			FileContents: part,
	//			FieldName:    "",
	//			FileMime:     "",
	//		},
	//	},
	//})

	// TODO: is this possible with grequests?
	r, err := http.NewRequest(http.MethodPost, fullPath, body)
	if err != nil {
		return "", err
	}

	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Authorization", c.secret)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()

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
