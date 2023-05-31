#!/usr/bin/env bash
set -xeuo pipefail

# TODO the updated-sasl-users.txt was changed so admin password did not change.
# This is to allow the <release>-console-test to pass since console does not auto
# reload on sasl password changes. Once this is fixed, the updated-sasl-users.txt should
# be changed so that full sasl changes can be observed and validated.

SECRET_NAME=${1-"some-users"}

# Create a secret object with updated user sasl data
kubectl create secret generic ${SECRET_NAME} \
--from-file=.github/updated-sasl-users.txt \
--dry-run=client -o yaml > ${SECRET_NAME}-updated.yaml.tmp

# Updated sasl secrete starts with annotations for post-install hooks
kubectl annotate -f ${SECRET_NAME}-updated.yaml.tmp \
helm.sh/hook-delete-policy="before-hook-creation" \
helm.sh/hook="post-install,post-upgrade" \
helm.sh/hook-weight="100" \
--local --dry-run=client -o yaml > ${SECRET_NAME}-updated.yaml

rm ${SECRET_NAME}-updated.yaml.tmp

echo created file ${SECRET_NAME}-updated.yaml
