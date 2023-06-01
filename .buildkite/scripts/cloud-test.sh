#!/bin/env bash

set -xeuo pipefail

set


PATH="$(realpath .local/bin):${PATH}"
bash -O extglob -c "rm -v charts/redpanda/ci/!(2)[0-9]-*"

envsubst < ./charts/redpanda/ci/21-eks-tiered-storage-with-creds-values.yaml.tpl > ./charts/redpanda/ci/21-eks-tiered-storage-with-creds-values.yaml

ct install --config .github/ct.yaml --upgrade --skip-missing-values | sed 's/>>> /--- /'



aws() {
    echo '--- testing that there is data in the s3 bucket'
    if (aws s3 ls "s3://${TEST_BUCKET}" --recursive --summarize | grep 'Total Objects: 0'); then
        echo "0 Objects in the bucket. Cloud-storage failed."
        exit 1
    fi
}

gke() {
  echo '--- testing that there is data in the gcloud bucket'
  if (gsutil ls -lR gs://${TEST_BUCKET} | tail -n 1 | grep 'TOTAL: 0 objects'); then
        echo "0 objects in the bucket. Cloud-storage failed."
        exit 1
  fi
}

${CLOUD_PROVIDER}
