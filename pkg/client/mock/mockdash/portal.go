package mockdash

import "github.com/TykTechnologies/tyk-operator/pkg/client/universal"

type mockPortal struct{}

func (p mockPortal) Policy() universal.Policy {
	return mockSecurityPolicy{}
}

func (mockPortal) Documentation() universal.Documentation {
	return Documentation{}
}

func (mockPortal) Catalogue() universal.Catalogue {
	return Catalogue{}
}

func (mockPortal) Configuration() universal.Configuration {
	return Configuration{}
}
