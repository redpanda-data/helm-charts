# Migration from Benthos

The options accepted by this chart in version `3.0.0` are backwards compatible with version `2.2.0` of the old Benthos based chart.

## Major Differences

- The default docker image has changed from `jeffail/benthos` to `docker.redpanda.com/redpandadata/connect`
- The `name` of every resource has changed. Specifically, occurrences of `benthos` have been replaced with `redpanda-connect`
- The following labels on all resources will have new values:
	- `helm.sh/chart`
	- `app.kubernetes.io/name`
	- `app.kubernetes.io/version`
- The default `root_path` for the HTTP server is now `/redpanda-connect` instead of `/benthos` which affects the configuration of the Ingress and Service resources
- The name of the file in the main ConfigMap has changed from `benthos.yaml` to `redpanda-connect.yaml`
- The name of the Container in the PodSpecs has changed from `benthos` to `connect`
- The mount path of the config file in the `connect` Container has changed from `/benthos.yaml` to `/redpanda-connect.yaml`

## Migration Process

The best way to switch to the new helm chart is to deploy a new release of the new Redpanda Connect chart along side the existing deployment of Benthos, then uninstall the old release once the new Pods are healthy and you have verified that your configuration is working correctly.

```
$ helm list -q
benthos-1730406308
$ helm repo add redpanda https://charts.redpanda.com/
$ helm install --generate-name -f values.yml redpanda/connect
... (verify that the new pods are working)
$ helm uninstall benthos-1730406308
```

If you cannot run multiple copies of Benthos/Redpanda Connect for whatever reason, you will have to uninstall the old Benthos release before installing the new chart.
