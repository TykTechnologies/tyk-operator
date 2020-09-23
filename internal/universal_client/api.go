package universal_client

import (
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

	allAPIs, err := c.Api().All()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get all apis")
	}
	apiID := ""
	for _, api := range allAPIs {
		if spec.Proxy.ListenPath == api.Proxy.ListenPath {
			apiID = api.APIID

			// Overwrite the ORGID if found
			spec.OrgID = api.OrgID
			spec.APIID = api.APIID
			break
		}
	}

	if apiID == "" {
		// Create
		apiID, err = c.Api().Create(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create api")
		}
	} else {
		// Update
		err = c.Api().Update(apiID, spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update api")
		}
	}

	_ = c.HotReload()

	api, err := c.Api().Get(apiID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get created api")
	}

	return api, nil
}
