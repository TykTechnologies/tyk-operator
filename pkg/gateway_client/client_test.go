package gateway_client

import (
	"testing"
)

func getClient() *Client {
	return NewClient("gateway-client-url", "api-key", true, "myorg")
}

func TestClient_HotReload(t *testing.T) {
	t.Skip("we don't run tests as we have no GW installation")
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
