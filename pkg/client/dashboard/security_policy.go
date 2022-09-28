package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

type SecurityPolicy struct{}

// All Returns all policies from the Dashboard
func (p SecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
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

// Get  find the Policy by id
func (p SecurityPolicy) Get(ctx context.Context, id string) (*v1.SecurityPolicySpec, error) {
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

// Create  creates a new policy using the def object
func (p SecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
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
func (p SecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
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
func (p SecurityPolicy) Delete(ctx context.Context, id string) error {
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
