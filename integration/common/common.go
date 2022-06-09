package common

import (
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

var Env environmet.Env

type ctxKey string

const (
	CtxNSKey     ctxKey = "namespaceName"
	CtxApiName   ctxKey = "apiName"
	CtxOpCtxName ctxKey = "opCtxName"

	DefaultWaitTimeout  = 30 * time.Second
	DefaultWaitInterval = 1 * time.Second

	OperatorNamespace = "tyk-operator-system"
	GatewayLocalhost  = "http://localhost:7000"

	TestApiDef      = "test-http"
	TestOperatorCtx = "mycontext"
)
