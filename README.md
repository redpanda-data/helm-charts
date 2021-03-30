# OLD REPO: Please use the official k8s operator



```

# new k8s operator tree.


https://github.com/vectorizedio/redpanda/tree/dev/src/go/k8s



```

> NOTE: This is no longer supported.
> 
> This repo is left here for older users that want to deploy their own containers without the help from an automated operator.
> 


# Redpanda Helm Chart

![Lint and Test Charts](https://github.com/vectorizedio/helm-charts/actions/workflows/lint-test.yml/badge.svg?branch=main)

***Status: Early Access***

This is the Helm Chart for [Redpanda](https://vectorized.io). 

## Requirements

* Helm >= 3.0
* Kubernetes >= 1.18

## Installation

### Local Installation

First, clone this repo:

```sh
git clone git@github.com:vectorizedio/helm-charts.git
```

Install the Helm Chart:

```sh
helm install --namespace redpanda --create-namespace redpanda ./redpanda
```
