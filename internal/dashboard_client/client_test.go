package dashboard_client

import (
	"testing"
)

// TODO: this needs to be pulled from K8s secrets
func getClient() *Client {
	return NewClient("dash-client-url", "api-key", true, "myorg")
}

func TestClient_HotReload(t *testing.T) {
	t.Skip("need to find out how to hot reload in dashboard")
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
