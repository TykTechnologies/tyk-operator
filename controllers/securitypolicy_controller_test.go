package controllers

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestUpdatePolicyStatus(t *testing.T) {
	is := is.New(t)

	api1 := &tykv1.ApiDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "api1",
		},
		Spec: tykv1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:            "api1",
				Active:          true,
				UseStandardAuth: true,
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
				Active:          true,
				UseStandardAuth: true,
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
					Name: "policy with access rights",
					AccessRightsArray: []*tykv1.AccessDefinition{
						{
							Name: "api1",
						},
						{
							Name: "api2",
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

			r.updatePolicyStatus(context.TODO(), test.Policy)
			t.Log(test.Policy.Status.LinkedAPIs)

			is.Equal(test.Policy.Status.LinkedAPIs, test.LinkedAPIs)
		})
	}
}

func TestUpdateStatusOfLinkedAPIs(t *testing.T) {
	// create policy - validate status is updated
	// update policy
	// add api to access rights - validate status is updated
	// remove api from access rights - validate status is updated
	// make access right empty

	is := is.New(t)
	api := &tykv1.ApiDefinition{}
	api1 := &tykv1.ApiDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "api1",
		},
		Spec: tykv1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:            "api1",
				Active:          true,
				UseStandardAuth: true,
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
				Active:          true,
				UseStandardAuth: true,
			},
		},
	}

	policy := &tykv1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "policy1",
		},
		Spec: tykv1.SecurityPolicySpec{
			Name: "policy1",
			AccessRightsArray: []*tykv1.AccessDefinition{
				{
					Name: "api1",
				},
			},
		},
	}

	c, err := NewFakeClient([]runtime.Object{api1, api2, policy})
	is.NoErr(err)

	r := SecurityPolicyReconciler{
		Client: c,
		Log:    log.NullLogger{},
	}

	err = r.updateStatusOfLinkedAPIs(context.Background(), policy, false)
	is.NoErr(err)

	c.Get(context.Background(), types.NamespacedName{Name: api1.Name}, api)

	is.True(len(api.Status.LinkedByPolicies) == 1)
	is.True(api.Status.LinkedByPolicies[0].Name == policy.Name)

	// Add api2 to access rights
	policy.Spec.AccessRightsArray = append(policy.Spec.AccessRightsArray, &tykv1.AccessDefinition{Name: "api2"})

	err = r.updateStatusOfLinkedAPIs(context.Background(), policy, false)
	is.NoErr(err)

	c.Get(context.Background(), types.NamespacedName{Name: api2.Name}, api)

	is.True(len(api.Status.LinkedByPolicies) == 1)
	is.True(api.Status.LinkedByPolicies[0].Name == policy.Name)

	// Remove api1 from access rights
	policy.Spec.AccessRightsArray = []*tykv1.AccessDefinition{{Name: "api2"}}
	policy.Status.LinkedAPIs = []model.Target{{Name: "api1"}, {Name: "api2"}}

	err = r.updateStatusOfLinkedAPIs(context.Background(), policy, false)
	is.NoErr(err)

	c.Get(context.Background(), types.NamespacedName{Name: api1.Name}, api)
	t.Log(policy.Status.LinkedAPIs)
	t.Log(api.Status.LinkedByPolicies)
	is.True(len(api.Status.LinkedByPolicies) == 0)

}
