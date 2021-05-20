package dashboard

import "github.com/TykTechnologies/tyk-operator/pkg/client/universal"

type Portal struct{}

func (p Portal) Policy() universal.Policy {
	return SecurityPolicy{}
}
