package dashboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/buger/jsonparser"
)

type TykOAS struct{}

func (oas TykOAS) Create(ctx context.Context, id, data string) error {
	quotedID := fmt.Sprintf("\"%s\"", id)
	byteData := []byte(data)

	var err error

	if byteData, err = jsonparser.Set(byteData, []byte(quotedID), "x-tyk-api-gateway", "info", "id"); err != nil {
		return err
	}

	resp, err := client.Post(ctx, endpointOAS, bytes.NewReader(byteData), client.AddHeaders(map[string]string{
		"Content-Type": "application/json",
	}))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return client.Error(resp)
	}

	return nil
}

func (oas TykOAS) Exists(ctx context.Context, id string) bool {
	resp, err := client.Get(ctx, client.Join(endpointOAS, id), nil)
	if err != nil {
		return false
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}

func (oas TykOAS) Update(ctx context.Context, id, data string) error {
	resp, err := client.Put(ctx, client.Join(endpointOAS, id), strings.NewReader(data))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return client.Error(resp)
	}

	return nil
}

func (oas TykOAS) Delete(ctx context.Context, id string) error {
	resp, err := client.Delete(ctx, client.Join(endpointOAS, id), nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return client.Error(resp)
	}

	return nil
}

func (oas TykOAS) Get(ctx context.Context, id string) (string, error) {
	resp, err := client.Get(ctx, client.Join(endpointOAS, id), nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", client.Error(resp)
	}

	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(rbody), nil
}
