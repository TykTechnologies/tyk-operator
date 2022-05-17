# GraphQL Federation with Tyk Operator

- [Introduction](#introduction)
- [Updating SubGraph](#tyk-licensing)
- [Deleting SubGraph](#installing-tyk)
    - [SubGraph without any reference](#subgraph-without-any-reference)
    - [SubGraph referenced in ApiDefinition](#subgraph-referenced-in-apidefinition)
    - [SubGraph referenced in SuperGraph](#subgraph-referenced-in-supergraph)
- [Deleting SuperGraph](#deleting-supergraph)
    - [SuperGraph without any reference](#supergraph-without-any-reference)
    - [SuperGraph referenced in ApiDefinition](#supergraph-referenced-in-apidefinition)

# Introduction

Tyk, with release *4.0* offers GraphQL federation that allows you to divide GQL implementation across multiple back-end 
services, while still exposing them all as a single graph for the consumers.

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

## Updating SubGraph

Updating SubGraph is an easy operation with new CRDs. Once you update your SubGraph CRD, Tyk Operator will update your 
SubGraph and Supergraph ApiDefinition that has a reference to the SubGraph you are updating.

As an end-user, once you've updated the SubGraph, you do not need to update the SuperGraph ApiDefinition fields. The operator
performs these tasks.

### Example

Let's assume that a developer responsible for the Users SubGraph would like to delete `username` field from the Users SubGraph.
Also, the Supergraph called Social Media already uses the Users Subgraph.

To achieve this, the developer should update the Users SubGraph CRD. Once the SubGraph CRD is updated, Tyk Operator will:
1. Update Users SubGraph CRD,
2. Update Social Media Supergraph ApiDefinition since it is referencing the Users SubGraph CRD.

## Deleting SubGraph

### SubGraph without any reference

If the subgraph is not referenced in any ApiDefinition CRD or SuperGraph CRD, it is easy to delete SubGraph CRDs as follows:
```bash
kubectl delete subgraphs.tyk.tyk.io <SUBGRAPH_NAME>
```

### SubGraph referenced in ApiDefinition

If you have a subgraph which is referenced in any ApiDefinition, Tyk Operator will not delete the SubGraph.

In order to delete this subgraph, corresponding ApiDefinition CRD must be updated, such that it has no reference to the 
subgraph in `graph_ref` field.

### SubGraph referenced in SuperGraph

Although the subgraph is not referenced in any ApiDefinition, if it is referenced in the SuperGraph, Tyk Operator will 
not delete the subgraph again.

In order to delete this subgraph, SuperGraph CRD should not have reference to corresponding subgraph in the `subgraphs_ref`.

## Deleting SuperGraph

### SuperGraph without any reference
If the supergraph is not referenced in any ApiDefinition CRD, it can be deleted as follows:
```bash
kubectl delete supergraphs.tyk.tyk.io <SUPERGRAPH_NAME>
```

### SuperGraph referenced in ApiDefinition
If a supergraph is referenced in any ApiDefinition, the Tyk Operator will not delete the SuperGraph CRD.

In order to delete this supergraph, ApiDefinition that has a reference to the supergraph must de-reference the supergraph
or be deleted.

