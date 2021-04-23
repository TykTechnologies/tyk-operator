module github.com/TykTechnologies/tyk-operator

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/cucumber/godog v0.10.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/pkg/errors v0.9.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	moul.io/http2curl/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/gateway-api v0.2.0
)
