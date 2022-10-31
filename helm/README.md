## Tyk Operator

Tyk Operator brings Full Lifecycle API Management capabilities to Kubernetes. Configure Ingress, APIs, Security Policies, Authentication, Authorization, Mediation and more - all using GitOps best practices with Custom Resources and Kubernetes-native primitives.

### Usage

```bash
helm repo add tyk-helm https://helm.tyk.io/public/helm/charts/
helm repo update
```

### Prerequisites

Before installing the Operator make sure you follow this guide and complete all 
steps from it, otherwise the Operator won't function properly: https://github.com/TykTechnologies/tyk-operator/blob/master/docs/installation/installation.md#tyk-operator-installation

**_NOTE_:** cert-manager is required as described [here](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#step-2-installing-cert-manager).
If you haven't installed `cert-manager` yet, you can install it as follows:
```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

### Installation
If you have fully functioning & bootstrapped Tyk Installation and cert-manager, 
you can install Tyk Operator as follows: 

```bash
helm install tyk-operator tyk-helm/tyk-operator
```

By default it will install latest stable release of operator.

You can install any other version by 
1. Setting `image.tag` in values.yml or with `--set {image.tag}={VERSION_TAG}` while doing the helm install. 
2. Installing CRDs of corresponding version. This is important as operator might not work otherwise. You can do so by running below command. 
```
kubectl apply -f https://raw.githubusercontent.com/TykTechnologies/tyk-operator/{VERSION_TAG}/helm/crds/crds.yaml
```

Replace `VERSION_TAG` with operator version tag.


> **_NOTE_:** If you want to install `latest` release of operator, replace `VERSION_TAG` with `master` while installing CRDs.
