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
	ctrl "sigs.k8s.io/controller-runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// Please note that SecurityPolicies API is available from v4 for Tyk GW (OSS). Therefore, tests for Tyk CE can only
// run versions greater than v4.

func TestSecurityPolicy(t *testing.T) {
	const (
		opNs    = "tyk-operator-system"
		apiName = "httpbin-policies"

		supportedMajorTykVersion = uint(4)
	)

	var (
		apiDef    *v1alpha1.ApiDefinition
		apiDefRec controllers.ApiDefinitionReconciler
		pol       *v1alpha1.SecurityPolicy
		policyRec controllers.SecurityPolicyReconciler
		tykEnv    environmet.Env
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
			if tykEnv.Environment.Mode == "ce" && v.Major() < 4 {
				t.Skip("Gateway API does not include Policy API in versions smaller than 4.")
			}

			// Create SecurityPolicy Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			policyRec = controllers.SecurityPolicyReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			apiDefRec = controllers.ApiDefinitionReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			apiDef = generateApiDef(testNs, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.ObjectMeta.Name = apiName
			})
			_, err = util.CreateOrUpdate(ctx, apiDefRec.Client, apiDef, func() error {
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
			_, err = util.CreateOrUpdate(ctx, policyRec.Client, pol, func() error {
				return nil
			})

			return ctx
		}).
		Assess("ApiDefinition and SecurityPolicy CRs must be linked",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				// Wait for reconciliation of ApiDefinition CR.
				err := wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(apiDef, func(object k8s.Object) bool {
						_, err := apiDefRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(apiDef)})
						return err == nil
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// Wait for reconciliation of SecurityPolicy CR; so that, the SecurityPolicy CR is updated.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(pol, func(object k8s.Object) bool {
						_, err := policyRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(pol)})
						if err != nil {
							return false
						}

						return pol.Spec.ID != "" && pol.Spec.OrgID == policyRec.Env.Org &&
							pol.Status.PolID == controllers.ParsePolicyID(policyRec.Env.Mode, &pol.Spec)
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// After reconciliation, check that ApiDefinition CR is updated according to SecurityPolicy CR.
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
