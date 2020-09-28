package universal_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/pkg/errors"
)

type UniversalSecurityPolicy interface {
	All() ([]v1.SecurityPolicySpec, error)
	Get(polId string) (*v1.SecurityPolicySpec, error)
	Create(def *v1.SecurityPolicySpec) (string, error)
	Update(def *v1.SecurityPolicySpec) error
	Delete(id string) error
}

var (
	PolicyCollisionError = errors.New("policy id collision detected")
	PolicyNotFoundError  = errors.New("policy not found")
)

func CreateOrUpdatePolicy(c UniversalClient, spec *v1.SecurityPolicySpec) (*v1.SecurityPolicySpec, error) {
	var err error

	pol, err := c.SecurityPolicy().Get(spec.ID)
	if err != nil {
		// should return nil, http.401 if policy doesn't exist
		if err != PolicyNotFoundError {
			return nil, errors.Wrap(err, "Unable to communicate with Client")
		}
	}

	if pol == nil {
		// Create
		_, err := c.SecurityPolicy().Create(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create policy")
		}
	} else {
		// Update
		spec.OrgID = pol.OrgID
		err = c.SecurityPolicy().Update(spec)
		pol, err = c.SecurityPolicy().Get(spec.ID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update api")
		}
	}

	_ = c.HotReload()

	return pol, nil
}
