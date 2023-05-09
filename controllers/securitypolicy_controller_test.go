package controllers

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestUpdatePolicyStatus(t *testing.T) {
	is := is.New(t)

	active := true
	useStandardAuth := true

	api1 := &tykv1.ApiDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "api1",
		},
		Spec: tykv1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:            "api1",
				Active:          &active,
				UseStandardAuth: &useStandardAuth,
			},
		},
	}

	api2 := &tykv1.ApiDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "api2",
		},
		Spec: tykv1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:            "api2",
				Active:          &active,
				UseStandardAuth: &useStandardAuth,
			},
		},
	}

	tests := map[string]struct {
		Policy     *tykv1.SecurityPolicy
		LinkedAPIs []model.Target
	}{
		"emtpy acess rights policy": {
			Policy: &tykv1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "policy without access rights",
				},
			},
		},
		"policy with access rights": {
			Policy: &tykv1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "policy with access rights",
				},
				Spec: tykv1.SecurityPolicySpec{
					SecurityPolicySpec: model.SecurityPolicySpec{
						Name: "policy with access rights",
						AccessRightsArray: []*model.AccessDefinition{
							{
								Name: "api1",
							},
							{
								Name: "api2",
							},
						},
					},
				},
			},
			LinkedAPIs: []model.Target{{Name: "api1"}, {Name: "api2"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := NewFakeClient([]runtime.Object{api1, api2, test.Policy})
			is.NoErr(err)

			r := &SecurityPolicyReconciler{
				Client: c,
				Log:    log.NullLogger{},
			}

			err = r.updatePolicyStatus(context.TODO(), test.Policy)
			is.NoErr(err)

			is.Equal(test.Policy.Status.LinkedAPIs, test.LinkedAPIs)
		})
	}
}
