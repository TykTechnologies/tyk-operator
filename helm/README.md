## Tyk Operator

Tyk Operator brings Full Lifecycle API Management capabilities to Kubernetes. Configure Ingress, APIs, Security Policies, Authentication, Authorization, Mediation and more - all using GitOps best practices with Custom Resources and Kubernetes-native primitives.

### Usage

```bash
helm repo add tyk-helm https://helm.tyk.io/public/helm/charts/
helm repo update
```

### Prerequisites

Before installing the Operator make sure you follow this guide and complete all steps from it, otherwise the Operator won't function properly: https://github.com/TykTechnologies/tyk-operator/blob/master/docs/installation/installation.md#tyk-operator-installation

**Note:** cert-manager is required as described [here](../docs/installation/installation.md#installing-cert-manager). If you haven't installed
cert-manager yet, you can install it as follows:
```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

### Installation
If you have fully functioning & bootstrapped Tyk Installation and cert-manager, you can install Tyk Operator as follows: 

```bash
helm install tyk-operator tyk-helm/tyk-operator
```
