#!/bin/bash

set -euo pipefail

CHART=$1
CHARTFILE="charts/${CHART}/Chart.yaml"
CHARTVALUES="charts/${CHART}/values.yaml"
GITHUB_OUTPUT="${GITHUB_OUTPUT:-/dev/null}"

get_chart_version() {
    VERSION=$(yq -r .version "${CHARTFILE}")
}

bump_patch_version() {
    NEW_VERSION=$(awk -F. '{$NF++;print}' OFS=. <<< "${VERSION}")
}

get_chart_appversion() {
    APPVERSION=$(yq -r .appVersion "${CHARTFILE}")
}

get_chart_app_constraint() {
    CONSTRAINT=$(yq -r .appVersion "${CHARTFILE}" | awk '{patch=split($1,version,".");version[patch]++; print "~"version[1]}' OFS=.)
}

get_repo() {
    REPO=$(yq -r .image.repository "${CHARTVALUES}")
}

get_chart_version
get_chart_appversion
get_chart_app_constraint
get_repo

LATEST=$(docker-tag-list -r "${REPO}" -c "${CONSTRAINT}" --latest)
if [[ "${APPVERSION}" == "${LATEST}" ]]; then
    echo "nothing to update"
    echo 'CHANGED=false' >> "${GITHUB_OUTPUT}"
    exit 0
fi
echo 'CHANGED=true' >> "${GITHUB_OUTPUT}"

# build the commit text
echo "update ${CHART} appVersion from ${APPVERSION} to ${LATEST}"
echo "TITLE=update ${CHART} appVersion from ${APPVERSION} to ${LATEST}" >> "${GITHUB_OUTPUT}"
echo
if [[ "$CHART" == "redpanda" ]]; then
    BODY_FILE=$(mktemp)
    gh --repo redpanda-data/redpanda release view "${LATEST}" --json body -t '{{ .body }}' >> "${BODY_FILE}"
    echo "BODY_FILE=${BODY_FILE}" >> "${GITHUB_OUTPUT}"
fi

bump_patch_version
sed -i "s@^version: .*\$@version: ${NEW_VERSION}@" "${CHARTFILE}"
sed -i "s@^appVersion: .*\$@appVersion: ${LATEST}@" "${CHARTFILE}"
sed -i "s@${REPO}:${APPVERSION}\$@${REPO}:${LATEST}@" "${CHARTFILE}"
