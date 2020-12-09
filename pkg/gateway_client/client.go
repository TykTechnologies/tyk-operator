package gateway_client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"github.com/levigross/grequests"
)

const (
	endpointAPIs     = "/tyk/apis"
	endpointCerts    = "/tyk/certs"
	endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/tyk/policies"
)

var (
	notFoundError = errors.New("api not found")
)

type ResponseMsg struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

func (c *Client) opts() *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.env.Auth,
			"content-type":        "application/json",
		},
		InsecureSkipVerify: c.env.InsecureSkipVerify,
	}
}

func NewClient(log logr.Logger, env environmet.Env) *Client {
	c := &Client{
		env: env,
	}
	return c
}

type Client struct {
	log logr.Logger
	env environmet.Env
}

func (c *Client) Api() universal_client.UniversalApi {
	return Api{Client: c}
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return SecurityPolicy{Client: c}
}

func (c *Client) HotReload() error {
	sess := grequests.NewSession(c.opts())

	fullPath := c.env.JoinURL(endpointReload)
	res, err := sess.Get(fullPath, c.opts())

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

// TODO: Webhook Requires implementation
func (c *Client) Webhook() universal_client.UniversalWebhook {
	panic("implement me")
}

// TODO: Organization requires implementation
func (c *Client) Organization() universal_client.UniversalOrganization {
	panic("implement me")
}

// TODO: Certificate Requires implementation
func (c *Client) Certificate() universal_client.UniversalCertificate {
	panic("implement me")
}
