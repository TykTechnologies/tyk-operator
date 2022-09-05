package controllers

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/matryer/is"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestShortHash(t *testing.T) {
	if expect, got := "c09b5928e", shortHash("foo.com"+"/httpbin22222"); got != expect {
		t.Errorf("expected %v got %v", expect, got)
	}
}

func TestTranslateHost(t *testing.T) {
	reconciler := IngressReconciler{}
	if expect, got := "foo.com", reconciler.translateHost("foo.com"); got != expect {
		t.Errorf("expected %v got %v", expect, got)
	}

	if expect, got := "{?:[^.]+}.foo.com", reconciler.translateHost("*.foo.com"); got != expect {
		t.Errorf("expected %v got %v", expect, got)
	}
}

func TestBuildAPIName(t *testing.T) {
	reconciler := IngressReconciler{}
	is := is.New(t)

	tests := map[string]struct {
		Namespace string
		Name      string
		Hash      string
		Result    string
	}{
		"empty fields": {
			Result: "--",
		},
		"empty namespace": {
			Name:   "test",
			Hash:   "test",
			Result: "-test-test",
		},
		"empty name": {
			Namespace: "test",
			Hash:      "test",
			Result:    "test--test",
		},
		"empty hash": {
			Name:      "test",
			Namespace: "test",
			Result:    "test-test-",
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			result := reconciler.buildAPIName(tc.Namespace, tc.Name, tc.Hash)
			is.Equal(result, tc.Result)
		})
	}
}

func TestCreateAPI(t *testing.T) {
	apiTemplate := v1alpha1.ApiDefinition{}
	ing := v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path: "/test",
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: "test-sv",
											Port: v1.ServiceBackendPort{
												Name:   "external",
												Number: 123,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	reconciler := IngressReconciler{
		Client: NewFakeClient(nil),
		Log:    log.NullLogger{},
		Scheme: scheme.Scheme,
		Env:    environmet.Env{},
	}

	is := is.New(t)

	err := reconciler.createAPI(context.TODO(), reconciler.Log, &apiTemplate, "default", &ing, reconciler.Env)
	is.NoErr(err)

	apiDef := &v1alpha1.ApiDefinition{}

	key := types.NamespacedName{
		Name:      reconciler.buildAPIName(ing.Namespace, ing.Name, shortHash(""+"/test")),
		Namespace: "default",
	}

	err = reconciler.Client.Get(context.TODO(), key, apiDef)
	is.NoErr(err)
}
