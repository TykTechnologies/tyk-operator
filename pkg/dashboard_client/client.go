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
				h.Header.Set("content-type", env.Auth)
			},
		},
	}
}

type Client struct {
	universal_client.Client
}

func (c *Client) Organization() universal_client.UniversalOrganization {
	return &Organization{c}
}

func (c *Client) Certificate() universal_client.UniversalCertificate {
	return &Cert{c}
}

func (c *Client) Webhook() universal_client.UniversalWebhook {
	return &Webhook{c}
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return &SecurityPolicy{c}
}

func (c *Client) Api() universal_client.UniversalApi {
	return &Api{c}
}

func (c *Client) HotReload() error {
	c.Log.Info("hot reload not implemented", "Action", "HotReload")
	return nil
}
