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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EndpointsReconciler reconciles a Endpoints object
type EndpointsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=,resources=endpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=,resources=endpoints/finalizers,verbs=update

func (r *EndpointsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	endpoints := &v1.Endpoints{}
	if err := r.Get(ctx, req.NamespacedName, endpoints); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !endpoints.DeletionTimestamp.IsZero() {
		// resource is being deleted
		// so we should delete the associated api definition resource

		util.RemoveFinalizer(endpoints, keys.EndpointsFinalizerName)
		r.Update(ctx, endpoints)

		apiDef := &v1alpha1.ApiDefinition{}
		apiDef.Name = req.Name
		apiDef.Namespace = req.Namespace
		if err := r.Delete(ctx, apiDef); err != nil {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{}, nil
	}

	if endpoints.Subsets == nil {
		return ctrl.Result{Requeue: true}, errors.New("endpoint not ready")
	}
	for _, subset := range endpoints.Subsets {
		if len(subset.NotReadyAddresses) != 0 {
			return ctrl.Result{Requeue: true}, errors.New("endpoint not ready")
		}
	}

	// we know that the endpoints is 100% ready
	discoveryMode, ok := endpoints.GetLabels()["discovery.tyk.io"]
	if !ok {
		return ctrl.Result{Requeue: false}, fmt.Errorf("endpoint discovery not enabled for %s", req.NamespacedName)
	}
	switch discoveryMode {
	case "oas":

		ip := endpoints.Subsets[0].Addresses[0].IP
		_ = ip
		port := endpoints.Subsets[0].Ports[0].Port

		url := fmt.Sprintf("http://%s:%d/spec.json", ip, port)
		url = "http://httpbin.org/spec.json"

		c := http.Client{Timeout: time.Duration(5) * time.Second}
		res, err := c.Get(url)
		if err != nil {
			return ctrl.Result{Requeue: true}, fmt.Errorf("unable to get schema for %s", req.NamespacedName)
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		_ = err
		println(string(body))

		versions := make(map[string]model.VersionInfo, 0)
		versions["Default"] = model.VersionInfo{Name: "Default"}

		discovered := &v1alpha1.ApiDefinition{
			Spec: v1alpha1.APIDefinitionSpec{
				APIDefinitionSpec: model.APIDefinitionSpec{
					Name:             fmt.Sprintf("%s #k8s", req.NamespacedName.String()),
					Protocol:         "http",
					UseKeylessAccess: true,
					// It was just discovered - not so sure it should automagically be active
					Active: false,
					Proxy: model.Proxy{
						TargetURL: fmt.Sprintf("http://%s.%s.svc:%d", req.Name, req.Namespace, port),
					},
					VersionData: model.VersionData{
						NotVersioned: true,
						Versions:     versions,
					},
				},
			},
		}

		discovered.Name = req.Name
		discovered.Namespace = req.Namespace
		discovered.Labels = endpoints.Labels
		discovered.Annotations = endpoints.Annotations

		util.AddFinalizer(endpoints, keys.EndpointsFinalizerName)
		err = r.Update(ctx, endpoints)
		if err != nil {
			println(err.Error())
		}

		if err := r.Create(ctx, discovered); err != nil {
			return ctrl.Result{Requeue: false}, nil
		}

	case "oas2graphql":
		// we need to introspect the oas and create a UDG
	case "graphql":
		// we need to introspect the graphql schema and create an API Defiinition inside Tyk
	case "subgraph":
		// we need to introspect the graphql schema and create a subgraph api definition. Propagate labels to the new apidefinition
	default:
		return ctrl.Result{Requeue: false}, errors.New("endpoint discovery mode unsupported")
	}

	return ctrl.Result{}, nil
}

func (r *EndpointsReconciler) endpointEventFilter() predicate.Predicate {
	discoveryEnabled := func(o runtime.Object) bool {
		switch e := o.(type) {
		case *v1.Endpoints:
			_, ok := e.GetLabels()["discovery.tyk.io"]
			if !ok {
				return false
			}
			return true
		default:
			return false
		}
	}
	_ = 1

	println("foovar")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return discoveryEnabled(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return discoveryEnabled(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return discoveryEnabled(e.Object)
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&v1.Endpoints{}).
		WithEventFilter(r.endpointEventFilter()).
		Complete(r)
}
