#!/usr/bin/env bash

mechanism=SCRAM-SHA-256

rp_args=(
    cluster
    config
    import
    -f /tmp/base-config/bootstrap.yaml
    --api-urls "${RP_API_URLS}"
)

if [[ "$RP_ADMIN_TLS_ENABLED" = true ]]; then
    rp_args+=(
        --admin-api-tls-enabled
        --admin-api-tls-truststore /etc/tls/certs/"${RP_ADMIN_TLS_CERT}"/ca.crt
    )
fi  
if [[ "$RP_KAFKA_TLS_ENABLED" = true ]]; then
    rp_args+=(
        --tls-enabled
        --tls-truststore /etc/tls/certs/"${RP_KAFKA_TLS_CERT}"/ca.crt
    )
fi  
if [[ "$RP_SASL_ENABLED" = true ]]; then
    parts=(${RP_BOOTSTRAP_USER//:/ })
    rp_user="${parts[0]}"
    rp_pw="${parts[1]}"

    rp_args+=(
        --user "${rp_user}"
        --password "${rp_pw}"
        --sasl-mechanism "${mechanism}"
    )
fi  

rpk "${rp_args[@]}"
