# Redpanda Console Helm Chart

This Helm chart allows you to deploy Redpanda Console to your Redpanda cluster.
You can install the chart by running the following commands:

```shell
helm repo add redpanda 'https://charts.redpanda.com/' 
helm repo update
helm install console redpanda/console -f myvalues.yaml
```

Have a look at the [values.yaml](./values.yaml) file to see the available options.
Additionally, there is an example configuration in the [examples](./examples) directory.
