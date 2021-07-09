package dashboard

import "github.com/TykTechnologies/tyk-operator/pkg/client/universal"

type Portal struct{}

func (p Portal) Policy() universal.Policy {
	return SecurityPolicy{}
}

func (Portal) Documentation() universal.Documentation {
	return Documentation{}
}

func (Portal) Catalogue() universal.Catalogue {
	return Catalogue{}
}

func (Portal) Configuration() universal.Configuration {
	return Configuration{}
}
