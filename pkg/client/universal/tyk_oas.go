package universal

import "context"

type TykOAS interface {
	Create(ctx context.Context, id, data string) error
	Exists(ctx context.Context, id string) bool
	Update(ctx context.Context, id, data string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (string, error)
}
