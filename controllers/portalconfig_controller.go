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
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
)

// PortalConfigReconciler reconciles a PortalConfig object
type PortalConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PortalConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PortalConfigReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	log := r.Log.WithValues("PortalConfig", req.NamespacedName.String())
	log.Info("Reconciling PortalConfig instance")
	defer func() {
		if err == nil {
			log.Info("Completed reconciling PortalConfig instance")
		}
	}()
	desired := &tykv1alpha1.PortalConfig{}
	if err = r.Get(ctx, req.NamespacedName, desired); err != nil {
		err = client.IgnoreNotFound(err) // Ignore not-found errors
		return
	}
	// set context for all api calls inside this reconciliation loop
	env, ctx := httpContext(ctx, r.Client, r.Env, desired, log)
	_, err = util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			return r.delete(ctx, desired, env, log)
		}
		util.AddFinalizer(desired, keys.PortalConfigurationFinalizerName)
		if desired.Status.ID == "" {
			return r.create(ctx, desired, env, log)
		}
		return r.update(ctx, desired, env, log)
	})
	return
}

func (r *PortalConfigReconciler) create(
	ctx context.Context,
	desired *v1alpha1.PortalConfig,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Creating portal configuration object")
	// Configuration is per organization. Since we can't delete this once created
	// we can assume that this will still be present in the dashboard after kubectl
	// delete command.
	//
	// We check if we have this object in dashboard already and we just update the
	// object to match the resource state
	conf, err := klient.Universal.Portal().Configuration().Get(ctx)
	if err != nil {
		res, err := klient.Universal.Portal().Configuration().Create(ctx, &desired.Spec.PortalModelPortalConfig)
		if err != nil {
			return err
		}
		desired.Status.ID = res.Message
		return r.Status().Update(ctx, desired)
	}
	log.Info("Found existing portal configuration")
	d := desired.Spec.PortalModelPortalConfig
	d.Id = conf.Id
	d.OrgID = conf.OrgID
	_, err = klient.Universal.Portal().Configuration().Update(ctx, &d)
	if err != nil {
		log.Error(err, "Failed updating portal configuration")
		return err
	}
	desired.Status.ID = conf.Id
	return r.Status().Update(ctx, desired)
}

func (r *PortalConfigReconciler) update(
	ctx context.Context,
	desired *v1alpha1.PortalConfig,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Updating portal configuration object")
	d := desired.Spec.PortalModelPortalConfig
	d.Id = desired.Status.ID
	d.OrgID = env.Org
	_, err := klient.Universal.Portal().Configuration().Update(ctx, &d)
	return err
}

func (r *PortalConfigReconciler) delete(
	ctx context.Context,
	desired *v1alpha1.PortalConfig,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Deleting portal configuration resource")
	util.RemoveFinalizer(desired, keys.PortalConfigurationFinalizerName)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PortalConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.PortalConfig{}).
		Complete(r)
}
