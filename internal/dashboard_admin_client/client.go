package dashboard_admin_client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
)

const (
	endpointOrganizations                    = "/admin/organisations"
	endpointUsers                            = "/admin/users"
	endpointDashboardUserPasswordResetFormat = "/api/users/%s/actions/reset"
)

var ErrNotFound = errors.New("NotFound")

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

func (c Client) OrganizationGet(id string) (*v1alpha1.OrganizationSpec, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointOrganizations, id)

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	var resMsg v1alpha1.OrganizationSpec
	if err := res.JSON(&resMsg); err != nil {
		return nil, err
	}

	return &resMsg, nil
}

func (c Client) OrganizationUpdate(id string, spec *v1alpha1.OrganizationSpec) (*v1alpha1.OrganizationSpec, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointOrganizations, id)

	_, err := sess.Put(fullPath, &grequests.RequestOptions{JSON: spec})
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (c Client) OrganizationCreate(spec *v1alpha1.OrganizationSpec) (string, error) {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointOrganizations)

	res, err := sess.Post(fullPath, &grequests.RequestOptions{JSON: spec})
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error creating org: %d", res.StatusCode)
	}

	var createOrgResponse CreateOrganizationResponse
	if err := res.JSON(&createOrgResponse); err != nil {
		return "", err
	}

	return createOrgResponse.Meta, nil
}

func (c Client) UserAdminCreate(reqBody CreateUserRequest) error {
	sess := grequests.NewSession(c.opts)

	fullPath := JoinUrl(c.url, endpointUsers)

	res, err := sess.Post(fullPath, &grequests.RequestOptions{JSON: reqBody})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error creating user: %d", res.StatusCode)
	}

	var createUserResponse CreateUserResponse
	if err := res.JSON(&createUserResponse); err != nil {
		return err
	}

	passwordReqBody := SetPasswordRequest{NewPassword: reqBody.Password}

	// use the user's API key to set their password, calling the dashboard api
	sess = grequests.NewSession(&grequests.RequestOptions{
		Headers: map[string]string{
			"authorization": createUserResponse.Meta.AccessKey,
			"content-type":  "application/json",
		},
	})

	fullPath = JoinUrl(c.url, fmt.Sprintf(endpointDashboardUserPasswordResetFormat, createUserResponse.Meta.ID))

	res, err = sess.Post(fullPath, &grequests.RequestOptions{JSON: passwordReqBody})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status code setting password")
	}

	return nil
}
