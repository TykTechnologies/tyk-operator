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
	"strings"

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

// Reconcile perform reconciliation logic for Ingress resource that is managed
// by the operator.
func (r *GatewayApiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//route := v1alpha1.HTTPRoute{}
	//hostnames := route.Spec.Hostnames
	return ctrl.Result{}, nil
}

func (r *GatewayApiReconciler) HostsToCustomDomains(hosts []v1alpha1.Hostname) (string, error) {
	hostnameToStrings := func(hostnames []v1alpha1.Hostname) []string {
		s := []string{}
		for _, name := range hostnames {
			s = append(s, string(name))
		}
		return s
	}
	built := strings.Join(hostnameToStrings(hosts), "|")
	if len(hosts) > 1 {
		built = fmt.Sprintf("{?:(%s)}", built)
	}

	//built = strings.Replace(built, "*", "{?:[^.]+}", -1)

	return built, nil
}

// SetupWithManager initializes ingress controller manager
func (r *GatewayApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		//Owns(&v1alpha1.ApiDefinition{}).
		Complete(r)
}
