# Change Log

## Redpanda Chart

### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

### [5.9.18](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.18) - 2024-12-20
#### Added
#### Changed
#### Fixed
* Fixed an issue with the helm chart when SASL and Connectors were enabled that caused a volume to be mounted incorrectly.
#### Removed

### [5.9.17](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.17) - 2024-12-17
#### Added
#### Changed
* Default for tiered storage cache to `none` which will defer tiered storage cache path to Redpanda process.
#### Fixed
#### Removed

### [5.9.16](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.16) - 2024-12-09
#### Added
#### Changed
* Update sidecar container redpanda-operator container tag
#### Fixed
#### Removed

### [5.9.15](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.15) - 2024-11-29
#### Added
#### Changed
#### Fixed
* ability to overwrite annotation and labels in Job metadata
#### Removed
* non-existent post-upgrade-job values of the non-existent resource (removed in 5.9.6)

### [5.9.14](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.14) - 2024-11-28
#### Added
#### Changed
* note to indicate Core count decreasing will be possible starting from 24.3 Redpanda version
#### Fixed
* Fixed the description of `-memory` and `--reserve-memory` in docs.
#### Removed

### [5.9.13](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.13) - 2024-11-27
#### Added
* overriding any PodSpec fields from `PodTemplate`
#### Changed
* Bump Redpanda operator side car container tag to v2.3.1-24.3.1
#### Fixed
#### Removed

### [5.9.12](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.12) - 2024-11-22
#### Added
#### Changed
* Chart version to update operator side-car container tag
#### Fixed
#### Removed

### [5.9.11](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.11) - 2024-11-21
#### Added
* Ability to generate Redpanda with Connector resources from go code
#### Changed
* Include all Connectors chart values in Redpanda chart values
#### Fixed
#### Removed

### [5.9.10](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.10) - 2024-11-14
#### Added
#### Changed
#### Fixed
* All occurrence of External Domain execution via tpl function
* Calculating Service typed LoadBalancer annotation based on external addresses (even single one)
* Fix connecting to the schema registry via rpk on nodes for versions of rpk that support a node-level rpk stanza.
#### Removed

### [5.9.9](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.9) - 2024-10-24
#### Added
* Strategic merge of Pod volumes and Container volumeMounts
#### Changed
* By default auto mount is disabled in ServiceAccount and Statefulset PodSpec
* Mount volume similar to auto mount functionality for ServiceAccount token when sidecar controllers are enabled
#### Fixed
* Passing console extra volume and volume mount in Redpanda chart
* implements `time.ParseDuration` in gotohelm (with limitations)
* updates the transpilation of `MustParseDuration` to properly re-serialize the provided duration
#### Removed

### [5.9.8](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.8) - 2024-10-23
#### Added
#### Changed
* Bump Redpanda app version
#### Fixed
* Increased the memory limits of `bootstrap-yaml-envsubst` to prevent hangs on aarch64 [#1564](https://github.com/redpanda-data/helm-charts/issues/1564).
#### Removed

### [5.9.7](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.7) - 2024-10-14
#### Added
#### Changed
* Bump Redpanda app version
#### Fixed
#### Removed

### [5.9.6](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.6) - 2024-10-09
#### Added
* Added the ability to override the name of the bootstrap user created when SASL authentication is enabled. [#1547](https://github.com/redpanda-data/helm-charts/pull/1547)
#### Changed
* The minimum Kubernetes version has been bumped to `1.25.0`
#### Fixed
* Chart render failures in tooling compiled with go < 1.19 (e.g. helm 3.10.x) have been fixed.
#### Removed
* `post_upgrade_job.*`, and the post-upgrade job itself, has been removed. All
  it's functionality has been consolidated into the `post_install_job`, which
  actually runs on both post-install and post-upgrade.

  The consolidated job now runs the redpanda-operator image, which may be
  controlled the same way as the additional controllers:
  `statefulset.controllers.{image,repository}`.

### [5.9.5](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.5) - 2024-09-26
#### Added
#### Changed
* Bump Redpanda container tag/application version [#1543](https://github.com/redpanda-data/helm-charts/pull/1543)
#### Fixed
* Connectors deployment [#1543](https://github.com/redpanda-data/helm-charts/pull/1543)
#### Removed

### [5.9.4](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.4) - 2024-09-17
#### Added
#### Changed
* Cluster configurations are no longer include in `redpanda.yaml` or the
  Redpanda Statefulset's configuration hash.

  This change makes it possible to update cluster configurations without
  initiating a rolling restart of the entire cluster.

  As has always been the case, users should consult `rpk cluster config status`
  to determine if a rolling restart needs to be manually performed due to
  cluster configuration changes.

  Cases requiring manual rolling restarts may increase as fewer chart
  operations will initiate rolling restart of the cluster.
#### Fixed
* Fix initialization of configurations using RestToConfig when the passed in rest.Config contain on-disk value files.
#### Removed
* All zero, empty, or default cluster configurations have been removed from
  `values.yaml` in favor of letting redpanda determine what the defaults will
  be. 

  Documentation of cluster configurations has also been removed in favor of
  linking to Redpanda's docs.

### [5.9.3](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.3) - 2024-09-11
#### Added
* Add basic bootstrap user support (#1513)
#### Changed
#### Fixed
* When specified, `truststore_file` is no longer propagated to client configurations.
* If provided, `config.cluster.default_topic_replications` is now respected regardless of the value of `statefulset.replicas`.
#### Removed

### [5.9.1](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.1) - 2024-8-19
#### Added
#### Changed
#### Fixed
* The `truststores` projected volume no longer duplicates entries when the same
  trust store is specified across multiple TLS configurations.
#### Removed

### [5.9.0](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.9.0) - 2024-08-09
#### Added
* `post_install_job.podTemplate` and `post_upgrade_job.podTemplate` have been
  added, which allow overriding various aspects of the corresponding
  `corev1.PodTemplate`. Notably, this field may be used to set labels and
  annotations on the Pod produced by the Job which was not previously possible.
* `statefulset.podTemplate` has benefited from the above additions as well.
  `statefulset.podTemplate.spec.securityContext` and
  `statefulset.podTemplate.spec.containers[*].securityContext` may be used to
  set/override the pod and container security contexts respectively.
* `appProtocol` added to the `listeners.admin` configuration
#### Changed
* The container name of the post-upgrade job is now statically set to
  `post-upgrade` to facilitate strategic merge patching.
* The container name of the post-install job is now statically set to
  `post-install` to facilitate strategic merge patching.
* `statefulset.securityContext`, `statefulset.podSecurityContext`,
  `post_upgrade_job.securityContext`, and `post_install_job.securityContext`
  have all been deprecated due to historically incorrect and confusing
  behavior. The desire to preserve backwards compatibility and not suddenly
  change sensitive fields has left us unable to cleanly correct said issues.
  `{statefulset,post_upgrade_job,post_install_job}.podTemplate` may be used to
  override either the Pod or Container security context.
#### Fixed
#### Removed

### [5.8.15](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.8.15) - 2024-08-08
#### Added
#### Changed
* Bump Redpanda version due to a bug in Redpanda
#### Fixed
* Fix mechanism check in superuser file creation
#### Removed

### [5.8.14](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.8.14) - 2024-08-07
#### Added
* unset `status` and `creationTimestamp` before rendering resource
#### Changed
* Convert connectors to go
* Bump redpanda, connectors, operator and console helm chart application version
#### Fixed
* Fix Redpanda node configuration generation, so that rpk can parse it
* Fix volume mounts in mTLS setup
* Correct boolean coalescing
#### Removed

### [5.8.13](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.8.13) - 2024-07-25
#### Added
#### Changed
* Updated `appVersion` to `v24.1.11`
#### Fixed
* Fixed a regression where `post_upgrade_job` would fail if TLS on the admin
  listener was disabled but had `cert` set to an invalid cert (e.g. `""`)
* Fixed mTLS configurations between Redpanda and Console [#1402](https://github.com/redpanda-data/helm-charts/pull/1402)
* Fixed a typo in `statefulset.securityContext.allowPriviledgeEscalation`. Both the correct
  and typoed name will be respected with the correct spelling taking
  precedence. [#1413](https://github.com/redpanda-data/helm-charts/issues/1413)
#### Removed
* Validation of `issuerRef` has been removed to permit external Issuers.
  [#1432](https://github.com/redpanda-data/helm-charts/issues/1432)

### [5.8.12](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-5.8.12) - 2024-07-10

#### Added

#### Changed
* `image.repository` longer needs to be the default value of
  `"docker.redpanda.com/redpandadata/redpanda"` to respect version checks of
  `image.tag`
  ([#1334](https://github.com/redpanda-data/helm-charts/issues/1334)).
* `post_upgrade_job.extraEnv` and `post_upgrade_job.extraEnvFrom` no longer accept string inputs.

    Previously, they accepted either strings or structured fields. As the types
    of this chart are reflected in the operator's CRD, we are bound by the
    constraints of Kubernetes' CRDs, which do not support fields with multiple
    types. We also noticed that the [CRD requires these fields to be structured
    types](https://github.com/redpanda-data/redpanda-operator/blob/9fa7a7848a22ece215be36dd17f0e4c2ba0002f7/src/go/k8s/api/redpanda/v1alpha2/redpanda_clusterspec_types.go#L597-L600)
    rather than strings. Too minimize the divergences between the two, we've
    opted to drop support for string inputs here but preserve them elsewhere.

    Updating these fields, if they are strings, is typically a case of needing
    to remove `|-`'s from one's values file.

    Before:
    ```yaml
    post_upgrade_job:
      extraEnv: |-
      - name: SPECIAL_LEVEL_KEY
          valueFrom:
            configMapKeyRef:
              name: special-config
              key: special.how
    ```

    After:
    ```yaml
    post_upgrade_job:
      extraEnv:
      - name: SPECIAL_LEVEL_KEY
        valueFrom:
          configMapKeyRef:
            name: special-config
            key: special.how
    ```

    If you were using a templated value and would like to see it added back,
    please [file us an
    issue](https://github.com/redpanda-data/helm-charts/issues/new/choose) and
    tell us about your use case!

#### Fixed
* Numeric node/broker configurations are now properly transcoded as numerics.

#### Removed

## Redpanda Operator Chart
### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/operator-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

### [0.4.38](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.38) - 2024-12-20
#### Added
#### Changed
* App version to match latest redpanda-operator release
#### Fixed
#### Removed

### [0.4.37](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.37) - 2024-12-18
#### Added
#### Changed
* App version to match latest redpanda-operator release
#### Fixed
#### Removed

### [0.4.36](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.36) - 2024-12-09
#### Added
#### Changed
* App version to match latest redpanda-operator release
#### Fixed
#### Removed

### [0.4.35](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.35) - 2024-12-04
#### Added
#### Changed
* to always mounting service account token regardless of auto mount property
#### Fixed
#### Removed

### [0.4.34](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.34) - 2024-11-27
#### Added
* overriding any PodSpec fields from `PodTemplate`
#### Changed
* Bump Redpanda Operator app version to latest release v2.3.2-24.3.1
#### Fixed
#### Removed

### [0.4.33](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.33) - 2024-11-22
#### Added
* Missing permissions for ClusterRoles, ClusterRoleBindings, Horizontal Pod Autoscaler, cert-manager/Certificate,
  cert-manager/Issuer, redpanda/Users, and redpanda/Schemas.
#### Changed
* Application version for newly operator release v2.3.0-24.3.1
#### Fixed
#### Removed

### [0.4.32](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.32) - 2024-10-31
#### Added
* Strategic merge of Pod volumes
* Add new Schema custom resource RBAC rules
#### Changed
* The minimum Kubernetes version has been bumped to `1.25.0`
* Bump operator version [v2.2.5-24.2.7](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.2.5-24.2.7)
* By default auto mount is disabled in ServiceAccount and Deployment PodSpec
* Mount volume similar to auto mount functionality for ServiceAccount token
#### Fixed
* `--configurator-tag` now correctly falls back to `.appVersion`
#### Removed

### [0.4.31](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.31) - 2024-10-7
#### Added
#### Changed
* Bump operator version [v2.2.4-24.2.5](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.2.4-24.2.5)
#### Fixed
#### Removed

### [0.4.30](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.30) - 2024-09-17
#### Added
* Add RBAC rules for the operator chart so it can manage users
#### Changed
#### Fixed
#### Removed

### [0.4.29](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.29) - 2024-09-11
#### Added
#### Changed
* Allow to overwrite `appsv1.Deployment.Spec.PodTemplate`
* Bump operator version [v2.2.2-24.2.4](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.2.2-24.2.4)
* Translate operator helm chart to go.
#### Fixed
#### Removed

### [0.4.28](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.28) - 2024-08-23
#### Added
#### Changed
* Bump operator version [v2.2.0-24.2.2](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.2.0-24.2.2)
#### Fixed
#### Removed

### [0.4.27](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.27) - 2024-08-08
#### Added
#### Changed
* Bump operator version [v2.1.29-24.2.2](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.1.29-24.2.2)
#### Fixed
#### Removed

### [0.4.26](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.26) - 2024-08-07
#### Added
#### Changed
* Bump operator version [v2.1.28-24.2.1](https://github.com/redpanda-data/redpanda-operator/releases/tag/v2.1.28-24.2.1)
#### Fixed
* Fix e2e operator tests
#### Removed

### [0.4.25](https://github.com/redpanda-data/helm-charts/releases/tag/operator-0.4.25) - 2024-07-17
#### Added
#### Changed
* Updated `appVersion` to `v2.1.26-24.1.9`
#### Fixed
* Added missing permissions for the NodeWatcher controller (`rbac.createAdditionalControllerCRs`)
#### Removed

## Connectors Chart
### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/connectors-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

### [0.1.14](https://github.com/redpanda-data/helm-charts/releases/tag/connectors-0.1.14) - 2024-11-20
#### Added
* Enabled flag that would be only used by Redpanda chart when partial values will be embedded into Redpanda values struct
#### Changed
* The minimum Kubernetes version has been bumped to `1.25.0`
* By default auto mount is disabled in ServiceAccount and Deployment PodSpec
* Use render function to generate all resources
#### Fixed
#### Removed

### [0.1.13](https://github.com/redpanda-data/helm-charts/releases/tag/connectors-0.1.13) - 2024-09-26
#### Added
#### Changed
* Test pod name will be stable (without randomization) [#1541](https://github.com/redpanda-data/helm-charts/pull/1541)
* Update connectors container tag/application version [#1541](https://github.com/redpanda-data/helm-charts/pull/1541)
#### Fixed
#### Removed

### [0.1.12](https://github.com/redpanda-data/helm-charts/releases/tag/connectors-0.1.12)
#### Added
#### Changed
#### Fixed
#### Removed

## Console Chart

### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/console-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

### [0.7.31](https://github.com/redpanda-data/helm-charts/releases/tag/console-0.7.31) - 2024-12-06
#### Added
#### Changed
* AppVersion for the new Console release
* By default auto mount is disabled in ServiceAccount and Deployment PodSpec
#### Fixed
#### Removed

### [0.7.30](https://github.com/redpanda-data/helm-charts/releases/tag/console-0.7.30) - 2024-10-14
#### Added
* Add Enabled flag that is used in Redpanda chart
* Add test example for oidc configuration [#1503](https://github.com/redpanda-data/helm-charts/pull/1503)
#### Changed
* Bump Console app version
* Align Console init container default value
* The minimum Kubernetes version has been bumped to `1.25.0`
#### Fixed
* License json tag to correctly set Console license [#1510](https://github.com/redpanda-data/helm-charts/pull/1510)
#### Removed

### [0.7.29](https://github.com/redpanda-data/helm-charts/releases/tag/console-0.7.29) - 2024-08-19
#### Added
#### Changed
#### Fixed
* Fixed empty tag for the console image if tag is not overridden in values [#1476](https://github.com/redpanda-data/helm-charts/issues/1476)
#### Removed

### [0.7.28](https://github.com/redpanda-data/helm-charts/releases/tag/console-0.7.28) - 2024-08-08
#### Added
#### Changed
#### Fixed
* Fixed kubeVersion to be able to deploy on AWS EKS clusters.
#### Removed

### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/console-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
* `initContainers.extraInitContainers` is now pre-processed as YAML by the
  chart. Invalid YAML will instead be rendered as an error messages instead of
  invalid YAML.

#### Fixed
#### Removed
* Support for Kubernetes versions < 1.21 have been dropped.

## Kminion Chart
### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/console-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

### [0.14.1](https://github.com/redpanda-data/helm-charts/releases/tag/kminion-0.14.1)
#### Added
#### Changed
#### Fixed
* Add serviceMonitor targetLabels parameter in Values.yaml
#### Removed

### [0.14.0](https://github.com/redpanda-data/helm-charts/releases/tag/kminion-0.14.0)
#### Added
#### Changed
#### Fixed
#### Removed

## Connect Chart

### [3.0.3](https://github.com/redpanda-data/helm-charts/releases/tag/connect-3.0.3)
#### Changed
* Added complete descriptions for all options in the `values.yaml`.

### [3.0.2](https://github.com/redpanda-data/helm-charts/releases/tag/connect-3.0.2)
#### Changed
* Bump Connect app version to 4.42.0
#### Fixed
* Fixed empty lines after labels when .Values.commonLabels is empty
* Fixed opentelemetry tracer configuration example, should be open_telemetry_collector

### [3.0.1](https://github.com/redpanda-data/helm-charts/releases/tag/connect-3.0.1)
#### Added
* Parameter to configure submitting anonymous telemetry data
#### Changed
* Bump Connect app version to 4.39.0

### [3.0.0](https://github.com/redpanda-data/helm-charts/releases/tag/connect-3.0.0)
#### Added
* Refreshed chart and migrated from [the old standalone repo](https://github.com/redpanda-data/redpanda-connect-helm-chart)
