package universal

import "github.com/TykTechnologies/tyk-operator/pkg/environmet"

type Client interface {
	Environment() environmet.Env
	HotReload() error
	Api() Api
	Portal() Portal
	Certificate() Certificate
}
