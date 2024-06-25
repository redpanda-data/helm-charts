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
			Template: corev1.PodTemplateSpec{
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
							Name:    fmt.Sprintf("%s-post-upgrade", Name(dot)),
							Image:   fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{PostUpgradeJobScript(dot)},
							Env:     values.PostUpgradeJob.ExtraEnv,
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
			},
		},
	}
}

func PostUpgradeJobScript(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	script := []string{`set -e`, ``}
	for key, value := range values.Config.Cluster {
		asInt64, isInt64 := helmette.AsIntegral[int64](value)

		if key == "default_topic_replications" && isInt64 {
			r := int64(values.Statefulset.Replicas)
			// This calculates the closest odd number less than or equal to r: 1=1, 2=1, 3=3, ...
			r = (r + (r % 2)) - 1
			asInt64 = helmette.Min(asInt64, int64(r))
		}

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

	if _, ok := values.Config.Cluster["storage_min_free_bytes"]; !ok {
		script = append(script, fmt.Sprintf("rpk cluster config set storage_min_free_bytes %d", values.Storage.StorageMinFreeBytes()))
	}

	if RedpandaAtLeast_23_2_1(dot) {
		service := values.Listeners.Admin
		cert := values.TLS.Certs.MustGet(service.TLS.Cert)

		caCert := ""
		if cert.CAEnabled {
			caCert = fmt.Sprintf("--cacert /etc/tls/certs/%s/ca.crt", service.TLS.Cert)
		}

		scheme := "http"
		if service.TLS.IsEnabled(&values.TLS) {
			scheme = "https"
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
