package universal_client

import (
	"context"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalWebhook interface {
	All(ctx context.Context) ([]tykv1alpha1.WebhookSpec, error)
	Get(ctx context.Context, id string) (*tykv1alpha1.WebhookSpec, error)
	Create(context.Context, *tykv1alpha1.WebhookSpec) error
	Update(ctx context.Context, def *tykv1alpha1.WebhookSpec) error
	Delete(ctx context.Context, id string) error
}

func CreateOrUpdateWebhook(ctx context.Context, c UniversalClient, spec *tykv1alpha1.WebhookSpec) error {
	return DoReload(ctx, c, func(u UniversalClient) error {
		w := c.Webhook()
		if spec.ID != "" {
			return w.Update(ctx, spec)
		}
		return w.Create(ctx, spec)
	})
}

// DoReload call HotReload if fn returns nil
func DoReload(ctx context.Context, c UniversalClient, fn func(u UniversalClient) error) error {
	if err := fn(c); err != nil {
		return nil
	}
	return c.HotReload(ctx)
}
