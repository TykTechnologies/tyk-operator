package organization

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/dashboard_admin_client"
	"github.com/pkg/errors"
)

type Organization struct {
	Client *dashboard_admin_client.Client
}

func (o Organization) CreateOrUpdate(spec v1alpha1.OrganizationSpec) error {
	oldOrg, err := o.Client.OrganizationGet(spec.ID)
	if err != nil {
		if !errors.Is(err, dashboard_admin_client.ErrNotFound) {
			return err
		}

		// create
		newOrgID, err := o.Client.OrganizationCreate(&spec)
		if err != nil {
			return err
		}

		_ = newOrgID
		return nil
	}

	_ = oldOrg

	// org exists - update it with new spec
	// TODO: find means of testing oldOrg with newOrg - we may not need to actually perform the update
	updatedOrg, err := o.Client.OrganizationUpdate(spec.ID, &spec)
	if err != nil {
		return err
	}

	_ = updatedOrg

	return nil
}
