package dashboard_client

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/levigross/grequests"
)

type Webhook struct {
	*Client
}

/**
Returns all webhooks from the Dashboard for this org
*/
func (w Webhook) All() ([]v1.WebhookSpec, error) {
	sess := grequests.NewSession(w.opts)

	fullPath := JoinUrl(w.url, endpointWebhooks)

	res, err := sess.Get(fullPath, w.opts)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response WebhookResponse
	if err := res.JSON(&response); err != nil {
		return nil, err
	}

	return response.Webhooks, nil
}

/**
  Attempts to find the webhook by the namespaced name combo.
  When creating a webhook, this is stored as the webhook's "name"

  If no webhook found, return "universal_client.WebhookNotFoundError"
*/
func (w Webhook) Get(namespacedName string) (*v1.WebhookSpec, error) {
	// Returns error if there was a problem fetching webhooks
	list, err := w.All()
	if err != nil {
		return nil, err
	}

	// Iterate through webhooks to find this hook
	for _, hook := range list {
		if hook.Name == namespacedName {
			return &hook, nil
		}
	}

	return nil, universal_client.WebhookNotFoundError
}

/*
	Creates a webhook.  Overwrites the Webhook "name" with the CRD's namespaced name
*/
func (w Webhook) Create(namespacedName string, def *v1.WebhookSpec) error {
	// Check if this webhook exists
	webhook, err := w.Get(namespacedName)
	// if webhook find gives "not found error", great, skip and create!
	if err != nil && err != universal_client.WebhookNotFoundError {
		return err
	} else if webhook != nil {
		return universal_client.WebhookCollisionError
	}

	def.Name = namespacedName
	def.ID = ""

	fullPath := JoinUrl(w.url, endpointWebhooks)

	sess := grequests.NewSession(w.opts)

	res, err := sess.Post(fullPath, &grequests.RequestOptions{JSON: def})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
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
func (w Webhook) Update(namespacedName string, def *v1.WebhookSpec) error {
	webhookToUpdate, err := w.Get(namespacedName)
	if err != nil {
		return err
	}

	if webhookToUpdate == nil {
		return notFoundError
	}

	// Add the unique webhook identifier to "name"
	def.Name = namespacedName
	def.ID = webhookToUpdate.ID       // Needed
	def.OrgID = webhookToUpdate.OrgID // Needed

	sess := grequests.NewSession(w.opts)

	fullPath := JoinUrl(w.url, endpointWebhooks, webhookToUpdate.ID)
	res, err := sess.Put(fullPath, &grequests.RequestOptions{JSON: def})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
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
func (w Webhook) Delete(namespacedName string) error {
	sess := grequests.NewSession(w.opts)

	webhook, err := w.Get(namespacedName)
	if err == universal_client.WebhookNotFoundError {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Unable to lookup webhook.")
	}

	delPath := JoinUrl(w.url, endpointWebhooks, webhook.ID)

	res, err := sess.Delete(delPath, nil)
	if err != nil {
		return err
	}

	// Status 200 is Webhook successfully deleted. 404 is no webhook found, which is desired by us
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("delete webhook API Returned error: %s", res.String())
}
