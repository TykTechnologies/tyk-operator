package universal_client

import (
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalWebhook interface {
	All() ([]tykv1alpha1.WebhookSpec, error)
	Get(id string) (*tykv1alpha1.WebhookSpec, error)
	Create(*tykv1alpha1.WebhookSpec) error
	Update(def *tykv1alpha1.WebhookSpec) error
	Delete(id string) error
}

func CreateOrUpdateWebhook(c UniversalClient, spec *tykv1alpha1.WebhookSpec) error {
	return DoReload(c, func(u UniversalClient) error {
		w := c.Webhook()
		if spec.ID != "" {
			return w.Update(spec)
		}
		return w.Create(spec)
	})
}

// DoReload call HotReload if fn returns nil
func DoReload(c UniversalClient, fn func(u UniversalClient) error) error {
	if err := fn(c); err != nil {
		return nil
	}
	return c.HotReload()
}
