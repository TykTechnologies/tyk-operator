# GraphQL Federation with Tyk Operator

Tyk, with release *4.0* offers GraphQL federation that allows you to divide GQL implementation across multiple back-end services, while still exposing them all as a single graph for the consumers.

> Tyk Operator does **_not_** yet fully support [GraphQL Federation](https://tyk.io/docs/getting-started/key-concepts/graphql-federation/). It is still in **_POC_** and under **_WIP_**.

- Create subgraphs

```bash
kubectl apply -f config/samples/09-federation/users-subgraph.yaml
kubectl apply -f config/samples/09-federation/posts-subgraph.yaml
```

- Create supergraph

```bash
kubectl apply -f config/samples/09-federation/supergraph.yaml 
```

**Note**: For POC purposes, `merged_sdl` of the super graph should be filled manually. We will update our APIs with the new endpoints required by Fedaration.