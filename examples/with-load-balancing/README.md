https://github.com/bitnami/charts/blob/master/bitnami/kafka/values.yaml#L698-L710
https://kubernetes.io/docs/concepts/services-networking/service/#type-loadbalancer
https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip

## External Load Balancing

This examples demonstrates load balancing in a local cluster via [MetalLB](https://metallb.org/).

### Create MetalLB config file

The first step is to find the available address pools so MetalLB can appropriately allocate IP addresses (see [this link](https://metallb.org/concepts/#address-allocation) for details).

```sh
NODES=$(kubectl describe node | grep InternalIP | awk '{print $2}')
kubectl get nodes -o wide | awk -v OFS='\t' '{print $6, $1}'
```

First, the MetaLB dependency needs to be installed in the cluster.

```sh
NODES=$(kubectl get nodes -o json | jq -r '.items[].status.addresses | select(.[].address | startswith("redpanda-worker")) | .[] | select(.type == "InternalIP").address')
SUBNET=$(echo "$NODES" | head -n1 | cut -d. -f 1,2).255
ADDRESSES="$SUBNET.1-$SUBNET.254"
```

Create the file `metallb-values.yaml` with the following contents:

```yaml
configInline:
  address-pools:
  - name: default
    protocol: layer2
    addresses:
    # replace the below line with the value of $ADDRESSES
    - 172.18.255.1-172.18.255.250
```

### Install metallb

```sh
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install -n metallb-system metallb bitnami/metallb -f metallb-values.yaml --create-namespace
```

### Configure and install Redpanda

Modify the `loadBalancer` section of values.yaml:

```yaml
loadBalancer:
  enabled: true
  # The parent zone for DNS entries.
  # See https://github.com/kubernetes-sigs/external-dns
  parentZone: redpanda.local
  # Additional annotations to apply to the created LoadBalancer services.
  # annotations:
    # For example:
    # cloud.google.com/load-balancer-type: "Internal"
    # service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

Then start redpanda:

```sh
helm -n redpanda install redpanda redpanda --create-namespace
```

For a local development environment running on Kind or Minikube, adjust your host system's /etc/hosts file to access the redpanda workers on your cluster:

```sh
172.18.255.2 redpanda-0.redpanda.kind
172.18.255.1 redpanda-1.redpanda.kind
172.18.255.3 redpanda-2.redpanda.kind
```

e.g.

```sh
rpk --brokers redpanda-0.redpanda.kind:9092 cluster info
```
