package dashboard_client

import (
	"testing"
)

// TODO: this needs to be pulled from K8s secrets
func getClient() *Client {
	return NewClient("https://655489409362.ngrok.io", "de2fc79499804c7072372b859e712b82", true)
}

func TestClient_HotReload(t *testing.T) {
	t.Skip("need to find out how to hot reload in dashboard")
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
