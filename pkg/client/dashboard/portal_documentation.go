package dashboard

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

const endpointDocumentation = "/api/portal/documentation/"

var _ universal.Documentation = Documentation{}

type Documentation struct{}

func (Documentation) Upload(
	ctx context.Context, o *model.APIDocumentation,
) (*model.Result, error) {
	return client.Result(client.PostJSON(ctx, endpointDocumentation, o))
}

func (Documentation) Delete(ctx context.Context, id string) (*model.Result, error) {
	return client.Result(client.Delete(
		ctx, client.Join(endpointDocumentation, id), nil,
	))
}
