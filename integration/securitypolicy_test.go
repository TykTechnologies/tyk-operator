package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSecurityPolicyForGraphQL(t *testing.T) {
	eval := is.New(t)
	const (
		queryName     = "Query"
		accountsField = "accounts"
		allField      = "*"
		fieldName     = "getMovers"
		limit         = int64(2)
	)

	var reqCtx context.Context

	securityPolicyForGraphQL := features.New("GraphQL specific Security Policy configurations").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}
			err := c.Client().Resources(operatorNamespace).Get(ctx, operatorSecret, operatorNamespace, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			if tykEnv.Mode == "ce" {
				t.Skip("SecurityPolicy API is not implemented in CE yet")
			}

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess(
			"Create a SecurityPolicy for GraphQL",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				apiDefCR, err := createTestAPIDef(ctx, testNs, nil, c)
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, apiDefCR)
				eval.NoErr(err)

				policyCR, err := createTestPolicy(ctx, c, testNs, func(policy *v1alpha1.SecurityPolicy) {
					policy.Spec.AccessRightsArray = []*v1alpha1.AccessDefinition{
						{
							Name:      apiDefCR.Name,
							Namespace: apiDefCR.Namespace,
							AllowedTypes: []v1alpha1.GraphQLType{
								{Name: queryName, Fields: []string{accountsField}},
							},
							RestrictedTypes: []v1alpha1.GraphQLType{{Name: queryName, Fields: []string{allField}}},
							FieldAccessRights: []v1alpha1.FieldAccessDefinition{
								{
									TypeName:  queryName,
									FieldName: fieldName,
									Limits:    v1alpha1.FieldLimits{MaxQueryDepth: limit},
								},
							},
						},
					}
				})
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, policyCR)
				eval.NoErr(err)

				return ctx
			}).
		Assess(
			"Validate fields are set properly on k8s and Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				var pol v1alpha1.SecurityPolicy

				err := c.Client().Resources().Get(ctx, testSecurityPolicy, testNs, &pol)
				eval.NoErr(err)

				validPolSpec := func(polSpec *v1alpha1.SecurityPolicySpec, k8s bool) bool {
					eval.Equal(len(polSpec.AccessRightsArray), 1)

					if k8s {
						eval.Equal(polSpec.AccessRightsArray[0].Name, testApiDef)
						eval.Equal(polSpec.AccessRightsArray[0].Namespace, testNs)
					}

					eval.Equal(len(polSpec.AccessRightsArray[0].AllowedTypes), 1)
					eval.Equal(polSpec.AccessRightsArray[0].AllowedTypes[0].Name, queryName)
					eval.Equal(len(polSpec.AccessRightsArray[0].AllowedTypes[0].Fields), 1)
					eval.Equal(polSpec.AccessRightsArray[0].AllowedTypes[0].Fields[0], accountsField)

					eval.Equal(len(polSpec.AccessRightsArray[0].RestrictedTypes), 1)
					eval.Equal(polSpec.AccessRightsArray[0].RestrictedTypes[0].Name, queryName)
					eval.Equal(len(polSpec.AccessRightsArray[0].RestrictedTypes[0].Fields), 1)
					eval.Equal(polSpec.AccessRightsArray[0].RestrictedTypes[0].Fields[0], allField)

					eval.Equal(len(polSpec.AccessRightsArray[0].FieldAccessRights), 1)
					eval.Equal(polSpec.AccessRightsArray[0].FieldAccessRights[0].TypeName, queryName)
					eval.Equal(polSpec.AccessRightsArray[0].FieldAccessRights[0].FieldName, fieldName)
					eval.Equal(polSpec.AccessRightsArray[0].FieldAccessRights[0].Limits.MaxQueryDepth, limit)

					return true
				}

				eval.True(validPolSpec(&pol.Spec, true))

				policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, pol.Status.PolID)
				eval.NoErr(err)

				eval.True(validPolSpec(policyOnTyk, false))

				return ctx
			}).
		Feature()

	testenv.Test(t, securityPolicyForGraphQL)
}
