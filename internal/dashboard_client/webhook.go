package dashboard_client

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"

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
func (p Webhook) All() ([]v1.WebhookSpec, error) {
	fullPath := JoinUrl(p.url, endpointWebhooks)

	res, err := grequests.Get(fullPath, p.opts)
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
  When creating a webhook, this is stored as the policy's "name"
*/
func (w Webhook) Get(namespacedName string) (*v1.WebhookSpec, error) {
	//Returns error if there was a mistake getting all the policies
	list, err := w.All()
	if err != nil {
		return nil, err
	}

	//Iterate through webhooks to find this hook
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
func (p Webhook) Create(def *v1.WebhookSpec, namespacedName string) (string, error) {
	// Check if this webhook exists
	webhook, err := p.Get(namespacedName)
	// if webhook find gives "not found error", great, skip and create!
	if err != nil && err != universal_client.WebhookNotFoundError {
		return "", err
	} else if webhook != nil {
		return "", universal_client.WebhookCollisionError
	}

	def.Name = namespacedName
	def.ID = ""

	// Create
	opts := p.opts
	opts.JSON = def
	fullPath := JoinUrl(p.url, endpointWebhooks)

	res, err := grequests.Post(fullPath, opts)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
		return "", err
	}

	if resMsg.Status != "OK" {
		return "", fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return "", nil
}

/**
Updates a Policy.  Adds the unique identifier namespaced-Name to the
policy's tags so subsequent CRUD opps are possible.
*/
func (p Webhook) Update(def *v1.WebhookSpec, namespacedName string) error {
	webhookToUpdate, err := p.Get(namespacedName)
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

	// Update
	opts := p.opts
	opts.JSON = def

	fullPath := JoinUrl(p.url, endpointWebhooks, webhookToUpdate.ID)
	res, err := grequests.Put(fullPath, opts)
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
	webhook, err := w.Get(namespacedName)
	if err == universal_client.WebhookNotFoundError {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Unable to lookup webhook.")
	}

	delPath := JoinUrl(w.url, endpointWebhooks, webhook.ID)

	res, err := grequests.Delete(delPath, w.opts)
	if err != nil {
		return err
	}

	// Status 200 is Webhook successfully deleted. 404 is no webhook found, which is desired by us
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("delete policy API Returned error: %s", res.String())
}
