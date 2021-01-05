# Redpanda Helm Chart

***Status: Early Access***

This is the Helm Chart for [Redpanda](https://vectorized.io). 

## Requirements

* Helm >= 3.0
* Kubernetes >= 1.18

## Installation

### Local Installation

First, clone this repo:

```
git clone git@github.com:vectorizedio/helm-charts.git
```

Install the Helm Chart:

```
helm install --namespace redpanda --create-namespace redpanda ./redpanda
```
