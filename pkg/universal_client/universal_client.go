package universal_client

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

type UniversalClient interface {
	Environment() environmet.Env
	HotReload(context.Context) error
	Api() UniversalApi
	SecurityPolicy() UniversalSecurityPolicy
	Webhook() UniversalWebhook
	Certificate() UniversalCertificate
	Organization() UniversalOrganization
}
