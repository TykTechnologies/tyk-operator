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

	"k8s.io/apimachinery/pkg/api/errors"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/dashboard_admin_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	AdminDashboardCient *dashboard_admin_client.Client
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=organizations/status,verbs=get;update;patch

func (r *OrganizationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("organization", req.NamespacedName)

	log.Info("fetching organization resource")
	desired := &tykv1alpha1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Tyk Organization resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get organization")
		return ctrl.Result{}, err
	}

	orgs, err := r.AdminDashboardCient.OrganizationAll()
	if err != nil {
		log.Error(err, "unable to list all orgs")
		return ctrl.Result{}, err
	}

	orgExists := false
	for _, org := range orgs {
		log.Info("org", "id", org.ID, "nane", org.OwnerName)
		// if org found
		orgExists = true
	}

	if orgExists {
		// update logic
	} else {
		// create logic
		_, _ = r.AdminDashboardCient.OrganizationCreate(&desired.Spec)
		// create an operator user for this org
		// grab the API key
		// store the orgID + apiKey in a K8s secret
	}

	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.Organization{}).
		Complete(r)
}
