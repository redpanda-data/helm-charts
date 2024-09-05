// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +gotohelm:filename=_post-install-upgrade-job.go.tpl
package redpanda

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func PostInstallUpgradeJob(dot *helmette.Dot) *batchv1.Job {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.PostInstallJob.Enabled {
		return nil
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-configuration", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels: helmette.Merge(
				FullLabels(dot),
				helmette.Default(map[string]string{}, values.PostInstallJob.Labels),
			),
			Annotations: helmette.Merge(
				// This is what defines this resource as a hook. Without this line, the
				// job is considered part of the release.
				map[string]string{
					"helm.sh/hook":               "post-install,post-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "-5",
				},
				helmette.Default(map[string]string{}, values.PostInstallJob.Annotations),
			),
		},
		Spec: batchv1.JobSpec{
			Template: StrategicMergePatch(values.PostInstallJob.PodTemplate, corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: fmt.Sprintf("%s-post-", dot.Release.Name),
					Labels: helmette.Merge(
						map[string]string{
							"app.kubernetes.io/name":      Name(dot),
							"app.kubernetes.io/instance":  dot.Release.Name,
							"app.kubernetes.io/component": fmt.Sprintf("%.50s-post-install", Name(dot)),
						},
						helmette.Default(map[string]string{}, values.CommonLabels),
					),
				},
				Spec: corev1.PodSpec{
					NodeSelector:     values.NodeSelector,
					Affinity:         postInstallJobAffinity(dot),
					Tolerations:      tolerations(dot),
					RestartPolicy:    corev1.RestartPolicyNever,
					SecurityContext:  PodSecurityContext(dot),
					ImagePullSecrets: helmette.Default(nil, values.ImagePullSecrets),
					Containers: []corev1.Container{
						{
							Name:      PostInstallContainerName,
							Image:     fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
							Env:       rpkEnvVars(dot, PostInstallUpgradeEnvironmentVariables(dot)),
							Command:   []string{"bash", "-c"},
							Args:      []string{},
							Resources: ptr.Deref(values.PostInstallJob.Resources, corev1.ResourceRequirements{}),
							// Note: this is a semantic change/fix from the template, which specified the merge in the incorrect order
							SecurityContext: ptr.To(helmette.MergeTo[corev1.SecurityContext](
								ptr.Deref(values.PostInstallJob.SecurityContext, corev1.SecurityContext{}),
								ContainerSecurityContext(dot),
							)),
							VolumeMounts: DefaultMounts(dot),
						},
					},
					Volumes:            DefaultVolumes(dot),
					ServiceAccountName: ServiceAccountName(dot),
				},
			}),
		},
	}

	var script []string
	script = append(script,
		`set -e`,
	)

	if RedpandaAtLeast_22_2_0(dot) {
		script = append(script,
			`if [[ -n "$REDPANDA_LICENSE" ]] then`,
			`  rpk cluster license set "$REDPANDA_LICENSE"`,
			`fi`,
		)
	}
	/* ### Here be dragons ###
	This block of bash configures cluster configuration settings by
	pulling them from environment variables.

	This allows us to support configurations from secrets or their raw
	values.

	WARNING: There is a small race condition here. `rpk cluster config import`
	will reset any values that are not specified. To work around this, we first
	export the the configuration. If there's a change to the configuration
	while we're updating the exported config on disk, said change will be reverted.

	TODO(chrisseto): Consolidate all cluster configuration setting to this job.
	*/
	script = append(script,
		// First: dump the existing cluster configuration.
		// We need to use config import to handle conditional configurations
		// (e.g. cloud_storage_enabled). Maintaining a DAG of configurations
		// is not an option for the helm chart.
		``, ``, ``, ``, // TODO: just WS-alignment with the original template; drop these
		`rpk cluster config export -f /tmp/cfg.yml`,
		``, ``,

		// Second: For each environment variable with the prefix RPK
		// ("${!RPK_@}"), use `rpk redpanda config set` to update the exported
		// config, ignoring any authentication environment variables.

		// Lots of Bash Jargon here:
		//     "${KEY#*RPK_}" => Strip the RPK_ prefix from KEY.
		//     "${config,,}" => config.toLower()
		//     "${!KEY}" => Dynamic variable resolution. ie: What is the value of the variable with a name equal to the value of $KEY?

		`for KEY in "${!RPK_@}"; do`,
		`  if ! [[ "$KEY" =~ ^(RPK_USER|RPK_PASS|RPK_SASL_MECHANISM)$ ]]; then`,
		`    config="${KEY#*RPK_}"`,
		`    rpk redpanda config set --config /tmp/cfg.yml "${config,,}" "${!KEY}"`,
		`  fi`,
		`done`,
		``, ``,

		// The updated file is then loaded via `rpk cluster config import` which
		// ensures that conditional configurations (cloud_storage_enabled)
		// "see" all their dependent keys.
		`rpk cluster config import -f /tmp/cfg.yml`,
		``,
	)
	job.Spec.Template.Spec.Containers[0].Args = append(job.Spec.Template.Spec.Containers[0].Args, helmette.Join("\n", script))

	return job
}

// was: post-install-job-affinity
// Set affinity for post_install_job, defaults to global affinity if not defined in post_install_job
func postInstallJobAffinity(dot *helmette.Dot) *corev1.Affinity {
	values := helmette.Unwrap[Values](dot.Values)

	if !helmette.Empty(values.PostInstallJob.Affinity) {
		return &values.PostInstallJob.Affinity
	}

	return helmette.MergeTo[*corev1.Affinity](values.PostInstallJob.Affinity, values.Affinity)
}

func tolerations(dot *helmette.Dot) []corev1.Toleration {
	values := helmette.Unwrap[Values](dot.Values)

	var result []corev1.Toleration
	for _, t := range values.Tolerations {
		result = append(result, helmette.MergeTo[corev1.Toleration](t))
	}
	return result
}

// PostInstallUpgradeEnvironmentVariables returns environment variables assigned to Redpanda
// container.
func PostInstallUpgradeEnvironmentVariables(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	envars := []corev1.EnvVar{}

	if license := GetLicenseLiteral(dot); license != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "REDPANDA_LICENSE",
			Value: license,
		})
	} else if secretReference := GetLicenseSecretReference(dot); secretReference != nil {
		envars = append(envars, corev1.EnvVar{
			Name: "REDPANDA_LICENSE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: secretReference,
			},
		})
	}

	if !values.Storage.IsTieredStorageEnabled() {
		return envars
	}

	tieredStorageConfig := values.Storage.GetTieredStorageConfig()

	ac, azureContainerExists := tieredStorageConfig["cloud_storage_azure_container"]
	asa, azureStorageAccountExists := tieredStorageConfig["cloud_storage_azure_storage_account"]
	if azureContainerExists && ac != nil && azureStorageAccountExists && asa != nil {
		envars = append(envars, addAzureSharedKey(tieredStorageConfig, values)...)
	} else {
		envars = append(envars, addCloudStorageSecretKey(tieredStorageConfig, values)...)
	}

	envars = append(envars, addCloudStorageAccessKey(tieredStorageConfig, values)...)

	for k, v := range tieredStorageConfig {
		if k == "cloud_storage_access_key" || k == "cloud_storage_secret_key" || k == "cloud_storage_azure_shared_key" {
			continue
		}

		if v == nil || helmette.Empty(v) {
			continue
		}

		// cloud_storage_cache_size can be represented as Resource.Quantity that why value can be converted
		// from value with SI suffix to bytes number.
		if k == "cloud_storage_cache_size" {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: helmette.ToJSON(helmette.UnmarshalInto[*resource.Quantity](v).Value()),
			})
			continue
		}

		if str, ok := v.(string); ok {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: str,
			})
		} else {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: helmette.MustToJSON(v),
			})
		}
	}

	return rpkEnvVars(dot, envars)
}

func addCloudStorageAccessKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	if v, ok := tieredStorageConfig["cloud_storage_access_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_ACCESS_KEY",
				Value: v.(string),
			},
		}
	} else if ak := values.Storage.Tiered.CredentialsSecretRef.AccessKey; ak.IsValid() {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: ak.Name},
						Key:                  ak.Key,
					},
				},
			},
		}
	}
	return []corev1.EnvVar{}
}

func addCloudStorageSecretKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	if v, ok := tieredStorageConfig["cloud_storage_secret_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_SECRET_KEY",
				Value: v.(string),
			},
		}
	} else if sk := values.Storage.Tiered.CredentialsSecretRef.SecretKey; sk.IsValid() {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: sk.Name},
						Key:                  sk.Key,
					},
				},
			},
		}
	}
	return []corev1.EnvVar{}
}

func addAzureSharedKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	// Preference Tiered Storage Config over credential secret reference
	if v, ok := tieredStorageConfig["cloud_storage_azure_shared_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				Value: v.(string),
			},
		}
	} else if sk := values.Storage.Tiered.CredentialsSecretRef.SecretKey; sk.IsValid() {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: sk.Name},
						Key:                  sk.Key,
					},
				},
			},
		}
	}

	return []corev1.EnvVar{}
}

func GetLicenseLiteral(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Enterprise.License != "" {
		return values.Enterprise.License
	}

	// Deprecated licenseKey fallback if Enterprise.License is not set
	return values.LicenseKey
}

func GetLicenseSecretReference(dot *helmette.Dot) *corev1.SecretKeySelector {
	values := helmette.Unwrap[Values](dot.Values)

	if !helmette.Empty(values.Enterprise.LicenseSecretRef) {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: values.Enterprise.LicenseSecretRef.Name,
			},
			Key: values.Enterprise.LicenseSecretRef.Key,
		}
		// Deprecated licenseSecretRef fallback if Enterprise.LicenseSecretRef is not set
	} else if !helmette.Empty(values.LicenseSecretRef) {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: values.LicenseSecretRef.SecretName,
			},
			Key: values.LicenseSecretRef.SecretKey,
		}
	}
	return nil
}
