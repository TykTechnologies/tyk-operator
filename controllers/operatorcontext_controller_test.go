package controllers_test

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/matryer/is"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestOperatorContextCreate(t *testing.T) {
	is := is.New(t)

	key := types.NamespacedName{
		Name:      "test",
		Namespace: "test-ns",
	}
	// Create a test object
	opCtx := v1alpha1.OperatorContext{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}

	cl := controllers.NewFakeClient([]runtime.Object{&opCtx})
	r := controllers.OperatorContextReconciler{
		Client: cl,
		Scheme: scheme.Scheme,
		Log:    log.NullLogger{},
	}

	req := reconcile.Request{}
	req.NamespacedName = key

	_, err := r.Reconcile(context.TODO(), req)
	is.NoErr(err)

	var result v1alpha1.OperatorContext

	err = cl.Get(context.TODO(), key, &result)
	if err != nil {
		t.Error(err)
	}

	is.True(len(result.Finalizers) != 0)
	is.True(result.Finalizers[0] == keys.OperatorContextFinalizerName)
}

func TestOperatorContextDelete(t *testing.T) {
	is := is.New(t)
	dummyName := "dummy"

	tests := map[string]struct {
		Name        string
		OpCtxStatus *v1alpha1.OperatorContextStatus
		ReturnError bool
	}{
		"without links": {
			Name:        "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{},
			ReturnError: false,
		},
		"with apiDef link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedApiDefinitions: []model.Target{{Name: dummyName}},
			},
			ReturnError: true,
		},
		"with security policy link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedSecurityPolicies: []model.Target{{Name: dummyName}},
			},
			ReturnError: true,
		},
		"with apiCatalogue link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedPortalAPICatalogues: []model.Target{{Name: dummyName}},
			},
			ReturnError: true,
		},
		"with portal config link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedPortalConfigs: []model.Target{{Name: dummyName}},
			},
			ReturnError: true,
		},
		"with apiDescription link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedApiDescriptions: []model.Target{{Name: dummyName}},
			},
			ReturnError: true,
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			key := types.NamespacedName{
				Name: tc.Name,
			}
			// Create a test object
			opCtx := v1alpha1.OperatorContext{
				ObjectMeta: v1.ObjectMeta{
					Name:              tc.Name,
					DeletionTimestamp: &v1.Time{Time: v1.Now().Time},
				},
				Status: *tc.OpCtxStatus,
			}

			cl := controllers.NewFakeClient([]runtime.Object{&opCtx})

			r := controllers.OperatorContextReconciler{
				Client: cl,
				Scheme: scheme.Scheme,
				Log:    log.NullLogger{},
			}

			req := reconcile.Request{}
			req.NamespacedName = key

			_, err := r.Reconcile(context.TODO(), req)
			is.Equal(err != nil, tc.ReturnError)
		})
	}
}
