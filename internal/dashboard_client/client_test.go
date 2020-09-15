package dashboard_client

import (
	"testing"
)

func getClient() *Client {
	return NewClient("https://62cf8c685845.ngrok.io", "de2fc79499804c7072372b859e712b82", true)
}

func TestClient_HotReload(t *testing.T) {
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}
}
