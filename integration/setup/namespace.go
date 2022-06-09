package setup

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/integration/common"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CreateNamespace(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
	name := envconf.RandomName("tyk-operator", 32)

	ctx = context.WithValue(ctx, common.CtxNSKey, name)

	nsObj := v1.Namespace{}
	nsObj.Name = name
	return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
}

func DeleteNamespace(ctx context.Context, envconf *envconf.Config, t *testing.T) (context.Context, error) {
	name := ctx.Value(common.CtxNSKey)

	nsObj := v1.Namespace{}
	nsObj.Name = name.(string)
	return ctx, envconf.Client().Resources().Delete(ctx, &nsObj)
}
