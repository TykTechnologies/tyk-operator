package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
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

func TestSecurityPolicyCreate(t *testing.T) {
	const opNs = "tyk-operator-system"

	polRec := &controllers.SecurityPolicyReconciler{}

	securityPolicyFeatures := features.New("Security Policy Create").
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

			polRec = &controllers.SecurityPolicyReconciler{
				Client: testCl,
				Log:    log.NullLogger{},
				Scheme: testCl.Scheme(),
				Env:    tykEnv,
			}

			return ctx
		}).
		Assess("", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)

			testNs, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			api, err := createTestAPIDef(ctx, testNs, nil, c)
			eval.NoErr(err)

			policy := v1alpha1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "sample-policy", Namespace: testNs},
				Spec: v1alpha1.SecurityPolicySpec{
					Name:              "existing policy",
					Active:            true,
					State:             "draft",
					AccessRightsArray: []*v1alpha1.AccessDefinition{{Name: api.Name, Namespace: api.Namespace}},
				},
			}

			_, err = util.CreateOrUpdate(ctx, polRec.Client, &policy, func() error {
				return nil
			})
			eval.NoErr(err)

			err = wait.For(func() (done bool, err error) {
				_, err = polRec.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&policy)})
				return err == nil, err
			})

			err = wait.For(
				conditions.New(c.Client().Resources()).ResourceMatch(&policy, func(object k8s.Object) bool {
					pol := object.(*v1alpha1.SecurityPolicy)

					eval.True(pol.Status.PolID == policy.Spec.MID)
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
