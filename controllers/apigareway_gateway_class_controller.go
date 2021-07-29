package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gwv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

type GatewayClassReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gatewayclasses,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gatewayclasses/status,verbs=get;update;patch

func (r *GatewayClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o gwv1alpha1.GatewayClass
	if err := r.Get(ctx, req.NamespacedName, &o); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	r.Log.Info("========> ", "ns", req.Namespace)

	// skip if not operator class
	if o.Spec.Controller != keys.GatewayAPI {
		return ctrl.Result{}, nil
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &o, func() error {
		if !o.DeletionTimestamp.IsZero() {
			controllerutil.RemoveFinalizer(&o, keys.GatewayAPIFinalizerName)
			return nil
		}
		controllerutil.AddFinalizer(&o, keys.GatewayAPIFinalizerName)
		status := r.validateParameters(
			ctx, o.Namespace,
			o.Spec.ParametersRef)
		status.LastTransitionTime = metav1.NewTime(time.Now())
		return r.setCondition(ctx, &o, status)
	})
	return ctrl.Result{}, err
}

func (r *GatewayClassReconciler) validateParameters(
	ctx context.Context,
	ns string,
	p *gwv1alpha1.ParametersReference,
) metav1.Condition {
	if p == nil {
		return metav1.Condition{
			Type:    string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
			Status:  metav1.ConditionFalse,
			Reason:  string(gwv1alpha1.GatewayClassNotAdmittedInvalidParameters),
			Message: "parametersRef is missing",
		}
	}
	if !r.isNSScoped(p) || !r.hasNS(p) {
		return metav1.Condition{
			Type:    string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
			Status:  metav1.ConditionFalse,
			Reason:  string(gwv1alpha1.GatewayClassNotAdmittedInvalidParameters),
			Message: "parametersRef must be Namespace scoped and namespace must be set",
		}
	}
	if p.Group != tykv1alpha1.GroupVersion.String() || p.Kind != "OperatorContext" {
		return metav1.Condition{
			Type:   string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
			Status: metav1.ConditionFalse,
			Reason: string(gwv1alpha1.GatewayClassNotAdmittedInvalidParameters),
			Message: fmt.Sprintf("group must be %v and kind must be %v",
				tykv1alpha1.GroupVersion.String(), "OperatorContext"),
		}
	}
	t := model.Target{Name: p.Name}

	if p.Namespace != nil && *p.Namespace != "" {
		t.Namespace = *p.Namespace
	}
	var o tykv1alpha1.OperatorContext
	r.Log.Info("Getting OperatorContext", "resource", t.String())
	if err := r.Get(ctx, t.NS(ns), &o); err != nil {
		return metav1.Condition{
			Type:    string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
			Status:  metav1.ConditionTrue,
			Reason:  string(gwv1alpha1.GatewayClassNotAdmittedInvalidParameters),
			Message: err.Error(),
		}
	}
	return metav1.Condition{
		Type:    string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
		Status:  metav1.ConditionTrue,
		Reason:  string(gwv1alpha1.GatewayClassConditionStatusAdmitted),
		Message: "successful configured gateway class",
	}
}

func (r *GatewayClassReconciler) isNSScoped(p *gwv1alpha1.ParametersReference) bool {
	if p.Scope == nil {
		return false
	}

	return *p.Scope == "Namespace"
}

func (r *GatewayClassReconciler) hasNS(p *gwv1alpha1.ParametersReference) bool {
	if p.Namespace == nil {
		return false
	}
	if *p.Namespace == "" {
		return false
	}
	return true
}

func (r *GatewayClassReconciler) setCondition(
	ctx context.Context,
	o *gwv1alpha1.GatewayClass,
	cond metav1.Condition) error {
	if len(o.Status.Conditions) == 0 {
		o.Status.Conditions = []metav1.Condition{cond}
	} else {
		updated := false
		for i := 0; i < len(o.Status.Conditions); i++ {
			if o.Status.Conditions[i].Type == cond.Type {
				a := o.Status.Conditions[i]
				b := cond
				if a.Status == b.Status {
					return nil
				}
				// check if anything changed
				o.Status.Conditions[i] = cond
				updated = true
				break
			}
		}
		if !updated {
			o.Status.Conditions = append(o.Status.Conditions, cond)
		}
	}
	return r.Status().Update(ctx, o)
}
func (r *GatewayClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1alpha1.GatewayClass{}).
		Complete(r)
}
