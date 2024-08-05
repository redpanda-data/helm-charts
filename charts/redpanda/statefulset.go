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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	// NOTE and tiered-storage-dir are NOT in this
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
	}...)

	if values.Statefulset.InitContainers.FSValidator.Enabled {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%.49s-fs-validator", fullname),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%.49s-fs-validator", fullname),
					DefaultMode: ptr.To[int32](0o775),
				},
			},
		})
	}

	if vol := values.Listeners.TrustStoreVolume(&values.TLS); vol != nil {
		volumes = append(volumes, *vol)
	}

	volumes = append(volumes, templateToVolumes(dot, values.Statefulset.ExtraVolumes)...)

	volumes = append(volumes, statefulSetVolumeDataDir(dot))

	if v := statefulSetVolumeTieredStorageDir(dot); v != nil {
		volumes = append(volumes, *v)
	}

	return volumes
}

func statefulSetVolumeDataDir(dot *helmette.Dot) corev1.Volume {
	values := helmette.Unwrap[Values](dot.Values)

	datadirSource := corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	if values.Storage.PersistentVolume.Enabled {
		datadirSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: "datadir",
			},
		}
	} else if values.Storage.HostPath != "" {
		datadirSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: values.Storage.HostPath,
			},
		}
	}
	return corev1.Volume{
		Name:         "datadir",
		VolumeSource: datadirSource,
	}
}

func statefulSetVolumeTieredStorageDir(dot *helmette.Dot) *corev1.Volume {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Storage.IsTieredStorageEnabled() {
		return nil
	}

	tieredType := values.Storage.TieredMountType()
	if tieredType == "none" || tieredType == "persistentVolume" {
		return nil
	}

	if tieredType == "hostPath" {
		return &corev1.Volume{
			Name: "tiered-storage-dir",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: values.Storage.GetTieredStorageHostPath(),
				},
			},
		}
	}

	return &corev1.Volume{
		Name: "tiered-storage-dir",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: values.Storage.CloudStorageCacheSize(),
			},
		},
	}
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

func StatefulSetInitContainers(dot *helmette.Dot) []corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	var containers []corev1.Container
	if c := statefulSetInitContainerTuning(dot); c != nil {
		containers = append(containers, *c)
	}
	if c := statefulSetInitContainerSetDataDirOwnership(dot); c != nil {
		containers = append(containers, *c)
	}
	if c := statefulSetInitContainerFSValidator(dot); c != nil {
		containers = append(containers, *c)
	}
	if c := statefulSetInitContainerSetTieredStorageCacheDirOwnership(dot); c != nil {
		containers = append(containers, *c)
	}
	containers = append(containers, *statefulSetInitContainerConfigurator(dot))
	containers = append(containers, templateToContainers(dot, values.Statefulset.InitContainers.ExtraInitContainers)...)
	return containers
}

func statefulSetInitContainerTuning(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Tuning.TuneAIOEvents {
		return nil
	}

	return &corev1.Container{
		Name:  "tuning",
		Image: fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
		Command: []string{
			`/bin/bash`,
			`-c`,
			`rpk redpanda tune all`,
		},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{`SYS_RESOURCE`},
			},
			Privileged: ptr.To(true),
			RunAsUser:  ptr.To(int64(0)),
			RunAsGroup: ptr.To(int64(0)),
		},
		VolumeMounts: append(append(CommonMounts(dot),
			templateToVolumeMounts(dot, values.Statefulset.InitContainers.Tuning.ExtraVolumeMounts)...),
			corev1.VolumeMount{
				Name:      Fullname(dot),
				MountPath: "/etc/redpanda",
			},
		),
		Resources: helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.InitContainers.Tuning.Resources),
	}
}

func statefulSetInitContainerSetDataDirOwnership(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Statefulset.InitContainers.SetDataDirOwnership.Enabled {
		return nil
	}

	uid, gid := securityContextUidGid(dot, "set-datadir-ownership")

	return &corev1.Container{
		Name:  "set-datadir-ownership",
		Image: fmt.Sprintf("%s:%s", values.Statefulset.InitContainerImage.Repository, values.Statefulset.InitContainerImage.Tag),
		Command: []string{
			`/bin/sh`,
			`-c`,
			fmt.Sprintf(`chown %d:%d -R /var/lib/redpanda/data`, uid, gid),
		},
		VolumeMounts: append(append(CommonMounts(dot),
			templateToVolumeMounts(dot, values.Statefulset.InitContainers.SetDataDirOwnership.ExtraVolumeMounts)...),
			corev1.VolumeMount{
				Name:      `datadir`,
				MountPath: `/var/lib/redpanda/data`,
			}),
		Resources: helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.InitContainers.SetDataDirOwnership.Resources),
	}
}

func securityContextUidGid(dot *helmette.Dot, containerName string) (int64, int64) {
	values := helmette.Unwrap[Values](dot.Values)

	uid := values.Statefulset.SecurityContext.RunAsUser
	if values.Statefulset.PodSecurityContext != nil && values.Statefulset.PodSecurityContext.RunAsUser != nil {
		uid = values.Statefulset.PodSecurityContext.RunAsUser
	}
	if uid == nil {
		panic(fmt.Sprintf(`%s container requires runAsUser to be specified`, containerName))
	}

	gid := values.Statefulset.SecurityContext.FSGroup
	if values.Statefulset.PodSecurityContext != nil && values.Statefulset.PodSecurityContext.FSGroup != nil {
		gid = values.Statefulset.PodSecurityContext.FSGroup
	}
	if gid == nil {
		panic(fmt.Sprintf(`%s container requires fsGroup to be specified`, containerName))
	}
	return *uid, *gid
}

func statefulSetInitContainerFSValidator(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Statefulset.InitContainers.FSValidator.Enabled {
		return nil
	}

	return &corev1.Container{
		Name:    "fs-validator",
		Image:   fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
		Command: []string{`/bin/sh`},
		Args: []string{
			`-c`,
			fmt.Sprintf(`trap "exit 0" TERM; exec /etc/secrets/fs-validator/scripts/fsValidator.sh %s & wait $!`,
				values.Statefulset.InitContainers.FSValidator.ExpectedFS,
			),
		},
		SecurityContext: ptr.To(ContainerSecurityContext(dot)),
		VolumeMounts: append(append(CommonMounts(dot),
			templateToVolumeMounts(dot, values.Statefulset.InitContainers.FSValidator.ExtraVolumeMounts)...),
			corev1.VolumeMount{
				Name:      fmt.Sprintf(`%.49s-fs-validator`, Fullname(dot)),
				MountPath: `/etc/secrets/fs-validator/scripts/`,
			},
			corev1.VolumeMount{
				Name:      `datadir`,
				MountPath: `/var/lib/redpanda/data`,
			},
		),
		Resources: helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.InitContainers.FSValidator.Resources),
	}
}

func statefulSetInitContainerSetTieredStorageCacheDirOwnership(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Storage.IsTieredStorageEnabled() {
		return nil
	}

	uid, gid := securityContextUidGid(dot, "set-tiered-storage-cache-dir-ownership")
	cacheDir := values.Storage.TieredCacheDirectory(dot)
	mounts := CommonMounts(dot)
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "datadir",
		MountPath: "/var/lib/redpanda/data",
	})
	if values.Storage.TieredMountType() != "none" {
		name := "tiered-storage-dir"
		if values.Storage.PersistentVolume != nil && values.Storage.PersistentVolume.NameOverwrite != "" {
			name = values.Storage.PersistentVolume.NameOverwrite
		}
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name,
			MountPath: cacheDir,
		})
	}
	mounts = append(mounts, templateToVolumeMounts(dot, values.Statefulset.InitContainers.SetTieredStorageCacheDirOwnership.ExtraVolumeMounts)...)

	return &corev1.Container{
		Name:  `set-tiered-storage-cache-dir-ownership`,
		Image: fmt.Sprintf(`%s:%s`, values.Statefulset.InitContainerImage.Repository, values.Statefulset.InitContainerImage.Tag),
		Command: []string{
			`/bin/sh`,
			`-c`,
			fmt.Sprintf(`mkdir -p %s; chown %d:%d -R %s`,
				cacheDir,
				uid, gid,
				cacheDir,
			),
		},
		VolumeMounts: mounts,
		Resources:    helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.InitContainers.SetTieredStorageCacheDirOwnership.Resources),
	}
}

func statefulSetInitContainerConfigurator(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	return &corev1.Container{
		Name:  fmt.Sprintf(`%.51s-configurator`, Name(dot)),
		Image: fmt.Sprintf(`%s:%s`, values.Image.Repository, Tag(dot)),
		Command: []string{
			`/bin/bash`,
			`-c`,
			`trap "exit 0" TERM; exec $CONFIGURATOR_SCRIPT "${SERVICE_NAME}" "${KUBERNETES_NODE_NAME}" & wait $!`,
		},
		Env: []corev1.EnvVar{
			{
				Name:  "CONFIGURATOR_SCRIPT",
				Value: "/etc/secrets/configurator/scripts/configurator.sh",
			},
			{
				Name: "SERVICE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
					ResourceFieldRef: nil,
					ConfigMapKeyRef:  nil,
					SecretKeyRef:     nil,
				},
			},
			{
				Name: "KUBERNETES_NODE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "spec.nodeName",
					},
				},
			},
			{
				Name: "HOST_IP_ADDRESS",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.hostIP",
					},
				},
			},
		},
		SecurityContext: ptr.To(ContainerSecurityContext(dot)),
		VolumeMounts: append(append(CommonMounts(dot),
			templateToVolumeMounts(dot, values.Statefulset.InitContainers.Configurator.ExtraVolumeMounts)...),
			corev1.VolumeMount{
				Name:      "config",
				MountPath: "/etc/redpanda",
			},
			corev1.VolumeMount{
				Name:      Fullname(dot),
				MountPath: "/tmp/base-config",
			},
			corev1.VolumeMount{
				Name:      fmt.Sprintf(`%.51s-configurator`, Fullname(dot)),
				MountPath: "/etc/secrets/configurator/scripts/",
			},
		),
		Resources: helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.InitContainers.Configurator.Resources),
	}
}

func StatefulSetContainers(dot *helmette.Dot) []corev1.Container {
	var containers []corev1.Container
	containers = append(containers, *statefulSetContainerRedpanda(dot))
	if c := statefulSetContainerConfigWatcher(dot); c != nil {
		containers = append(containers, *c)
	}
	if c := statefulSetContainerControllers(dot); c != nil {
		containers = append(containers, *c)
	}
	return containers
}

func statefulSetContainerRedpanda(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	internalAdvertiseAddress := fmt.Sprintf("%s.%s", "$(SERVICE_NAME)", InternalDomain(dot))

	container := &corev1.Container{
		Name:  Name(dot),
		Image: fmt.Sprintf(`%s:%s`, values.Image.Repository, Tag(dot)),
		Env:   StatefulSetRedpandaEnv(dot),
		Lifecycle: &corev1.Lifecycle{
			// finish the lifecycle scripts with "true" to prevent them from terminating the pod prematurely
			PostStart: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						`/bin/bash`,
						`-c`,
						helmette.Join("\n", []string{
							fmt.Sprintf(`timeout -v %d bash -x /var/lifecycle/postStart.sh`,
								values.Statefulset.TerminationGracePeriodSeconds/2,
							),
							`true`,
							``,
						}),
					},
				},
			},
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						`/bin/bash`,
						`-c`,
						helmette.Join("\n", []string{
							fmt.Sprintf(`timeout -v %d bash -x /var/lifecycle/preStop.sh`,
								values.Statefulset.TerminationGracePeriodSeconds/2,
							),
							`true # do not fail and cause the pod to terminate`,
							``,
						}),
					},
				},
			},
		},
		StartupProbe: &corev1.Probe{
			// the startupProbe checks to see that the admin api is listening and that the broker has a node_id assigned. This
			// check is only used to delay the start of the liveness and readiness probes until it passes.
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						`/bin/sh`,
						`-c`,
						helmette.Join("\n", []string{
							`set -e`,
							fmt.Sprintf(`RESULT=$(curl --silent --fail -k -m 5 %s "%s://%s/v1/status/ready")`,
								adminTLSCurlFlags(dot),
								adminInternalHTTPProtocol(dot),
								adminApiURLs(dot),
							),
							`echo $RESULT`,
							`echo $RESULT | grep ready`,
							``,
						}),
					},
				},
			},
			InitialDelaySeconds: values.Statefulset.StartupProbe.InitialDelaySeconds,
			PeriodSeconds:       values.Statefulset.StartupProbe.PeriodSeconds,
			FailureThreshold:    values.Statefulset.StartupProbe.FailureThreshold,
		},
		LivenessProbe: &corev1.Probe{
			// the livenessProbe just checks to see that the admin api is listening and returning 200s.
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						`/bin/sh`,
						`-c`,
						fmt.Sprintf(`curl --silent --fail -k -m 5 %s "%s://%s/v1/status/ready"`,
							adminTLSCurlFlags(dot),
							adminInternalHTTPProtocol(dot),
							adminApiURLs(dot),
						),
					},
				},
			},
			InitialDelaySeconds: values.Statefulset.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       values.Statefulset.LivenessProbe.PeriodSeconds,
			FailureThreshold:    values.Statefulset.LivenessProbe.FailureThreshold,
		},
		Command: []string{
			`rpk`,
			`redpanda`,
			`start`,
			fmt.Sprintf(`--advertise-rpc-addr=%s:%d`,
				internalAdvertiseAddress,
				values.Listeners.RPC.Port,
			),
		},
		VolumeMounts: append(StatefulSetVolumeMounts(dot),
			templateToVolumeMounts(dot, values.Statefulset.ExtraVolumeMounts)...),
		SecurityContext: ptr.To(ContainerSecurityContext(dot)),
		Resources:       corev1.ResourceRequirements{},
	}

	if !helmette.Dig(values.Config.Node, false, `recovery_mode_enabled`).(bool) {
		// the readiness probe just checks that the cluster is healthy according to rpk cluster health.
		// It's ok that this cluster-wide check affects all the pods as it's only used for the
		// PodDisruptionBudget and we don't want to roll any pods if the Redpanda cluster isn't healthy.
		// https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#pod-disruption-budgets
		// All services set `publishNotReadyAddresses:true` to prevent this from affecting cluster access
		container.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						`/bin/sh`,
						`-c`,
						helmette.Join("\n", []string{
							`set -x`,
							`RESULT=$(rpk cluster health)`,
							`echo $RESULT`,
							`echo $RESULT | grep 'Healthy:.*true'`,
							``,
						}),
					},
				},
			},
			InitialDelaySeconds: values.Statefulset.ReadinessProbe.InitialDelaySeconds,
			TimeoutSeconds:      values.Statefulset.ReadinessProbe.TimeoutSeconds,
			PeriodSeconds:       values.Statefulset.ReadinessProbe.PeriodSeconds,
			SuccessThreshold:    values.Statefulset.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    values.Statefulset.ReadinessProbe.FailureThreshold,
		}
	}

	// admin http kafka schemaRegistry rpc
	container.Ports = append(container.Ports, corev1.ContainerPort{
		Name:          "admin",
		ContainerPort: values.Listeners.Admin.Port,
	})
	for externalName, external := range values.Listeners.Admin.External {
		if external.IsEnabled() {
			// The original template used
			// $external.port > 0 &&
			// [ $external.enabled ||
			//   (values.External.Enabled && (dig "enabled" true $external)
			// ]
			// ... which is equivalent to the above check
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("admin-%.8s", helmette.Lower(externalName)),
				ContainerPort: external.Port,
			})
		}
	}
	container.Ports = append(container.Ports, corev1.ContainerPort{
		Name:          "http",
		ContainerPort: values.Listeners.HTTP.Port,
	})
	for externalName, external := range values.Listeners.HTTP.External {
		if external.IsEnabled() {
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("http-%.8s", helmette.Lower(externalName)),
				ContainerPort: external.Port,
			})
		}
	}
	container.Ports = append(container.Ports, corev1.ContainerPort{
		Name:          "kafka",
		ContainerPort: values.Listeners.Kafka.Port,
	})
	for externalName, external := range values.Listeners.Kafka.External {
		if external.IsEnabled() {
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("kafka-%.8s", helmette.Lower(externalName)),
				ContainerPort: external.Port,
			})
		}
	}
	container.Ports = append(container.Ports, corev1.ContainerPort{
		Name:          "rpc",
		ContainerPort: values.Listeners.RPC.Port,
	})
	container.Ports = append(container.Ports, corev1.ContainerPort{
		Name:          "schemaregistry",
		ContainerPort: values.Listeners.SchemaRegistry.Port,
	})
	for externalName, external := range values.Listeners.SchemaRegistry.External {
		if external.IsEnabled() {
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("schema-%.8s", helmette.Lower(externalName)),
				ContainerPort: external.Port,
			})
		}
	}

	if values.Storage.IsTieredStorageEnabled() && values.Storage.TieredMountType() != "none" {
		name := "tiered-storage-dir"
		if values.Storage.PersistentVolume != nil && values.Storage.PersistentVolume.NameOverwrite != "" {
			name = values.Storage.PersistentVolume.NameOverwrite
		}
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				Name:      name,
				MountPath: values.Storage.TieredCacheDirectory(dot),
			},
		)
	}

	container.Resources.Limits = helmette.UnmarshalInto[corev1.ResourceList](map[string]any{
		"cpu":    values.Resources.CPU.Cores,
		"memory": values.Resources.Memory.Container.Max,
	})

	if values.Resources.Memory.Container.Min != nil {
		container.Resources.Requests = helmette.UnmarshalInto[corev1.ResourceList](map[string]any{
			"cpu":    values.Resources.CPU.Cores,
			"memory": *values.Resources.Memory.Container.Min,
		})
	}

	return container
}

// adminApiURLs was: admin-api-urls
func adminApiURLs(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	return fmt.Sprintf(`${SERVICE_NAME}.%s:%d`,
		InternalDomain(dot),
		values.Listeners.Admin.Port,
	)
}

func statefulSetContainerConfigWatcher(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Statefulset.SideCars.ConfigWatcher.Enabled {
		return nil
	}

	return &corev1.Container{
		Name:    "config-watcher",
		Image:   fmt.Sprintf(`%s:%s`, values.Image.Repository, Tag(dot)),
		Command: []string{`/bin/sh`},
		Args: []string{
			`-c`,
			`trap "exit 0" TERM; exec /etc/secrets/config-watcher/scripts/sasl-user.sh & wait $!`,
		},
		Resources:       helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.SideCars.ConfigWatcher.Resources),
		SecurityContext: values.Statefulset.SideCars.ConfigWatcher.SecurityContext,
		VolumeMounts: append(
			append(CommonMounts(dot),
				corev1.VolumeMount{
					Name:      "config",
					MountPath: "/etc/redpanda",
				},
				corev1.VolumeMount{
					Name:      fmt.Sprintf(`%s-config-watcher`, Fullname(dot)),
					MountPath: "/etc/secrets/config-watcher/scripts",
				},
			),
			templateToVolumeMounts(dot, values.Statefulset.SideCars.ConfigWatcher.ExtraVolumeMounts)...,
		),
	}
}

func statefulSetContainerControllers(dot *helmette.Dot) *corev1.Container {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.RBAC.Enabled || !values.Statefulset.SideCars.Controllers.Enabled {
		return nil
	}

	return &corev1.Container{
		Name: "redpanda-controllers",
		Image: fmt.Sprintf(`%s:%s`,
			values.Statefulset.SideCars.Controllers.Image.Repository,
			values.Statefulset.SideCars.Controllers.Image.Tag,
		),
		Command: []string{`/manager`},
		Args: []string{
			`--operator-mode=false`,
			fmt.Sprintf(`--namespace=%s`, dot.Release.Namespace),
			fmt.Sprintf(`--health-probe-bind-address=%s`,
				values.Statefulset.SideCars.Controllers.HealthProbeAddress,
			),
			fmt.Sprintf(`--metrics-bind-address=%s`,
				values.Statefulset.SideCars.Controllers.MetricsAddress,
			),
			fmt.Sprintf(`--additional-controllers=%s`,
				helmette.Join(",", values.Statefulset.SideCars.Controllers.Run),
			),
		},
		Env: []corev1.EnvVar{
			{
				Name:  "REDPANDA_HELM_RELEASE_NAME",
				Value: dot.Release.Name,
			},
		},
		Resources:       helmette.UnmarshalInto[corev1.ResourceRequirements](values.Statefulset.SideCars.Controllers.Resources),
		SecurityContext: values.Statefulset.SideCars.Controllers.SecurityContext,
	}
}

func templateToVolumeMounts(dot *helmette.Dot, template string) []corev1.VolumeMount {
	result := helmette.Tpl(template, dot)
	return helmette.UnmarshalYamlArray[corev1.VolumeMount](result)
}

func templateToVolumes(dot *helmette.Dot, template string) []corev1.Volume {
	result := helmette.Tpl(template, dot)
	return helmette.UnmarshalYamlArray[corev1.Volume](result)
}

func templateToContainers(dot *helmette.Dot, template string) []corev1.Container {
	result := helmette.Tpl(template, dot)
	return helmette.UnmarshalYamlArray[corev1.Container](result)
}

func StatefulSet(dot *helmette.Dot) *appsv1.StatefulSet {
	values := helmette.Unwrap[Values](dot.Values)

	if !RedpandaAtLeast_22_2_0(dot) && !values.Force {
		sv := semver(dot)
		panic(fmt.Sprintf("Error: The Redpanda version (%s) is no longer supported \nTo accept this risk, run the upgrade again adding `--force=true`\n", sv))
	}
	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: StatefulSetPodLabelsSelector(dot),
			},
			ServiceName:         ServiceName(dot),
			Replicas:            ptr.To(values.Statefulset.Replicas),
			UpdateStrategy:      helmette.UnmarshalInto[appsv1.StatefulSetUpdateStrategy](values.Statefulset.UpdateStrategy),
			PodManagementPolicy: "Parallel",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      StatefulSetPodLabels(dot),
					Annotations: StatefulSetPodAnnotations(dot, statefulSetChecksumAnnotation(dot)),
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: ptr.To(values.Statefulset.TerminationGracePeriodSeconds),
					SecurityContext:               PodSecurityContext(dot),
					ServiceAccountName:            ServiceAccountName(dot),
					ImagePullSecrets:              helmette.Default(nil, values.ImagePullSecrets),
					InitContainers:                StatefulSetInitContainers(dot),
					Containers:                    StatefulSetContainers(dot),
					Volumes:                       StatefulSetVolumes(dot),
					TopologySpreadConstraints:     statefulSetTopologySpreadConstraints(dot),
					NodeSelector:                  statefulSetNodeSelectors(dot),
					Affinity:                      statefulSetAffinity(dot),
					PriorityClassName:             values.Statefulset.PriorityClassName,
					Tolerations:                   statefulSetTolerations(dot),
				},
			},
			VolumeClaimTemplates: nil, // Set below
		},
	}

	// VolumeClaimTemplates
	if values.Storage.PersistentVolume.Enabled || (values.Storage.IsTieredStorageEnabled() && values.Storage.TieredMountType() == "persistentVolume") {
		if t := volumeClaimTemplateDatadir(dot); t != nil {
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *t)
		}
		if t := volumeClaimTemplateTieredStorageDir(dot); t != nil {
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *t)
		}
	}

	return ss
}

func semver(dot *helmette.Dot) string {
	return strings.TrimPrefix(Tag(dot), "v")
}

// statefulSetChecksumAnnotation was statefulset-checksum-annotation
// statefulset-checksum-annotation calculates a checksum that is used
// as the value for the annotation, "checksum/config". When this value
// changes, kube-controller-manager will roll the pods.
//
// Append any additional dependencies that require the pods to restart
// to the $dependencies list.
func statefulSetChecksumAnnotation(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)
	var dependencies []any
	dependencies = append(dependencies, ConfigMapsWithoutSeedServer(dot))
	if values.External.Enabled {
		dependencies = append(dependencies, ptr.Deref(values.External.Domain, ""))
		if helmette.Empty(values.External.Addresses) {
			dependencies = append(dependencies, "")
		} else {
			dependencies = append(dependencies, values.External.Addresses)
		}
	}
	return helmette.Sha256Sum(helmette.ToJSON(dependencies))
}

// statefulSetTolerations was statefulset-tolerations
func statefulSetTolerations(dot *helmette.Dot) []corev1.Toleration {
	values := helmette.Unwrap[Values](dot.Values)
	return helmette.Default(values.Tolerations, values.Statefulset.Tolerations)
}

// statefulSetNodeSelectors was statefulset-nodeselectors
func statefulSetNodeSelectors(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	return helmette.Default(values.Statefulset.NodeSelector, values.NodeSelector)
}

// statefulSetAffinity was statefulset-affinity
// Set affinity for statefulset, defaults to global affinity if not defined in statefulset
func statefulSetAffinity(dot *helmette.Dot) *corev1.Affinity {
	values := helmette.Unwrap[Values](dot.Values)

	affinity := &corev1.Affinity{}

	if !helmette.Empty(values.Statefulset.NodeAffinity) {
		affinity.NodeAffinity = ptr.To(helmette.UnmarshalInto[corev1.NodeAffinity](values.Statefulset.NodeAffinity))
	} else if !helmette.Empty(values.Affinity.NodeAffinity) {
		affinity.NodeAffinity = ptr.To(helmette.UnmarshalInto[corev1.NodeAffinity](values.Affinity.NodeAffinity))
	}

	if !helmette.Empty(values.Statefulset.PodAffinity) {
		affinity.PodAffinity = ptr.To(helmette.UnmarshalInto[corev1.PodAffinity](values.Statefulset.PodAffinity))
	} else if !helmette.Empty(values.Affinity.PodAffinity) {
		affinity.PodAffinity = ptr.To(helmette.UnmarshalInto[corev1.PodAffinity](values.Affinity.PodAffinity))
	}

	if !helmette.Empty(values.Statefulset.PodAntiAffinity) {
		affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
		if values.Statefulset.PodAntiAffinity.Type == "hard" {
			affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []corev1.PodAffinityTerm{
				{
					TopologyKey: values.Statefulset.PodAntiAffinity.TopologyKey,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: StatefulSetPodLabelsSelector(dot),
					},
				},
			}
		} else if values.Statefulset.PodAntiAffinity.Type == "soft" {
			affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
				{
					Weight: values.Statefulset.PodAntiAffinity.Weight,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: values.Statefulset.PodAntiAffinity.TopologyKey,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: StatefulSetPodLabelsSelector(dot),
						},
					},
				},
			}
		} else if values.Statefulset.PodAntiAffinity.Type == "custom" {
			affinity.PodAntiAffinity = ptr.To(helmette.UnmarshalInto[corev1.PodAntiAffinity](values.Statefulset.PodAntiAffinity.Custom))
		}
	} else if !helmette.Empty(values.Affinity.PodAntiAffinity) {
		affinity.PodAntiAffinity = ptr.To(helmette.UnmarshalInto[corev1.PodAntiAffinity](values.Affinity.PodAntiAffinity))
	}

	return affinity
}

func volumeClaimTemplateDatadir(dot *helmette.Dot) *corev1.PersistentVolumeClaim {
	values := helmette.Unwrap[Values](dot.Values)
	if !values.Storage.PersistentVolume.Enabled {
		return nil
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "datadir",
			Labels: helmette.Merge(map[string]string{
				`app.kubernetes.io/name`:      Name(dot),
				`app.kubernetes.io/instance`:  dot.Release.Name,
				`app.kubernetes.io/component`: Name(dot),
			},
				values.Storage.PersistentVolume.Labels,
				values.CommonLabels,
			),
			Annotations: helmette.Default(nil, values.Storage.PersistentVolume.Annotations),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: helmette.UnmarshalInto[corev1.ResourceList](map[string]any{
					"storage": values.Storage.PersistentVolume.Size,
				}),
			},
		},
	}

	if !helmette.Empty(values.Storage.PersistentVolume.StorageClass) {
		if values.Storage.PersistentVolume.StorageClass == "-" {
			pvc.Spec.StorageClassName = ptr.To("")
		} else {
			pvc.Spec.StorageClassName = ptr.To(values.Storage.PersistentVolume.StorageClass)
		}
	}

	return pvc
}

func volumeClaimTemplateTieredStorageDir(dot *helmette.Dot) *corev1.PersistentVolumeClaim {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Storage.IsTieredStorageEnabled() || values.Storage.TieredMountType() != "persistentVolume" {
		return nil
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: helmette.Default("tiered-storage-dir", values.Storage.PersistentVolume.NameOverwrite),
			Labels: helmette.Merge(map[string]string{
				`app.kubernetes.io/name`:      Name(dot),
				`app.kubernetes.io/instance`:  dot.Release.Name,
				`app.kubernetes.io/component`: Name(dot),
			},
				values.Storage.TieredPersistentVolumeLabels(),
				values.CommonLabels,
			),
			Annotations: helmette.Default(nil, values.Storage.TieredPersistentVolumeAnnotations()),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: helmette.UnmarshalInto[corev1.ResourceList](map[string]any{
					"storage": values.Storage.GetTieredStorageConfig()[`cloud_storage_cache_size`],
				}),
			},
		},
	}

	if sc := values.Storage.TieredPersistentVolumeStorageClass(); sc == "-" {
		pvc.Spec.StorageClassName = ptr.To("")
	} else if !helmette.Empty(sc) {
		pvc.Spec.StorageClassName = ptr.To(sc)
	}

	return pvc
}

func statefulSetTopologySpreadConstraints(dot *helmette.Dot) []corev1.TopologySpreadConstraint {
	values := helmette.Unwrap[Values](dot.Values)

	// XXX: Was protected with this: semverCompare ">=1.16-0" .Capabilities.KubeVersion.GitVersion
	// but that version is beyond EOL; and the chart as a whole wants >= 1.21

	var result []corev1.TopologySpreadConstraint
	labelSelector := &metav1.LabelSelector{
		MatchLabels: StatefulSetPodLabelsSelector(dot),
	}
	for _, v := range values.Statefulset.TopologySpreadConstraints {
		result = append(result,
			corev1.TopologySpreadConstraint{
				MaxSkew:           v.MaxSkew,
				TopologyKey:       v.TopologyKey,
				WhenUnsatisfiable: v.WhenUnsatisfiable,
				LabelSelector:     labelSelector,
			},
		)
	}

	return result
}

// StorageTieredConfig was: storage-tiered-config
// Wrap this up since there are helm tests that require it
func StorageTieredConfig(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)
	return values.Storage.GetTieredStorageConfig()
}
