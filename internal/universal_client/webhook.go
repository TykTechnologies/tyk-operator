package universal_client

import (
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/pkg/errors"
)

type UniversalWebhook interface {
	All() ([]tykv1alpha1.WebhookSpec, error)
	Get(namespacedName string) (*tykv1alpha1.WebhookSpec, error)
	Create(namespacedName string, def *tykv1alpha1.WebhookSpec) error
	Update(namespacedName string, def *tykv1alpha1.WebhookSpec) error
	Delete(namespacedName string) error
}

var (
	WebhookCollisionError = errors.New("webhook already exists")
	WebhookNotFoundError  = errors.New("webhook not found")
)

func applyDefaultsWebhooks(spec *tykv1alpha1.WebhookSpec) {
}

func CreateOrUpdateWebhook(c UniversalClient, spec *tykv1alpha1.WebhookSpec, namespacedName string) error {
	var err error

	webhook, err := c.Webhook().Get(namespacedName)
	if err != nil && err != WebhookNotFoundError {
		return errors.Wrap(err, "Unable to communicate with Client")
	}

	// Create
	if webhook == nil {

		err = c.Webhook().Create(namespacedName, spec)
		if err != nil {
			return errors.Wrap(err, "unable to create webhook")
		}

	} else { // Update

		err = c.Webhook().Update(namespacedName, spec)
		if err != nil {
			return errors.Wrap(err, "unable to update webhook")
		}

	}

	_ = c.HotReload()

	return nil
}
