package dashboard_client

import (
	"errors"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"github.com/levigross/grequests"
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

func (c *Client) opts() *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"authorization": c.env.Auth,
			"content-type":  "application/json",
		},
		InsecureSkipVerify: c.env.InsecureSkipVerify,
	}
}

func NewClient(log logr.Logger, env environmet.Env) *Client {
	c := &Client{
		log: log,
		env: env,
	}
	return c
}

type Client struct {
	env environmet.Env
	log logr.Logger
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
	c.log.Info("hot reload not implemented", "Action", "HotReload")
	return nil
}
