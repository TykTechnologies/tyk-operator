package klient

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/dashboard"
	"github.com/TykTechnologies/tyk-operator/pkg/client/gateway"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

var _ universal.Client = (*Client)(nil)

var Universal = Client{}

func get(ctx context.Context) universal.Client {
	r := client.GetContext(ctx)
	if r.Env.Mode == "pro" {
		return dashboard.Client{}
	}

	return gateway.Client{}
}

// Client implements universal.Client but picks the correct client dynamically based on context.Context
type Client struct{}

func (Client) HotReload(ctx context.Context) error {
	return get(ctx).HotReload(ctx)
}

func (Client) Api() universal.Api {
	return Api{}
}

func (Client) Portal() universal.Portal {
	return Portal{}
}

func (Client) Certificate() universal.Certificate {
	return Certificate{}
}

func (Client) OAS() universal.OAS {
	return OAS{}
}

type Api struct{}

func (Api) Create(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error) {
	return get(ctx).Api().Create(ctx, def)
}

func (Api) Get(ctx context.Context, id string) (*model.APIDefinitionSpec, error) {
	return get(ctx).Api().Get(ctx, id)
}

func (Api) Update(ctx context.Context, spec *model.APIDefinitionSpec) (*model.Result, error) {
	return get(ctx).Api().Update(ctx, spec)
}

func (Api) Delete(ctx context.Context, id string) (*model.Result, error) {
	return get(ctx).Api().Delete(ctx, id)
}

func (Api) List(ctx context.Context, options ...model.ListAPIOptions) (*model.APIDefinitionSpecList, error) {
	return get(ctx).Api().List(ctx, options...)
}

type Portal struct{}

func (Portal) Policy() universal.Policy {
	return Policy{}
}

func (Portal) Documentation() universal.Documentation {
	return Documentation{}
}

func (Portal) Catalogue() universal.Catalogue {
	return Catalogue{}
}

func (Portal) Configuration() universal.Configuration {
	return Configuration{}
}

type Policy struct{}

func (Policy) All(ctx context.Context) ([]v1alpha1.SecurityPolicySpec, error) {
	return get(ctx).Portal().Policy().All(ctx)
}

func (Policy) Get(ctx context.Context, id string) (*v1alpha1.SecurityPolicySpec, error) {
	return get(ctx).Portal().Policy().Get(ctx, id)
}

func (Policy) Create(ctx context.Context, def *v1alpha1.SecurityPolicySpec) error {
	return get(ctx).Portal().Policy().Create(ctx, def)
}

func (Policy) Update(ctx context.Context, def *v1alpha1.SecurityPolicySpec) error {
	return get(ctx).Portal().Policy().Update(ctx, def)
}

func (Policy) Delete(ctx context.Context, id string) error {
	return get(ctx).Portal().Policy().Delete(ctx, id)
}

type Documentation struct{}

func (Documentation) Upload(ctx context.Context, o *model.APIDocumentation) (*model.Result, error) {
	return get(ctx).Portal().Documentation().Upload(ctx, o)
}

func (Documentation) Delete(ctx context.Context, id string) (*model.Result, error) {
	return get(ctx).Portal().Documentation().Delete(ctx, id)
}

type Catalogue struct{}

func (Catalogue) Get(ctx context.Context) (*model.APICatalogue, error) {
	return get(ctx).Portal().Catalogue().Get(ctx)
}

func (Catalogue) Create(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return get(ctx).Portal().Catalogue().Create(ctx, o)
}

func (Catalogue) Update(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return get(ctx).Portal().Catalogue().Update(ctx, o)
}

type Configuration struct{}

func (Configuration) Get(ctx context.Context) (*model.PortalModelPortalConfig, error) {
	return get(ctx).Portal().Configuration().Get(ctx)
}

func (Configuration) Create(
	ctx context.Context, o *model.PortalModelPortalConfig,
) (*model.Result, error) {
	return get(ctx).Portal().Configuration().Create(ctx, o)
}

func (Configuration) Update(
	ctx context.Context, o *model.PortalModelPortalConfig,
) (*model.Result, error) {
	return get(ctx).Portal().Configuration().Update(ctx, o)
}

type Certificate struct{}

func (Certificate) All(ctx context.Context) ([]string, error) {
	return get(ctx).Certificate().All(ctx)
}

func (Certificate) Upload(ctx context.Context, key, crt []byte) (id string, err error) {
	return get(ctx).Certificate().Upload(ctx, key, crt)
}

func (Certificate) Delete(ctx context.Context, id string) error {
	return get(ctx).Certificate().Delete(ctx, id)
}

func (Certificate) Exists(ctx context.Context, id string) bool {
	return get(ctx).Certificate().Exists(ctx, id)
}

type OAS struct{}

func (OAS) Delete(ctx context.Context, id string) error {
	return get(ctx).OAS().Delete(ctx, id)
}

func (OAS) Create(ctx context.Context, data string) (*model.Result, error) {
	return get(ctx).OAS().Create(ctx, data)
}

func (OAS) Update(ctx context.Context, id string, data string) error {
	return get(ctx).OAS().Update(ctx, id, data)
}
