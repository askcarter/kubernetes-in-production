# Creating and Managing Node Pools

In Kubernetes, a [Node](https://github.com/kubernetes/kubernetes/blob/release-1.2/docs/design/architecture.md#the-kubernetes-node)

In this lab you will:
* [Use labels and nodeSelectors to assign Pods to specific Nodes](http://kubernetes.io/docs/user-guide/node-selection/)
* [Set up a Node Pool to manage multiple types of Nodes in one Cluster](https://cloudplatform.googleblog.com/2016/05/introducing-Google-Container-Engine-GKE-node-pools.html)

## Tutorial:  Assign a Pod to a Specific Node
```
kubectl get nodes --show-labels
```

```
kubectl label nodes <node-name> nodeclass=high-mem
```

Now that we have labeled Node, we can use that to assign our Pods to the node by adding a nodeSelector to `deployments/persistent-db.yaml`:

Explore our deployment and examine the `nodeSelector` field:
```
cat deployments/persistent-db.yaml
```

## Exercise: Verify that our node was assigned properly
```
kubectl create -f deployments/persistent-db.yaml
```

```
kubectl get pods -o wide
```

## Tutorial:  Create a High Memory NodePool

The first thing we need to do is create a cluster for use during this workshop.
```
gcloud container clusters create workshop
```

Now if we look at our node pools we'll see we have 'default' nodes set up.
```
gcloud container node-pools list --cluster=workshop
```

The next step is to create our 'high memory' Node as part of our cluster.
```
gcloud container node-pools create high-mem --cluster=workshop --machine-type=custom-2-12288 --disk-size=200 --num-nodes=1
```

If we view all of our Nodes now, we'll see the high memory ones, too.
```
kubectl get nodes
```
It is now possible to delete the default-pool, leaving only the high memory nodes.
```
gcloud container node-pools delete default-pool --cluster=workshop
```

```
kubectl get nodes
```

## Exercise:  Create a new node pool to replace the deleted default pool.

> Use `gcloud container node-pools help` to learn the proper command syntax

## Exercise:  Assign the dbd pod to our high-mem node.
When this node is created it gets a label, we'll use that label in our NodeSelector field to assign our DB pod to this node.

View the labels on our 'high-mem' node
```
kubectl get nodes --show-labels
```

You'll see a label that looks like `cloud.google.com/gke-nodepool=`.  Update our deployment to assign our pod to this Node:

```
nano deployments/persistent-db.yaml
```

```
kubectl apply -f deployments/persistent-db.yaml
```
