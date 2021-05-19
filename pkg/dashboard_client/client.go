package dashboard_client

import (
	"errors"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
)

const (
	endpointAPIs  = "/api/apis"
	endpointCerts = "/api/certs"
	//endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/api/portal/policies"
	endpointWebhooks = "/api/hooks"
)

const (
	XAuthorization = "authorization"
	XContentType   = "content-type"
)

var (
	notFoundError = errors.New("api not found")
)

type ResponseMsg struct {
	Key     string `json:"Key,omitempty"`
	Status  string `json:"Status,omitempty"`
	Action  string `json:"Action,omitempty"`
	Message string `json:"Message,omitempty"`
	Meta    string `json:"Meta,omitempty"`
}

func NewClient(log logr.Logger, env environmet.Env) *Client {
	return &Client{
		Client: universal_client.Client{
			Log: log,
			Env: env,
			BeforeRequest: func(h *http.Request) {
				h.Header.Set("authorization", env.Auth)
				h.Header.Set("content-type", "application/json")
			},
		},
	}
}

type Client struct {
	universal_client.Client
}

func (c *Client) Certificate() universal_client.Certificate {
	return &Cert{c}
}

func (c *Client) SecurityPolicy() universal_client.Policy {
	return &SecurityPolicy{c}
}

func (c *Client) Api() universal_client.Api {
	return &Api{c}
}

func (c *Client) HotReload() error {
	return nil
}
