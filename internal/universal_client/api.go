package universal_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/pkg/errors"
)

type UniversalApi interface {
	Get(apiID string) (*v1.APIDefinitionSpec, error)
	All() ([]v1.APIDefinitionSpec, error)
	Create(def *v1.APIDefinitionSpec) (apiID string, err error)
	Update(apiID string, def *v1.APIDefinitionSpec) error
	Delete(id string) error
}

func CreateOrUpdateAPI(c UniversalClient, spec *v1.APIDefinitionSpec) (*v1.APIDefinitionSpec, error) {
	var err error

	// should return nil, nil http.BadRequest if api doesn't exist
	api, err := c.Api().Get(spec.APIID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to communicate with Client")
	}

	if api == nil {
		// Create
		insertedId, err := c.Api().Create(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create api")
		}
		api, err = c.Api().Get(insertedId)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get inserted api")
		}

		// if this is an existing CRD and the "api_id" does not match between the two APIs, that means
		// the newly created one was unintentionally deleted.  The next action updates the "api_id" to the old one in order to resync
		if api.APIID != spec.APIID {
			spec.OrgID = api.OrgID
			err = c.Api().Update(api.APIID, spec)
			if err != nil {
				return nil, errors.Wrap(err, "unable to update api")
			}
		}

	} else {
		// Update
		spec.OrgID = api.OrgID
		spec.ID = api.ID
		spec.APIID = api.APIID
		err = c.Api().Update(api.APIID, spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update api")
		}
	}

	api, err = c.Api().Get(spec.APIID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get final api form")
	}

	_ = c.HotReload()

	return api, nil
}
