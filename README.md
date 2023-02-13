# Redpanda Kubernetes Helm Charts

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![Release Charts](https://github.com/Redpanda/helm-charts/workflows/Release%20Charts/badge.svg?branch=main) [![Releases downloads](https://img.shields.io/github/downloads/Redpanda/helm-charts/total.svg)](https://github.com/Redpanda/helm-charts/releases)

This functionality is in beta and is subject to change. The code is provided as-is with no warranties. Beta features are not subject to the support SLA of official GA features.

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```console
helm repo add redpanda https://charts.redpanda.com/
helm repo update
```

You can then run `helm search repo redpanda` to see the charts.

You can install the chart by running

```console
helm install redpanda redpanda/redpanda \
    --namespace redpanda \
    --create-namespace
```

## Contributing

The source code of all [Redpanda](https://github.com/redpanda-data/) community [Helm](https://github.com/redpanda-data/helm-charts/) charts can be found on Github: <https://github.com/redpanda-data/helm-charts/>

<!-- Keep full URL links to repo files because this README syncs from main to gh-pages.  -->
We'd love to have you contribute! Please refer to our [contribution guidelines](https://github.com/Redpanda/helm-charts/blob/main/CONTRIBUTING.md) for details.

## License

<!-- Keep full URL links to repo files because this README syncs from main to gh-pages.  -->
[Apache 2.0 License](https://github.com/Redpanda/helm-charts/blob/main/LICENSE).

## Helm charts build status

![Release Charts](https://github.com/Redpanda/helm-charts/workflows/Release%20Charts/badge.svg?branch=main)
