#!/usr/bin/env bash
set -xeuo pipefail

DOMAIN=${1-}

rm tls.crt tls.csr tls.key \
  ca.crt ca.key cert.conf csr.conf \
  ca.srl || true

# if domain was clean, we wanted to just clean everything.
if [ "${DOMAIN}" = "clean" ]; then
  rm external-tls-secret.yaml internal-tls-secret.yaml || true
  exit 0
fi

# check here if we have openssl, exit immediately if not
if ! command -v openssl &> /dev/null
then
  echo "cannot run without 'openssl' installed"
  exit 1
fi

# assume we are creating an external tls secret first
SECRET_NAME=external-tls-secret
ALT_NAMES="DNS.1 = ${DOMAIN}
DNS.2 = *.${DOMAIN}
"

# if we do not have a domain, then assume we are creating an internal tls secret
if [ -z ${DOMAIN} ]; then
   echo "internal tls requested"
   SECRET_NAME=internal-tls-secret
   ALT_NAMES='
DNS.1 = "*.svc.cluster.local"
DNS.2 = "svc.cluster.local"
DNS.3 = "*.svc.cluster.local."
DNS.4 = "svc.cluster.local."
   '
fi

# create CA crt here
openssl req -x509 -sha256 -days 365 -nodes -newkey rsa:2048 -subj "/C=US" -keyout ca.key -out ca.crt
# create TLS key
openssl genrsa -out tls.key 2048

# creating Certificate Signing Request (CSR) configuration file
cat > csr.conf <<EOF
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = US

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
${ALT_NAMES}

EOF

# with the CSR create a TLS CSR
openssl req -new -key tls.key -out tls.csr -config csr.conf

# create a Certificate Configuration file
cat > cert.conf <<EOF

authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
${ALT_NAMES}

EOF

# Use the configuration file to create the tls.crt and sign it with the ca.crt
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 365 -sha256 -extfile cert.conf

# Create a secret object store to file at first
kubectl create secret generic ${SECRET_NAME} \
--from-file=ca.crt=ca.crt \
--from-file=tls.crt=tls.crt \
--from-file=tls.key=tls.key \
--dry-run=client -o yaml > ${SECRET_NAME}.yaml.tmp

kubectl annotate -f ${SECRET_NAME}.yaml.tmp \
helm.sh/hook-delete-policy="before-hook-creation" \
helm.sh/hook="pre-install,pre-upgrade" \
helm.sh/hook-weight="-100" \
--local --dry-run=none -o yaml > ${SECRET_NAME}.yaml

rm ${SECRET_NAME}.yaml.tmp

echo ${SECRET_NAME}
