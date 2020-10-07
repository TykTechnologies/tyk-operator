package gateway_client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
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
		orgID:              orgID,
		insecureSkipVerify: false,
		opts: &grequests.RequestOptions{
			Headers: map[string]string{
				"x-tyk-authorization": auth,
				"content-type":        "application/json",
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
	opts               *grequests.RequestOptions
}

func (c *Client) Webhook() universal_client.UniversalWebhook {
	panic("implement me")
}

func (c *Client) Api() universal_client.UniversalApi {
	return Api{Client: c}
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return SecurityPolicy{Client: c}
}

func (c *Client) HotReload() error {
	fullPath := JoinUrl(c.url, endpointReload)
	res, err := grequests.Get(fullPath, c.opts)

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
