#!/usr/bin/env bash
set -xeuo pipefail

# Create a secret object store to file at first
kubectl create secret generic redpanda-license \
--from-literal=license-key=$1 \
--dry-run=client -o yaml > redpanda-license.yaml.tmp

kubectl annotate -f redpanda-license.yaml.tmp \
helm.sh/hook-delete-policy="before-hook-creation" \
helm.sh/hook="pre-install" \
helm.sh/hook-weight="-100" \
--local --dry-run=none -o yaml > redpanda-license.yaml

rm redpanda-license.yaml.tmp