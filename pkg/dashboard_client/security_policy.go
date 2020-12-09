package dashboard_client

import (
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type SecurityPolicy struct {
	*Client
}

var (
	SecPolPrefix = "SecurityPolicy"
)

/**
The unique identifier of this policy in the Dashboard
we prefix this with "SecurityPolicy-"
This is stored in the Policy's "tags"
*/
func GetPolicyK8SName(nameSpacedName string) string {
	return SecPolPrefix + "-" + nameSpacedName
}

// All Returns all policies from the Dashboard
func (p SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	res, err := p.Client.Get(p.Env.JoinURL(endpointPolicies), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response PoliciesResponse
	if err := universal_client.JSON(res, &response); err != nil {
		return nil, err
	}
	return response.Policies, nil
}

// Get  find the Policy by id
func (p SecurityPolicy) Get(id string) (*v1.SecurityPolicySpec, error) {
	res, err := p.Client.Get(p.Env.JoinURL(endpointPolicies, id), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var o v1.SecurityPolicySpec
	if err := universal_client.JSON(res, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

/*
	Creates a policy.  Creates it with a custom "id" field.
	Valid payload needs to include that.

	1.  If this is an existing policy, will look it up via the "id" field OR
	2.  create a policy and preserves the "id" field.
*/
func (p SecurityPolicy) Create(def *v1.SecurityPolicySpec) error {
	res, err := p.Client.PostJSON(p.Env.JoinURL(endpointPolicies), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}
	var msg ResponseMsg
	if err := universal_client.JSON(res, &msg); err != nil {
		return err
	}
	switch strings.ToLower(msg.Status) {
	case "ok":
		def.MID = msg.Message
		return nil
	default:
		return universal_client.Error(res)
	}
}

/**
Updates a Policy.  The Dashboard requires that the "MID"
is included in both the Payload as well as the endpoint,
so be sure to pass a valid Policy that includes a "MID" (looked up) and "ID" (the custom one used)
*/
func (p SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	res, err := p.Client.PutJSON(p.Env.JoinURL(endpointPolicies, def.MID), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}
	return universal_client.JSON(res, def)
}

/**
Tries to delete a Policy by first attempting to do a lookup on it
with the Operator friendly "id" field.
If found, will be used to lookup the mongo id "_id", which is used to do the
delete.

If policy does not exist, move on, nothing to delete.
*/
func (p SecurityPolicy) Delete(id string) error {
	res, err := p.Client.Delete(p.Env.JoinURL(endpointPolicies, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}
	return nil
}
