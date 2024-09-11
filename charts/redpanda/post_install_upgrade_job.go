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

// bootstrapYamlTemplater returns an initcontainer that will template
// environment variables into ${base-config}/boostrap.yaml and output it to
// ${config}/.bootstrap.yaml.
func bootstrapYamlTemplater(dot *helmette.Dot) corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	env := values.Storage.Tiered.CredentialsSecretRef.AsEnvVars(values.Storage.GetTieredStorageConfig())

	image := fmt.Sprintf(`%s:%s`,
		values.Statefulset.SideCars.Controllers.Image.Repository,
		values.Statefulset.SideCars.Controllers.Image.Tag,
	)

	return corev1.Container{
		Name:  "bootstrap-yaml-envsubst",
		Image: image,
		Command: []string{
			"/redpanda-operator",
			"envsubst",
			"/tmp/base-config/bootstrap.yaml",
			"--output",
			"/tmp/config/.bootstrap.yaml",
		},
		Env: env,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("25Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("25Mi"),
			},
		},
		SecurityContext: &corev1.SecurityContext{
			// NB: RunAsUser and RunAsGroup will be inherited from the
			// PodSecurityContext of consumers.
			AllowPrivilegeEscalation: ptr.To(false),
			ReadOnlyRootFilesystem:   ptr.To(true),
			RunAsNonRoot:             ptr.To(true),
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "config", MountPath: "/tmp/config/"},
			{Name: "base-config", MountPath: "/tmp/base-config/"},
		},
	}
}

func PostInstallUpgradeJob(dot *helmette.Dot) *batchv1.Job {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.PostInstallJob.Enabled {
		return nil
	}

	image := fmt.Sprintf(`%s:%s`,
		values.Statefulset.SideCars.Controllers.Image.Repository,
		values.Statefulset.SideCars.Controllers.Image.Tag,
	)

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
					NodeSelector:                 values.NodeSelector,
					Affinity:                     postInstallJobAffinity(dot),
					Tolerations:                  tolerations(dot),
					RestartPolicy:                corev1.RestartPolicyNever,
					SecurityContext:              PodSecurityContext(dot),
					ImagePullSecrets:             helmette.Default(nil, values.ImagePullSecrets),
					InitContainers:               []corev1.Container{bootstrapYamlTemplater(dot)},
					AutomountServiceAccountToken: ptr.To(false),
					Containers: []corev1.Container{
						{
							Name:  PostInstallContainerName,
							Image: image,
							Env:   PostInstallUpgradeEnvironmentVariables(dot),
							// See sync-cluster-config in the operator for exact details. Roughly, it:
							// 1. Sets the redpanda license
							// 2. Sets the redpanda cluster config
							// 3. Restarts schema-registry (see https://github.com/redpanda-data/redpanda-operator/issues/232)
							// Upon the post-install run, the clusters's
							// configuration will be re-set (that is set again
							// not reset) which is an unfortunate but ultimately acceptable side effect.
							Command: []string{
								"/redpanda-operator",
								"sync-cluster-config",
								"--redpanda-yaml", "/tmp/base-config/redpanda.yaml",
								"--bootstrap-yaml", "/tmp/config/.bootstrap.yaml",
							},
							Resources: ptr.Deref(values.PostInstallJob.Resources, corev1.ResourceRequirements{}),
							// Note: this is a semantic change/fix from the template, which specified the merge in the incorrect order
							SecurityContext: ptr.To(helmette.MergeTo[corev1.SecurityContext](
								ptr.Deref(values.PostInstallJob.SecurityContext, corev1.SecurityContext{}),
								ContainerSecurityContext(dot),
							)),
							VolumeMounts: append(
								CommonMounts(dot),
								corev1.VolumeMount{Name: "config", MountPath: "/tmp/config"},
								corev1.VolumeMount{Name: "base-config", MountPath: "/tmp/base-config"},
							),
						},
					},
					Volumes: append(
						CommonVolumes(dot),
						corev1.Volume{
							Name: "base-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: Fullname(dot),
									},
								},
							},
						},
						corev1.Volume{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					),
					ServiceAccountName: ServiceAccountName(dot),
				},
			}),
		},
	}

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

	// include any authentication envvars as well.
	return bootstrapEnvVars(dot, envars)
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
