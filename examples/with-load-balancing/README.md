## External Load Balancing

Load balancing can be demonstrated with a local cluster. [MetalLB](https://metallb.org/) is used in this example.

First the MetaLB dependency needs to be installed to the cluster (this could be added as a conditional dependency to the chart):

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

The Redpanda cluster can then be installed via the helm chart. The load balancer values file is layered onto the base values.yaml:

```sh
helm install redpanda redpanda -f examples/with-load-balancing/values_add_lb.yaml -n redpanda
```

For a local [kind](https://kind.sigs.k8s.io/) development environment adjust your /etc/hosts of your host machine to access the redpanda workers on your kind cluster.

```sh
172.18.255.2    redpanda-0.redpanda.kind
172.18.255.1    redpanda-1.redpanda.kind
172.18.255.3    redpanda-2.redpanda.kind
```

e.g.

```sh
rpk --brokers redpanda-0.redpanda.kind:9092 cluster info
```
