# Change Log

## Redpanda Chart

### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/redpanda-FILLMEIN) - YYYY-MM-DD

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

## Console Chart
### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/console-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed

## Kminion Chart
### [Unreleased](https://github.com/redpanda-data/helm-charts/releases/tag/console-FILLMEIN) - YYYY-MM-DD
#### Added
#### Changed
#### Fixed
#### Removed
