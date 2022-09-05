package controllers

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewFakeClient(objs []runtime.Object) client.Client {
	scheme := scheme.Scheme
	v1alpha1.AddToScheme(scheme)

	if objs == nil {
		return fake.NewClientBuilder().WithScheme(scheme).Build()
	}

	cliBuilder := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)

	return cliBuilder.Build()
}
