package universal

import "context"

type Client interface {
	HotReload(context.Context) error
	Api() Api
	Portal() Portal
	Certificate() Certificate
	OAS() OAS
}
