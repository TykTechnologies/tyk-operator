package mockdash

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

const (
	endpointCerts = "/api/certs"
	// endpointReload   = "/tyk/reload/group"
	endpointPolicies = "/api/portal/policies"
)

var _ universal.Client = (*Client)(nil)

const (
	XAuthorization = "authorization"
	XContentType   = "content-type"
)

type ResponseMsg struct {
	Key     string `json:"Key,omitempty"`
	Status  string `json:"Status,omitempty"`
	Action  string `json:"Action,omitempty"`
	Message string `json:"Message,omitempty"`
	Meta    string `json:"Meta,omitempty"`
}

type Client struct{}

func (c Client) Certificate() universal.Certificate {
	return Cert{}
}

func (c Client) Portal() universal.Portal {
	return Portal{}
}

func (c Client) Api() universal.Api {
	return &mockDashApi{}
}

func (c Client) HotReload(context.Context) error {
	return nil
}
