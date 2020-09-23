package universal_client

import (
	"time"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
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
		time.Sleep(time.Second)
		created, err := c.Api().Get(insertedId)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get inserted api")
		}
		time.Sleep(time.Second)
		oldAPIID := created.APIID
		created.APIID = spec.APIID

		err = c.Api().Update(oldAPIID, created)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update API ID")
		}
		time.Sleep(time.Second)
	} else {
		// Update
		spec.OrgID = api.OrgID
		spec.ID = api.ID
		err = c.Api().Update(api.APIID, spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update api")
		}
	}

	_ = c.HotReload()

	api, err = c.Api().Get(spec.APIID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get created api")
	}

	return api, nil
}
