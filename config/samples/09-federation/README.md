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

## Deleting SubGraph

### SubGraph without any reference

If the subgraph is not referenced in any ApiDefinition CRD or SuperGraph CRD, it is easy to delete SubGraph CRDs as follows:
```bash
kubectl delete subgraphs.tyk.tyk.io <SUBGRAPH_NAME>
```

### SubGraph referenced in ApiDefinition

If you have a subgraph which is referenced in any ApiDefinition, Tyk Operator will not delete the SubGraph.

In order to delete this subgraph, corresponding ApiDefinition CRD must be updated, such that it has no reference to the subgraph in `graph_ref` field.

### SubGraph referenced in SuperGraph

Although the subgraph is not referenced in any ApiDefinition, if it is referenced in the SuperGraph, Tyk Operator will not delete the subgraph again.

In order to delete this subgraph, SuperGraph CRD should not have reference to corresponding subgraph in the `subgraphs_ref`.

## Deleting SuperGraph

### SuperGraph without any reference
If the supergraph is not referenced in any ApiDefinition CRD, it can be deleted as follows:
```bash
kubectl delete supergraphs.tyk.tyk.io <SUPERGRAPH_NAME>
```

### SuperGraph referenced in ApiDefinition
If supergraph is referenced any ApiDefinition, Tyk Operator will not delete the SuperGraph CRD.
