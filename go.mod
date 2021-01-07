module github.com/TykTechnologies/tyk-operator

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/cucumber/godog v0.10.0
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/pkg/errors v0.9.1
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	moul.io/http2curl/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)

replace github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
