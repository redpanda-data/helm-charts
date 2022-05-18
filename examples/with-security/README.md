The chart uses a layered values.yaml files to demonstrate the different permutations of configuration. 

## Method 1: No TLS and No SASL

If no TLS or SASL is required, simply invoke:

```sh
> cd helm-charts
> helm install redpanda redpanda -n redpanda --create-namespace 
```

This can be followed by a basic helm test that will change corresponding to your installation.

```sh
> helm test redpanda -n redpanda
```
The output should indicate the success of the tests and some example commands you can try that do not require either SASL or TLS configuration.

## Method 2: TLS but No SASL

If TLS is required, invoke the following:

```sh
> helm install redpanda redpanda -f examples/with-security/values_add_tls.yaml -n redpanda --create-namespace
```

When this command is invoked the self signed issuers are created for each service by default following the model of the Redpanda [operator](https://www.redpanda.com). These self-signed Issuers are used to create per service root Issuers whome accordingly create keys and ca certs separately for each service. This behaviour is for example only. A later example will demonstrate how to utilise your own custom Issuer.

The following test command will ensure that the kafka api, panda proxy and schema registy all have basic access using TLS.

```sh
> helm test redpanda -n redpanda
``` 

Note that specification of the subdomain in the configuration is automatically detected and included in the SAN. For example the following amendment to the `kafka_api`

```
    kafka_api:
      - name: kafka 
        port: 9092
        external:
        	enabled: true
        	subdomain: "streaming.rockdata.io"
```

Will generate the external nodeport and the following entries in the certificate

```
spec:                                                                         
  commonName: redpanda-kafka-cert
dnsNames:
  - redpanda-cluster.redpanda.redpanda.svc.cluster.local                                                                                                                      
  - '*.redpanda-cluster.redpanda.redpanda.svc.cluster.local'                                                                                                           
  - redpanda.redpanda.svc.cluster.local                                                                                                     
  - '*.redpanda.redpanda.svc.cluster.local'
  - streaming.rockdata.io
  - '*.streaming.rockdata.io'
```

Whereby the generated nodeport can be accessed with TLS via `redpanda-<x>.streaming.rockdata.io` for example.


## Method 3: TLS and SASL Enabled

To further include SASL protection for your cluster the following command can be run, layering SASL configuration on top of the basic configuration and TLS configuration additively. For an extensive reference for Redpanda rpk ACL commands please visit [here](https://docs.redpanda.com/docs/reference/rpk-commands/#rpk-acl).

```sh
> helm install redpanda redpanda -f examples/with-security/values_add_tls.yaml -f examples/with-security/values_add_sasl.yaml -n redpanda --create-namespace
```

In this case both TLS and SASL should now be enabled with a default admin user and test password (please dont use this password for your deployment).

The installation can be further tested with the following command:

```sh
> helm test redpanda -n redpanda
```

## Issuer override

The default behaviour of this chart is to create an Issuer per service by iterating the following list in the values file. NOTE: the creation of Issuers is bound to this list, not to the enablement of services (this may change in the future); therefore, an Issuer can be added by merely appending to this list e.g. `- name: myservice`.

```
certIssuers:
  - name: kafka
  - name: proxy
  - name: schema
  - name: admin
```

The certs-issuers.yaml iterates this list performing simple template substitution to generate firsta self-signed Issuer then that self-signed Issuer issues its own service based root Certificate.

The self-signed root certificate is then used to create a <release>-<service>-root-issuer. For example (A):

```sh
> kubectl get issuers -o wide

redpanda-admin-root-issuer                             True    Signing CA verified   2m20s
redpanda-admin-selfsigned-issuer                       True                          2m20s
redpanda-kafka-root-issuer                             True    Signing CA verified   2m20s
redpanda-kafka-selfsigned-issuer                       True                          2m20s
redpanda-proxy-root-issuer                             True    Signing CA verified   2m20s
redpanda-proxy-selfsigned-issuer                       True                          2m20s
redpanda-schema-root-issuer                            True    Signing CA verified   2m20s
redpanda-schema-selfsigned-issuer                      True                          2m20s
```

If required the issuer of the service cert issuers can be specified. This is possible to change per service individually if required.

For example; in this case a self-signed issuer for the rockdata.io company is generated as follows:

```sh
kubectl apply -f - << EOF
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rockdata-io-selfsigned-issuer
spec:
  selfSigned: {}
EOF
```

Therefore values yaml can be modified as follows to specifically override the kafka issuer (NOTE: it is likely that the issuerRef specification will be enriched in the future):

```
certIssuers:
  - name: kafka
    issuerRef: rockdata-io-selfsigned-issuer
  - name: proxy
  - name: schema
  - name: admin
```

This can be easily tested by using using the following:

```sh
> helm install redpanda . -f values_add_custom_issuer.yaml -n redpanda
```

The following helm tests that interact with the Redpanda cluster via TLS as before should pass.

```sh
> helm test redpanda -n redpanda
```

Note the creation of the custom Issuer in the output below.

```sh
> kubectl get issuers

redpanda-admin-root-issuer                             True    36m
redpanda-admin-selfsigned-issuer                       True    36m
redpanda-kafka-root-issuer                             True    36m
redpanda-proxy-root-issuer                             True    36m
redpanda-proxy-selfsigned-issuer                       True    36m
redpanda-schema-root-issuer                            True    36m                                                                 
redpanda-schema-selfsigned-issuer                      True    36m
rockdata-io-selfsigned-issuer                          True    37m
```
