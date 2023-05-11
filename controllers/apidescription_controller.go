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
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
)

// APIDescriptionReconciler reconciles a APIDescription object
type APIDescriptionReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Universal universal.Client
	Env       environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *APIDescriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("APICatalogue", req.NamespacedName.String())

	log.Info("Reconciling APIDescription instance")

	defer func() {
		if err == nil {
			log.Info("Successfully reconciled APIDescription")
		}
	}()

	desired := &tykv1alpha1.APIDescription{}

	if err = r.Get(ctx, req.NamespacedName, desired); err != nil {
		err = client.IgnoreNotFound(err) // Ignore not-found errors
		return
	}
	// set context for all api calls inside this reconciliation loop
	env, ctx, err := HttpContext(ctx, r.Client, r.Env, desired, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			return r.delete(ctx, desired, env, log)
		}

		util.AddFinalizer(desired, keys.PortalAPIDescriptionFinalizerName)

		return r.sync(ctx, desired, env, log)
	})

	return
}

func (r *APIDescriptionReconciler) delete(
	ctx context.Context,
	desired *v1alpha1.APIDescription,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Deleting APIDescription resource")
	// we find all api catalogues referencing this and update it to reflect the
	// change
	log.Info("Fetching APICatalogueList ...")

	var ls v1alpha1.PortalAPICatalogueList

	err := r.List(ctx, &ls, &client.ListOptions{
		Namespace: desired.Namespace,
	})
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	log.Info("Fetching APICatalogueList ...Ok", "count", len(ls.Items))

	namespace := desired.Namespace
	target := model.Target{
		Name:      desired.Name,
		Namespace: &namespace,
	}

	for _, catalogue := range ls.Items {
		for _, desc := range catalogue.Spec.APIDescriptionList {
			if desc.APIDescriptionRef != nil && target.Equal(*desc.APIDescriptionRef) {
				cat_ns := catalogue.Namespace

				return fmt.Errorf("Unable to delete api description due to partal catalogue dependency %q",
					model.Target{Name: catalogue.Name, Namespace: &cat_ns}.String(),
				)
			}
		}
	}

	util.RemoveFinalizer(desired, keys.PortalAPIDescriptionFinalizerName)

	return nil
}

func (r *APIDescriptionReconciler) sync(
	ctx context.Context,
	desired *v1alpha1.APIDescription,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Syncing changes to catalogues resource")
	// we find all api catalogues referencing this and update it to reflect the
	// change
	log.Info("Fetching APICatalogueList ...")

	var ls v1alpha1.PortalAPICatalogueList

	err := r.List(ctx, &ls, &client.ListOptions{
		Namespace: desired.Namespace,
	})
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	log.Info("Fetching APICatalogueList ...Ok", "count", len(ls.Items))

	namespace := desired.Namespace
	target := model.Target{
		Name:      desired.Name,
		Namespace: &namespace,
	}

	for _, catalogue := range ls.Items {
		if catalogue.Status.ID == "" {
			// Skip all unpublished catalogues
			continue
		}

		for _, desc := range catalogue.Spec.APIDescriptionList {
			if desc.APIDescriptionRef != nil &&
				target.Equal(*desc.APIDescriptionRef) {
				if catalogue.Labels == nil {
					catalogue.Labels = map[string]string{}
				}
				// we need to trigger an update to the catalogue. We are incrementing
				// updates label for this
				v, _ := strconv.Atoi(catalogue.Labels["updates"])
				catalogue.Labels["updates"] = strconv.Itoa(v + 1)
				namespace := catalogue.Namespace
				target := model.Target{Name: catalogue.Name, Namespace: &namespace}

				log.Info("Updating catalogue", "resource", target.String())

				if err := r.Update(ctx, &catalogue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIDescriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.APIDescription{}).
		Complete(r)
}
