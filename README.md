# Kubernetes Helm Chart

***Status: Early Access***

This is the Helm Chart for [Redpanda](https://vectorized.io). A Helm Chart is a collection of files used to describe a set of Kubernetes resources, and may be used to deploy and manage Redpanda in a cloud such as AWS or GCP.

For AWS see the [Redpanda EKS documentation](eks/README.md).

### Local Installation

For development purposes, it can be installed locally on Linux or Mac OS.

* The containers Kubernetes uses are typically Docker ones, so install that first: https://docs.docker.com/get-docker/

* To create a Kubernetes cluster, you need a tool such as `kind`. The install instructions are here: https://kind.sigs.k8s.io/. (minikube is also popular but it is only good for single-node clusters.)

* To install the Redpanda Helm Chart you will need `Helm`: https://helm.sh/docs/intro/install/

* To run extra commands against the cluster you need `kubectl`: https://kubernetes.io/docs/tasks/tools/install-kubectl/

* Clone this repository (`helm-charts`).

### Local Usage

* Create the cluster:

  ```
  $ kind create cluster
  ```

* Install the Helm Chart:

  ```
  $ helm install --namespace redpanda --create-namespace redpanda ./redpanda
  ```

  This will read the Helm Chart in the redpanda folder and deploy the Redpanda Docker image (the latest from Docker Hub) to pods. For now it just creates a basic 3 node Redpanda cluster. The `redpanda` namespace will be created if it does not already exist.

* At this stage you will get further instructions in the terminal about how to run `rpk` in the cluster.

### Development

* Show what Helm has currently deployed (remembering to use the `redpanda` namespace as usual):

  ```
  helm -n redpanda list
  ```

* Show how many replicas in the cluster were successful in starting up and are available:

  ```
  kubectl -n redpanda get statefulsets
  ```

* Show the pods (groups of containers) that are running:

  ```
  kubectl -n redpanda get pods
  ```

   You will see pod names, e.g. `redpanda-0`,  `redpanda-1`, `redpanda-2`.

* To see the log for Redpanda on one of the pods, run:

  ```
  kubectl -n redpanda-0 logs <pod-name>
  ```

* For troubleshooting, this command is also useful:

  ```
  kubectl -n redpanda-0 describe pod <pod-name>
  ```

* To log into a pod:

  ```
  kubectl -n redpanda exec --stdin --tty <pod-name> /bin/bash
  ```

* To clean up, run:

  ```
  helm -n redpanda uninstall redpanda
  ```

  The cluster will still be running, just without Redpanda.

* To experiment, edit the `redpanda/values.yaml` file, and run the install again:

  ```
  helm install redpanda ./redpanda
  ```
