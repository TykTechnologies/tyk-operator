package universal_client

import (
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/pkg/errors"
)

type UniversalWebhook interface {
	All() ([]tykv1alpha1.WebhookSpec, error)
	Get(namespacedName string) (*tykv1alpha1.WebhookSpec, error)
	Create(namespacedName string, def *tykv1alpha1.WebhookSpec) (string, error)
	Update(namespacedName string, def *tykv1alpha1.WebhookSpec) error
	Delete(namespacedName string) error
}

var (
	WebhookCollisionError = errors.New("webhook already exists")
	WebhookNotFoundError  = errors.New("webhook not found")
)

func applyDefaultsWebhooks(spec *tykv1alpha1.WebhookSpec) {
}

func CreateOrUpdateWebhook(c UniversalClient, spec *tykv1alpha1.WebhookSpec, namespacedName string) (*tykv1alpha1.WebhookSpec, error) {
	var err error

	webhook := tykv1alpha1.WebhookSpec{}
	applyDefaultsWebhooks(spec)

	_ = c.HotReload()

	return &webhook, err
}
