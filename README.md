# Redpanda Helm Chart

![Lint and Test Charts](https://github.com/vectorizedio/helm-charts/actions/workflows/lint-test.yml/badge.svg?branch=main)

## We have two different helm projects

Please note that we have two helm charts: `redpanda` (this project) and `redpanda-operator` ([here](https://github.com/redpanda-data/redpanda/tree/dev/src/go/k8s/helm-chart/charts/redpanda-operator)). These are two separate projects!

This helm chart (`redpanda`) focuses on providing a helm chart that deploys a Redpanda cluster according to the configuration in a values.yaml. Once deployed, you continue to use the helm command and modify [`values.yaml`](https://github.com/redpanda-data/helm-charts/blob/main/redpanda/values.yaml) to change and/or upgrade your Redpanda deployment.

The `redpanda-operator` chart installs a golang-based operator that will deploy and manage your Redpanda cluster. Helm is primarily used only to deploy the operator, and from there you would interact with the operator and/or `kubectl` in order to modify your Redpanda cluster. `redpanda-operator` is released alongside Redpanda (see the latest release [here](https://github.com/redpanda-data/redpanda/releases)). For now, much of our site's helm documentation focuses on the `redpanda-operator` (see [here](https://docs.redpanda.com/docs/quickstart/kubernetes-qs-cloud/).

We will continue to provide both projects going forward, so feel free to use which ever you prefer! But keep in mind that they are separate, incompatible projects, and instructions for one will not apply to the other. A good rule of thumb is that if you see mention of the word "operator" in some resource, it's not related to this `repdanda` helm chart. This helm chart has no operator and no custom resource definitions (CRDs).

## v2 in development

We are making major changes to this chart to provide a number of new features and simplify how features can be managed from within `values.yaml`. For now this means we will have two versions of this chart. This version (main branch) is the older one, and the next version is in the [v2 branch](https://github.com/redpanda-data/helm-charts/tree/v2). Please see [this project](https://github.com/redpanda-data/helm-charts/projects/1) for details on what is being worked on and planned, and feel free to provide any feedback by [opening a new issue](https://github.com/redpanda-data/helm-charts/issues/new).

We also plan to update our docs site with details on this helm chart. Right now much of the helm-related documentation focuses on `redpanda-operator`, but that will change as we finalize the initial v2 release of this chart.

Once we are comfortable with the v2 branch

## Overview 

This is the Helm Chart for [Redpanda](https://redpanda.com). It provides the ability to set up a multi node redpanda cluster with optional:

- TLS 
- TLS and SASL 
- External access.

The chart uses a layered values.yaml files to demonstrate the different permutations of configuration. 

## Requirements

* Helm >= 3.0
* Kubernetes >= 1.18
* Cert-Manager
* MetalLB (optional)

## Installation

First, clone this repo:

```
git clone https://github.com/redpanda-data/helm-charts.git
cd helm-charts/redpanda
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

kind create cluster --name redpanda --config=tri-node-config.yaml

kubectl config current-context

kubectl get nodes -o wide
```

If you intend to install TLS, then you are required to install [cert-manager](https://cert-manager.io/docs).

Cert-manager installation information can be found [here](https://cert-manager.io/docs/installation/)

Use this command to install cert-manager with Helm:

```sh
helm repo add jetstack https://charts.jetstack.io && \
helm repo update && \
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.8.0 \
  --set installCRDs=true
```

## Method 1: No TLS and No SASL

If no TLS or SASL is required, simply invoke:

```sh
helm install redpanda . -n redpanda --create-namespace 
```

This can be followed by a basic helm test that will change corresponding to your installation.

```sh
helm test redpanda -n redpanda
```
The output should indicate the success of the tests and some example commands you can try that do not require either SASL or TLS configuration.

## Method 2: TLS but No SASL

If TLS is required, invoke the following:

```sh
helm install redpanda . -f values_add_tls.yaml -n redpanda --create-namespace
```

When this command is invoked the self signed issuers are created for each service by default following the model of the Redpanda [operator](https://www.redpanda.com). These self-signed Issuers are used to create per service root Issuers whome accordingly create keys and ca certs separately for each service. This behaviour is for example only. A later example will demonstrate how to utilise your own custom Issuer.

The following test command will ensure that the kafka api, panda proxy and schema registy all have basic access using TLS.

```sh
helm test redpanda -n redpanda
``` 

Note that specification of the subdomain in the configuration is automatically detected and included in the SAN. For example the following amendment to the `kafka_api`

```
    kafka_api:
      - name: kafka 
        port: 9092
        external:
        	enabled: true
        	subdomain: "streaming.rockdata.io"
```

Will generate the external nodeport and the following entries in the certificate

```
spec:                                                                         
  commonName: redpanda-kafka-cert
dnsNames:
  - redpanda-cluster.redpanda.redpanda.svc.cluster.local                                                                                                                      
  - '*.redpanda-cluster.redpanda.redpanda.svc.cluster.local'                                                                                                           
  - redpanda.redpanda.svc.cluster.local                                                                                                     
  - '*.redpanda.redpanda.svc.cluster.local'
  - streaming.rockdata.io
  - '*.streaming.rockdata.io'
```

Whereby the generated nodeport can be accessed with TLS via `redpanda-<x>.streaming.rockdata.io` for example.


## Method 3: TLS and SASL Enabled

To further include SASL protection for your cluster the following command can be run, layering SASL configuration on top of the basic configuration and TLS configuration additively. For an extensive reference for Redpanda rpk ACL commands please visit [here](https://docs.redpanda.com/docs/reference/rpk-commands/#rpk-acl).

```sh
helm install redpanda . -f values_add_tls.yaml -f values_add_sasl.yaml -n redpanda --create-namespace
```

In this case both TLS and SASL should now be enabled with a default admin user and test password (please dont use this password for your deployment).

The installation can be further tested with the following command:

```sh
helm test redpanda -n redpanda
```

## Issuer override

The default behaviour of this chart is to create an Issuer per service by iterating the following list in the values file. NOTE: the creation of Issuers is bound to this list, not to the enablement of services (this may change in the future); therefore, an Issuer can be added by merely appending to this list e.g. `- name: myservice`.

```
certIssuers:
  - name: kafka
  - name: proxy
  - name: schema
  - name: admin
```

The certs-issuers.yaml iterates this list performing simple template substitution to generate firsta self-signed Issuer then that self-signed Issuer issues its own service based root Certificate.

The self-signed root certificate is then used to create a <release>-<service>-root-issuer. For example (A):

```sh
> kubectl get issuers -o wide

redpanda-admin-root-issuer                             True    Signing CA verified   2m20s
redpanda-admin-selfsigned-issuer                       True                          2m20s
redpanda-kafka-root-issuer                             True    Signing CA verified   2m20s
redpanda-kafka-selfsigned-issuer                       True                          2m20s
redpanda-proxy-root-issuer                             True    Signing CA verified   2m20s
redpanda-proxy-selfsigned-issuer                       True                          2m20s
redpanda-schema-root-issuer                            True    Signing CA verified   2m20s
redpanda-schema-selfsigned-issuer                      True                          2m20s
```

If required the issuer of the service cert issuers can be specified. This is possible to change per service individually if required.

For example; in this case a self-signed issuer for the rockdata.io company is generated as follows:

```sh
kubectl apply -f - << EOF
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rockdata-io-selfsigned-issuer
spec:
  selfSigned: {}
EOF
```

Therefore values yaml can be modified as follows to specifically override the kafka issuer (NOTE: it is likely that the issuerRef specification will be enriched in the future):

```
certIssuers:
  - name: kafka
    issuerRef: rockdata-io-selfsigned-issuer
  - name: proxy
  - name: schema
  - name: admin
```

This can be easily tested by using using the following:

```sh
helm install redpanda . -f values_add_custom_issuer.yaml -n redpanda
```

The following helm tests that interact with the Redpanda cluster via TLS as before should pass.

```sh
helm test redpanda -n redpanda
```

Note the creation of the custom Issuer in the output below.

```sh
> kubectl get issuers

redpanda-admin-root-issuer                             True    36m
redpanda-admin-selfsigned-issuer                       True    36m
redpanda-kafka-root-issuer                             True    36m
redpanda-proxy-root-issuer                             True    36m
redpanda-proxy-selfsigned-issuer                       True    36m
redpanda-schema-root-issuer                            True    36m                                                                 
redpanda-schema-selfsigned-issuer                      True    36m
rockdata-io-selfsigned-issuer                          True    37m
```

# External Access

## Created Services 

| Type | headless | load balanced |node ports | externally load balanced |
| :--- | :---: | :---: | :---: | :---: |
| Kafka API | y | n | y | y |
| Admin API | y | n | y | WIP |
| Schema Registry | y | y  | y | WIP |
| PandaProxy API | y | y  | y | WIP |

The chart will create the headless service as in the internal connectivity case, and can also create further services to support external connectivity:

A load-balanced ClusterIP service that is used as an entrypoint for the Pandaproxy.

A Nodeport service used to expose each API to the node's external network. Make sure that the node is externally accesible.


In addition an external load balancer can be specified - see APPENDIX 1.


## APPENDIX 1: External Load Balancing Demo

An external load balancer can be demonstrated with a local kind cluster.

In this example [MetalLB](https://metallb.org/) is utilised.

First the MetaLB dependency needs to be installed to the cluster (this could be added as a conditional dependency to the chart). In this case:

```sh
# TODO - add the other method of achieving this   
NODES=$(kubectl get nodes -o json | jq -r '.items[].status.addresses | select(.[].address | startswith("redpanda-worker")) | .[] | select(.type == "InternalIP").address')
SUBNET=$(echo "$NODES" | head -n1 | cut -d. -f 1,2).255
ADDRESSES="$SUBNET.1-$SUBNET.254"
   
# Install metallb
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install \
  --create-namespace \
  --namespace metallb-system \
  metallb bitnami/metallb \
  -f metallb-values.yaml \
  --set configInline.address-pools[0].addresses[0]="$ADDRESSES"
```

```
kubectl apply -f - << EOF
configInline:
  address-pools:
    - name: default
      protocol: layer2
      addresses:
        - 172.18.255.1-172.18.255.250
EOF 
```

The Redpanda cluster can then be installed via the helm chart. In this case with the demonstration load balancer values file layered onto the base values.yaml.

```sh
helm install redpanda . -f values_add_lb.yaml -n redpanda
```

For a local [kind](https://kind.sigs.k8s.io/) development environment adjust your /etc/hosts of your host machine to access the redpanda workers on your kind cluster.

```
172.18.255.2    redpanda-0.redpanda.kind
172.18.255.1    redpanda-1.redpanda.kind
172.18.255.3    redpanda-2.redpanda.kind
```

e.g.

```sh
rpk --brokers redpanda-0.redpanda.kind:9092 cluster info
```

## Troubleshooting

TBD


