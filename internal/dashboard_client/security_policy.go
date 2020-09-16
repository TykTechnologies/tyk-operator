package dashboard_client

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/levigross/grequests"
)

var (
	policyCollisionError = errors.New("policy id collision detected")
	policyNotFound       = errors.New("policy not found")
)

type SecurityPolicy struct {
	*Client
}

func (a SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	fullPath := JoinUrl(a.url, endpointPolicies)

	res, err := grequests.Get(fullPath, a.opts)
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

func (a SecurityPolicy) Get(polId string) (*v1.SecurityPolicySpec, error) {
	// get all policies and find the matching ID
	list, err := a.All()
	if err != nil {
		return nil, err
	}
	for _, pol := range list {
		if pol.ID == polId {
			return &pol, nil
		}
	}

	return nil, policyNotFound
}

func (a SecurityPolicy) Create(def *v1.SecurityPolicySpec) error {
	// get all policies and check exists/collisions
	list, err := a.All()
	if err != nil {
		return err
	}
	for _, pol := range list {
		if pol.ID == def.ID {
			return policyCollisionError
		}
	}

	// Create
	opts := a.opts
	opts.JSON = def
	fullPath := JoinUrl(a.url, endpointPolicies)

	res, err := grequests.Post(fullPath, opts)
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

func (a SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	polToUpdate, err := a.Get(def.ID)
	if err != nil {
		return err
	}

	if polToUpdate == nil {
		return notFoundError
	}

	// Update
	opts := a.opts
	opts.JSON = def
	def.MID = polToUpdate.MID

	fullPath := JoinUrl(a.url, endpointPolicies, polToUpdate.MID)
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

func (a SecurityPolicy) Delete(id string) error {
	delPath := JoinUrl(a.url, endpointPolicies, id)

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
