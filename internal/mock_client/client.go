package mock_client

import (
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
)

func NewClient() *Client {
	c := &Client{}

	return c
}

type Client struct{}

func (c *Client) Api() universal_client.UniversalApi {
	return Api{Client: c}
}

func (c *Client) HotReload() error {
	return nil
}

func (c *Client) SecurityPolicy() universal_client.UniversalSecurityPolicy {
	return SecurityPolicy{Client: c}
}
