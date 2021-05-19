package universal_client

import "github.com/TykTechnologies/tyk-operator/pkg/environmet"

type UniversalClient interface {
	Environment() environmet.Env
	HotReload() error
	Api() Api
	SecurityPolicy() Policy
	Certificate() Certificate
}
