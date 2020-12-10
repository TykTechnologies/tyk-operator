package universal_client

import (
	"os"
	"strings"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/pkg/errors"
)

type UniversalApi interface {
	Get(apiID string) (*v1.APIDefinitionSpec, error)
	All() ([]v1.APIDefinitionSpec, error)
	Create(def *v1.APIDefinitionSpec) error
	Update(def *v1.APIDefinitionSpec) error
	Delete(id string) error
}

func CreateOrUpdateAPI(c UniversalClient, spec *v1.APIDefinitionSpec) error {
	var err error

	// should return nil, nil if api doesn't exist
	api, err := c.Api().Get(spec.APIID)
	if err != nil {
		return errors.Wrap(err, "Unable to communicate with Client")
	}

	if spec.APIID == "" {
		// Create
		err := c.Api().Create(spec)
		if err != nil {
			return errors.Wrap(err, "unable to create api")
		}

		_ = c.HotReload()
		// todo: replace this once we replace it in main.go
		spec.OrgID = strings.TrimSpace(os.Getenv(environmet.TykORG))

		// Update "api_id" to preserve it
		err = c.Api().Update(spec)
		if err != nil {
			return errors.Wrap(err, "unable to update api")
		}

		_ = c.HotReload()
	} else {
		// Update
		spec.OrgID = api.OrgID
		err = c.Api().Update(spec)
		if err != nil {
			return errors.Wrap(err, "unable to update api")
		}
		_ = c.HotReload()
	}

	return nil
}
