package mockgw

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/pkg/client/mock/internal/helper"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

// mockSecurityPolicy provides api for accessing policies on the tyk gateway
type mockSecurityPolicy struct{}

func (a mockSecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
	res, err := client.Get(ctx, client.Join(fmt.Sprintf("%s?p=-2", endpointPolicies)), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get all Policies, API Returned error: %d", res.StatusCode)
	}

	var response PoliciesResponse
	if err := client.JSON(res, &response); err != nil {
		return nil, err
	}

	return response.Policies, nil
}

func (a mockSecurityPolicy) Get(ctx context.Context, id string) (*v1.SecurityPolicySpec, error) {
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

func (a mockSecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
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
		def.MID = msg.Key
		return nil
	default:
		return client.Error(res)
	}
}

func (a mockSecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
	err := helper.UpdateSecurityPolicyAnnotations(ctx)
	if err != nil {
		return err
	}

	res, err := client.PutJSON(ctx, client.Join(endpointPolicies, def.MID), def)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	return nil
}

func (a mockSecurityPolicy) Delete(ctx context.Context, id string) error {
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
