package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

type SecurityPolicy struct {
	*Client
}

// All Returns all policies from the Dashboard
func (p SecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
	res, err := p.Client.Get(p.Env.JoinURL(endpointPolicies), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response PoliciesResponse
	if err := universal.JSON(res, &response); err != nil {
		return nil, err
	}
	return response.Policies, nil
}

// Get  find the Policy by id
func (p SecurityPolicy) Get(ctx context.Context, id string) (*v1.SecurityPolicySpec, error) {
	res, err := p.Client.Get(p.Env.JoinURL(endpointPolicies, id), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var o v1.SecurityPolicySpec
	if err := universal.JSON(res, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// Create  creates a new policy using the def object
func (p SecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
	res, err := p.Client.PostJSON(p.Env.JoinURL(endpointPolicies), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal.Error(res)
	}
	var msg ResponseMsg
	if err := universal.JSON(res, &msg); err != nil {
		return err
	}
	switch strings.ToLower(msg.Status) {
	case "ok":
		def.MID = msg.Message
		return nil
	default:
		return universal.Error(res)
	}
}

// Update updates a resource object def
func (p SecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
	res, err := p.Client.PutJSON(p.Env.JoinURL(endpointPolicies, def.MID), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal.Error(res)
	}
	return universal.JSON(res, def)
}

// Delete deletes the resource by ID
func (p SecurityPolicy) Delete(ctx context.Context, id string) error {
	res, err := p.Client.Delete(p.Env.JoinURL(endpointPolicies, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal.Error(res)
	}
	return nil
}
