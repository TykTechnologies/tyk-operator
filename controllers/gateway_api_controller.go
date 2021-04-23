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
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha1"
)

type GatewayApiReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

func (r *GatewayApiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	route := v1alpha1.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, &route); err != nil {
		return ctrl.Result{}, err
	}

	apiDefs := []v1.ApiDefinition{}
	for _, hostname := range route.Spec.Hostnames {

		apiDef := v1.ApiDefinition{}
		customDomain, err := r.hostToCustomDomain(hostname)
		if err != nil {
			// TODO: do something
		}
		apiDef.Spec.Domain = customDomain

		apiDefs = append(apiDefs, apiDef)
	}

	return ctrl.Result{}, nil
}

func (r *GatewayApiReconciler) hostToCustomDomain(hostname v1alpha1.Hostname) (string, error) {
	if hostname == "*" {
		return "", nil
	}
	built := strings.Replace(string(hostname), "*", "{?:[^.]+}", 1)

	return built, nil
}

// SetupWithManager initializes ingress controller manager
func (r *GatewayApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		//Owns(&v1alpha1.ApiDefinition{}).
		Complete(r)
}
