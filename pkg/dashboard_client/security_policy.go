package dashboard_client

import (
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
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

// All Returns all policies from the Dashboard
func (p SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	sess := grequests.NewSession(p.opts())
	fullPath := p.env.JoinURL(endpointPolicies)

	res, err := sess.Get(fullPath, nil)
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

// Get  find the Policy by id
func (p SecurityPolicy) Get(id string) (*v1.SecurityPolicySpec, error) {
	sess := grequests.NewSession(p.opts())
	fullPath := p.env.JoinURL(endpointPolicies, id)
	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}
	var o v1.SecurityPolicySpec
	if err := res.JSON(&o); err != nil {
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
	o := p.opts()
	sess := grequests.NewSession(o)
	fullPath := p.env.JoinURL(endpointPolicies)
	res, err := sess.Post(fullPath, &grequests.RequestOptions{JSON: def})
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}
	var msg ResponseMsg
	if err := res.JSON(&msg); err != nil {
		return err
	}
	switch strings.ToLower(msg.Status) {
	case "ok":
		def.MID = msg.Message
		return nil
	default:
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}
}

/**
Updates a Policy.  The Dashboard requires that the "MID"
is included in both the Payload as well as the endpoint,
so be sure to pass a valid Policy that includes a "MID" (looked up) and "ID" (the custom one used)
*/
func (p SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	sess := grequests.NewSession(p.opts())

	fullPath := p.env.JoinURL(endpointPolicies, def.MID)
	res, err := sess.Put(fullPath, &grequests.RequestOptions{JSON: def})
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}
	return res.JSON(def)
}

/**
Tries to delete a Policy by first attempting to do a lookup on it
with the Operator friendly "id" field.
If found, will be used to lookup the mongo id "_id", which is used to do the
delete.

If policy does not exist, move on, nothing to delete.
*/
func (p SecurityPolicy) Delete(id string) error {
	sess := grequests.NewSession(p.opts())

	pol, err := p.Get(id)
	if err == universal_client.PolicyNotFoundError {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Unable to delete policy.")
	}

	delPath := p.env.JoinURL(endpointPolicies, pol.MID)

	res, err := sess.Delete(delPath, p.opts())
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("delete policy API Returned error: %s", res.String())
}
