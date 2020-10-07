package dashboard_admin_client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/levigross/grequests"
)

const (
	endpointOrganizations = "/admin/organisations"
)

type Client struct {
	url                string
	secret             string
	insecureSkipVerify bool
	log                logr.Logger
	opts               *grequests.RequestOptions
}

func NewClient(url string, auth string, insecureSkipVerify bool) *Client {
	c := &Client{
		url:                url,
		insecureSkipVerify: insecureSkipVerify,
		opts: &grequests.RequestOptions{
			Headers: map[string]string{
				"admin-auth":   auth,
				"content-type": "application/json",
			},
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	return c
}

func JoinUrl(parts ...string) string {
	l := len(parts)
	if l == 1 {
		return parts[0]
	}
	ps := make([]string, l)
	for i, part := range parts {
		if i == 0 {
			ps[i] = strings.TrimRight(part, "/")
		} else {
			ps[i] = strings.TrimLeft(part, "/")
		}
	}
	return strings.Join(ps, "/")
}

func (c Client) OrganizationAll() ([]v1alpha1.OrganizationSpec, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointOrganizations)

	// -2 means get all pages
	queryStruct := struct {
		Pages int `url:"p"`
	}{
		Pages: -2,
	}

	sess.RequestOptions.QueryStruct = queryStruct

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var orgsResponse OrganizationsResponse
	if err := res.JSON(&orgsResponse); err != nil {
		return nil, err
	}

	return orgsResponse.Organizations, nil
}

func (c Client) OrganizationCreate() ([]v1alpha1.OrganizationSpec, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointOrganizations)

	// -2 means get all pages
	queryStruct := struct {
		Pages int `url:"p"`
	}{
		Pages: -2,
	}

	sess.RequestOptions.QueryStruct = queryStruct

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var orgsResponse OrganizationsResponse
	if err := res.JSON(&orgsResponse); err != nil {
		return nil, err
	}

	return orgsResponse.Organizations, nil
}
