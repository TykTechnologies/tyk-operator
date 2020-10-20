package dashboard_client

import (
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
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

/**
Returns all policies from the Dashboard
*/
func (p SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	fullPath := JoinUrl(p.url, endpointPolicies)

	res, err := grequests.Get(fullPath, p.opts)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var response PoliciesResponse
	if err := res.JSON(&response); err != nil {
		return nil, err
	}

	return response.Policies, nil
}

/**
  Attempts to find the Policy by the namespaced name combo.
  When creating an API, we store this unique combination in the
  policy's tags.
*/
func (p SecurityPolicy) Get(id string) (*v1.SecurityPolicySpec, error) {
	// Returns error if there was a mistake getting all the policies
	list, err := p.All()
	if err != nil {
		return nil, err
	}

	// Iterate through policies to find the policy that stores this
	// unique identifier as the "id"
	for _, pol := range list {
		if id == pol.ID {
			return &pol, nil
		}
	}

	return nil, universal_client.PolicyNotFoundError
}

/*
	Creates a policy.  Creates it with a custom "id" field.
	Valid payload needs to include that.

	1.  If this is an existing policy, will look it up via the "id" field OR
	2.  create a policy and preserves the "id" field.
*/
func (p SecurityPolicy) Create(def *v1.SecurityPolicySpec) (string, error) {
	// Create
	opts := p.opts
	opts.JSON = def
	fullPath := JoinUrl(p.url, endpointPolicies)

	res, err := grequests.Post(fullPath, opts)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
		return "", err
	}

	if resMsg.Status != "OK" {
		return "", fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return resMsg.Message, nil
}

/**
Updates a Policy.  The Dashboard requires that the "MID"
is included in both the Payload as well as the endpoint,
so be sure to pass a valid Policy that includes a "MID" (looked up) and "ID" (the custom one used)
*/
func (p SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {

	// Update
	opts := p.opts
	opts.JSON = def

	fullPath := JoinUrl(p.url, endpointPolicies, def.MID)
	res, err := grequests.Put(fullPath, opts)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
		return err
	}

	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

/**
Tries to delete a Policy by first attempting to do a lookup on it
with the Operator friendly "id" field.
If found, will be used to lookup the mongo id "_id", which is used to do the
delete.

If policy does not exist, move on, nothing to delete.
*/
func (p SecurityPolicy) Delete(policyId string) error {
	pol, err := p.Get(policyId)
	if err == universal_client.PolicyNotFoundError {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Unable to delete policy.")
	}

	delPath := JoinUrl(p.url, endpointPolicies, pol.MID)

	res, err := grequests.Delete(delPath, p.opts)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("delete policy API Returned error: %s", res.String())
}
