package dashboard_client

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type Webhook struct {
	*Client
}

/**
Returns all webhooks from the Dashboard for this org
*/
func (w Webhook) All(ctx context.Context) ([]v1.WebhookSpec, error) {

	res, err := w.Client.Get(ctx, toURL(endpointWebhooks), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response WebhookResponse
	if err := universal_client.JSON(res, &response); err != nil {
		return nil, err
	}
	return response.Webhooks, nil
}

/**
  Attempts to find the webhook by the namespaced name combo.
  When creating a webhook, this is stored as the webhook's "name"

  If no webhook found, return "universal_client.WebhookNotFoundError"
*/
func (w Webhook) Get(ctx context.Context, hookID string) (*v1.WebhookSpec, error) {
	res, err := w.Client.Get(ctx, toURL(endpointWebhooks, hookID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var h v1.WebhookSpec
	err = universal_client.JSON(res, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

/*
	Creates a webhook.  Overwrites the Webhook "name" with the CRD's namespaced name
*/
func (w Webhook) Create(ctx context.Context, def *v1.WebhookSpec) error {
	res, err := w.Client.PostJSON(ctx, toURL(endpointWebhooks), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

/**
Updates a Webhook.  Adds the unique identifier namespaced-Name to the
webhook's "name" so subsequent CRUD opps are possible.
*/
func (w Webhook) Update(ctx context.Context, def *v1.WebhookSpec) error {
	res, err := w.Client.PutJSON(ctx, toURL(endpointWebhooks, def.ID), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

/**
Tries to delete a Webhook by first attempting to do a lookup on it.
If webhook does not exist, move on, nothing to delete
*/
func (w Webhook) Delete(ctx context.Context, id string) error {
	res, err := w.Client.Delete(ctx, toURL(endpointWebhooks, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		return nil
	default:
		return universal_client.Error(res)
	}
}
