package controllers

import (
	"context"
	"encoding/base64"
	"fmt"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Helper function to check string exists in a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Helper function to remove string from slice of string
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// addString returns a string slice with s in it. If s is in slice then slice
// will be returned
func addString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

func addTarget(slice []tykv1.Target, s tykv1.Target) (result []tykv1.Target) {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

func removeTarget(slice []tykv1.Target, s tykv1.Target) (result []tykv1.Target) {
	for _, item := range slice {
		if item.Equal(s) {
			continue
		}
		result = append(result, item)
	}
	return
}

func encodeNS(decoded string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(decoded))
}

// labelsForGateway returns the labels for selecting the resources
// belonging to the given gateway CR name.
func labelsForGateway(name string) map[string]string {
	return map[string]string{"app": "gateway", "gateway_cr": name}
}

func annotationsForIngress() map[string]string {
	return map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "tcp",
		"service.beta.kubernetes.io/aws-load-balancer-type":             "nlb",
	}
}

func (r *GatewayReconciler) ensureSecret(ctx context.Context, log logr.Logger, request reconcile.Request, instance *tykv1.Gateway, s *corev1.Secret) (*reconcile.Result, error) {
	found := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Create(ctx, s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		log.Error(err, "Failed to get Secret")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *GatewayReconciler) ensureDeployment(ctx context.Context, log logr.Logger, request reconcile.Request, instance *tykv1.Gateway, dep *appsv1.Deployment, requiredReplicas int32) (*reconcile.Result, error) {
	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &ctrl.Result{}, err
		} else {
			// Deployment was successful
			return &reconcile.Result{Requeue: true}, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}

	log.Info("checking spec matches status")
	log.Info(fmt.Sprintf("spec: requiredReplicas: %d, replicas: %d", requiredReplicas, *found.Spec.Replicas))
	if *found.Spec.Replicas != requiredReplicas {
		found.Spec.Replicas = &requiredReplicas
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return &ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return &ctrl.Result{Requeue: true}, nil
	}

	return nil, nil
}

func (r *GatewayReconciler) ensureService(ctx context.Context, log logr.Logger, request reconcile.Request, instance *tykv1.Gateway, s *corev1.Service) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the service
		log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Create(ctx, s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}
