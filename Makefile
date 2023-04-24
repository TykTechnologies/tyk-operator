# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

TYK_VERSION ?= v4.3

# Image URL to use all building/pushing image targets
IMG ?= tyk-operator:latest

TAG = $(lastword $(subst :, ,$(IMG)))

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
#The name of the kind cluster used for development
CLUSTER_NAME ?= kind

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
#test: generate fmt vet manifests
#	go test ./... -coverprofile cover.out
# Run tests

# skip bdd when doing unit testing
UNIT_TEST=$(shell go list ./... | grep -v bdd)

test: generate fmt vet manifests
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	setup-envtest --arch=amd64 use
	go test ${UNIT_TEST}  -coverprofile test_coverage.out --timeout 30m


manager: generate fmt vet	## build manager binary
	go build -o bin/manager main.go

run: generate fmt vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	TYK_URL=${TYK_URL} TYK_MODE=${TYK_MODE} TYK_TLS_INSECURE_SKIP_VERIFY=${TYK_TLS_INSECURE_SKIP_VERIFY} TYK_AUTH=${TYK_AUTH} TYK_ORG=${TYK_ORG} ENABLE_WEBHOOKS=${ENABLE_WEBHOOKS} go run ./main.go

log: ## This will print logs of Tyk Operator pod.
	$(eval POD=$(shell kubectl get pod -l control-plane=tyk-operator-controller-manager -n tyk-operator-system -o name))
	kubectl logs -n tyk-operator-system ${POD} -c manager -f

install: manifests kustomize	## Install CRDs into a cluster
	$(KUSTOMIZE) build config/crd | kubectl apply -f -


uninstall: manifests kustomize	## Uninstall CRDs from a cluster
	$(KUSTOMIZE) build config/crd | kubectl delete -f -


deploy: manifests kustomize ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

helm: kustomize ## Update helm charts
	$(KUSTOMIZE) version
	$(KUSTOMIZE) build config/crd > ./helm/crds/crds.yaml
	$(KUSTOMIZE) build config/helm |go run hack/helm/pre_helm.go > ./helm/templates/all.yaml

manifests: controller-gen ## Generate manifests
	$(CONTROLLER_GEN) --version
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases


fmt: ## Run go fmt against code
	go fmt ./...
	gofmt -s -w .


vet: ## Run go vet against code
	go vet ./...

golangci-lint: ## Run golangci-lint linter
	golangci-lint run

linters: fmt vet golangci-lint ## Run all linters once

generate: controller-gen ## Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."


docker-build: test ## Build the docker image
	docker build . -t ${IMG}


docker-build-notest: ## Build the docker image
	docker build . -t ${IMG}


docker-push: ## Push the docker image
	docker push ${IMG}


release: ## Make release
# fail if version is not specified
ifndef VERSION
	$(error VERSION is not specified)
endif

	git checkout master
	yq eval -i '.version="${VERSION}"' helm/Chart.yaml
	yq eval -i '.image.tag="v${VERSION}"' helm/values.yaml
	git add helm/Chart.yaml helm/values.yaml
	git commit -m "version to: v${VERSION}"
	git push origin master && git tag v${VERSION} && git push --tags


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.4.1

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
    test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: bundle
bundle: manifests kustomize	## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle


.PHONY: bundle-build
bundle-build:	## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: cross-build-image
cross-build-image: ## Build docker image
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager.linux main.go
	docker build -f cross.Dockerfile . -t ${IMG}

.PHONY: install-cert-manager
install-cert-manager: ## Install cert manager
	@echo "===> installing cert-manager"
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.0.4/cert-manager.yaml
	kubectl rollout status  deployment/cert-manager -n cert-manager
	kubectl rollout status  deployment/cert-manager-cainjector -n cert-manager
	kubectl rollout status  deployment/cert-manager-webhook -n cert-manager

.PHONY: install-operator-helm
install-operator-helm: cross-build-image manifests helm	## Install operator using helm
	@echo "===> installing operator with helm"
	go run hack/cluster/load_image.go -image ${IMG} -cluster=${CLUSTER_NAME}
	helm install ci ./helm --values ./ci/helm_values.yaml --set image.tag=${TAG} -n tyk-operator-system --wait

.PHONY: scrap
scrap: generate manifests helm cross-build-image ## Re-install operator with helm
	@echo "===> re installing operator with helm"
	go run hack/cluster/load_image.go -image ${IMG} -cluster=${CLUSTER_NAME}
	helm uninstall ci -n tyk-operator-system
	kubectl apply -f ./helm/crds
	helm install ci ./helm --values ./ci/helm_values.yaml -n tyk-operator-system --wait

.PHONY: setup-pro
setup-pro:	## Install Tyk Pro
	helm repo add tyk-helm https://helm.tyk.io/public/helm/charts/
	helm repo update
	go run hack/bootstrap/create/main.go -debug  -mode=pro -cluster=${CLUSTER_NAME} -tyk_version=$(TYK_VERSION)

.PHONY: setup-ce
setup-ce:	## Install Tyk CE
	helm repo add tyk-helm https://helm.tyk.io/public/helm/charts/
	helm repo update
	go run hack/bootstrap/create/main.go -debug -mode=ce -cluster=${CLUSTER_NAME} -tyk_version=$(TYK_VERSION)


.PHONY: boot-pro
boot-pro: setup-pro install-operator-helm	## Install Tyk Pro and Operator
	@echo "******** Successful boot strapped pro dev env ************"

.PHONY: boot-ce
boot-ce:setup-ce install-operator-helm	## Install Tyk CE and  Operator
	@echo "******** Successful boot strapped ce dev env ************"

.PHONY: bdd
bdd:
	go test -timeout 400s -coverprofile bdd_coverage.out -v  ./bdd

.PHONY: test-all
test-all: test bdd ## Run tests

.PHONY: create-kind-cluster
create-kind-cluster:	## Create kind cluster
	kind create cluster --config hack/kind.yaml --name=${CLUSTER_NAME}

.PHONY: clean
clean:	## Delete kind cluster
	kind delete cluster --name=${CLUSTER_NAME}

.PHONY: install-venom
install-venom:
ifeq (, $(venom version))
	@echo "Installing venom"
	sudo curl https://github.com/ovh/venom/releases/download/v1.0.1/venom.linux-amd64 -L -o /usr/local/bin/venom && sudo chmod +x /usr/local/bin/venom
else
	@echo "Venom is already installed"
endif
	
.PHONY: run-venom-tests
run-venom-tests: install-venom ## Run Venom integration tests
	cd venom-tests && IS_TTY=true venom run

help:
	@fgrep -h "##" Makefile | fgrep -v fgrep |sed -e 's/\\$$//' |sed -e 's/:/-:/'| sed -e 's/:.*##//'