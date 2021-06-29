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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	uc "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
)

// PortalAPICatalogueReconciler reconciles a PortalAPICatalogue object
type PortalAPICatalogueReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Universal universal.Client
	Env       environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalapicatalogues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalapicatalogues/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=portalapicatalogues/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PortalAPICatalogue object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PortalAPICatalogueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("PortalAPICatalogue", req.NamespacedName.String())

	log.Info("Reconciling PortalAPICatalogue instance")
	defer func() {
		if err == nil {
			log.Info("Successfully reconciled PortalAPICatalogue")
		} else {
			result = ctrl.Result{RequeueAfter: queueAfter}
		}

	}()
	desired := &tykv1alpha1.PortalAPICatalogue{}
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
		if desired.Spec.OrgID == "" {
			desired.Spec.OrgID = env.Org
		}
		util.AddFinalizer(desired, keys.PortalAPICatalogueFinalizerName)
		if desired.Status.ID != "" {
			return r.update(ctx, desired, env, log)
		}
		return r.create(ctx, desired, env, log)
	})
	return
}

func (r *PortalAPICatalogueReconciler) model(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
	log logr.Logger,
) (*model.APICatalogue, error) {
	m := &model.APICatalogue{
		OrgId: desired.Spec.OrgID,
		Email: desired.Spec.Email,
	}
	// We keep track of APIDescriptionList to ID so that we can detect if we have
	// un published on the catalogue
	active := map[string]struct{}{}
	for _, t := range desired.Spec.APIDescriptionList {
		var a v1alpha1.APIDescription
		if err := r.Get(ctx, t.NS(), &a); err != nil {
			return nil, err
		}

		if a.Spec.PolicyRef != nil {
			// update security policy
			log.Info("Updating PolicyID")
			var sec v1alpha1.SecurityPolicy
			if err := r.Get(ctx, a.Spec.PolicyRef.NS(), &sec); err != nil {
				return nil, err
			}
			a.Spec.PolicyID = sec.Status.PolID
		}
		if a.Spec.PolicyID == "" {
			return nil, fmt.Errorf("%s missing policy_id", t)
		}

		if err := r.sync(ctx, desired, env, &t, &a); err != nil {
			return nil, err
		}
		d := a.Spec.APIDescription
		d.Documentation = desired.Status.Documentation[t.String()]
		m.APIS = append(m.APIS, d)
		active[t.String()] = struct{}{}
	}

	// Try to delete all un published catalogues
	for target, id := range desired.Status.Documentation {
		_, ok := active[target]
		if !ok {
			// This target is no longer part of the catalogue, we delete the
			// documentation.
			log.Info("Deleting un published catalogue", "Target", target)
			_, err := r.Universal.Portal().Documentation().Delete(ctx, id)
			if err != nil {
				if !uc.IsNotFound(err) {
					return nil, err
				}
			}
		}
	}
	return m, nil
}

func (r *PortalAPICatalogueReconciler) sync(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
	t *model.Target,
	a *v1alpha1.APIDescription,
) error {
	d := &model.APIDocumentation{
		DocumentationType: a.Spec.APIDocumentation.DocumentationType,
		Documentation:     a.Spec.APIDocumentation.Documentation,
		APIID:             a.Spec.PolicyID,
	}
	if desired.Status.Documentation == nil {
		desired.Status.Documentation = make(map[string]string)
	}
	if a.Spec.Documentation != "" {
		_, err := r.Universal.Portal().Documentation().Delete(ctx, a.Spec.Documentation)
		if err != nil {
			return err
		}
	}
	// need to update the status of the catalogue to track new config
	// upload new documentation
	res, err := r.Universal.Portal().Documentation().Upload(ctx, d)
	if err != nil {
		return err
	}
	a.Spec.Documentation = res.Message
	desired.Status.Documentation[t.String()] = res.Message
	return nil
}

func (r *PortalAPICatalogueReconciler) init(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
) (id string, err error) {
	cat, err := r.Universal.Portal().Catalogue().Get(ctx)
	if err != nil {
		if uc.IsNotFound(err) {
			result, err := r.Universal.Portal().Catalogue().Create(ctx, &model.APICatalogue{
				OrgId: env.Org,
			})
			if err != nil {
				return "", err
			}
			return result.Message, nil
		}
		return "", err
	}
	return cat.Id, nil
}

func (r *PortalAPICatalogueReconciler) create(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
	log logr.Logger,
) error {
	// create an empty catalogue for the org
	catalogueID, err := r.init(ctx, desired, env)
	if err != nil {
		return err
	}
	m, err := r.model(ctx, desired, env, log)
	if err != nil {
		return err
	}
	m.Id = catalogueID
	_, err = r.Universal.Portal().Catalogue().Update(ctx, m)
	if err != nil {
		return err
	}
	desired.Status.ID = catalogueID
	return r.Status().Update(ctx, desired)
}

func (r *PortalAPICatalogueReconciler) update(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
	log logr.Logger,
) error {
	m, err := r.model(ctx, desired, env, log)
	if err != nil {
		return err
	}
	m.Id = desired.Status.ID
	_, err = r.Universal.Portal().Catalogue().Update(ctx, m)
	if err != nil {
		return err
	}
	return r.Status().Update(ctx, desired)
}

func (r *PortalAPICatalogueReconciler) delete(
	ctx context.Context,
	desired *tykv1alpha1.PortalAPICatalogue,
	env environmet.Env,
	log logr.Logger,
) error {
	log.Info("Deleting PortalAPICatalogue")
	// There is no actual DELETE api for catalogue. What we can do is we can update
	// the catalogue with zero APIDescription and remove the finalizer.
	_, err := r.Universal.Portal().Catalogue().Update(ctx, &model.APICatalogue{
		Id:    desired.Status.ID,
		OrgId: env.Org,
	})
	if err != nil {
		return err
	}
	util.RemoveFinalizer(desired, keys.PortalAPICatalogueFinalizerName)
	log.Info("Deleting documentation published by this catalogue")
	for target, id := range desired.Status.Documentation {
		log.Info("Deleting", "Target", target)
		_, err := r.Universal.Portal().Documentation().Delete(ctx, id)
		if err != nil {
			if !uc.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PortalAPICatalogueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.PortalAPICatalogue{}).
		Complete(r)
}
