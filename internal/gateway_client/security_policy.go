package gateway_client

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/levigross/grequests"
)

var (
	policyCollisionError = errors.New("policy id collision detected")
)

type SecurityPolicy struct {
	*Client
}

// todo: needs testing
func (a SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	fullPath := JoinUrl(a.url, endpointPolicies)

	res, err := grequests.Get(fullPath, a.opts)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var list []v1.SecurityPolicySpec
	if err := res.JSON(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// todo: needs testing
func (a SecurityPolicy) Get(namespacedName string) (*v1.SecurityPolicySpec, error) {
	// todo: check does this namespaced name work?
	fullPath := JoinUrl(a.url, endpointPolicies, namespacedName)
	res, err := grequests.Get(fullPath, a.opts)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var retrievedPol v1.SecurityPolicySpec
	if err := res.JSON(&retrievedPol); err != nil {
		return nil, err
	}

	return &retrievedPol, nil
}

// todo: needs testing
func (a SecurityPolicy) Create(def *v1.SecurityPolicySpec, namespacedName string) (string, error) {
	// Replace this with a GET ONE once that is fixed
	// get all policies
	list, err := a.All()
	if err != nil {
		return "", err
	}
	// check exists / collisions
	for _, pol := range list {
		if pol.ID == def.ID {
			return "", policyCollisionError
		}
	}

	// Create
	opts := a.opts
	opts.JSON = def
	fullPath := JoinUrl(a.url, endpointPolicies)

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

	if resMsg.Status != "ok" {
		return "", fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return resMsg.Key, nil
}

// todo: needs testing
func (a SecurityPolicy) Update(def *v1.SecurityPolicySpec, namespacedName string) error {
	// Replace this with a GET ONE once that is fixed
	list, err := a.All()
	if err != nil {
		return err
	}

	var polToUpdate *v1.SecurityPolicySpec
	for _, pol := range list {
		if pol.ID == def.ID {
			polToUpdate = &pol
			break
		}
	}

	if polToUpdate == nil {
		return notFoundError
	}

	// Update
	opts := a.opts
	opts.JSON = def
	fullPath := JoinUrl(a.url, endpointPolicies, polToUpdate.ID)

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

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

// todo: needs testing
func (a SecurityPolicy) Delete(namespacedName string) error {
	// todo: check does this namespaced name work?
	delPath := JoinUrl(a.url, endpointPolicies, namespacedName)

	res, err := grequests.Delete(delPath, a.opts)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	if res.StatusCode == http.StatusInternalServerError {
		// Tyk returns internal server error if api is already deleted
		return nil
	}

	return fmt.Errorf("API Returned error: %s", res.String())
}
