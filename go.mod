module github.com/TykTechnologies/tyk-operator

go 1.15

require (
	cloud.google.com/go v0.45.1 // indirect
	github.com/cucumber/godog v0.10.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.6
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/kind v0.9.0
)
