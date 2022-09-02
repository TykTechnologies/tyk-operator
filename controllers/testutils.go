package controllers

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewFakeClient(opCtx runtime.Object) client.Client {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.GroupVersion, opCtx)

	objs := []runtime.Object{
		opCtx,
	}

	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
