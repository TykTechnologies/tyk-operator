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
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	webhookDefFinalizerName = "finalizers.tyk.io/webhook"
)

// WebhookReconciler reconciles a Webhook object
type WebhookReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=webhooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=webhooks/status,verbs=get;update;patch

func (r *WebhookReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("webhook", req.NamespacedName)

	r.Log.Info("fetching webhook desired spec")

	// Lookup webhook object
	des := &tykv1alpha1.Webhook{}
	if err := r.Get(ctx, req.NamespacedName, des); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}
	if des.Name == "" {
		des.Name = req.NamespacedName.String()
	}
	r.Recorder.Event(des, "Normal", "Webhook", "Reconciling")

	// If object is being deleted
	if !des.ObjectMeta.DeletionTimestamp.IsZero() {

		// If still need to delete from Tyk
		if util.ContainsFinalizer(des, webhookDefFinalizerName) {

			if err := r.UniversalClient.Webhook().Delete(ctx, req.NamespacedName.String()); err != nil {
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			util.RemoveFinalizer(des, webhookDefFinalizerName)
			if err := r.Update(context.Background(), des); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !util.ContainsFinalizer(des, webhookDefFinalizerName) {
		des.ObjectMeta.Finalizers = append(des.ObjectMeta.Finalizers, webhookDefFinalizerName)
		err := r.Update(ctx, des)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update
	err := universal_client.CreateOrUpdateWebhook(ctx, r.UniversalClient, &des.Spec)
	if err != nil {
		r.Log.Error(err, "CreateOrUpdateWebhook failure")
		r.Recorder.Event(des, "Error", "Webhook", "Create or Update Webhook")
		return ctrl.Result{Requeue: true}, err
	}

	// if webhook_id not there, add it, this is new object.
	if des.Status.WebhookID == "" {
		des.Status.WebhookID = des.Spec.ID
		if err := r.Status().Update(ctx, des); err != nil {
			r.Log.Error(err, "Could not update ID")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *WebhookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.Webhook{}).
		Complete(r)
}
