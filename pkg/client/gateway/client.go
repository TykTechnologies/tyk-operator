package gateway

import (
	"context"
	"fmt"

	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

const (
	endpointAPIs   = "/tyk/apis"
	endpointReload = "/tyk/reload/group"
	endpointCerts  = "/tyk/certs"
)

var _ universal.Client = (*Client)(nil)

type ResponseMsg struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

type Client struct{}

func (c Client) Api() universal.Api {
	return Api{}
}

func (c Client) Portal() universal.Portal {
	return Portal{}
}

func (c Client) HotReload(ctx context.Context) error {
	res, err := client.Get(ctx, endpointReload, nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var resMsg ResponseMsg

	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return nil
}

func (c Client) Certificate() universal.Certificate {
	return Cert{}
}
