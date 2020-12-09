package dashboard_client

import (
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type Kase = universal_client.Kase

func newKlient(c universal_client.Client) universal_client.UniversalClient {
	x := NewClient(c.Log, c.Env)
	x.Do = c.Do
	return x
}
