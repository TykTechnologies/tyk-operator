package gateway_client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
)

const (
	endpointAPIs   = "/tyk/apis"
	endpointCerts  = "/tyk/certs"
	endpointReload = "/tyk/reload/group"
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

func NewClient(log logr.Logger, env environmet.Env) *Client {
	c := &Client{
		Client: universal_client.Client{
			Log: log,
			Env: env,
			BeforeRequest: func(h *http.Request) {
				h.Header.Set("x-tyk-authorization", env.Auth)
				h.Header.Set("content-type", "application/json")
			},
		},
	}
	return c
}

type Client struct {
	universal_client.Client
}

func (c *Client) Api() universal_client.UniversalApi {
	return &Api{c}
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return SecurityPolicy{}
}

func (c *Client) HotReload() error {
	res, err := c.Get(c.Env.JoinURL(endpointReload), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

// Webhook TODO: Requires implementation
func (c *Client) Webhook() universal_client.UniversalWebhook {
	panic("implement me")
}

// Organization TODO: requires implementation
func (c *Client) Organization() universal_client.UniversalOrganization {
	panic("implement me")
}

// Certificate TODO: Requires implementation
func (c *Client) Certificate() universal_client.UniversalCertificate {
	panic("implement me")
}

func (c *Client) PortalCatalogue() universal_client.UniversalPortalAPI {
	panic("implement me")
}
