package dashboard_client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/levigross/grequests"
)

const (
	endpointAPIs     = "/api/apis"
	endpointCerts    = "/tyk/certs"
	endpointReload   = "/tyk/reload/group"
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

	c.Api = &Api{c}
	c.SecurityPolicy = &SecurityPolicy{c}

	return c
}

type Client struct {
	url                string
	secret             string
	insecureSkipVerify bool
	opts               *grequests.RequestOptions
	Api                *Api
	SecurityPolicy     *SecurityPolicy
}

func (c Client) HotReload() error {
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
