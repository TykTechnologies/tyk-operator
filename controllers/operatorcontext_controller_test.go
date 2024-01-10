package controllers_test

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/google/uuid"
	"github.com/matryer/is"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestOperatorContextCreate(t *testing.T) {
	eval := is.New(t)
	t.Parallel()

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

	cl, err := controllers.NewFakeClient([]runtime.Object{&opCtx})
	eval.NoErr(err)

	r := controllers.OperatorContextReconciler{
		Client: cl,
		Scheme: scheme.Scheme,
		Log:    log.NullLogger{},
	}

	req := reconcile.Request{}
	req.NamespacedName = key

	_, err = r.Reconcile(context.TODO(), req)
	eval.NoErr(err)

	var result v1alpha1.OperatorContext

	err = cl.Get(context.TODO(), key, &result)
	if err != nil {
		t.Error(err)
	}

	eval.True(len(result.Finalizers) != 0)
	eval.True(result.Finalizers[0] == keys.OperatorContextFinalizerName)
}

func TestOperatorContextDelete(t *testing.T) {
	eval := is.New(t)
	dummyName := "dummy"
	ctx := context.TODO()

	cl, err := controllers.NewFakeClient(nil)
	eval.NoErr(err)

	r := controllers.OperatorContextReconciler{
		Client: cl,
		Scheme: scheme.Scheme,
		Log:    log.NullLogger{},
	}

	t.Parallel()

	tests := map[string]struct {
		Name        string
		OpCtxStatus *v1alpha1.OperatorContextStatus
		Error       error
	}{
		"without links": {
			Name:        "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{},
			Error:       nil,
		},
		"with apiDef link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedApiDefinitions: []model.Target{{Name: dummyName}},
			},
			Error: controllers.ErrOperatorContextIsStillInUse,
		},
		"with security policy link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedSecurityPolicies: []model.Target{{Name: dummyName}},
			},
			Error: controllers.ErrOperatorContextIsStillInUse,
		},
		"with apiCatalogue link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedPortalAPICatalogues: []model.Target{{Name: dummyName}},
			},
			Error: controllers.ErrOperatorContextIsStillInUse,
		},
		"with portal config link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedPortalConfigs: []model.Target{{Name: dummyName}},
			},
			Error: controllers.ErrOperatorContextIsStillInUse,
		},
		"with apiDescription link": {
			Name: "optCtx",
			OpCtxStatus: &v1alpha1.OperatorContextStatus{
				LinkedApiDescriptions: []model.Target{{Name: dummyName}},
			},
			Error: controllers.ErrOperatorContextIsStillInUse,
		},
	}

	for n, tc := range tests {
		tc := tc

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			key := types.NamespacedName{
				Name: tc.Name + uuid.New().String(),
			}
			// Create a test object
			opCtx := v1alpha1.OperatorContext{
				ObjectMeta: v1.ObjectMeta{
					Name:              key.Name,
					DeletionTimestamp: &v1.Time{Time: v1.Now().Time},
				},
				Status: *tc.OpCtxStatus,
			}

			err := cl.Create(ctx, &opCtx)
			eval.NoErr(err)

			req := reconcile.Request{}
			req.NamespacedName = key

			_, err = r.Reconcile(ctx, req)
			eval.Equal(err, tc.Error)
		})
	}
}
