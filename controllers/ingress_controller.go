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
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ingressFinalizerName = "finalizers.tyk.io/ingress"

// CertificateReconciler reconciles a CertificateSecret object
type IngressReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespacedName := req.NamespacedName

	log := r.Log.WithValues("Ingress", namespacedName.String())

	desired := &v1beta1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if *desired.Spec.IngressClassName != "ingress.tyk.io" {
		// it's not for us
		return ctrl.Result{}, nil
	}

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
			apiDef := &v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespacedName.Name,
					Namespace: namespacedName.Namespace,
				},
			}
			if err := r.Delete(ctx, apiDef); err != nil {
				log.Error(err, "unable to delete apidefinition resource")
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, ingressFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, ingressFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	for _, tls := range desired.Spec.TLS {
		desiredSecret := &v1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Namespace: namespacedName.Namespace,
			Name:      tls.SecretName,
		}
		if err := r.Get(ctx, secretNamespacedName, desiredSecret); err != nil {
			log.Error(err, "secret doesnt exist yet")
			return ctrl.Result{}, err
		}
		// TODO: store certs in Tyk cert storage
		log.Info("storing secret in certificate storage", "name", secretNamespacedName.String())
	}

	for _, rule := range desired.Spec.Rules {
		hostName := rule.Host

		for _, p := range rule.HTTP.Paths {
			if p.Path == "dummy" {
				// This is a hack in case we only want to create apis to handle ingress resources
				// for the acme challenge
				continue
			}
			apiDefNamespacedName := types.NamespacedName{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			}
			apiDef := &v1alpha1.ApiDefinition{}
			spec := v1alpha1.APIDefinitionSpec{
				Name:   namespacedName.Name,
				Active: true,
				Proxy: v1alpha1.Proxy{
					StripListenPath: true,
					ListenPath:      p.Path,
					TargetURL:       fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", p.Backend.ServiceName, namespacedName.Namespace, p.Backend.ServicePort.IntValue()),
				},
				Protocol:         "http",
				Domain:           hostName,
				UseKeylessAccess: true,
			}

			if err := r.Get(ctx, apiDefNamespacedName, apiDef); err != nil {
				// doesn't exist
				desiredApiDef := &v1alpha1.ApiDefinition{
					ObjectMeta: ctrl.ObjectMeta{
						Name:      namespacedName.Name,
						Namespace: namespacedName.Namespace,
					},
					Spec: spec,
				}
				err := r.Client.Create(ctx, desiredApiDef)
				if err != nil {
					log.Error(err, "unable to create api definition")
					return ctrl.Result{}, err
				}
			} else {
				// update
				apiDef.Spec = spec
				err := r.Client.Update(ctx, apiDef)
				if err != nil {
					log.Error(err, "unable to create api definition")
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		Complete(r)
}
