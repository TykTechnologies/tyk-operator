package gateway_client

import (
	"testing"
)

// TODO: create a gateway deployment & implement cleanup function
func getClient() *Client {
	return NewClient("gateway-client-url", "api-key", true, "myorg")
}

func TestClient_HotReload(t *testing.T) {
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
