# OLD REPO: Please use the official k8s operator



```

# new k8s operator tree.


https://github.com/vectorizedio/redpanda/tree/dev/src/go/k8s



```

> NOTE: This is no longer supported.
> 
> This repo is left here for older users that want to deploy their own containers without the help from an automated operator.
> 


# Redpanda Helm Chart

![Lint and Test Charts](https://github.com/vectorizedio/helm-charts/actions/workflows/lint-test.yml/badge.svg?branch=main)

***Status: Early Access***

This is the Helm Chart for [Redpanda](https://redpanda.com). It provides the ability to set up a multi node redpanda cluster with optional:

- TLS 
- TLS and SASL 
- external access.

## Requirements

* Helm >= 3.0
* Kubernetes >= 1.18
* Cert-Manager

## Installation

First, clone this repo:

```
git clone https://github.com/redp01/helm-charts-1.git
cd helm-charts-1/redpanda
```

If required a multi node kind cluster can be created. Kind is shown here as an example; however, it is likely that you will have your own Kubernetes cluster e.g. GKE.

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

kind create cluster --name rp01rdmi --config=tri-node-config.yaml

kubectl get nodes
```

If you intend to install tls, then you are required to install cert-manager.

```sh
kubectl apply -f https://raw.githubusercontent.com/redp01/helm-charts-1/fixup-certs/certs/cert-manager.yaml
```

Depending on your intent with regards to securing the cluster it can be installed via 3 main methods whereby the successive application of values config layers is applied to fully build out the secuirty features.

If no TLS or SASl is required, simply invoke:

```sh
helm install redpanda .
```

or the equivalent command depending in your namespace for example.

This can be followed by a basic helm test that will change corresponding to your installation.

```sh
helm test redpanda
```
The output should indicate the success of the tests and some example commands you can try.

If TLS is required, invoke the following:

```sh
```

When this command is invoked the self signed issuers are created for each service (by default). These issuers are used to create per service keys and ca certs. This behaviour is for example, in a production installation 
it is likely that the issuerRef etc etc etc.

The following test commmand will ensure that the kafka api, panda proxy and schema registy all have basic access using TLS.

```sh
helm test redpanda
``` 

To further include SASL protection for your cluster the following command can be run, layering SASL configuration on top of the basic and TLS configuration additively.

```sh
helm install redpanda . -f values_add_tls.yaml -f values_add_sasl.yaml
```

In this case both TLS and SASL should now be enabled with a default admin user and test password (please dont use this password for your deployment).

The installation can be further tested with the following command:

```sh
helm test redpanda
```

#Troubleshooting




### Local Installation

First, clone this repo:

```sh
git clone git@github.com:vectorizedio/helm-charts.git
```

Install the Helm Chart:

```sh
helm install --namespace redpanda --create-namespace redpanda ./redpanda
```
