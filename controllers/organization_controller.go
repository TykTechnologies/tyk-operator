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
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/dashboard_admin_client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	AdminDashboardCient *dashboard_admin_client.Client
}

const orgConfigMapName = "tyk-org-ids"

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=organizations/status,verbs=get;update;patch

func (r *OrganizationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("organization", req.NamespacedName)

	// should this configmap logic be elsewhere?
	configMap := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: orgConfigMapName, Namespace: req.Namespace}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "ConfigMap not found… attempting to create...\n")
		} else {
			log.Error(err, "Error fetching configmap, retrying\n")
			return reconcile.Result{}, err
		}

		configMap.Name = orgConfigMapName
		configMap.Namespace = req.Namespace

		err = r.Client.Create(ctx, configMap)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Try again with config map
		return reconcile.Result{Requeue: true}, nil
	}

	orgNamespacedName := req.NamespacedName.String()
	log.Info("fetching organization resource")

	// Lookup organization object
	desired := &tykv1alpha1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// If object is being deleted
	const organizationFinalzerName = "finalizers.tyk.io/organization"
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if containsString(desired.ObjectMeta.Finalizers, organizationFinalzerName) {
			// our finalizer is present, so lets handle our external dependency

			// Delete org
			if err := r.AdminDashboardCient.OrganizationDelete(configMap.Data["orgId."+req.Name]); err != nil {
				log.Error(err, "unable to delete org", "nameSpacedName", orgNamespacedName)
				return reconcile.Result{}, err
			}

			// todo: not working
			// Delete Operator User
			if err := r.AdminDashboardCient.UserAdminDelete(configMap.Data["userId.operator"]); err != nil {
				log.Error(err, "unable to delete operator user", "operator-id", configMap.Data["userId.operator"])
				return reconcile.Result{}, err
			}

			// Delete config map entries
			// TODO: username is unique in each organization
			// ex. UserId.operator can exist in multiple orgs, so need to prefix or suffix
			// with org ID to create universal
			delete(configMap.Data, "userId.operator")
			delete(configMap.Data, "apiKey.operator")
			delete(configMap.Data, "orgId."+req.Name)

			if err = r.Update(ctx, configMap); err != nil {
				log.Error(err, "unable to delete configmap entries")
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, organizationFinalzerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, organizationFinalzerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, organizationFinalzerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update Below
	//_, err := r.AdminDashboardCient.OrganizationAll()
	//if err != nil {
	//	log.Error(err, "unable to list all orgs")
	//	return ctrl.Result{}, err
	//}

	orgExists := false
	//for _, org := range orgs {
	//	log.Info("org", "id", org.ID, "nane", org.OwnerName)
	//	// if org found
	//	orgExists = true
	//}

	if orgExists {
		// update logic
	} else {
		// create logic
		orgId, err := r.AdminDashboardCient.OrganizationCreate(&desired.Spec)
		if err != nil {
			log.Error(err, "unable to create org")
			return ctrl.Result{}, err
		}

		// create an operator user for this org
		// grab the API key

		operatorUser := r.createDashboardOperatorUser(orgId)
		createdUserInfo, err := r.AdminDashboardCient.UserAdminCreate(operatorUser)
		if err != nil {
			log.Error(err, "unable to create operator user")
			return ctrl.Result{}, err
		}

		log.Info("userapikey: " + createdUserInfo.AccessKey)
		log.Info("userid: " + createdUserInfo.ID)

		// store the orgID, userId, and apiKey in a K8s secret
		// userId used for cascade delete if org is deleted
		// org id and api key needed for api calls
		configMap.Data = make(map[string]string)
		configMap.Data["orgId."+desired.Name] = orgId
		configMap.Data["userId.operator"] = createdUserInfo.ID
		configMap.Data["apiKey.operator"] = createdUserInfo.AccessKey

		// Update config map
		_ = r.Client.Update(ctx, configMap)
	}

	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) createDashboardOperatorUser(orgId string) dashboard_admin_client.CreateUserRequest {
	operatorUser := dashboard_admin_client.CreateUserRequest{}
	operatorUser.OrgID = orgId
	operatorUser.Active = true
	operatorUser.EmailAddress = "operator@tyk.com"
	operatorUser.FirstName = "Operator"
	operatorUser.LastName = "User"
	operatorUser.Password = "password"

	return operatorUser
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.Organization{}).
		Complete(r)
}
