package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

type OAS struct{}

const oasEndpoint = "/api/apis/oas"

func (o OAS) Create(ctx context.Context, data string) (*model.Result, error) {
	reader := strings.NewReader(data)
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

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("failed to create OAS Api: %s", result.Message)
	}

	return result, nil
}

func (o OAS) Update(ctx context.Context, id, data string) error {
	reader := strings.NewReader(data)
	result := &model.Result{}
	url := fmt.Sprintf("%s/%s", oasEndpoint, id)

	octx := client.GetContext(ctx)

	octx.Log.Info("updating OAS Api", "id", id)

	resp, err := client.Put(ctx, url, reader)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update OAS Api: %s", result.Message)
	}

	return nil
}

func (o OAS) Delete(ctx context.Context, id string) error {
	result := &model.Result{}
	url := fmt.Sprintf("%s/%s", oasEndpoint, id)

	octx := client.GetContext(ctx)

	octx.Log.Info("deleting OAS Api", "id", id)

	resp, err := client.Delete(ctx, url, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete OAS Api: %s", result.Message)
	}

	return nil
}
