package gateway

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
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
		HTTPClient: universal.HTTPClient{
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
	universal.HTTPClient
}

func (c *Client) Api() universal.Api {
	return &Api{c}
}

func (c *Client) SecurityPolicy() universal.Policy {
	return SecurityPolicy{}
}

func (c *Client) HotReload() error {
	res, err := c.Get(c.Env.JoinURL(endpointReload), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := universal.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

// TODO: Certificate Requires implementation
func (c *Client) Certificate() universal.Certificate {
	panic("implement me")
}
