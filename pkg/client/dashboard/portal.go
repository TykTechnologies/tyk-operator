package dashboard

import "github.com/TykTechnologies/tyk-operator/pkg/client/universal"

type Portal struct {
	*Client
}

func (p Portal) Policy() universal.Policy {
	return &SecurityPolicy{p.Client}
}
