/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Please note that SecurityPolicies API is available from v4.1 for Tyk GW (OSS). Therefore, tests for Tyk CE can only
run versions above v4.0.

*/
package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/version"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSecurityPolicy(t *testing.T) {
	const (
		opNs    = "tyk-operator-system"
		apiName = "httpbin-policies"

		// minPolicyApiVersion indicates minimum version of Tyk Gateway (CE) that includes Policy API.
		minPolicyApiVersion = "v4.1"
	)

	var (
		apiDef *v1alpha1.ApiDefinition
		pol    *v1alpha1.SecurityPolicy
		tykEnv environmet.Env
		cl     ctrl.Client
	)

	securityPolicyFeatures := features.New("SecurityPolicy CR must establish correct links").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			v, err := version.ParseGeneric(tykEnv.TykVersion)
			eval.NoErr(err)
			if tykEnv.Environment.Mode == "ce" &&
				v.Major() < 4 || v.LessThan(version.MustParseGeneric(minPolicyApiVersion)) {
				t.Skip("Gateway API does not include the Policy API in versions smaller than v4.1.")
			}

			// Create SecurityPolicy Reconciler.
			cl, err = createTestClient(c.Client())
			eval.NoErr(err)

			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			apiDef = generateApiDef(testNs, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.ObjectMeta.Name = apiName
			})
			_, err = util.CreateOrUpdate(ctx, cl, apiDef, func() error {
				return nil
			})
			eval.NoErr(err)

			// Generate SecurityPolicy CR and create it.
			pol = generateSecurityPolicy(testNs, func(policy *v1alpha1.SecurityPolicy) {
				policy.Spec.AccessRightsArray = []*v1alpha1.AccessDefinition{
					{
						Name:      apiDef.ObjectMeta.Name,
						Namespace: testNs,
						Versions:  []string{"Default"},
					},
				}
			},
			)
			_, err = util.CreateOrUpdate(ctx, cl, pol, func() error {
				return nil
			})
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition and SecurityPolicy CRs must be linked",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				err := wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(pol, func(object k8s.Object) bool {
						return pol.Spec.ID != "" && pol.Spec.OrgID == tykEnv.Org &&
							pol.Status.PolID == controllers.ParsePolicyID(tykEnv.Mode, &pol.Spec)
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(apiDef, func(object k8s.Object) bool {
						apiDefObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						return len(apiDefObj.Status.LinkedByPolicies) == 1 &&
							apiDefObj.Status.LinkedByPolicies[0].Name == testSecurityPolicyName &&
							apiDefObj.Status.LinkedByPolicies[0].Namespace == testNs
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, securityPolicyFeatures)
}
