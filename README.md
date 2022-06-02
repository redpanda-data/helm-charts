# Redpanda Helm Chart

![Lint and Test Charts](https://github.com/redpanda-data/helm-charts/actions/workflows/lint-test.yml/badge.svg?branch=main)

***Status: Early Access***

This is the helm chart for [Redpanda](https://redpanda.com). It provides the ability to set up a multi node redpanda cluster with optional:

- TLS and/or SASL
- external access
- load balancing
- [tiered storage](https://docs.redpanda.com/docs/data-management/tiered-storage/)

See the [values.yaml](./redpanda/values.yaml) file to see all possible properties.

Multiple [examples](./examples/) show how to configure the chart to enable various features of Redpanda and Kubernetes.

## Prerequisites

### Required software

* Helm >= 3.0
* Kubernetes >= 1.18
* Cert-Manager
* MetalLB (optional)

## Preparation

> For now this chart is not being uploaded to a repo since it is in active development. This will change soon.

Clone this repo:

```sh
> git clone https://github.com/redpanda-data/helm-charts.git
> cd helm-charts
```

### Create cluster

It is likely that you will have your own Kubernetes cluster (e.g. local, GKE, EKS, etc.). But a local multi-node cluster can be created using one of the following instructions for either [Kind](#create-cluster-via-kind) or [Minikube](#create-cluster-via-minikube) (only use one of these options :D).

#### Option 1: Minikube

[Install minikube](https://k8s-docs.netlify.app/en/docs/tasks/tools/install-minikube/) (if needed), then start a single node cluster:

```sh
> minikube start
```

#### Option 2: Kind

[Install Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (if needed), then create a Kind cluster config and start a new cluster with it:

```sh
> cat <<EOF > tri-node-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF

> kind create cluster --name redpanda --config=tri-node-config.yaml
> kubectl config current-context
> kubectl get nodes -o wide
```

### Install cert-manager

[cert-manager](https://cert-manager.io/docs/installation/) is almost always needed (especially if you intend to use TLS). To install via helm:

```sh
> helm repo add jetstack https://charts.jetstack.io && \
helm repo update && \
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.8.0 \
  --set installCRDs=true
```

## Redpanda

At this point you have a cluster and other pre-requisites available. We are now ready to install Redpanda into the cluster. Most of the time you will want Redpanda to be contained in its own namespace. This can be done with the following command:

```sh
> helm install redpanda redpanda -n redpanda --create-namespace
```

The above command uses the default values from `values.yaml` to create multiple kubernetes objects in the redpanda namespace:

```sh
> kubectl get all -A --field-selector=metadata.namespace=redpanda
NAMESPACE   NAME             READY   STATUS    RESTARTS   AGE
redpanda    pod/redpanda-0   1/1     Running   0          48s
redpanda    pod/redpanda-1   1/1     Running   0          48s
redpanda    pod/redpanda-2   1/1     Running   0          48s

NAMESPACE   NAME                        TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                        AGE
redpanda    service/redpanda            ClusterIP   None             <none>        9092/TCP,9644/TCP,8082/TCP                                     48s
redpanda    service/redpanda-cluster    ClusterIP   10.100.155.122   <none>        8083/TCP,18081/TCP                                             48s
redpanda    service/redpanda-external   NodePort    10.109.201.86    <none>        9093:32005/TCP,9644:30494/TCP,8083:30658/TCP,18081:31127/TCP   48s

NAMESPACE   NAME                        READY   AGE
redpanda    statefulset.apps/redpanda   3/3     48s
```

## Next steps

Now you are ready to customize your configuration however you like. Check the [examples](./examples) folder for guides on enabling various Redpanda features.

## Cleanup

Once you are done with your Redpanda cluster, the following command will uninstall all objects created in the redpanda namespace by the helm chart:

```sh
> helm uninstall redpanda -n redpanda
```

You may also want to delete the cluster. With Kind:

```sh
> kind delete cluster --name redpanda
```

Or with Minikube:

```sh
> minikube delete
```

## External Access

### Created Services 

| Type | headless | load balanced |node ports | externally load balanced |
| :--- | :---: | :---: | :---: | :---: |
| Kafka API | y | n | y | y |
| Admin API | y | n | y | WIP |
| Schema Registry | y | y  | y | WIP |
| PandaProxy API | y | y  | y | WIP |

The chart will create the headless service as in the internal connectivity case, and can also create further services to support external connectivity:

A load-balanced ClusterIP service that is used as an entrypoint for the Pandaproxy.

A Nodeport service used to expose each API to the node's external network. Make sure that the node is externally accesible.
