#!/usr/bin/env bash
set -euo pipefail

# TODO the updated-sasl-users.txt was changed so admin password did not change.
# This is to allow the <release>-console-test to pass since console does not auto
# reload on sasl password changes. Once this is fixed, the updated-sasl-users.txt should
# be changed so that full sasl changes can be observed and validated.

SECRET_NAME=${1-"redpanda-license"}
LICENSE_KEY=${2-""}

# Create a secret object with updated user sasl data
kubectl create secret generic ${SECRET_NAME} \
  --from-literal=a-license-key=${LICENSE_KEY} \
  --dry-run=client -o yaml > ${SECRET_NAME}.yaml.tmp

# Annotate with before-hook-creation
kubectl annotate -f ${SECRET_NAME}.yaml.tmp \
helm.sh/hook-delete-policy="before-hook-creation" \
--local --dry-run=client -o yaml > ${SECRET_NAME}.yaml

rm ${SECRET_NAME}.yaml.tmp

mv ${SECRET_NAME}.yaml ./charts/redpanda/templates/

