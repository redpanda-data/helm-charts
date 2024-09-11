# Change Log

## Redpanda Chart

### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
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

### [0.14.0](https://github.com/redpanda-data/helm-charts/releases/tag/kminion-0.14.0)
#### Added
#### Changed
#### Fixed
#### Removed
