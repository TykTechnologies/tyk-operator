package dashboard_client

import (
	"errors"
	"strings"

	"github.com/go-logr/logr"

	"github.com/levigross/grequests"
)

const (
	endpointAPIs     = "/api/apis"
	endpointCerts    = "/tyk/certs"
	endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/tyk/policies"
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

func NewClient(url string, auth string, insecureSkipVerify bool) *Client {
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
	insecureSkipVerify bool
	log                logr.Logger
	opts               *grequests.RequestOptions
}

func (c Client) Api() *Api {
	return &Api{Client: &c}
}

func (c Client) SecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{Client: &c}
}

func (c Client) HotReload() error {
	c.log.WithValues("Action", "HotReload")
	c.log.Info("not implemented")

	return nil
}
