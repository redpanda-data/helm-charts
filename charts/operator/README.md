# Redpanda Operator

![Version: 0.4.11](https://img.shields.io/badge/Version-0.4.11-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.1.10-23.2.18](https://img.shields.io/badge/AppVersion-v2.1.10--23.2.18-informational?style=flat-square)

## Installation

### Prerequisite

To deploy operator with webhooks (enabled by default) please install
cert manager. Please follow
[the installation guide](https://cert-manager.io/docs/installation/)

The cert manager needs around 1 minute to be ready. The helm chart
will create Issuer and Certificate custom resource. The
webhook of cert-manager will prevent from creating mentioned
resources. To verify that cert manager is ready please follow
[the verifying the installation](https://cert-manager.io/docs/installation/kubernetes/#verifying-the-installation)

The operator by default exposes metrics endpoint. By leveraging prometheus
operator ServiceMonitor custom resource metrics can be automatically
discovered.

1. Install Redpanda operator CRDs:

```sh
kubectl kustomize https://github.com/redpanda-data/redpanda-operator//src/go/k8s/config/crd | kubectl apply -f -
```

> The CRDs are decoupled from helm chart, so that helm release can be managed with fewer privileges.
> The CRDs need to be installed by someone with cluster-level privileges, but once installed the
> user only needs access to a namespace.

### Helm installation

1. Install the Redpanda operator:

```sh
helm repo add redpanda https://charts.redpanda.com
helm repo update redpanda
helm install --namespace redpanda --create-namespace redpanda-operator redpanda/operator
```

Other instruction will be visible after installation.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| additionalCmdFlags | list | `[]` | Allows setting additional flags to the operator command |
| affinity | object | `{}` | Allows to specify affinity for Redpanda Operator PODs |
| clusterDomain | string | `"cluster.local"` |  |
| commonLabels | string | `nil` | Allows to assign labels to the resources created by this helm chart |
| config.apiVersion | string | `"controller-runtime.sigs.k8s.io/v1alpha1"` |  |
| config.health.healthProbeBindAddress | string | `":8081"` |  |
| config.kind | string | `"ControllerManagerConfig"` |  |
| config.leaderElection.leaderElect | bool | `true` |  |
| config.leaderElection.resourceName | string | `"aa9fc693.vectorized.io"` |  |
| config.metrics.bindAddress | string | `"127.0.0.1:8080"` |  |
| config.webhook.port | int | `9443` |  |
| configurator.pullPolicy | string | `"IfNotPresent"` |  |
| configurator.repository | string | `"docker.redpanda.com/redpandadata/configurator"` | Repository that Redpanda configurator image is available |
| fullnameOverride | string | `""` | Override the fully qualified app name |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for kube-rbac-proxy image |
| image.repository | string | `"docker.redpanda.com/redpandadata/redpanda-operator"` | Repository in which the kube-rbac-proxy image is available |
| imagePullSecrets | list | `[]` | Redpanda Operator container registry pullSecret (ex: specify docker registry credentials) |
| kubeRbacProxy.image.pullPolicy | string | `"IfNotPresent"` |  |
| kubeRbacProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` |  |
| kubeRbacProxy.image.tag | string | `"v0.14.0"` |  |
| logLevel | string | `"info"` | Set Redpanda Operator log level (debug, info, error, panic, fatal) |
| monitoring | object | `{"deployPrometheusKubeStack":false,"enabled":false}` | Add service monitor to the deployment |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Redpanda Operator on specific nodes |
| podAnnotations | object | `{}` | Allows setting additional annotations for Redpanda Operator PODs |
| podLabels | object | `{}` | Allows setting additional labels for Redpanda Operator PODs |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| rbac.createAdditionalControllerCRs | bool | `false` | Enable to create additional rbac cluster roles needed to run additional-controllers, set to true to opt-in |
| rbac.createRPKBundleCRs | bool | `false` | Specified to create additional rbac cluster roles needed when you will set 'rbac.enabled' to the Redpanda Spec (See redpanda chart values file) |
| replicaCount | int | `1` | Number of instances of Redpanda Operator |
| resources | object | `{}` | Set resources requests/limits for Redpanda Operator PODs |
| scope | string | `"Namespace"` | change the scope and therefore the resource the controller will manage only "Cluster" and "Namespace" supported |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set name is generated using the fullname template |
| tolerations | list | `[]` | Allows to schedule Redpanda Operator on tainted nodes |
| webhook.enabled | bool | `false` |  |
| webhookSecretName | string | `"webhook-server-cert"` |  |
