# development environment

Prerequisites

- `docker`
- `kubectl`
- [kind](https://kind.sigs.k8s.io/)

We use local registry to publish and consume our operator images in development.

All commands on this guide must be executed at the root of the operator project.

## Setup the cluster

```
go run dev/main.go up
```

This will create a 4 node cluster with one control plane node and 3 worker nodes.
The script also takes care of setting up the local registry to which we will be
publishing our development images.

## Deleting the cluster

```
go run dev/main.go down
```

This will delete the development cluster. Be aware of directories mounted on nodes
They will not be deleted because you might want to inspect/persist the data after
shutting down the cluster. They are all in `tmp/{clusterName}/node-{0..3}`