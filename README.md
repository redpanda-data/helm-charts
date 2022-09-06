# Redpanda Helm Chart

![Lint and Test Charts](https://github.com/vectorizedio/helm-charts/actions/workflows/lint-test.yml/badge.svg?branch=main)

## We have two different helm projects

Please note that we have two helm charts: `redpanda` (this project) and `redpanda-operator` ([here](https://github.com/redpanda-data/redpanda/tree/dev/src/go/k8s/helm-chart/charts/redpanda-operator)). These are two separate projects!

This helm chart (`redpanda`) focuses on providing a helm chart that deploys a Redpanda cluster according to the configuration in a values.yaml. Once deployed, you continue to use the helm command and modify [`values.yaml`](https://github.com/redpanda-data/helm-charts/blob/main/redpanda/values.yaml) to change and/or upgrade your Redpanda deployment.

The `redpanda-operator` chart installs a kubernetes operator that will deploy and manage a Redpanda cluster. The future state of the operator is in flux and may change in the near future. Helm is primarily used in that project only to deploy the operator, and from there you would interact with the operator and/or `kubectl` in order to modify your Redpanda cluster. `redpanda-operator` is released alongside Redpanda (see the latest release [here](https://github.com/redpanda-data/redpanda/releases)). For now, much of our site's helm documentation focuses on the `redpanda-operator` (see [here](https://docs.redpanda.com/docs/quickstart/kubernetes-qs-cloud/)). We are improving our documentation to have more extensive coverage of both the `redpanda-operator` and this `redpanda` helm chart.

This is the recommended chart. Feel free to use which ever helm chart you prefer, but keep in mind that they are separate, incompatible projects, and instructions for one will not apply to the other. A good rule of thumb is that if you see mention of the word "operator" in some resource, it's not related to this helm chart. This chart has no operator and no custom resource definitions (CRDs).

## Overview

This is the Helm Chart for [Redpanda](https://redpanda.com). It provides the ability to set up a multi node redpanda cluster with the following optional features:

- Schema registry (enabled by default)
- REST (aka PandaProxy, enabled by default)
- TLS
- SASL
- External access
- Load balancing

We have included an [examples folder](./examples/README.md) with more details on how to use this helm chart. Each example focuses on specific features like the ones listed above. We recommend completing the instructions in this README before continuing steps in any of these examples.

We have also included many comments through the [values.yaml](./redpanda/values.yaml) file. Please take a look at this file for more details.

## Prerequisites

### Required software

* Helm >= 3.0
* Kubernetes >= 1.21
* Cert-Manager (optional, needed for TLS)
* MetalLB (optional)

## Preparation

First, clone this repo:

```sh
git clone https://github.com/redpanda-data/helm-charts.git
cd helm-charts/redpanda
```

### Create cluster

It is likely that you will have your own Kubernetes cluster (e.g. local, GKE, EKS, etc.). But a local multi-node cluster can be created using one of the following instructions for either [Minikube](#option-1-minikube) or [Kind](#option-2-kind) (only use one of these options :D).

#### Option 1: Minikube

[Install minikube](https://k8s-docs.netlify.app/en/docs/tasks/tools/install-minikube/) (if needed), then start a 4-node cluster:

```sh
minikube start --nodes 4 --memory=3000m --extra-config=apiserver.service-node-port-range=8081-65535
```

This command starts minikube with 4 nodes (1 control plane, 3 worker nodes, 3G memory each) and will extend the NodePort range to include default ports for Redpanda services. This assumes the default memory size of 2.5Gi for each container is being used in `values.yaml`. You should modify the memory size given to each node based on your configuration and available memory.

Extending the NodePort range is optional, but it could be useful if using the default NodePort external service. Having a NodePort range that includes all Redpanda services allows you to assign a single port per listener which gets re-used by the external service.

#### Option 2: Kind

[Install Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (if needed), then create a Kind cluster config and start a new cluster with it:

```sh
cat <<EOF > tri-node-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF
kind create cluster --name redpanda --config=tri-node-config.yaml
kubectl config current-context
kubectl get nodes -o wide
```

### Install cert-manager

[cert-manager](https://cert-manager.io/docs/installation/) is needed if you intend to use TLS. To install via helm:

```sh
helm repo add jetstack https://charts.jetstack.io && \
helm repo update && \
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true
```

## Redpanda

At this point you have a cluster and other pre-requisites available and are now ready to install Redpanda into the cluster.
Install Redpanda using this chart into a namespace (eg. redpanda-ns) using the default values:

```sh
helm install redpanda redpanda -n redpanda-ns --create-namespace
```

Inspect the resources installed:

```sh
> kubectl get all -A --field-selector=metadata.namespace=redpanda-ns
NAMESPACE    NAME             READY   STATUS    RESTARTS   AGE
redpanda-ns  pod/redpanda-0   1/1     Running   0          48s
redpanda-ns  pod/redpanda-1   1/1     Running   0          48s
redpanda-ns  pod/redpanda-2   1/1     Running   0          48s
NAMESPACE    NAME                        TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                        AGE
redpanda-ns  service/redpanda            ClusterIP   None             <none>        9092/TCP,9644/TCP,8082/TCP                                     48s
redpanda-ns  service/redpanda-cluster    ClusterIP   10.100.155.122   <none>        8083/TCP,18081/TCP                                             48s
redpanda-ns  service/redpanda-external   NodePort    10.109.201.86    <none>        9093:32005/TCP,9644:30494/TCP,8083:30658/TCP,18081:31127/TCP   48s
NAMESPACE    NAME                        READY   AGE
redpanda-ns  statefulset.apps/redpanda   3/3     48s
```

## Next steps

Installing the way we have here will only use the default configuration options for this chart. Many times, you will want to customize the chart to use your preferred configuration.

To see what options are configurable on a chart, use helm show values:

```sh
helm show values redpanda
```

```yaml
# ...

# Common parameters
#
# Override redpanda.name template
nameOverride: ""
# Override redpanda.fullname template
fullnameOverride: ""
# Default kuberentes cluster domain
clusterDomain: cluster.local
# Additional labels added to all Kubernetes objects
commonLabels: {}

# Redpanda parameters
#
image:
  repository: vectorized/redpanda
  # Redpanda version. This determines the installed version (not Chart.appVersion)
  tag: v22.1.6
  # The imagePullPolicy will default to Always when the tag is 'latest'
  pullPolicy: IfNotPresent

# ...
```

You can override any of these settings in a yaml formatted file and pass that file during installation:
```sh
echo '{image.pullPolicy: Always}' > myvalues.yaml
helm install -f myvalues.yaml redpanda
```

The above will be merged with the default values, overriding just the pullPolicy, setting it to "Always". The rest of the defaults will be unchanged.

There are two ways to pass configuration data during install:

- --values (or -f): Specify a YAML file with overrides. This can be specified multiple times and the rightmost file will take precedence
- --set: Specify overrides on the command line.

The installation can be customized by providing values in yaml or json, or by overriding the defaults on the command line using the `--set` flag. Check the [examples](./examples) folder for guides on enabling various Redpanda features.

If both are used, --set values are merged into --values with higher precedence. For more information, see the [helm documentation][helm_customizing]

Many times you will be able apply updates without needing to re-install the entire cluster. If you make a change that only impacts a single service (for example), then running the following command will only restart that service and leave the rest of the cluster running with the same state:

```
helm -n redpanda-ns upgrade -f myvalues.yaml redpanda ./redpanda
```

## Cleanup

Once you are done with your Redpanda cluster, the following command will uninstall all objects created in the redpanda namespace by the helm chart:

```sh
> helm uninstall redpanda -n redpanda-ns
```

You may also want to delete the cluster. With Kind:

```sh
> kind delete cluster --name redpanda
```

Or with Minikube:

```sh
> minikube delete
```

[helm_customizing]: https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing
