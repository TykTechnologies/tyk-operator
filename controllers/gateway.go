package controllers

import (
	"context"
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *GatewayReconciler) gatewaySecret(g *tykv1.Gateway) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tykGatewaySecretName,
			Namespace: g.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"secret":      g.Spec.Config.Secret,
			"node_secret": g.Spec.Config.NodeSecret,
		},
	}
	controllerutil.SetControllerReference(g, secret, r.Scheme)
	return secret
}

func (r *GatewayReconciler) gatewayDeployment(g *tykv1.Gateway) *appsv1.Deployment {
	labels := labelsForGateway("tyk")
	replicas := g.Spec.Size

	secret := envVarFromSecret("SECRET", "secret")
	nodeSecret := envVarFromSecret("NODE_SECRET", "node_secret")

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.Name,
			Namespace: g.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf(tykGatewayImageName),
							Name:  tykGatewayDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "proxy",
								},
								{
									ContainerPort: 8081,
									Name:          "admin",
								},
							},
							Env: []corev1.EnvVar{
								secret,
								nodeSecret,
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(g, dep, r.Scheme)
	return dep
}

func (r *GatewayReconciler) gatewayService(v *tykv1.Gateway) *corev1.Service {
	ls := labelsForGateway("tyk")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tykGatewayServiceName,
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Port: v.Spec.Config.ListenPort,
				},
			},
			ClusterIP: "None",
		},
	}

	controllerutil.SetControllerReference(v, s, r.Scheme)
	return s
}

// Returns whether or not the MySQL deployment is running
func (r *GatewayReconciler) isGatewayUp(ctx context.Context, log logr.Logger, g *tykv1.Gateway) bool {
	deployment := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{
		Name:      g.Name,
		Namespace: g.Namespace,
	}, deployment)

	if err != nil {
		log.Error(err, "Deployment tyk-gateway not found")
		return false
	}

	if deployment.Status.ReadyReplicas > 0 {
		return true
	}

	return false
}

func (r *GatewayReconciler) updateGatewayStatus(ctx context.Context, log logr.Logger, g *tykv1.Gateway) error {
	// Update the Gateway status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(g.Namespace),
		client.MatchingLabels(labelsForGateway(g.Name)),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Gateway.Namespace", g.Namespace, "Gateway.Name", g.Name)
		return err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, g.Status.Nodes) {
		g.Status.Nodes = podNames
		err := r.Status().Update(ctx, g)
		if err != nil {
			log.Error(err, "Failed to update gateway status")
			return err
		}
	}

	return nil
}

func envVarFromSecret(gatewayKey string, secretKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: prefixEnvKey(gatewayKey),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: tykGatewaySecretName},
				Key:                  secretKey,
			},
		},
	}
}

func prefixEnvKey(key string) string {
	return "TYK_GW_" + key
}
