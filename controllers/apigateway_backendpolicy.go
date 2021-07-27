package controllers

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

type BackendPolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=backendpolicies,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=backendpolicies/status,verbs=get;update;patch

func (r *BackendPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *BackendPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1alpha1.BackendPolicy{}).
		Complete(r)
}
