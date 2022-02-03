package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	apiDefWithJSONValidationName = "apidef-json-validation"
	apiDefListenPath             = "/validation"
	defaultVersion               = "Default"
	validationSchemaKey          = "key"
	validationSchemaValue        = "value"
	defaultTimeout               = 1 * time.Minute
)

func TestApiDefinitionCreate(t *testing.T) {
	eps := &model.ExtendedPathsSet{
		ValidateJSON: []model.ValidatePathMeta{{
			ErrorResponseCode: 422,
			Path:              "/get",
			Method:            http.MethodGet,
			Schema: &model.JSONValidationSchema{Unstructured: unstructured.Unstructured{
				Object: map[string]interface{}{validationSchemaKey: validationSchemaValue},
			}},
		}},
	}

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with JSON Schema Validation support.
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefWithJSONValidationName
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefListenPath,
					TargetURL:       "http://httpbin.org",
					StripListenPath: true,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have ValidateJSON field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithJSONValidationName, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck

					// 'validate_json' field must exist in the ApiDefinition object.
					if len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.ValidateJSON) != 1 {
						return false
					}

					// Check if 'schema' field is written into ApiDefinition as expected.
					return apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.ValidateJSON[0].
						Schema.UnstructuredContent()[validationSchemaKey] == validationSchemaValue
				}), wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).
		Assess("ApiDefinition must verify user requests",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					resp, getErr := http.Get(fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath))
					if getErr != nil {
						t.Log(getErr)
						return false, nil
					}

					if resp.StatusCode != 200 {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}
