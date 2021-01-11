module github.com/TykTechnologies/tyk-operator

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/cucumber/godog v0.10.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	moul.io/http2curl/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.7.0
)

replace github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
