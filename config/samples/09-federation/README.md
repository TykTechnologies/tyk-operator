# GraphQL Federation with Tyk Operator

Tyk Operator does **not** yet fully support [GraphQL Federation](https://tyk.io/docs/getting-started/key-concepts/graphql-federation/).


- Create sample APIs used by the Federation examples.

```bash
kubectl apply -f config/samples/09-federation/apis.yaml
```

This will create Services and Deployments of the sample GraphQL APIs that will be used by the Federation.

- Create subgraphs

```bash
kubectl apply -f config/samples/09-federation/users-subgraph.yaml
kubectl apply -f config/samples/09-federation/posts-subgraph.yaml
```

- Create supergraph

```bash
kubectl apply -f config/samples/09-federation/supergraph.yaml 
```
