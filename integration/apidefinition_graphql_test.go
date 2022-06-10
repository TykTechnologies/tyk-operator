package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/integration/internal/common"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestAPIDefinition_GraphQL_ExecutionMode(t *testing.T) {
	createAPI := features.New("Create GraphQL API").
		Assess("validate_executionMode", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS := ctx.Value(common.CtxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			tests := map[string]struct {
				ExecutionMode string
				ReturnErr     bool
			}{
				"invalid execution mode": {
					ExecutionMode: "invalid",
					ReturnErr:     true,
				},
				"empty execution mode": {
					ExecutionMode: "",
					ReturnErr:     true,
				},
				"valid execution engine mode": {
					ExecutionMode: "executionEngine",
					ReturnErr:     false,
				},
				"valid proxy only mode": {
					ExecutionMode: "proxyOnly",
					ReturnErr:     false,
				},
			}

			for n, tc := range tests {
				t.Run(n, func(t *testing.T) {
					_, err := common.CreateTestAPIDef(ctx, testNS, func(ad *v1alpha1.ApiDefinition) {
						ad.Name = fmt.Sprintf("%s-%s", ad.Name, uuid.New().String())
						ad.Spec.Name = ad.Name
						ad.Spec.GraphQL = &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: model.GraphQLExecutionMode(tc.ExecutionMode),
						}
					}, c)

					t.Log("Error=", err)
					is.Equal(tc.ReturnErr, err != nil)
				})
			}

			return ctx
		}).Feature()

	testenv.Test(t, createAPI)
}
