package dashboard_client

import (
	"errors"
	"strings"

	"github.com/TykTechnologies/tyk-operator/internal/universal_client"

	"github.com/go-logr/logr"

	"github.com/levigross/grequests"
)

const (
	endpointAPIs = "/api/apis"
	//endpointCerts    = "/tyk/certs"
	//endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/api/portal/policies"
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

func NewClient(url string, auth string, insecureSkipVerify bool, orgID string) *Client {
	c := &Client{
		url:                url,
		insecureSkipVerify: false,
		opts: &grequests.RequestOptions{
			Headers: map[string]string{
				"authorization": auth,
				"content-type":  "application/json",
			},
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	return c
}

type Client struct {
	url                string
	secret             string
	orgID              string
	insecureSkipVerify bool
	log                logr.Logger
	opts               *grequests.RequestOptions
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return &SecurityPolicy{c}
}

func (c *Client) Api() universal_client.UniversalApi {
	return &Api{c}
}

func (c *Client) HotReload() error {
	//c.log.WithValues("Action", "HotReload")
	//c.log.Info("not implemented")

	return nil
}
