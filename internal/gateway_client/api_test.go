package gateway_client

import (
	"testing"
)

func TestApi_All(t *testing.T) {
	c := getClient()
	apis, err := c.Api.All()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, a := range apis {
		t.Log("api:", a.APIID, a.Slug, a.Proxy.ListenPath)
	}
}

func TestApi_Create(t *testing.T) {
	t.Fatal("no test")
}

func TestApi_Update(t *testing.T) {
	t.Fatal("no test")
}

func TestApi_Delete(t *testing.T) {
	t.Fatal("no test")
}
