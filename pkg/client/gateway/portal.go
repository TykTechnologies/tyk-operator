package gateway

import "github.com/TykTechnologies/tyk-operator/pkg/client/universal"

type Portal struct{}

func (Portal) Policy() universal.Policy {
	return SecurityPolicy{}
}
