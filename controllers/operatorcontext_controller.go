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
	"strconv"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
)

// OperatorContextReconciler reconciles a OperatorContext object
type OperatorContextReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts/status,verbs=get;update;patch

func (r *OperatorContextReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("OperatorContextReconciler", req.NamespacedName.String())
	log.Info("Reconciling OperatorContext")
	desired := &v1alpha1.OperatorContext{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.DeletionTimestamp.IsZero() {
			controllerutil.RemoveFinalizer(desired, keys.OperatorContextFinalizerName)
			return nil
		}
		if desired.Spec.Env == nil {
			desired.Spec.Env = &v1alpha1.Env{}
		}
		controllerutil.AddFinalizer(desired, keys.OperatorContextFinalizerName)
		if desired.Spec.FromSecret != nil {
			secret := &v1.Secret{}
			if err := r.Get(ctx, types.NamespacedName{
				Name:      desired.Spec.FromSecret.Name,
				Namespace: desired.Spec.FromSecret.Namespace,
			}, secret); err != nil {
				return err
			}
			e := desired.Spec.Env
			e.Mode = v1alpha1.OperatorMode(secret.Data[environmet.TykMode])
			e.Auth = string(secret.Data[environmet.TykAuth])
			e.ORG = string(secret.Data[environmet.TykORG])
			e.URL = string(secret.Data[environmet.TykURL])
			e.InsecureSkipVerify, _ = strconv.ParseBool(string(secret.Data[environmet.SkipVerify]))
		}
		return nil
	})
	log.Info("done")
	return ctrl.Result{}, err
}

func (r *OperatorContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.OperatorContext{}).
		Complete(r)
}
