package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/matryer/is"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func verifyPolicyApiVersion(t *testing.T, tykEnv *environmet.Env) {
	v, err := version.ParseGeneric(tykEnv.TykVersion)
	if err != nil {
		t.Fatal("failed to parse Tyk Version")
	}

	if tykEnv.Mode == "ce" && !v.AtLeast(minPolicyGwVersion) {
		t.Skip("Security Policies API in CE mode requires at least Tyk v4.1")
	}
}

func TestSecurityPolicyStatusIsUpdated(t *testing.T) {
	eval := is.New(t)

	api1Name := "test-api-1-status"
	api2Name := "test-api-2-status"
	policyName := "test-policy"

	policyCreate := features.New("SecurityPolicy status is updated").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			verifyPolicyApiVersion(t, &tykEnv)

			api1, err := createTestAPIDef(ctx, c, testNs, func(ad *v1alpha1.ApiDefinition) {
				ad.Name = api1Name
				ad.Spec.Name = api1Name
				ad.Spec.Proxy.ListenPath = "/test-api-1"
			})
			eval.NoErr(err)

			api2, err := createTestAPIDef(ctx, c, testNs, func(ad *v1alpha1.ApiDefinition) {
				ad.Name = api2Name
				ad.Spec.Name = api2Name
				ad.Spec.Proxy.ListenPath = "/test-api-2"
			})
			eval.NoErr(err)

			// ensure API is created on Tyk before creating policy
			err = waitForTykResourceCreation(c, api1)
			eval.NoErr(err)

			// ensure API is created on Tyk before creating policy
			err = waitForTykResourceCreation(c, api2)
			eval.NoErr(err)

			pol, err := createTestPolicy(ctx, c, testNs, func(policy *v1alpha1.SecurityPolicy) {
				policy.Name = policyName
				policy.Spec.Name = policyName + testNs
				policy.Spec.AccessRightsArray = []*v1alpha1.AccessDefinition{{Name: api1Name, Namespace: testNs}}
			})
			eval.NoErr(err)

			err = waitForTykResourceCreation(c, pol)
			eval.NoErr(err)

			return ctx
		}).Assess("validate links are created properly",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			var pol v1alpha1.SecurityPolicy
			// check status of policy
			err := c.Client().Resources().Get(ctx, policyName, testNs, &pol)
			eval.NoErr(err)

			eval.True(len(pol.Status.LinkedAPIs) != 0)
			eval.Equal(pol.Status.LinkedAPIs[0].Name, api1Name)

			var api v1alpha1.ApiDefinition
			// check status of ApiDefinition
			err = c.Client().Resources().Get(ctx, api1Name, testNs, &api)
			eval.NoErr(err)

			eval.True(len(api.Status.LinkedByPolicies) != 0)
			eval.Equal(api.Status.LinkedByPolicies[0].Name, policyName)

			return ctx
		}).Assess("Add a new api in the access rights",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			var updatePolicy v1alpha1.SecurityPolicy
			err := c.Client().Resources().Get(ctx, policyName, testNs, &updatePolicy)
			eval.NoErr(err)

			updatePolicy.Spec.AccessRightsArray = append(updatePolicy.Spec.AccessRightsArray,
				&v1alpha1.AccessDefinition{Name: api2Name, Namespace: testNs})

			err = c.Client().Resources().Update(ctx, &updatePolicy)
			eval.NoErr(err)

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
			eval.NoErr(err)

			eval.True(pol.Status.LinkedAPIs[0].Name == api2Name || pol.Status.LinkedAPIs[1].Name == api2Name)

			var api v1alpha1.ApiDefinition
			err = c.Client().Resources().Get(ctx, api2Name, testNs, &api)
			eval.NoErr(err)

			eval.True(len(api.Status.LinkedByPolicies) != 0)
			eval.Equal(api.Status.LinkedByPolicies[0].Name, policyName)

			return ctx
		}).Assess("Delete access rights", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		var updatePolicy v1alpha1.SecurityPolicy

		testNs, ok := ctx.Value(ctxNSKey).(string)
		eval.True(ok)

		err := c.Client().Resources().Get(ctx, policyName, testNs, &updatePolicy)
		eval.NoErr(err)

		updatePolicy.Spec.AccessRightsArray = nil

		err = c.Client().Resources().Update(ctx, &updatePolicy)
		eval.NoErr(err)

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
		eval.NoErr(err)

		var api v1alpha1.ApiDefinition

		err = c.Client().Resources().Get(ctx, api1Name, testNs, &api)
		eval.NoErr(err)

		eval.True(api.Status.LinkedByPolicies == nil)

		err = c.Client().Resources().Get(ctx, api2Name, testNs, &api)
		eval.NoErr(err)

		eval.True(api.Status.LinkedByPolicies == nil)

		return ctx
	}).Feature()

	testenv.Test(t, policyCreate)
}

// minPolicyGwVersion represents the minimum Tyk Gateway version required to use Policy API.
var minPolicyGwVersion = version.MustParseGeneric("v4.1.0")

// createPolicyOnTyk creates given spec and reloads Tyk. Although reloading is not required for Pro,
// it is needed for CE.
func createPolicyOnTyk(ctx context.Context, spec *v1alpha1.SecurityPolicySpec) error {
	err := wait.For(func() (done bool, err error) {
		err = klient.Universal.Portal().Policy().Create(ctx, spec)
		if err != nil {
			return false, err
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			return false, err
		}

		return true, nil
	}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// updatePolicyOnTyk updates given spec and reloads Tyk. Although reloading is not required for Pro,
// it is needed for CE.
func updatePolicyOnTyk(ctx context.Context, spec *v1alpha1.SecurityPolicySpec) error {
	err := wait.For(func() (done bool, err error) {
		err = klient.Universal.Portal().Policy().Update(ctx, spec)
		if err != nil {
			return false, err
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			return false, err
		}

		return true, nil
	}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// deletePolicyOnTyk deletes SecurityPolicy identified by given ID and reloads Tyk.
// Although reloading is not required for Pro, it is needed for CE.
func deletePolicyOnTyk(ctx context.Context, id string) error {
	err := wait.For(func() (done bool, err error) {
		err = klient.Universal.Portal().Policy().Delete(ctx, id)
		if err != nil {
			return false, err
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			return false, err
		}

		return true, nil
	}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

func TestSecurityPolicyMigration(t *testing.T) {
	const (
		initialK8sPolicyTag   = "sample-tag"
		initialK8sPolicyRate  = 50
		initialK8sPolicyState = "deny"
		existingPolicyID      = "my-testing-id"
	)

	var (
		spec = v1alpha1.SecurityPolicySpec{
			ID:     existingPolicyID,
			Name:   "existing-spec",
			State:  "draft",
			Rate:   34,
			Active: true,
			Tags:   []string{"testing"},
		}
		polRec   controllers.SecurityPolicyReconciler
		policyCR v1alpha1.SecurityPolicy
		reqCtx   context.Context

		hasSameValues = func(m v1alpha1.OperatorContextMode, k8s, tyk *v1alpha1.SecurityPolicySpec, k8sID string) bool {
			if m == "pro" {
				return k8s.MID == tyk.MID &&
					k8sID == tyk.MID &&
					len(k8s.Tags) == len(tyk.Tags) &&
					len(tyk.Tags) == 1 &&
					tyk.Tags[0] == initialK8sPolicyTag &&
					tyk.Rate == initialK8sPolicyRate &&
					tyk.State == initialK8sPolicyState
			}

			return k8s.MID == tyk.ID &&
				k8sID == tyk.ID &&
				len(k8s.Tags) == len(tyk.Tags) &&
				len(tyk.Tags) == 1 &&
				tyk.Tags[0] == initialK8sPolicyTag &&
				tyk.Rate == initialK8sPolicyRate
		}
	)

	securityPolicyMigrationFeatures := features.New("Existing Security Policy Migration to K8s").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			verifyPolicyApiVersion(t, &tykEnv)

			testCl, err := createTestClient(c.Client())
			eval.NoErr(err)

			polRec = controllers.SecurityPolicyReconciler{
				Client: testCl,
				Log:    log.NullLogger{},
				Scheme: testCl.Scheme(),
				Env:    tykEnv,
			}

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: polRec.Env,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Migrate a simple Security Policy from Dashboard to K8s",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				err := createPolicyOnTyk(reqCtx, &spec)
				eval.NoErr(err)

				policyCR = v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "sample-policy", Namespace: testNs},
					Spec: v1alpha1.SecurityPolicySpec{
						Name:   envconf.RandomName("sample-policy", 32),
						ID:     spec.MID,
						Active: true,
						State:  initialK8sPolicyState,
						Rate:   initialK8sPolicyRate,
						Tags:   []string{initialK8sPolicyTag},
					},
				}

				err = c.Client().Resources().Create(ctx, &policyCR)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)
						eval.True(len(policyOnK8s.Status.PolID) > 0)

						policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)

						eval.True(
							hasSameValues(polRec.Env.Mode, &policyOnK8s.Spec, policyOnTyk, policyOnK8s.Status.PolID),
						)
						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("Create a drift between k8s and Dashboard",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// In order to create a drift between k8s and Dashboard, first delete the SecurityPolicy controlled
				// by Operator from Dashboard. Now, in the next reconciliation, Operator should recreate same
				// SecurityPolicy on Tyk Side based on the SecurityPolicy CR on k8s.

				eval := is.New(t)

				// Delete an existing Policy from Dashboard to create drift between Tyk and K8s state.
				err := deletePolicyOnTyk(reqCtx, policyCR.Status.PolID)
				eval.NoErr(err)

				previousPolicyID := policyCR.Status.PolID

				// Ensure that policy is deleted from Tyk.
				err = wait.For(func() (done bool, err error) {
					_, err = klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
					if tykClient.IsNotFound(err) {
						return true, nil
					}

					return false, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				// After reconciliation, Operator should detect drift and recreate non-existing policy on Tyk side.
				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						newSpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)
						eval.True(hasSameValues(polRec.Env.Mode, &policyOnK8s.Spec, newSpec, policyOnK8s.Status.PolID))

						// Ensure that the Policy is accessible via the previous ID
						newSpec, err = klient.Universal.Portal().Policy().Get(reqCtx, previousPolicyID)
						eval.NoErr(err)
						eval.True(hasSameValues(polRec.Env.Mode, &policyOnK8s.Spec, newSpec, policyOnK8s.Status.PolID))

						return true
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Must be restored from Dashboard updates to K8s state on the next reconciliation",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// Assume that the user updated the Policy from Dashboard, which creates a drift between k8s state
				// and Tyk. In order to establish consistency, Operator must update Security Policy based on the
				// k8s state on next reconciliation - so that k8s remains as a source of truth.
				eval := is.New(t)

				copySpec := policyCR.Spec.DeepCopy()
				copySpec.Name = "Updating Existing Policy"

				err := updatePolicyOnTyk(reqCtx, copySpec)
				eval.NoErr(err)

				err = wait.For(func() (done bool, err error) {
					// Ensure that policy is updated accordingly on Tyk Side.
					newCopySpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
					if err != nil {
						return false, err
					}

					eval.True(newCopySpec != nil)
					eval.Equal(newCopySpec.Name, copySpec.Name)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				// Ensure that reconciliation brings updated Security Policy back to the k8s state. In the
				// reconciliation, the operator must realize that the SecurityPolicy CR does not exist on Tyk
				// and it must create a SecurityPolicy based on the spec stored in k8s.
				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						newCopySpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)

						// Ensure that the latest Policy of Dashboard is created according to k8s state during
						// reconciliation.
						eval.True(
							hasSameValues(polRec.Env.Mode, &policyOnK8s.Spec, newCopySpec, policyOnK8s.Status.PolID),
						)
						return true
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, securityPolicyMigrationFeatures)
}

func TestSecurityPolicy(t *testing.T) {
	var (
		reqCtx   context.Context
		policyCR v1alpha1.SecurityPolicy
		apiDefCR *v1alpha1.ApiDefinition
		tykEnv   environmet.Env
	)

	securityPolicyFeatures := features.New("Create Security Policy from scratch").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			verifyPolicyApiVersion(t, &tykEnv)

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Access ApiDefinition CR",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				var err error
				apiDefCR, err = createTestAPIDef(ctx, c, testNs, nil)
				eval.NoErr(err)

				policyCR = v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "sample-policy", Namespace: testNs},
					Spec: v1alpha1.SecurityPolicySpec{
						Name:   envconf.RandomName("sample-policy", 32),
						Active: true,
						State:  "draft",
						AccessRightsArray: []*v1alpha1.AccessDefinition{{
							Name:      apiDefCR.Name,
							Namespace: apiDefCR.Namespace,
						}},
					},
				}

				// Ensure API is created before creating a policy
				err = wait.For(conditions.New(c.Client().Resources()).ResourceMatch(apiDefCR, func(object k8s.Object) bool {
					return apiDefCR.Status.ApiID != ""
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				// Create the SecurityPolicy on k8s after creating the Policy.
				err = c.Client().Resources(testNs).Create(ctx, &policyCR)
				eval.NoErr(err)

				// Ensure that policy is created on Tyk
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						// Ensure that policy is created on Tyk
						policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						if err != nil {
							t.Logf("Failed to find Policy '%v' on Tyk, err: %v", policyOnK8s.Status.PolID, err)
							return false
						}

						eval.True(policyOnK8s.Status.PolID == policyCR.Spec.MID)
						eval.Equal(policyOnK8s.Spec.Name, policyOnTyk.Name)
						eval.Equal(len(policyOnK8s.Spec.AccessRightsArray), 1)

						if tykEnv.Mode == "pro" {
							eval.Equal(len(policyOnK8s.Spec.AccessRightsArray), len(policyOnTyk.AccessRightsArray))
							eval.Equal(
								policyOnK8s.Spec.AccessRightsArray[0].APIID,
								policyOnTyk.AccessRightsArray[0].APIID,
							)
							eval.Equal(policyOnK8s.Status.PolID, policyOnTyk.MID)
						} else {
							ad, exists := policyOnTyk.AccessRights[policyOnK8s.Spec.AccessRightsArray[0].APIID]
							eval.True(exists)
							eval.Equal(policyOnK8s.Spec.AccessRightsArray[0].APIID, ad.APIID)
							eval.Equal(policyOnK8s.Status.PolID, policyOnTyk.ID)
						}

						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("Delete SecurityPolicy and check k8s and Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// Deleting SecurityPolicy CR must delete corresponding SecurityPolicy from Tyk and
				// update ApiDefinition CR status accordingly.
				err := c.Client().Resources(testNs).Delete(ctx, &policyCR)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceDeleted(&policyCR),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// Ensure that the policy is deleted successfully from Tyk.
				err = wait.For(func() (done bool, err error) {
					_, err = klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
					if tykClient.IsNotFound(err) {
						return true, nil
					}

					return false, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(apiDefCR, func(object k8s.Object) bool {
						apiDefOnK8s, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						// Ensure that ApiDefinition CR removed policy link after deleting
						// the SecurityPolicy CR.
						eval.Equal(len(apiDefOnK8s.Status.LinkedByPolicies), 0)

						return true
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

func TestSecurityPolicyForGraphQL(t *testing.T) {
	eval := is.New(t)

	const (
		queryName     = "Query"
		accountsField = "accounts"
		allField      = "*"
		fieldName     = "getMovers"
		limit         = int64(2)
	)

	var (
		reqCtx                    context.Context
		tykEnv                    environmet.Env
		apiDefID                  string
		minGraphQLPolicyGwVersion = version.MustParseGeneric("v4.3.0")
	)

	securityPolicyForGraphQL := features.New("GraphQL specific Security Policy configurations").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			v, err := version.ParseGeneric(tykEnv.TykVersion)
			eval.NoErr(err)

			if tykEnv.Mode == "ce" && !v.AtLeast(minGraphQLPolicyGwVersion) {
				t.Skip("GraphQL specific Security Policies API in CE mode requires at least Tyk v4.3.0")
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

				apiDefCR, err := createTestAPIDef(ctx, c, testNs, nil)
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, apiDefCR)
				eval.NoErr(err)

				var createdResource v1alpha1.ApiDefinition
				err = c.Client().Resources(testNs).Get(ctx, apiDefCR.Name, testNs, &createdResource)
				eval.NoErr(err)
				apiDefID = createdResource.Status.ApiID

				policyCR, err := createTestPolicy(ctx, c, testNs, func(policy *v1alpha1.SecurityPolicy) {
					policy.Spec.AccessRightsArray = []*v1alpha1.AccessDefinition{
						{
							Name:      apiDefCR.Name,
							Namespace: apiDefCR.Namespace,
							AllowedTypes: []v1alpha1.GraphQLType{
								{Name: queryName, Fields: []string{accountsField}},
							},
							RestrictedTypes: []v1alpha1.GraphQLType{
								{Name: queryName, Fields: []string{allField}},
							},
							DisableIntrospection: true,
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

				validPolSpec := func(mode v1alpha1.OperatorContextMode, s *v1alpha1.SecurityPolicySpec, k8s bool) bool {
					if k8s {
						eval.Equal(len(s.AccessRightsArray), 1)
						eval.Equal(s.AccessRightsArray[0].Name, testApiDef)
						eval.Equal(s.AccessRightsArray[0].Namespace, testNs)

						eval.Equal(len(s.AccessRightsArray[0].AllowedTypes), 1)
						eval.Equal(s.AccessRightsArray[0].AllowedTypes[0].Name, queryName)
						eval.Equal(len(s.AccessRightsArray[0].AllowedTypes[0].Fields), 1)
						eval.Equal(s.AccessRightsArray[0].AllowedTypes[0].Fields[0], accountsField)

						eval.Equal(s.AccessRightsArray[0].DisableIntrospection, true)
					} else {
						if mode == "ce" {
							ad, exists := s.AccessRights[apiDefID]
							eval.True(exists)

							// allowed_types and disable_introspection are only available via GW API v4.3
							eval.Equal(len(ad.AllowedTypes), 1)
							eval.Equal(ad.AllowedTypes[0].Name, queryName)
							eval.Equal(len(ad.AllowedTypes[0].Fields), 1)
							eval.Equal(ad.AllowedTypes[0].Fields[0], accountsField)

							eval.Equal(ad.DisableIntrospection, true)
						} else {
							eval.Equal(len(s.AccessRightsArray[0].RestrictedTypes), 1)
							eval.Equal(s.AccessRightsArray[0].RestrictedTypes[0].Name, queryName)
							eval.Equal(len(s.AccessRightsArray[0].RestrictedTypes[0].Fields), 1)
							eval.Equal(s.AccessRightsArray[0].RestrictedTypes[0].Fields[0], allField)

							eval.Equal(len(s.AccessRightsArray[0].FieldAccessRights), 1)
							eval.Equal(s.AccessRightsArray[0].FieldAccessRights[0].TypeName, queryName)
							eval.Equal(s.AccessRightsArray[0].FieldAccessRights[0].FieldName, fieldName)
							eval.Equal(s.AccessRightsArray[0].FieldAccessRights[0].Limits.MaxQueryDepth, limit)
						}
					}

					return true
				}

				eval.True(validPolSpec(tykEnv.Mode, &pol.Spec, true))

				policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, pol.Status.PolID)
				eval.NoErr(err)

				eval.True(validPolSpec(tykEnv.Mode, policyOnTyk, false))

				return ctx
			}).
		Feature()

	testenv.Test(t, securityPolicyForGraphQL)
}
