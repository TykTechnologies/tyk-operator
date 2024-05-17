package controllers

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
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
	eval := is.New(t)

	t.Parallel()

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
		tc := tc

		t.Run(n, func(t *testing.T) {
			t.Parallel()
			result := reconciler.buildAPIName(tc.Namespace, tc.Name, tc.Hash)
			eval.Equal(result, tc.Result)
		})
	}
}

func TestCreateAPI(t *testing.T) {
	eval := is.New(t)
	apiTemplate := v1alpha1.ApiDefinition{}
	ns := "default"
	path := "test"

	t.Parallel()

	ing := v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: ns,
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path: path,
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

	client, err := NewFakeClient(nil)
	eval.NoErr(err)

	reconciler := IngressReconciler{
		Client: client,
		Log:    log.NullLogger{},
		Scheme: scheme.Scheme,
		Env:    environment.Env{},
	}

	err = reconciler.reconcileClassicApiDefinition(
		context.TODO(),
		reconciler.Log,
		&apiTemplate,
		"default",
		&ing,
		&reconciler.Env,
	)
	eval.NoErr(err)

	apiDef := &v1alpha1.ApiDefinition{}

	key := types.NamespacedName{
		Name:      reconciler.buildAPIName(ing.Namespace, ing.Name, shortHash(""+path)),
		Namespace: ns,
	}

	err = reconciler.Client.Get(context.TODO(), key, apiDef)
	eval.NoErr(err)
}
