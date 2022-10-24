# Redpanda Helm Chart

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/redpanda-data)](https://artifacthub.io/packages/search?repo=redpanda-data)

This Helm chart (`redpanda`) deploys a Redpanda cluster.
Once deployed, you continue to use the Helm command and override values to change and/or upgrade your Redpanda deployment.
The defaults are in [values.yaml][values].

## Requirements

- Helm v3.5.0 or newer
- Kubernetes 1.21.0 or newer
- Cert-manager 1.9.0 or newer (optional for TLS support)

## Overview

This is the Helm Chart for [Redpanda](https://redpanda.com). It provides the ability to set up a multi node redpanda cluster with the following optional features:

- Schema registry (enabled by default)
- REST (aka PandaProxy, enabled by default)
- TLS
- SASL
- External access

See the [examples folder][examples] with more details on how to use this helm chart.
Each example focuses on specific features like the ones listed above.
We recommend completing the instructions in the [60-Second Guide for Kubernetes][kubernetes-qs-dev] before continuing steps in any of these examples.

The [values.yaml][values] file is documented throughout.
Please see this file for more details.

## Installation

See the [60-Second Guide for Kubernetes][kubernetes-qs-dev]

## Contributing

If you have improvements that can be made to this Helm chart, please consider becoming a contributor.
See our [Contributing][contributing] document for more details.

[values]: https://github.com/redpanda-data/helm-charts/blob/main/redpanda/values.yaml
[examples]: https://github.com/redpanda-data/helm-charts/blob/main/examples/README.md
[contributing]: https://github.com/redpanda-data/helm-charts/blob/main/CONTRIBUTING.md
[kubernetes-qs-dev]: https://docs.redpanda.com/docs/quickstart/kubernetes-qs-dev/


