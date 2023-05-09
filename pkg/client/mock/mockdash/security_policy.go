package mockdash

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type mockSecurityPolicy struct{}

// All Returns all policies from the Dashboard
func (p mockSecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
	res, err := client.Get(ctx, client.Join(fmt.Sprintf("%s?p=-2", endpointPolicies)), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response PoliciesResponse
	if err := client.JSON(res, &response); err != nil {
		return nil, err
	}

	return response.Policies, nil
}

// Get find the Policy by id.
func (p mockSecurityPolicy) Get(ctx context.Context, id string) (*v1.SecurityPolicySpec, error) {
	res, err := client.Get(ctx, client.Join(endpointPolicies, id), nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var o v1.SecurityPolicySpec

	if err := client.JSON(res, &o); err != nil {
		return nil, err
	}

	return &o, nil
}

// Create creates a new policy using the def object
func (p mockSecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
	res, err := client.PostJSON(ctx, client.Join(endpointPolicies), def)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	var msg ResponseMsg
	if err := client.JSON(res, &msg); err != nil {
		return err
	}

	switch strings.ToLower(msg.Status) {
	case "ok":
		def.MID = msg.Message
		return nil
	default:
		return client.Error(res)
	}
}

// Update updates a resource object def
func (p mockSecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
	conf := config.GetConfigOrDie()
	scheme := runtime.NewScheme()

	cl, err := ctrl.New(conf, ctrl.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	labels := map[string]string{"mock_test": "tyk"}
	polList := v1alpha1.SecurityPolicyList{}

	err = cl.List(ctx, &polList, ctrl.MatchingLabels(labels))
	if err != nil {
		fmt.Println("failed: ", err)
		return err
	} else {
		for _, item := range polList.Items {
			annotations := item.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
				annotations["mock_test"] = strconv.Itoa(0)
			} else {
				uc, err := strconv.Atoi(annotations["mock_test"])
				if err != nil {
					fmt.Println("cannot convert: ", err)
					return err
				}

				uc++
				annotations["mock_test"] = strconv.Itoa(uc)
				item.SetAnnotations(annotations)
			}
			err = cl.Update(ctx, &item)
			if err != nil {
				fmt.Println("cannot update: ", err)
				return err
			}
		}
	}

	res, err := client.PutJSON(ctx, client.Join(endpointPolicies, def.MID), def)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	return client.JSON(res, def)
}

// Delete deletes the resource by ID
func (p mockSecurityPolicy) Delete(ctx context.Context, id string) error {
	res, err := client.Delete(ctx, client.Join(endpointPolicies, id), nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	return nil
}
