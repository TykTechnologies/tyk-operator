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

func CreateOrUpdatePolicy(c UniversalClient, spec *v1.SecurityPolicySpec) (*v1.SecurityPolicySpec, error) {
	var err error

	// should return nil, nil http.BadRequest if api doesn't exist
	pol, err := c.SecurityPolicy().Get(spec.ID)
	if err != nil && err.Error() != errors.New("policy not found").Error() {
		return nil, errors.Wrap(err, "Unable to communicate with Client")
	}

	if pol == nil {
		// Create
		insertedId, err := c.SecurityPolicy().Create(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create policy")
		}
		println("my id: " + insertedId)
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
