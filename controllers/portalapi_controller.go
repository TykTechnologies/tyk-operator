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
	"fmt"
	"time"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PortalAPIReconciler reconciles a PortalAPI object
type PortalAPIReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=portalapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=portalapis/status,verbs=get;update;patch

func (r *PortalAPIReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespacedName := req.NamespacedName

	log := r.Log.WithValues("PortalApi", namespacedName.String())
	log.Info("Reconciling PortalApi")
	desired := &tykv1alpha1.PortalAPI{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	var queue bool
	var queueA time.Duration
	_, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			e, err := r.delete(ctx, desired)
			queueA = e
			return err
		}
		// if a PolicyID exists - use that, otherwise, overwrite it with that from the SecurityPolicy object
		if desired.Spec.PolicyID == "" {
			desired.Spec.PolicyID = encodeNS(fmt.Sprintf("%s/%s", desired.Spec.SecurityPolicy.Namespace, desired.Spec.SecurityPolicy.Name))
		}
		util.AddFinalizer(desired, keys.PortalAPIFinalizerName)

		desired.Spec.Version = "v2"

		// try create
		err := r.UniversalClient.PortalCatalogue().Create(&desired.Spec)
		if err != nil {
			// create failed - try to update instead
			return r.UniversalClient.PortalCatalogue().Update(&desired.Spec)
		}
		return nil
	})
	if err == nil {
		log.Info("Completed reconciling ApiDefinition instance")
	}
	return ctrl.Result{Requeue: queue, RequeueAfter: queueA}, err
}

func (r *PortalAPIReconciler) delete(ctx context.Context, desired *tykv1alpha1.PortalAPI) (time.Duration, error) {
	r.Log.Info("resource being deleted")
	if util.ContainsFinalizer(desired, keys.PortalAPIFinalizerName) {
		r.Log.Info("deleting")
		err := r.UniversalClient.PortalCatalogue().Delete(desired.Spec.PolicyID)
		if err != nil {
			r.Log.Error(err, "unable to delete", "policy_id", desired.Spec.PolicyID)
			return 0, err
		}

		r.Log.Info("removing finalizer")
		util.RemoveFinalizer(desired, keys.PortalAPIFinalizerName)
	}
	return 0, nil
}

func (r *PortalAPIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.PortalAPI{}).
		Complete(r)
}
