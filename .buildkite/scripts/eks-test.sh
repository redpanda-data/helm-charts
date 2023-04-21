#!/bin/env bash

set -xeuo pipefail

bash -O extglob -c "rm -v charts/redpanda/ci/!(01-)*"

echo ct install --config .github/ct.yaml --upgrade