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

	v1 "k8s.io/api/core/v1"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/api/networking/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ingressClassName = "tyk"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=ingresses/status,verbs=get;update;patch

func (r *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("ingress", req.NamespacedName)

	namespacedName := req.NamespacedName

	desired := v1beta1.Ingress{}
	err := r.Get(ctx, namespacedName, &desired)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if desired.Spec.IngressClassName == nil || *desired.Spec.IngressClassName != ingressClassName {
		// it's not for us
		return ctrl.Result{}, nil
	}

	tlsSecretName := func(host string) *string {
		for _, tlsEntry := range desired.Spec.TLS {
			for _, h := range tlsEntry.Hosts {
				if h == host {
					return &tlsEntry.SecretName
				}
			}
		}
		return nil
	}

	for _, rule := range desired.Spec.Rules {
		host := rule.Host
		tlsSecret := tlsSecretName(host)
		for _, path := range rule.HTTP.Paths {
			// we have an ingress - create the APIDefintion CRD
			apiDefDesired := &v1alpha1.ApiDefinition{}

			if tlsSecret != nil {
				apiDefDesired.Spec.Protocol = "https"
			} else {
				apiDefDesired.Spec.Protocol = "http"
			}

			apiDefDesired.Spec.UseKeylessAccess = true
			apiDefDesired.Spec.Active = true

			svcNamespacedName := types.NamespacedName{
				Namespace: namespacedName.Namespace,
				Name:      path.Backend.ServiceName,
			}
			svc := &v1.Service{}
			// TODO: handle error
			_ = r.Get(ctx, svcNamespacedName, svc)
			svc.Spec.
				apiDefDesired.Spec.Proxy.TargetURL = fmt.Sprintf("%s.%s.svc:%s", path.Backend.ServiceName, namespacedName.Namespace, path.Backend.ServicePort.String())
			apiDefDesired.Spec.Proxy.ListenPath = path.Path
			apiDefDesired.Spec.Proxy.StripListenPath = true
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		Complete(r)
}
