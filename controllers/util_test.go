package controllers

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewFakeClient(objs []runtime.Object) (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if objs == nil {
		return fake.NewClientBuilder().WithScheme(scheme).Build(), nil
	}

	cliBuilder := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...)

	return cliBuilder.Build(), nil
}
