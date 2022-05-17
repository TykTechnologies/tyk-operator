# GraphQL Federation with Tyk Operator

- [Introduction](#introduction)
- [New CRDs](#new-crds)
    - [SubGraph](#subgraph)
    - [SuperGraph](#supergraph)
- [Updating SubGraph](#updating-subgraph)
- [Deleting SubGraph](#deleting-subgraph)
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

## New CRDs

GraphQL Federation uses concepts of Subgraph and Supergraph.

**Subgraph** is a representation of a back-end service and defines a distinct GraphQL schema. It can be queried directly as a separate service or it can be federated into a larger schema of a supergraph.

**Supergraph** is a composition of several subgraphs that allows the execution of a query across multiple services in the backend.

Tyk Operator introduces new Custom Resources called SubGraph and SuperGraph, that allows users to create graphs.

### SubGraph

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: SubGraph
metadata:
  name: users-subgraph
spec:
  subgraph:
    schema: |
      <Schema of your SubGraph>
    sdl: |
      <SDL of the SubGraph>
```

SubGraph CRD takes `schema` and `sdl` values for your subgraph. However, creating SubGraph CRD is inadequate to use it as 
regular ApiDefinition. SubGraph must be referred by ApiDefinition object through `graphql.graph_ref` field, as follows:
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: subgraph-api
spec:
  name: Federation - Subgraph
  ... 
  graphql:
    enabled: true
    execution_mode: subgraph
    graph_ref: users-subgraph
    version: "2"
    playground:
      enabled: false
      path: ""
  proxy:
    target_url: http://users.default.svc:4001/query
    listen_path: /users-subgraph/
    disable_strip_slash: true
```

#### Rules
An ApiDefinition must adhere to the following rules in order to represent an ApiDefinition for your SubGraph CRDs.

1. ApiDefinition and SubGraph must be in the same namespace,
2. `graphql.execution_mode` must be set to `subgraph`,
3. `graphql.graph_ref` must be set to the name of the SubGraph resource that you would like to refer.

Tyk Operator will update your ApiDefinition according to the SubGraph CRD. Any changes to SubGraph CRD will update corresponding
ApiDefinition by Tyk Operator. Therefore, as an end-user, you do not need to update your ApiDefinition manually after 
updating the SubGraph.

### SuperGraph

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: SuperGraph
metadata:
  name: social-media-supergraph
spec:
  subgraph_refs:
    - name: users-subgraph
      namespace: default
  schema: |-
    <Schema of your Supergraph>
```


SuperGraph CRD takes subgraph_refs and schema values for your supergraph. `subgraph_refs` is an array of SubGraph CRD 
references which expects the name and namespace of the referenced subgraph. 

Tyk Operator will update your SuperGraph ApiDefinition when one of the subgraphs that you reference in `subgraph_refs` changes.

We should create an ApiDefinition that represents the ApiDefinition for your SuperGraph CRD, just as we did with the SubGraph.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: federation-supergraph
spec:
  name: Federated - Social Media APIS
  ...
  graphql:
    execution_mode: supergraph
    graph_ref: social-media-supergraph
    enabled: true
    version: "2"
    playground:
      enabled: true
      path: /playground
  proxy:
    target_url: ""
    strip_listen_path: true
    listen_path: /social-media-apis-federated/
```

#### Rules

An ApiDefinition must adhere to the following rules in order to represent an ApiDefinition for your SuperGraph CRDs.

1. ApiDefinition and SuperGraph must be in the same namespace,
2. `graphql.execution_mode` must be set to `supergraph`,
3. `graphql.graph_ref` must be set to the name of the SuperGraph resource that you would like to refer.

When you make an update on one of the subgraphs defined under your SuperGraph CRD's `subgraph_refs` field, Tyk Operator
will update
- Subgraph CRD,
- Subgraph's ApiDefinition,
- SuperGraph's ApiDefinition.

Therefore, once you make an update on SubGraph CRD, you do not need to update your supergraph. It will be updated by Tyk Operator.

With this approach, multiple teams can work on SubGraph CRDs and Tyk Operator will update the relevant SuperGraph ApiDefinition.

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

