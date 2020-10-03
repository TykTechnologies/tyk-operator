package controllers

import (
	"context"
	"fmt"
	"time"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	tykGatewayImageName        = "tykio/tyk-gateway:v3.0.0"
	tykGatewaySecretName       = "tyk-gateway-secret"
	tykGatewayDeploymentName   = "tyk-gateway"
	tykGatewayServiceName      = "tyk-service"
	tykGatewayAdminServiceName = "tyk-admin-service"
)

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=gateways/status,verbs=get;update;patch
func (r *GatewayReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("gateway", req.NamespacedName)

	log.Info("fetching gateway instance")
	gateway := &tykv1.Gateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Tyk Gateway resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Tyk")
		return ctrl.Result{}, err
	}

	var err error
	var result *ctrl.Result

	log.Info("ensuring secrets are set")
	result, err = r.ensureSecret(ctx, log, req, gateway, r.gatewaySecret(gateway))
	if result != nil {
		return *result, err
	}

	switch gateway.Spec.Kind {
	case "Deployment":
		deployment := r.gatewayDeployment(gateway)
		log.Info("ensuring deployment exists")
		result, err = r.ensureDeployment(ctx, log, req, gateway, deployment, gateway.Spec.Size)
		if result != nil {
			return *result, err
		}
	case "DaemonSet":
		// TODO
	}

	log.Info("ensuring admin exists")
	result, err = r.ensureService(ctx, log, req, gateway, r.gatewayAdminService(gateway))
	if result != nil {
		return *result, err
	}

	log.Info("ensuring proxy service exists")
	result, err = r.ensureService(ctx, log, req, gateway, r.gatewayService(gateway))
	if result != nil {
		return *result, err
	}

	log.Info("checking gateway is up")
	if !r.isGatewayUp(ctx, log, gateway) {
		delay := time.Second * time.Duration(5)
		log.Info(fmt.Sprintf("tyk isn't running, waiting for %s", delay))

		return ctrl.Result{
			RequeueAfter: delay,
		}, nil
	}

	log.Info("updating gateway status")
	if err = r.updateGatewayStatus(ctx, log, gateway); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GatewayReconciler) discoverServices(ctx context.Context, log logr.Logger) (*reconcile.Result, error) {
	serviceList := &corev1.ServiceList{}

	if err := r.List(ctx, serviceList, nil); err != nil {
		log.Error(err, "unable to get service list")
		return nil, err
	}

	for _, svc := range serviceList.Items {
		log.Info("svc: ", svc.Name)
	}

	return nil, nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.Gateway{}).
		Complete(r)
}
