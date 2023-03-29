# Redpanda Helm Chart

[![Nightly Test](https://github.com/redpanda-data/helm-charts/actions/workflows/nightly.yaml/badge.svg)](https://github.com/redpanda-data/helm-charts/actions/workflows/nightly.yaml) [![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/redpanda-data)](https://artifacthub.io/packages/search?repo=redpanda-data)

The Redpanda Helm chart deploys a Redpanda cluster in Kubernetes, and provides the following features:

- Schema registry (enabled by default)
- REST (aka PandaProxy, enabled by default)
- TLS
- SASL
- External access

## Requirements

- Helm version 3.6.0 or later
- Kubernetes version 1.21.0 or later
- Cert-manager 1.9.0 or later (required for TLS support only)

## Installation

To get started, see the [Redpanda documentation][kubernetes-qs-dev].

## Configuration

The Redpanda Helm chart is configured in the [`values.yaml`][values] file. To customize your deployment, you can override the default values in your own YAML file with the `--values` option or in the command line with the --set option. For example, you can do the following:

- Specify which Kubernetes components to deploy.

- Configure the deployed Kubernetes components.

To learn how to override the default values in the `values.yaml` file, see the [Helm documentation][helm].

All configuration options for the Redpanda Helm chart are documented in the [`values.yaml`][values] file.

## Upgrading Chart

```bash
helm upgrade [RELEASE_NAME] redpanda/redpanda
```

### From 2.6.x onwards

In order to enable dedicated persistent volume for tiered storage cache, the `storage.tieredStoragePersistentVolume.enabled` need to be set to `true`.
The `helm upgrade` will fail with the following error.
```bash
helm upgrade --namespace redpanda redpanda/redpanda \
  --set storage.tieredStoragePersistentVolume.enabled=true \
  --set storage.tieredConfig.cloud_storage_enabled=true \
  --set storage.tieredConfig.cloud_storage_cache_directory=/some/path/for-tiered-storage
Error: UPGRADE FAILED: cannot patch "redpanda" with kind StatefulSet: StatefulSet.apps "redpanda" is invalid: spec: Forbidden: updates to statefulset spec for fields other than 'replicas', 'template', 'updateStrategy', 'persistentVolumeClaimRetentionPolicy' and 'minReadySeconds' are forbidden
```

To be able to add PersistentVolume for tiered storage cache please delete StatefulSet with cascade orphan option to leave Pods behind.
```bash
kubectl delete sts redpanda --cascade=orphan
```

The `helm upgrade` should be able to succeed, but you need to manually do rolling update starting from ordinal 0 to ordinal N.

```bash
kubectl delete pod redpanda-0
```

Please wait for deleted Pod to be restarted and become ready in order to move to next Pod.

## Contributing

If you have improvements that can be made to this Helm chart, please consider becoming a contributor.
To contribute to the Helm chart, see our [contribution guidelines][contributing].

[redpanda]: https://redpanda.com
[helm]: https://helm.sh/docs/chart_template_guide/values_files/
[values]: https://github.com/redpanda-data/helm-charts/blob/main/charts/redpanda/values.yaml
[contributing]: https://github.com/redpanda-data/helm-charts/blob/main/CONTRIBUTING.md
[kubernetes-qs-dev]: https://docs.redpanda.com/docs/quickstart/kubernetes-qs-dev/
