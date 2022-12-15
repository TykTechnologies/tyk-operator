package integration

import (
	"context"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSecurityPolicyStatusIsUpdated(t *testing.T) {
	is := is.New(t)

	api1Name := "test-api-1-status"
	api2Name := "test-api-2-status"
	policyName := "test-policy"

	mode := os.Getenv("TYK_MODE")
	if mode == "ce" {
		t.Skip("Skipping security policy test in CE mode")
	}

	policyCreate := features.New("SecurityPolicy status is updated").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNs, ok := ctx.Value(ctxNSKey).(string)
			is.True(ok)

			_, err := createTestAPIDef(ctx, c, testNs, func(ad *v1alpha1.ApiDefinition) {
				ad.Name = api1Name
				ad.Spec.Name = api1Name
				ad.Spec.Proxy.ListenPath = "/test-api-1"
			})
			is.NoErr(err)

			_, err = createTestAPIDef(ctx, c, testNs, func(ad *v1alpha1.ApiDefinition) {
				ad.Name = api2Name
				ad.Spec.Name = api2Name
				ad.Spec.Proxy.ListenPath = "/test-api-2"
			})
			is.NoErr(err)

			// ensure API is created on Tyk before creating policy
			apiDef := &v1alpha1.ApiDefinition{ObjectMeta: metav1.ObjectMeta{Name: api1Name, Namespace: testNs}}
			err = waitForTykResourceCreation(c, apiDef)
			is.NoErr(err)

			_, err = createTestPolicy(ctx, testNs, func(policy *v1alpha1.SecurityPolicy) {
				policy.Name = policyName
				policy.Spec.Name = policyName + testNs
				policy.Spec.AccessRightsArray = []*v1alpha1.AccessDefinition{{Name: api1Name, Namespace: testNs}}
			}, c)
			is.NoErr(err)

			pol := v1alpha1.SecurityPolicy{ObjectMeta: metav1.ObjectMeta{Name: policyName, Namespace: testNs}}
			err = waitForTykResourceCreation(c, &pol)
			is.NoErr(err)

			return ctx
		}).Assess("validate links are created properly",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var pol v1alpha1.SecurityPolicy
			var api v1alpha1.ApiDefinition

			testNs, ok := ctx.Value(ctxNSKey).(string)
			is.True(ok)

			// check status of policy
			err := c.Client().Resources().Get(ctx, policyName, testNs, &pol)
			is.NoErr(err)

			is.True(len(pol.Status.LinkedAPIs) != 0)
			is.Equal(pol.Status.LinkedAPIs[0].Name, api1Name)

			// check status of ApiDefinition
			err = c.Client().Resources().Get(ctx, api1Name, testNs, &api)
			is.NoErr(err)

			is.True(len(api.Status.LinkedByPolicies) != 0)
			is.Equal(api.Status.LinkedByPolicies[0].Name, policyName)

			return ctx
		}).Assess("Add new api in the access rights",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var updatePolicy v1alpha1.SecurityPolicy
			testNs, ok := ctx.Value(ctxNSKey).(string)
			is.True(ok)

			err := c.Client().Resources().Get(ctx, policyName, testNs, &updatePolicy)
			is.NoErr(err)

			updatePolicy.Spec.AccessRightsArray = append(updatePolicy.Spec.AccessRightsArray,
				&v1alpha1.AccessDefinition{Name: api2Name, Namespace: testNs})

			err = c.Client().Resources().Update(ctx, &updatePolicy)
			is.NoErr(err)

			var pol v1alpha1.SecurityPolicy
			pol.Name = policyName
			pol.Namespace = testNs

			// wait until status of policy is updated.
			err = wait.For(conditions.New(c.Client().Resources()).ResourceMatch(&pol, func(object k8s.Object) bool {
				pol, ok := object.(*v1alpha1.SecurityPolicy)
				if !ok {
					return false
				}

				if len(pol.Status.LinkedAPIs) == 2 {
					return true
				}

				return false
			}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
			is.NoErr(err)

			is.True(pol.Status.LinkedAPIs[0].Name == api2Name || pol.Status.LinkedAPIs[1].Name == api2Name)

			var api v1alpha1.ApiDefinition
			err = c.Client().Resources().Get(ctx, api2Name, testNs, &api)
			is.NoErr(err)

			is.True(len(api.Status.LinkedByPolicies) != 0)
			is.Equal(api.Status.LinkedByPolicies[0].Name, policyName)

			return ctx
		}).Assess("Delete access rights", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		var updatePolicy v1alpha1.SecurityPolicy

		testNs, ok := ctx.Value(ctxNSKey).(string)
		is.True(ok)

		err := c.Client().Resources().Get(ctx, policyName, testNs, &updatePolicy)
		is.NoErr(err)

		updatePolicy.Spec.AccessRightsArray = nil

		err = c.Client().Resources().Update(ctx, &updatePolicy)
		is.NoErr(err)

		var pol v1alpha1.SecurityPolicy

		pol.Name = policyName
		pol.Namespace = testNs

		err = wait.For(conditions.New(c.Client().Resources()).ResourceMatch(&pol, func(object k8s.Object) bool {
			pol, ok := object.(*v1alpha1.SecurityPolicy)
			if !ok {
				return false
			}

			return pol.Status.LinkedAPIs == nil
		}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
		is.NoErr(err)

		var api v1alpha1.ApiDefinition

		err = c.Client().Resources().Get(ctx, api1Name, testNs, &api)
		is.NoErr(err)

		is.True(api.Status.LinkedByPolicies == nil)

		err = c.Client().Resources().Get(ctx, api2Name, testNs, &api)
		is.NoErr(err)

		is.True(api.Status.LinkedByPolicies == nil)

		return ctx
	}).Feature()

	testenv.Test(t, policyCreate)
}
