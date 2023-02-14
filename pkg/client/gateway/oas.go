package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

type OAS struct{}

const oasEndpoint = "/tyk/apis/oas"

func (o OAS) Create(ctx context.Context, data []byte) (*model.Result, error) {
	reader := bytes.NewReader(data)
	result := &model.Result{}

	octx := client.GetContext(ctx)

	octx.Log.Info("creating OAS Api")

	resp, err := client.Post(ctx, oasEndpoint, reader)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (o OAS) Update(ctx context.Context, id string, data []byte) (*model.Result, error) {
	reader := bytes.NewReader(data)
	result := &model.Result{}
	url := fmt.Sprintf("%s/%s", oasEndpoint, id)

	octx := client.GetContext(ctx)

	octx.Log.Info("updating OAS Api", "id", id)

	resp, err := client.Put(ctx, url, reader)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (o OAS) Delete(ctx context.Context, id string) (*model.Result, error) {
	result := &model.Result{}
	url := fmt.Sprintf("%s/%s", oasEndpoint, id)

	octx := client.GetContext(ctx)

	octx.Log.Info("deleting OAS Api", "id", id)

	resp, err := client.Delete(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (o OAS) Get(ctx context.Context, id string) ([]byte, error) {
	var data []byte

	url := fmt.Sprintf("%s/%s", oasEndpoint, id)

	octx := client.GetContext(ctx)
	octx.Log.Info("getting OAS Api", "id", id)

	resp, err := client.Get(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
