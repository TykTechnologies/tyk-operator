package dashboard_client

import (
	"fmt"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/pkg/errors"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/levigross/grequests"
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
func (p SecurityPolicy) Get(namespacedName string) (*v1.SecurityPolicySpec, error) {
	// Returns error if there was a mistake getting all the policies
	list, err := p.All()
	if err != nil {
		return nil, err
	}

	// Iterate through policies to find the policy that stores this
	// unique identifier in its' tags
	policyName := GetPolicyK8SName(namespacedName)
	for _, pol := range list {
		for _, tag := range pol.Tags {
			if tag == policyName {
				return &pol, nil
			}
		}
	}

	return nil, universal_client.PolicyNotFoundError
}

/*
	Creates a policy.  Adds the namespaced name to the Policy's tags in order to
	uniquely identify it.  This allows us to do CRUDs without having to store
	the mongo BSON ID after creating.
*/
func (p SecurityPolicy) Create(def *v1.SecurityPolicySpec, namespacedName string) (string, error) {
	// Check if this policy exists and check exists/collisions
	pol, err := p.Get(namespacedName)
	// if policy not found error, great, skip and create!
	if err != nil && err != universal_client.PolicyNotFoundError {
		return "", err
	} else if pol != nil {
		return "", universal_client.PolicyCollisionError
	}

	// Add the unique policy identifier
	def.Tags = append(def.Tags, GetPolicyK8SName(namespacedName))

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
Updates a Policy.  Adds the unique identifier namespaced-Name to the
policy's tags so subsequent CRUD opps are possible.
*/
func (p SecurityPolicy) Update(def *v1.SecurityPolicySpec, namespacedName string) error {
	polToUpdate, err := p.Get(namespacedName)
	if err != nil {
		return err
	}

	if polToUpdate == nil {
		return notFoundError
	}

	// Update
	opts := p.opts
	opts.JSON = def
	def.MID = polToUpdate.MID

	// Add the unique policy identifier
	def.Tags = append(def.Tags, GetPolicyK8SName(namespacedName))

	fullPath := JoinUrl(p.url, endpointPolicies, polToUpdate.MID)
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
Tries to delete a Policy by first attempting to do a lookup on it.
If policy does not exist, move on, nothing to delete
*/
func (p SecurityPolicy) Delete(namespacedName string) error {
	pol, err := p.Get(namespacedName)
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
