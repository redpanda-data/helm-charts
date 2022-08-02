## External access via NodePort

NodePort is the default method for providing external access for the helm chart, as it can work without requiring any additional installation or service. External access via NodePort requires that worker nodes in your cluster have an externally accessible IP address. If this is not the case, then you should use another method such as LoadBalancer or Ingress.

The example commands below use the namespace `redpanda-ns`. Either use this or replace it with whatever namespace you want to run the Redpanda cluster within.

### Initial steps

External access via a NodePort service is enabled by default, but there are some helpful verification tasks listed below.

Most of these tasks are optional, but there is one required task that will almost always need to be performed before external access works: you must ensure the external host names being advertised by Redpanda are resolvable on your system. More details on this step can be found starting [here](#verify-advertised-listeners).

These are steps that should be performed regardless of your environment:

1. [Verify the number of replicas or your environment](#number-of-replicas)
2. [Verify the external hostname](#custom-external-hostname)
3. [Determine node IP addresses](#determine-node-ip-addresses)
4. [Match brokers to nodes](#match-broker-to-node)
5. [Verify advertised listeners](#verify-advertised-listeners)
6. [Make hostnames resolvable](#ensure-external-hostnames-are-resolvable)
7. [Verify NodePort configuration](#verify-nodeport-configuration)
8. [Verify external client connectivity](#verify-external-client-connectivity)

#### Number of replicas

These instructions assume you have 3 replicas, one pod (and by extension Redpanda broker) per node. 3 is the default replica value within the statefulset configuration, `statefulset.replicas`.

We recommend setting `statefulset.replicas` to the number of nodes in the cluster so there is one pod per node, especially if you plan to use a NodePort service for external access. This is because the NodePort service will open the same port across all nodes that have a Redpanda broker available, and multiple brokers attempting to use the same port on a single node will cause issues when the NodePort service balances traffic across the brokers. In this scenario, external clients would have inconsistent access to specific brokers on this node.

#### Custom external hostname

The default externally accessible Redpanda broker hostnames follow the pattern `redpanda-x.local` (where `x` is the broker ID starting with 0). These values are derived from a combination of the chart name and the external domain. The default chart name is `redpanda` (from name in `Chart.yaml`), and this can be overridden with setting `nameOverride` in `values.yaml`. The external domain is at `external.domain` in `values.yaml`.

So if, for example, you wanted to have the custom external address of `broker-x.redpanda.com`, you would use the following values:

```yaml
nameOverride: broker
external:
  domain: redpanda.com
```

#### Determine node IP addresses

The next step is to determine the IP addresses for each worker node in your cluster that will run a Redpanda broker:

```
> kubectl get nodes -o wide
NAME           STATUS   ROLES           AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
minikube       Ready    control-plane   25h   v1.24.1   192.168.49.2   <none>        Ubuntu 20.04.4 LTS   5.15.0-41-generic   docker://20.10.17
minikube-m02   Ready    <none>          25h   v1.24.1   192.168.49.3   <none>        Ubuntu 20.04.4 LTS   5.15.0-41-generic   docker://20.10.17
minikube-m03   Ready    <none>          25h   v1.24.1   192.168.49.4   <none>        Ubuntu 20.04.4 LTS   5.15.0-41-generic   docker://20.10.17
minikube-m04   Ready    <none>          25h   v1.24.1   192.168.49.5   <none>        Ubuntu 20.04.4 LTS   5.15.0-41-generic   docker://20.10.17
```

In the above example there are three worker nodes. We will need to match the IP addresses of these nodes with the Redpanda broker that ends up running on that node due to the pod anti-affinity configuration.

#### Match broker to node

Once you have deployed the Redpanda cluster (see [this section](../../README.md#redpanda) for details if needed), you can see which brokers are running on which node with the following command:

```
> kubectl -n redpanda-ns get pods -o wide
NAME         READY   STATUS    RESTARTS   AGE   IP            NODE           NOMINATED NODE   READINESS GATES
redpanda-0   1/1     Running   0          14m   10.244.1.11   minikube-m02   <none>           <none>
redpanda-1   1/1     Running   0          14m   10.244.2.11   minikube-m03   <none>           <none>
redpanda-2   1/1     Running   0          14m   10.244.3.11   minikube-m04   <none>           <none>
```

In the output above, redpanda pods 0 through 2 are deployed sequentially on nodes m02 through m04. While this example shows pods and nodes lining up sequentially, this is not always the case! Also the node names shown above are related to a local minikube install, and your names will likely be different depending on your environment.

#### Ensure external hostnames are resolvable

You must ensure the hostnames being advertised to external clients by the Redpanda brokers are resolvable by your system. The steps for this task varies depending on your environment, but one way to do this is to add entries to `/etc/hosts` for each worker node running a Redpanda pod. Given the above details you would add the following lines:

```
192.168.49.3 redpanda-0.local # minikube-m02
192.168.49.4 redpanda-1.local # minikube-m03
192.168.49.5 redpanda-2.local # minikube-m04
```

#### Verify advertised listeners

By default there are two Kafka listeners configured in `values.yaml`: internal and external. Both have an advertised listener value that is given to Redpanda at startup.

To verify the __external__ Kafka advertised listener, you can use a locally installed Redpanda CLI:

```
> rpk cluster info --brokers redpanda-0.local:31092
BROKERS
=======
ID    HOST              PORT
0*    redpanda-0.local  31092
1     redpanda-1.local  31092
2     redpanda-2.local  31092
```

For instructions on installing `rpk`, see these instructions for [Linux](https://docs.redpanda.com/docs/quickstart/quick-start-linux/#install-and-run-redpanda) or [macOS](https://docs.redpanda.com/docs/quickstart/quick-start-macos/#installing-rpk).

You may also want to verify the __internal__ Kafka advertised listener, which can be done with the rpk CLI included within any of the Redpanda brokers:

```
> kubectl -n redpanda-ns exec -it redpanda-0 -c redpanda -- rpk cluster info
BROKERS
=======
ID    HOST                                                PORT
0*    redpanda-0.redpanda.redpanda-ns.svc.cluster.local.  9093
1     redpanda-1.redpanda.redpanda-ns.svc.cluster.local.  9093
2     redpanda-2.redpanda.redpanda-ns.svc.cluster.local.  9093
```

The output shows the FQDN for each broker on the internal cluster network. You can also access each broker internally with its short hostname `redpanda-x` (where `x` is the broker ID).

#### Verify NodePort configuration

The NodePort service is named `redpanda-external`, and you can get a summary of its configuration with the following command:

```
> kubectl -n redpanda-ns get svc redpanda-external
NAME                TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                       AGE
redpanda-external   NodePort   10.100.47.125   <none>        9644:31644/TCP,9092:31092/TCP,8082:30082/TCP,8081:30081/TCP   9m35s
```

We can see there are four ports opened on each node where a Redpanda broker is running for the following services: admin API (9644), kafka (9092), HTTP proxy (8082), and schema registry (8081). It is fine to have different ports listed; just make sure to use the correct external ports for your external clients. In the above example, we see the redpanda-external service has the node port 31092 forwarding to the external Kafka API listener at port 9092.

#### Verify external client connectivity

Now we can use the `rpk` CLI as an external client to create a topic, and then produce/consume events.

First create a topic:

```
> rpk topic create external-access -r 3 --brokers redpanda-0.local:31092
TOPIC            STATUS
external-access  OK
```

Then in one terminal start the folllowing client which will consume messages to the previously created topic:
```
> rpk topic consume external-access --brokers redpanda-0.local:31092
```

Run the producer in another terminal:

```
> rpk topic consume external-access --brokers redpanda-0.local:31092
```

In the above terminal where the producer is running, type some text and press enter. You should see an event printed on the consumer side similar to the following:

```
{
  "topic": "external-access",
  "value": "test",
  "timestamp": 1658286544200,
  "partition": 0,
  "offset": 1
}
```

### Possible issues and resolutions

If you've followed along without issue to this point, you have a multi-node Redpanda cluster running with external access enabled via the NodePort service. However there are some common issues that you could end up facing, and this section covers some of those issues along with their resolution.

#### Unable to use the rpk cluster health command

The Redpanda documentation points to a useful command `rpk cluster health`, and you may want to try it out. But depending on your configuration, you may be the following error:

```
> rpk cluster health --api-urls redpanda-0.local:9644
unable to request cluster health: request   failed: Not Found, body: "{\"message\": \"Not found\", \"code\": 404}"
```

This is due to using a Redpanda version that doesn't support the cluster health command (ie. `21.11.x` or older). To get the Redpanda cluster version, either check `image.tag` in `values.yaml`, or run the following command to get the version from a running cluster:

```
> kubectl -n redpanda-ns exec -it redpanda-0 -c redpanda -- rpk version
v21.11.16 (rev c0ff554)
```

In the above example, the version is older than the cluster health command and it is not available. Instead you can get some of the same details using `rpk cluster info`.

#### Unable to create a topic

When attempting to create a topic, you may run into this issue:

```
> rpk topic create topic1 -r 3 --brokers redpanda-0.local:9092
TOPIC   STATUS
topic1  NOT_CONTROLLER: This is not the correct controller for this cluster.
```

This is due to a mismatch between the brokers and the node IPs. This is an easy issue to run into, especially if you restart your cluster environment or switch between environments regularly. Verify that your host entries correctly targets the proper IP addresses for each worker node (see instructions [here](#match-broker-to-node) and [here](#ensure-external-hostnames-are-resolvable)).
