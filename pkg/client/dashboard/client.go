package dashboard

import (
	"errors"
	"net/http"

	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
)

const (
	endpointAPIs  = "/api/apis"
	endpointCerts = "/api/certs"
	//endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/api/portal/policies"
	endpointWebhooks = "/api/hooks"
)

var _ universal.Client = (*Client)(nil)

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
		HTTP: client.HTTP{
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
	client.HTTP
}

func (c *Client) Certificate() universal.Certificate {
	return &Cert{c}
}

func (c *Client) Portal() universal.Portal {
	return &Portal{c}
}

func (c *Client) Api() universal.Api {
	return &Api{c}
}

func (c *Client) HotReload() error {
	return nil
}
