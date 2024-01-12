package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

// SecurityPolicy provides api for accessing policies on the tyk gateway
type SecurityPolicy struct{}

func (a SecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
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

func (a SecurityPolicy) Get(ctx context.Context, id string) (*v1.SecurityPolicySpec, error) {
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

func (a SecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
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
		if def.MID == nil {
			def.MID = new(string)
		}

		return nil
	default:
		return client.Error(res)
	}
}

func (a SecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
	res, err := client.PutJSON(ctx, client.Join(endpointPolicies, *def.ID), def)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	return nil
}

func (a SecurityPolicy) Delete(ctx context.Context, id string) error {
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
