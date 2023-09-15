#!/bin/bash

set -euo pipefail

CHART=$1
CHARTFILE="charts/${CHART}/Chart.yaml"
CHARTVALUES="charts/${CHART}/values.yaml"

get_chart_version() {
    VERSION=$(awk '/^version:/{print $2}' "${CHARTFILE}")
}

bump_patch_version() {
    NEW_VERSION=$(awk -F. '{$NF++;print}' OFS=. <<< "${VERSION}")
}

get_chart_appversion() {
    APPVERSION=$(awk '/^appVersion:/{print $2}' "${CHARTFILE}")
}

get_chart_app_constraint() {
    CONSTRAINT=$(awk '/^appVersion:/{patch=split($2,version,".");version[patch]++; print "~"version[1],version[2]}' OFS=. "${CHARTFILE}")
}

get_repo() {
    REPO=$(awk '/^  repository:/{print $2}' "${CHARTVALUES}")
}

get_chart_version
get_chart_appversion
get_chart_app_constraint
get_repo

LATEST=$(docker-tag-list -r "${REPO}" -c "${CONSTRAINT}" --latest)
if [[ "${APPVERSION}" == "${LATEST}" ]]; then
    echo "nothing to update"
    exit 0
fi

echo "update ${CHART} appVersion from ${APPVERSION} to ${LATEST}"
bump_patch_version
sed -i "s/^version: .*$/version: ${NEW_VERSION}/" "${CHARTFILE}"
sed -i "s/^appVersion: .*$/appVersion: ${LATEST}/" "${CHARTFILE}"
sed -i "s/${APPVERSION}/${LATEST}/" "${CHARTFILE}"