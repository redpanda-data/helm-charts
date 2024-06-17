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
// +gotohelm:filename=_statefulset.go.tpl
package redpanda

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	// RedpandaContainerName is the user facing name of the redpanda container
	// in the redpanda StatefulSet. While the name of the container can
	// technically change, this is the name that is used to locate the
	// [corev1.Container] that will be smp'd into the redpanda container.
	RedpandaContainerName = "redpanda"
	// TrustStoreMountPath is the absolute path at which the
	// [corev1.VolumeProjection] of truststores will be mounted to the redpanda
	// container. (Without a trailing slash)
	TrustStoreMountPath = "/etc/truststores"
)

// StatefulSetRedpandaEnv returns the environment variables for the Redpanda
// container of the Redpanda Statefulset.
func StatefulSetRedpandaEnv(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	// Ideally, this would just be a part of the strategic merge patch. While
	// we're moving the chart into go in a piecemeal fashion there isn't a "top
	// level" location to perform the merge so we're instead required to
	// Implement aspects of the SMP by hand.
	var userEnv []corev1.EnvVar
	for _, container := range values.Statefulset.PodTemplate.Spec.Containers {
		if container.Name == RedpandaContainerName {
			userEnv = container.Env
		}
	}

	// TODO(chrisseto): Actually implement this as a strategic merge patch.
	// EnvVar's are "last in wins" so there's not too much of a need to fully
	// implement a patch for this usecase.
	return append([]corev1.EnvVar{
		{
			Name: "SERVICE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
	}, userEnv...)
}

// StatefulSetPodLabelsSelector returns the label selector for the Redpanda StatefulSet.
// If this helm release is an upgrade, the existing statefulset's label selector will be used as it's an immutable field.
func StatefulSetPodLabelsSelector(dot *helmette.Dot) map[string]string {
	// StatefulSets cannot change their selector. Use the existing one even if it's broken.
	// New installs will get better selectors.
	if dot.Release.IsUpgrade {
		if existing, ok := helmette.Lookup[appsv1.StatefulSet](dot, dot.Release.Namespace, Fullname(dot)); ok && len(existing.Spec.Selector.MatchLabels) > 0 {
			return existing.Spec.Selector.MatchLabels
		}
	}

	values := helmette.Unwrap[Values](dot.Values)

	additionalSelectorLabels := map[string]string{}
	if values.Statefulset.AdditionalSelectorLabels != nil {
		additionalSelectorLabels = values.Statefulset.AdditionalSelectorLabels
	}

	component := fmt.Sprintf("%s-statefulset",
		strings.TrimSuffix(helmette.Trunc(51, Name(dot)), "-"))

	defaults := map[string]string{
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/instance":  dot.Release.Name,
		"app.kubernetes.io/name":      Name(dot),
	}

	return helmette.Merge(additionalSelectorLabels, defaults)
}

// StatefulSetPodLabels returns the label that includes label selector for the Redpanda PodTemplate.
// If this helm release is an upgrade, the existing statefulset's pod template labels will be used as it's an immutable field.
func StatefulSetPodLabels(dot *helmette.Dot) map[string]string {
	// StatefulSets cannot change their selector. Use the existing one even if it's broken.
	// New installs will get better selectors.
	if dot.Release.IsUpgrade {
		if existing, ok := helmette.Lookup[appsv1.StatefulSet](dot, dot.Release.Namespace, Fullname(dot)); ok && len(existing.Spec.Template.ObjectMeta.Labels) > 0 {
			return existing.Spec.Template.ObjectMeta.Labels
		}
	}

	values := helmette.Unwrap[Values](dot.Values)

	statefulSetLabels := map[string]string{}
	if values.Statefulset.PodTemplate.Labels != nil {
		statefulSetLabels = values.Statefulset.PodTemplate.Labels
	}

	defaults := map[string]string{
		"redpanda.com/poddisruptionbudget": Fullname(dot),
	}

	return helmette.Merge(statefulSetLabels, StatefulSetPodLabelsSelector(dot), defaults, FullLabels(dot))
}

// StatefulSetPodAnnotations returns the annotation for the Redpanda PodTemplate.
func StatefulSetPodAnnotations(dot *helmette.Dot, configMapChecksum string) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	configMapChecksumAnnotation := map[string]string{
		"config.redpanda.com/checksum": configMapChecksum,
	}

	if values.Statefulset.PodTemplate.Annotations != nil {
		return helmette.Merge(values.Statefulset.PodTemplate.Annotations, configMapChecksumAnnotation)
	}

	return helmette.Merge(values.Statefulset.Annotations, configMapChecksumAnnotation)
}

// StatefulSetVolumes returns the [corev1.Volume]s for the Redpanda StatefulSet.
func StatefulSetVolumes(dot *helmette.Dot) []corev1.Volume {
	fullname := Fullname(dot)
	volumes := CommonVolumes(dot)
	values := helmette.Unwrap[Values](dot.Values)

	// NOTE extraVolumes, datadir, and tiered-storage-dir are NOT in this
	// function. TODO: Migrate them into this function.
	volumes = append(volumes, []corev1.Volume{
		{
			Name: "lifecycle-scripts",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%.50s-sts-lifecycle", fullname),
					DefaultMode: ptr.To[int32](0o775),
				},
			},
		},
		{
			Name: fullname,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: fullname},
				},
			},
		},
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: fmt.Sprintf("%.51s-configurator", fullname),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%.51s-configurator", fullname),
					DefaultMode: ptr.To[int32](0o775),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-config-watcher", fullname),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-config-watcher", fullname),
					DefaultMode: ptr.To[int32](0o775),
				},
			},
		},
		{
			Name: fmt.Sprintf("%.49s-fs-validator", fullname),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%.49s-fs-validator", fullname),
					DefaultMode: ptr.To[int32](0o775),
				},
			},
		},
	}...)

	if vol := values.Listeners.TrustStoreVolume(&values.TLS); vol != nil {
		volumes = append(volumes, *vol)
	}

	return volumes
}

// StatefulSetRedpandaMounts returns the VolumeMounts for the Redpanda
// Container of the Redpanda StatefulSet.
func StatefulSetVolumeMounts(dot *helmette.Dot) []corev1.VolumeMount {
	mounts := CommonMounts(dot)
	values := helmette.Unwrap[Values](dot.Values)

	// NOTE extraVolumeMounts and tiered-storage-dir are still handled in helm.
	// TODO: Migrate them into this function.
	mounts = append(mounts, []corev1.VolumeMount{
		{Name: "config", MountPath: "/etc/redpanda"},
		{Name: Fullname(dot), MountPath: "/tmp/base-config"},
		{Name: "lifecycle-scripts", MountPath: "/var/lifecycle"},
		{Name: "datadir", MountPath: "/var/lib/redpanda/data"},
	}...)

	if len(values.Listeners.TrustStores(&values.TLS)) > 0 {
		mounts = append(
			mounts,
			corev1.VolumeMount{Name: "truststores", MountPath: TrustStoreMountPath, ReadOnly: true},
		)
	}

	return mounts
}
