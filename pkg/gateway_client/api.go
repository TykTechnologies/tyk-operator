package gateway_client

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/levigross/grequests"
)

var (
	apiCollisionError = errors.New("api id, listen path or slug collision")
)

type Api struct {
	*Client
}

func (a Api) All() ([]v1.APIDefinitionSpec, error) {
	sess := grequests.NewSession(a.opts())

	fullPath := a.env.JoinURL(endpointAPIs)

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var list []v1.APIDefinitionSpec
	if err := res.JSON(&list); err != nil {
		return nil, err
	}

	return list, nil
}

func (a Api) Get(apiID string) (*v1.APIDefinitionSpec, error) {
	sess := grequests.NewSession(a.opts())

	fullPath := a.env.JoinURL(endpointAPIs, apiID)

	res, err := sess.Get(fullPath, a.opts())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway API Returned error: %d", res.StatusCode)
	}

	var spec v1.APIDefinitionSpec
	if err := res.JSON(&spec); err != nil {
		return nil, err
	}
	a.log.Info("======> received ", res.String())
	return &spec, nil
}

func (a Api) Create(def *v1.APIDefinitionSpec) error {
	// get all apis
	list, err := a.All()
	if err != nil {
		return err
	}

	// check exists / collisions
	for _, api := range list {
		if api.APIID == def.APIID {
			return apiCollisionError
		}

		if api.Proxy.ListenPath == def.Proxy.ListenPath {
			return apiCollisionError
		}

	}

	// Create
	opts := a.opts()

	def.OrgID = a.env.Org
	opts.JSON = def
	fullPath := a.env.JoinURL(endpointAPIs)

	sess := grequests.NewSession(a.opts())

	res, err := sess.Post(fullPath, opts)
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
	def.APIID = resMsg.Key
	return nil
}

func (a Api) Update(def *v1.APIDefinitionSpec) error {
	list, err := a.All()
	if err != nil {
		return err
	}

	var apiToUpdate *v1.APIDefinitionSpec
	for _, api := range list {
		if api.APIID == def.APIID {
			apiToUpdate = &api
			break
		}
	}

	if apiToUpdate == nil {
		return notFoundError
	}

	// Update
	opts := a.opts()
	opts.JSON = def
	fullPath := a.env.JoinURL(endpointAPIs, apiToUpdate.APIID)

	sess := grequests.NewSession(a.opts())

	res, err := sess.Put(fullPath, opts)
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

func (a Api) Delete(id string) error {
	sess := grequests.NewSession(a.opts())

	delPath := a.env.JoinURL(endpointAPIs, id)

	res, err := sess.Delete(delPath, a.opts())
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
