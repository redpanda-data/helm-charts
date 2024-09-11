package redpanda

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func PostUpgrade(dot *helmette.Dot) *batchv1.Job {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.PostUpgradeJob.Enabled {
		return nil
	}

	labels := helmette.Default(map[string]string{}, values.PostUpgradeJob.Labels)
	annotations := helmette.Default(map[string]string{}, values.PostUpgradeJob.Annotations)

	annotations = helmette.Merge(map[string]string{
		"helm.sh/hook":               "post-upgrade",
		"helm.sh/hook-delete-policy": "before-hook-creation",
		"helm.sh/hook-weight":        "-10",
	}, annotations)

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-post-upgrade", Name(dot)),
			Namespace:   dot.Release.Namespace,
			Labels:      helmette.Merge(FullLabels(dot), labels),
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: values.PostUpgradeJob.BackoffLimit,
			Template: StrategicMergePatch(values.PostUpgradeJob.PodTemplate, corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: dot.Release.Name,
					Labels: helmette.Merge(map[string]string{
						"app.kubernetes.io/name":      Name(dot),
						"app.kubernetes.io/instance":  dot.Release.Name,
						"app.kubernetes.io/component": fmt.Sprintf("%s-post-upgrade", helmette.Trunc(50, Name(dot))),
					}, values.CommonLabels),
				},
				Spec: corev1.PodSpec{
					NodeSelector:       values.NodeSelector,
					Affinity:           helmette.MergeTo[*corev1.Affinity](values.PostUpgradeJob.Affinity, values.Affinity),
					Tolerations:        values.Tolerations,
					RestartPolicy:      corev1.RestartPolicyNever,
					SecurityContext:    PodSecurityContext(dot),
					ServiceAccountName: ServiceAccountName(dot),
					ImagePullSecrets:   helmette.Default(nil, values.ImagePullSecrets),
					Containers: []corev1.Container{
						{
							Name:    PostUpgradeContainerName,
							Image:   fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{PostUpgradeJobScript(dot)},
							Env:     rpkEnvVars(dot, values.PostUpgradeJob.ExtraEnv),
							EnvFrom: values.PostUpgradeJob.ExtraEnvFrom,
							SecurityContext: ptr.To(helmette.MergeTo[corev1.SecurityContext](
								ptr.Deref(values.PostUpgradeJob.SecurityContext, corev1.SecurityContext{}),
								ContainerSecurityContext(dot),
							)),
							Resources:    values.PostUpgradeJob.Resources,
							VolumeMounts: DefaultMounts(dot),
						},
					},
					Volumes: DefaultVolumes(dot),
				},
			}),
		},
	}
}

func PostUpgradeJobScript(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	script := []string{`set -e`, ``}
	for key, value := range values.Config.Cluster {
		asInt64, isInt64 := helmette.AsIntegral[int64](value)

		if asBool, ok := value.(bool); ok && asBool {
			script = append(script, fmt.Sprintf("rpk cluster config set %s %t", key, asBool))
		} else if asStr, ok := value.(string); ok && asStr != "" {
			script = append(script, fmt.Sprintf("rpk cluster config set %s %s", key, asStr))
		} else if isInt64 && asInt64 > 0 {
			script = append(script, fmt.Sprintf("rpk cluster config set %s %d", key, asInt64))
		} else if asSlice, ok := value.([]any); ok && len(asSlice) > 0 {
			script = append(script, fmt.Sprintf(`rpk cluster config set %s "[ %s ]"`, key, helmette.Join(",", asSlice)))
		} else if !helmette.Empty(value) {
			script = append(script, fmt.Sprintf("rpk cluster config set %s %v", key, value))
		}
	}

	// If default_topic_replications is not set and we have at least 3 Brokers,
	// upgrade from redpanda's default of 1 to 3 so, when possible, topics are
	// HA by default.
	// See also:
	// - https://github.com/redpanda-data/helm-charts/issues/583
	// - https://github.com/redpanda-data/helm-charts/issues/1501
	if _, ok := values.Config.Cluster["default_topic_replications"]; !ok && values.Statefulset.Replicas >= 3 {
		script = append(script, "rpk cluster config set default_topic_replications 3")
	}

	if _, ok := values.Config.Cluster["storage_min_free_bytes"]; !ok {
		script = append(script, fmt.Sprintf("rpk cluster config set storage_min_free_bytes %d", values.Storage.StorageMinFreeBytes()))
	}

	if RedpandaAtLeast_23_2_1(dot) {
		service := values.Listeners.Admin

		caCert := ""
		scheme := "http"

		if service.TLS.IsEnabled(&values.TLS) {
			scheme = "https"
			caCert = fmt.Sprintf("--cacert %q", service.TLS.ServerCAPath(&values.TLS))
		}

		url := fmt.Sprintf("%s://%s:%d/v1/debug/restart_service?service=schema-registry", scheme, InternalDomain(dot), int64(service.Port))

		script = append(
			script,
			`if [ -d "/etc/secrets/users/" ]; then`,
			`    IFS=":" read -r USER_NAME PASSWORD MECHANISM < <(grep "" $(find /etc/secrets/users/* -print))`,
			`    curl -svm3 --fail --retry "120" --retry-max-time "120" --retry-all-errors --ssl-reqd \`,
			fmt.Sprintf(`    %s \`, caCert),
			`    -X PUT -u ${USER_NAME}:${PASSWORD} \`,
			fmt.Sprintf(`    %s || true`, url),
			`fi`,
		)
	}

	script = append(script, "")

	return helmette.Join("\n", script)
}
