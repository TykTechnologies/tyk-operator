package dashboard_client

import (
	"testing"
)

// TODO: this needs to be pulled from K8s secrets
func getClient() *Client {
	return NewClient("http://localhost:3000", "80f26e6383ae4eda63048f64ff37ae1e", true, "5e9d9544a1dcd60001d0ed20")
}

func TestClient_HotReload(t *testing.T) {
	t.Skip("skip as we have no dash client")
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
