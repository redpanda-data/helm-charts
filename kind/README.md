# Redpanda on KinD

## TLDR

```sh
./make.sh
```

## Details

### Local Registry

There is a [local registry](https://kind.sigs.k8s.io/docs/user/local-registry/).

The registry can be used like this.

1. First we'll pull an image `docker pull vectorized/redpanda:latest`
2. Then we'll tag the image to use the local registry `docker tag vectorized/redpanda:latest localhost:5000/redpanda:latest`
3. Then we'll push it to the registry `docker push localhost:5000/redpanda:latest`
4. And now we can use the image `kubectl create deployment hello-server --image=localhost:5000/redpanda:latest`

If you build your own image and tag it like `localhost:5000/redpanda:latest` and then use it in kubernetes as `localhost:5000/redpanda:latest`.

### LoadBalancer

[metallb](https://metallb.universe.tf/) is used.

This works on Ubuntu - YMMV.

Useful info for Mac: https://www.thehumblelab.com/kind-and-metallb-on-mac/

I don't know of a non-intrusive way to get DNS lookups from outside the cluster.

For now modifying `/etc/hosts` is suggested, installation will describe how.

### Redpanda

3-node cluster.

Each node has a PVC.

Testing a rolling restart:

```sh
kubectl -n redpanda rollout restart statefulset redpanda
```
