// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_deployment.go.tpl
package operator

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

const (
	// Injected bound service account token expiration which triggers monitoring of its time-bound feature.
	// Reference
	// https://github.com/kubernetes/kubernetes/blob/ae53151cb4e6fbba8bb78a2ef0b48a7c32a0a067/pkg/serviceaccount/claims.go#L38-L39
	tokenExpirationSeconds = 60*60 + 7

	// ServiceAccountVolumeName is the prefix name that will be added to volumes that mount ServiceAccount secrets
	// Reference
	// https://github.com/kubernetes/kubernetes/blob/c6669ea7d61af98da3a2aa8c1d2cdc9c2c57080a/plugin/pkg/admission/serviceaccount/admission.go#L52-L53
	ServiceAccountVolumeName = "kube-api-access"

	// DefaultAPITokenMountPath is the path that ServiceAccountToken secrets are automounted to.
	// The token file would then be accessible at /var/run/secrets/kubernetes.io/serviceaccount
	// Reference
	// https://github.com/kubernetes/kubernetes/blob/c6669ea7d61af98da3a2aa8c1d2cdc9c2c57080a/plugin/pkg/admission/serviceaccount/admission.go#L55-L57
	//nolint: gosec
	DefaultAPITokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

func Deployment(dot *helmette.Dot) *appsv1.Deployment {
	values := helmette.Unwrap[Values](dot.Values)

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        Fullname(dot),
			Labels:      Labels(dot),
			Namespace:   dot.Release.Namespace,
			Annotations: values.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(values.ReplicaCount),
			Selector: &metav1.LabelSelector{
				MatchLabels: SelectorLabels(dot),
			},
			Strategy: *values.Strategy,
			Template: StrategicMergePatch(&corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      values.PodTemplate.Metadata.Labels,
					Annotations: values.PodTemplate.Metadata.Annotations,
				},
				Spec: values.PodTemplate.Spec,
			},
				corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: values.PodAnnotations,
						Labels:      helmette.Merge(SelectorLabels(dot), values.PodLabels),
					},
					Spec: corev1.PodSpec{
						AutomountServiceAccountToken:  ptr.To(false),
						TerminationGracePeriodSeconds: ptr.To(int64(10)),
						ImagePullSecrets:              values.ImagePullSecrets,
						ServiceAccountName:            ServiceAccountName(dot),
						NodeSelector:                  values.NodeSelector,
						Tolerations:                   values.Tolerations,
						Volumes:                       operatorPodVolumes(dot),
						Containers:                    operatorContainers(dot, nil),
					},
				}),
		},
	}

	// Values.Affinity should be deprecated.
	if !helmette.Empty(values.Affinity) {
		dep.Spec.Template.Spec.Affinity = values.Affinity
	}

	return dep
}

func operatorContainers(dot *helmette.Dot, podTerminationGracePeriodSeconds *int64) []corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	return []corev1.Container{
		{
			Name:            "manager",
			Image:           containerImage(dot),
			ImagePullPolicy: values.Image.PullPolicy,
			Command:         []string{"/manager"},
			Args:            operatorArguments(dot),
			SecurityContext: &corev1.SecurityContext{AllowPrivilegeEscalation: ptr.To(false)},
			Ports: []corev1.ContainerPort{
				{
					Name:          "webhook-server",
					ContainerPort: 9443,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			VolumeMounts:   operatorPodVolumesMounts(dot),
			LivenessProbe:  livenessProbe(dot, podTerminationGracePeriodSeconds),
			ReadinessProbe: readinessProbe(dot, podTerminationGracePeriodSeconds),
			Resources:      values.Resources,
		},
		{
			Name: "kube-rbac-proxy",
			Args: []string{
				"--secure-listen-address=0.0.0.0:8443",
				"--upstream=http://127.0.0.1:8080/",
				"--logtostderr=true",
				fmt.Sprintf("--v=%d", values.KubeRBACProxy.LogLevel),
			},
			Image:           fmt.Sprintf("%s:%s", values.KubeRBACProxy.Image.Repository, *values.KubeRBACProxy.Image.Tag),
			ImagePullPolicy: values.KubeRBACProxy.Image.PullPolicy,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: 8443,
					Name:          "https",
				},
			},
			VolumeMounts: kubeRBACProxyVolumeMounts(dot),
		},
	}
}

func kubeRBACProxyVolumeMounts(dot *helmette.Dot) []corev1.VolumeMount {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.ServiceAccount.Create {
		return nil
	}

	mountName := ServiceAccountVolumeName
	for _, vol := range operatorPodVolumes(dot) {
		if strings.HasPrefix(ServiceAccountVolumeName+"-", vol.Name) {
			mountName = vol.Name
		}
	}

	return []corev1.VolumeMount{{
		Name:      mountName,
		ReadOnly:  true,
		MountPath: DefaultAPITokenMountPath,
	}}
}

func livenessProbe(dot *helmette.Dot, podTerminationGracePeriodSeconds *int64) *corev1.Probe {
	values := helmette.Unwrap[Values](dot.Values)

	if values.LivenessProbe != nil {
		return &corev1.Probe{
			InitialDelaySeconds:           helmette.Default(15, values.LivenessProbe.InitialDelaySeconds), // TODO what to do with this??
			PeriodSeconds:                 helmette.Default(20, values.LivenessProbe.PeriodSeconds),
			TimeoutSeconds:                values.LivenessProbe.TimeoutSeconds,
			SuccessThreshold:              values.LivenessProbe.SuccessThreshold,
			FailureThreshold:              values.LivenessProbe.FailureThreshold,
			TerminationGracePeriodSeconds: helmette.Default(podTerminationGracePeriodSeconds, values.LivenessProbe.TerminationGracePeriodSeconds),
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz/",
					Port: intstr.FromInt32(8081),
				},
			},
		}
	}
	return &corev1.Probe{
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz/",
				Port: intstr.FromInt32(8081),
			},
		},
	}
}

func readinessProbe(dot *helmette.Dot, podTerminationGracePeriodSeconds *int64) *corev1.Probe {
	values := helmette.Unwrap[Values](dot.Values)

	if values.LivenessProbe != nil {
		return &corev1.Probe{
			InitialDelaySeconds:           helmette.Default(5, values.ReadinessProbe.InitialDelaySeconds),
			PeriodSeconds:                 helmette.Default(10, values.ReadinessProbe.PeriodSeconds),
			TimeoutSeconds:                values.ReadinessProbe.TimeoutSeconds,
			SuccessThreshold:              values.ReadinessProbe.SuccessThreshold,
			FailureThreshold:              values.ReadinessProbe.FailureThreshold,
			TerminationGracePeriodSeconds: helmette.Default(podTerminationGracePeriodSeconds, values.ReadinessProbe.TerminationGracePeriodSeconds),
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/readyz",
					Port: intstr.FromInt32(8081),
				},
			},
		}
	}

	return &corev1.Probe{
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readyz",
				Port: intstr.FromInt32(8081),
			},
		},
	}
}

func containerImage(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	tag := dot.Chart.AppVersion
	if !helmette.Empty(values.Image.Tag) {
		tag = *values.Image.Tag
	}

	return fmt.Sprintf("%s:%s", values.Image.Repository, tag)
}

func configuratorTag(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if !helmette.Empty(values.Configurator.Tag) {
		return *values.Configurator.Tag
	}
	return dot.Chart.AppVersion
}

func isWebhookEnabled(dot *helmette.Dot) bool {
	values := helmette.Unwrap[Values](dot.Values)

	return values.Webhook.Enabled && values.Scope == Cluster
}

func operatorPodVolumes(dot *helmette.Dot) []corev1.Volume {
	values := helmette.Unwrap[Values](dot.Values)

	vol := []corev1.Volume{}

	if values.ServiceAccount.Create {
		vol = append(vol, kubeTokenAPIVolume(ServiceAccountVolumeName))
	}

	if !values.Webhook.Enabled {
		return vol
	}

	vol = append(vol, corev1.Volume{
		Name: "cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: ptr.To(int32(420)),
				SecretName:  values.WebhookSecretName,
			},
		},
	})

	return vol
}

// kubeTokenAPIVolume is a slightly changed variant of
// https://github.com/kubernetes/kubernetes/blob/c6669ea7d61af98da3a2aa8c1d2cdc9c2c57080a/plugin/pkg/admission/serviceaccount/admission.go#L484-L524
// Upstream creates Projected Volume Source, but this function returns Volume with provided name.
// Also const are renamed.
func kubeTokenAPIVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				// explicitly set default value, see https://github.com/kubernetes/kubernetes/issues/104464
				DefaultMode: ptr.To(corev1.ProjectedVolumeSourceDefaultMode),
				Sources: []corev1.VolumeProjection{
					{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Path:              "token",
							ExpirationSeconds: ptr.To(int64(tokenExpirationSeconds)),
						},
					},
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "kube-root-ca.crt",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "ca.crt",
								},
							},
						},
					},
					{
						DownwardAPI: &corev1.DownwardAPIProjection{
							Items: []corev1.DownwardAPIVolumeFile{
								{
									Path: "namespace",
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func operatorPodVolumesMounts(dot *helmette.Dot) []corev1.VolumeMount {
	values := helmette.Unwrap[Values](dot.Values)

	volMount := []corev1.VolumeMount{}

	if values.ServiceAccount.Create {
		mountName := ServiceAccountVolumeName
		for _, vol := range operatorPodVolumes(dot) {
			if strings.HasPrefix(ServiceAccountVolumeName+"-", vol.Name) {
				mountName = vol.Name
			}
		}

		volMount = append(volMount, corev1.VolumeMount{
			Name:      mountName,
			ReadOnly:  true,
			MountPath: DefaultAPITokenMountPath,
		})
	}

	if !values.Webhook.Enabled {
		return volMount
	}

	volMount = append(volMount, corev1.VolumeMount{
		Name:      "cert",
		MountPath: "/tmp/k8s-webhook-server/serving-certs",
		ReadOnly:  true,
	})

	return volMount
}

func operatorArguments(dot *helmette.Dot) []string {
	values := helmette.Unwrap[Values](dot.Values)

	args := []string{
		"--health-probe-bind-address=:8081",
		"--metrics-bind-address=127.0.0.1:8080",
		"--leader-elect",
		fmt.Sprintf("--configurator-tag=%s", configuratorTag(dot)),
		fmt.Sprintf("--configurator-base-image=%s", values.Configurator.Repository),
		fmt.Sprintf("--webhook-enabled=%t", isWebhookEnabled(dot)),
	}

	if values.Scope == Namespace {
		args = append(args,
			fmt.Sprintf("--namespace=%s", dot.Release.Namespace),
			fmt.Sprintf("--log-level=%s", values.LogLevel),
		)
	}

	return append(args, values.AdditionalCmdFlags...)
}
