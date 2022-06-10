package setup

import (
	"context"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func MultiTenancy(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	return c1, nil
}

func TeardownMultiTenancy(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	return c1, nil
}
