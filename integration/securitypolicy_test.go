package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestSecurityPolicyMigration(t *testing.T) {
	const (
		opNs                  = "tyk-operator-system"
		initialK8sPolicyTag   = "sample-tag"
		initialK8sPolicyRate  = 50
		initialK8sPolicyState = "deny"
	)

	spec := v1alpha1.SecurityPolicySpec{
		Name:   "existing-spec",
		State:  "draft",
		Rate:   34,
		Active: true,
		Tags:   []string{"testing"},
	}

	hasSameValues := func(k8sSpec, tykSpec *v1alpha1.SecurityPolicySpec, k8sStatusID string) bool {
		return k8sSpec.MID == tykSpec.MID &&
			k8sStatusID == tykSpec.MID &&
			len(k8sSpec.Tags) == len(tykSpec.Tags) &&
			len(tykSpec.Tags) == 1 &&
			tykSpec.Tags[0] == initialK8sPolicyTag &&
			tykSpec.Rate == initialK8sPolicyRate &&
			tykSpec.State == initialK8sPolicyState
	}

	var (
		polRec   controllers.SecurityPolicyReconciler
		policyCR v1alpha1.SecurityPolicy
		reqCtx   context.Context
	)

	securityPolicyMigrationFeatures := features.New("Existing Security Policy Migration to K8s").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			if tykEnv.Mode == "ce" {
				t.Skip("SecurityPolicy API is not implemented in CE yet")
			}

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

				err := klient.Universal.Portal().Policy().Create(reqCtx, &spec)
				eval.NoErr(err)

				policyCR = v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "sample-policy", Namespace: testNs},
					Spec: v1alpha1.SecurityPolicySpec{
						ID:     spec.MID,
						Active: true,
						State:  initialK8sPolicyState,
						Rate:   initialK8sPolicyRate,
						Tags:   []string{initialK8sPolicyTag},
					},
				}

				_, err = util.CreateOrUpdate(ctx, polRec.Client, &policyCR, func() error {
					return nil
				})
				eval.NoErr(err)

				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				})
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)
						eval.True(len(policyOnK8s.Status.PolID) > 0)

						policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)

						eval.True(hasSameValues(&policyOnK8s.Spec, policyOnTyk, policyOnK8s.Status.PolID))
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
				err := klient.Universal.Portal().Policy().Delete(reqCtx, policyCR.Status.PolID)
				eval.NoErr(err)

				previousPolicyID := policyCR.Status.PolID

				// Ensure that policy is deleted from Tyk.
				_, err = klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
				eval.True(tykClient.IsNotFound(err))

				// After reconciliation, Operator should detect drift and recreate non-existing policy on Tyk side.
				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				})
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						newSpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)
						eval.True(hasSameValues(&policyOnK8s.Spec, newSpec, policyOnK8s.Status.PolID))

						// Ensure that the Policy is accessible via the previous ID
						newSpec, err = klient.Universal.Portal().Policy().Get(reqCtx, previousPolicyID)
						eval.NoErr(err)
						eval.True(hasSameValues(&policyOnK8s.Spec, newSpec, policyOnK8s.Status.PolID))

						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("Must be restored from Dashboard updates to K8s state on the next reconciliation",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// Assume that the user updated the Policy from Dashboard, which creates a drift between k8s state
				// and Tyk. In order to establish consistency, Operator must update Security Policy based on the
				// k8s state on next reconciliation - so that k8s remains as a source of truth.
				eval := is.New(t)

				err := wait.For(func() (done bool, err error) {
					copySpec := policyCR.Spec.DeepCopy()
					copySpec.Name = "Updating Existing Policy"

					err = klient.Universal.Portal().Policy().Update(reqCtx, copySpec)
					eval.NoErr(err)

					// Ensure that policy is updated accordingly on Tyk Side.
					newCopySpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
					eval.NoErr(err)
					eval.True(newCopySpec != nil)
					eval.Equal(newCopySpec.Name, copySpec.Name)

					return true, nil
				})
				eval.NoErr(err)

				// Ensure that reconciliation brings updated Security Policy back to the k8s state. In the
				// reconciliation, the operator must realize that the SecurityPolicy CR does not exist on Tyk
				// and it must create a SecurityPolicy based on the spec stored in k8s.
				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				})
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						policyOnK8s, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						newCopySpec, err := klient.Universal.Portal().Policy().Get(reqCtx, policyOnK8s.Status.PolID)
						eval.NoErr(err)

						// Ensure that the latest Policy of Dashboard is created according to k8s state during
						// reconciliation.
						eval.True(hasSameValues(&policyOnK8s.Spec, newCopySpec, policyOnK8s.Status.PolID))
						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, securityPolicyMigrationFeatures)
}

func TestSecurityPolicy(t *testing.T) {
	const opNs = "tyk-operator-system"

	var (
		reqCtx   context.Context
		policyCR v1alpha1.SecurityPolicy
		apiDefCR *v1alpha1.ApiDefinition
		polRec   controllers.SecurityPolicyReconciler
	)

	securityPolicyFeatures := features.New("Create Security Policy from scratch").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			if tykEnv.Mode == "ce" {
				t.Skip("SecurityPolicy API is not implemented in CE yet")
			}

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
		Assess("Access ApiDefinition CR",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				var err error
				apiDefCR, err = createTestAPIDef(ctx, testNs, nil, c)
				eval.NoErr(err)

				policyCR = v1alpha1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "sample-policy", Namespace: testNs},
					Spec: v1alpha1.SecurityPolicySpec{
						Name:   envconf.RandomName("existing-name", 32),
						Active: true,
						State:  "draft",
						AccessRightsArray: []*v1alpha1.AccessDefinition{{
							Name:      apiDefCR.Name,
							Namespace: apiDefCR.Namespace,
						}},
					},
				}

				err = c.Client().Resources().Create(ctx, &policyCR)
				eval.NoErr(err)

				err = wait.For(func() (done bool, err error) {
					_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policyCR)})
					return err == nil, err
				},
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&policyCR, func(object k8s.Object) bool {
						pol, ok := object.(*v1alpha1.SecurityPolicy)
						eval.True(ok)

						return pol.Status.PolID != ""
					}), wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(policyCR.Status.PolID == policyCR.Spec.MID)
				eval.Equal(len(policyCR.Spec.AccessRightsArray), 1)

				// Ensure that policy is created on Tyk
				policyOnTyk, err := klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
				eval.NoErr(err)

				eval.Equal(policyCR.Status.PolID, policyOnTyk.MID)
				eval.Equal(policyCR.Spec.Name, policyOnTyk.Name)
				eval.Equal(len(policyCR.Spec.AccessRightsArray), len(policyOnTyk.AccessRightsArray))
				eval.Equal(policyCR.Spec.AccessRightsArray[0].APIID, policyOnTyk.AccessRightsArray[0].APIID)

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

				err = wait.For(conditions.New(c.Client().Resources()).ResourceDeleted(&policyCR))
				eval.NoErr(err)

				// Ensure that the policy is deleted successfully from Tyk.
				err = wait.For(func() (done bool, err error) {
					_, err = klient.Universal.Portal().Policy().Get(reqCtx, policyCR.Status.PolID)
					if tykClient.IsNotFound(err) {
						return true, nil
					}

					return false, err
				})
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
