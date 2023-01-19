# Redpanda Helm Chart

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/redpanda-data)](https://artifacthub.io/packages/search?repo=redpanda-data)

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

For examples of using the Helm chart, see the [`examples/` directory][examples]. Each example focuses on a specific feature.

## Configuration

The Redpanda Helm chart is configured in the [`values.yaml`][values] file. To customize your deployment, you can override the default values in your own YAML file with the `--values` option or in the command line with the --set option. For example, you can do the following:

- Specify which Kubernetes components to deploy.

- Configure the deployed Kubernetes components.

To learn how to override the default values in the `values.yaml` file, see the [Helm documentation][helm].

All configuration options for the Redpanda Helm chart are documented in the [`values.yaml`][values] file.

## Contributing

If you have improvements that can be made to this Helm chart, please consider becoming a contributor.
To contribute to the Helm chart, see our [contribution guidelines][contributing].

[redpanda]: https://redpanda.com
[helm]: https://helm.sh/docs/chart_template_guide/values_files/
[values]: https://github.com/redpanda-data/helm-charts/blob/main/charts/redpanda/values.yaml
[examples]: https://github.com/redpanda-data/helm-charts/blob/main/examples/README.md
[contributing]: https://github.com/redpanda-data/helm-charts/blob/main/CONTRIBUTING.md
[kubernetes-qs-dev]: https://docs.redpanda.com/docs/quickstart/kubernetes-qs-dev/
