package gateway_client

import (
	"testing"
)

func TestApi_All(t *testing.T) {
	t.SkipNow()
	c := getClient()
	apis, err := c.Api().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, a := range apis {
		t.Log("api:", a.APIID, a.Proxy.ListenPath)
	}
}

func TestApi_Create(t *testing.T) {
	t.SkipNow()
}

func TestApi_Update(t *testing.T) {
	t.SkipNow()
}

func TestApi_Delete(t *testing.T) {
	t.SkipNow()
}
