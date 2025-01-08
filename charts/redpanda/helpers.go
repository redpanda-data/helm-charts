// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_helpers.go.tpl
package redpanda

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

const (
	//nolint:stylecheck
	redpanda_22_2_0 = ">=22.2.0-0 || <0.0.1-0"
	//nolint:stylecheck
	redpanda_22_3_0 = ">=22.3.0-0 || <0.0.1-0"
	//nolint:stylecheck
	redpanda_23_1_1 = ">=23.1.1-0 || <0.0.1-0"
	//nolint:stylecheck
	redpanda_23_1_2 = ">=23.1.2-0 || <0.0.1-0"
	//nolint:stylecheck
	redpanda_22_3_atleast_22_3_13 = ">=22.3.13-0,<22.4"
	//nolint:stylecheck
	redpanda_22_2_atleast_22_2_10 = ">=22.2.10-0,<22.3"
	//nolint:stylecheck
	redpanda_23_2_1 = ">=23.2.1-0 || <0.0.1-0"
	//nolint:stylecheck
	redpanda_23_3_0 = ">=23.3.0-0 || <0.0.1-0"
)

// Create chart name and version as used by the chart label.
func ChartLabel(dot *helmette.Dot) string {
	return cleanForK8s(strings.ReplaceAll(fmt.Sprintf("%s-%s", dot.Chart.Name, dot.Chart.Version), "+", "_"))
}

// Expand the name of the chart
func Name(dot *helmette.Dot) string {
	if override, ok := dot.Values["nameOverride"].(string); ok && override != "" {
		return cleanForK8s(override)
	}
	return cleanForK8s(dot.Chart.Name)
}

// Create a default fully qualified app name.
// We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
func Fullname(dot *helmette.Dot) string {
	if override, ok := dot.Values["fullnameOverride"].(string); ok && override != "" {
		return cleanForK8s(override)
	}
	return cleanForK8s(dot.Release.Name)
}

// full helm labels + common labels
func FullLabels(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	labels := map[string]string{}
	if values.CommonLabels != nil {
		labels = values.CommonLabels
	}

	defaults := map[string]string{
		"helm.sh/chart":                ChartLabel(dot),
		"app.kubernetes.io/name":       Name(dot),
		"app.kubernetes.io/instance":   dot.Release.Name,
		"app.kubernetes.io/managed-by": dot.Release.Service,
		"app.kubernetes.io/component":  Name(dot),
	}

	return helmette.Merge(labels, defaults)
}

// Create the name of the service account to use
func ServiceAccountName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)
	serviceAccount := values.ServiceAccount

	if serviceAccount.Create && serviceAccount.Name != "" {
		return serviceAccount.Name
	} else if serviceAccount.Create {
		return Fullname(dot)
	} else if serviceAccount.Name != "" {
		return serviceAccount.Name
	}

	return "default"
}

// Use AppVersion if image.tag is not set
func Tag(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	tag := string(values.Image.Tag)
	if tag == "" {
		tag = dot.Chart.AppVersion
	}

	pattern := "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"

	if !helmette.RegexMatch(pattern, tag) {
		// This error message is for end users. This can also occur if
		// AppVersion doesn't start with a 'v' in Chart.yaml.
		panic("image.tag must start with a 'v' and be a valid semver")
	}

	return tag
}

// Create a default service name
func ServiceName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Service != nil && values.Service.Name != nil {
		return cleanForK8s(*values.Service.Name)
	}

	return Fullname(dot)
}

// Generate internal fqdn
func InternalDomain(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	service := ServiceName(dot)
	ns := dot.Release.Namespace
	domain := strings.TrimSuffix(values.ClusterDomain, ".")

	return fmt.Sprintf("%s.%s.svc.%s.", service, ns, domain)
}

// check if client auth is enabled for any of the listeners
func TLSEnabled(dot *helmette.Dot) bool {
	values := helmette.Unwrap[Values](dot.Values)

	if values.TLS.Enabled {
		return true
	}

	listeners := []string{"kafka", "admin", "schemaRegistry", "rpc", "http"}
	for _, listener := range listeners {
		tlsCert := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "tls", "cert")
		tlsEnabled := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "tls", "enabled")
		if !helmette.Empty(tlsEnabled) && !helmette.Empty(tlsCert) {
			return true
		}

		external := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "external")
		if helmette.Empty(external) {
			continue
		}

		keys := helmette.Keys(external.(map[string]any))
		for _, key := range keys {
			enabled := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "external", key, "enabled")
			tlsCert := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "external", key, "tls", "cert")
			tlsEnabled := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "external", key, "tls", "enabled")

			if !helmette.Empty(enabled) && !helmette.Empty(tlsCert) && !helmette.Empty(tlsEnabled) {
				return true
			}
		}
	}

	return false
}

func ClientAuthRequired(dot *helmette.Dot) bool {
	listeners := []string{"kafka", "admin", "schemaRegistry", "rpc", "http"}
	for _, listener := range listeners {
		required := helmette.Dig(dot.Values.AsMap(), false, "listeners", listener, "tls", "requireClientAuth")
		if !helmette.Empty(required) {
			return true
		}
	}
	return false
}

// mounts that are common to most containers
func DefaultMounts(dot *helmette.Dot) []corev1.VolumeMount {
	return append([]corev1.VolumeMount{
		{
			Name:      "base-config",
			MountPath: "/etc/redpanda",
		},
	}, CommonMounts(dot)...)
}

// mounts that are common to all containers
func CommonMounts(dot *helmette.Dot) []corev1.VolumeMount {
	values := helmette.Unwrap[Values](dot.Values)

	mounts := []corev1.VolumeMount{}

	if sasl := values.Auth.SASL; sasl.Enabled && sasl.SecretRef != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "users",
			MountPath: "/etc/secrets/users",
			ReadOnly:  true,
		})
	}

	if TLSEnabled(dot) {
		certNames := helmette.Keys(values.TLS.Certs)
		helmette.SortAlpha(certNames)

		for _, name := range certNames {
			cert := values.TLS.Certs[name]

			if !ptr.Deref(cert.Enabled, true) {
				continue
			}

			mounts = append(mounts, corev1.VolumeMount{
				Name:      fmt.Sprintf("redpanda-%s-cert", name),
				MountPath: fmt.Sprintf("%s/%s", certificateMountPoint, name),
			})
		}

		adminTLS := values.Listeners.Admin.TLS
		if adminTLS.RequireClientAuth {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      "mtls-client",
				MountPath: fmt.Sprintf("%s/%s-client", certificateMountPoint, Fullname(dot)),
			})
		}
	}

	return mounts
}

func DefaultVolumes(dot *helmette.Dot) []corev1.Volume {
	return append([]corev1.Volume{
		{
			Name: "base-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: Fullname(dot),
					},
				},
			},
		},
	}, CommonVolumes(dot)...)
}

// volumes that are common to all pods
func CommonVolumes(dot *helmette.Dot) []corev1.Volume {
	volumes := []corev1.Volume{}
	values := helmette.Unwrap[Values](dot.Values)

	if TLSEnabled(dot) {
		certNames := helmette.Keys(values.TLS.Certs)
		helmette.SortAlpha(certNames)

		for _, name := range certNames {
			cert := values.TLS.Certs[name]

			if !ptr.Deref(cert.Enabled, true) {
				continue
			}

			volumes = append(volumes, corev1.Volume{
				Name: fmt.Sprintf("redpanda-%s-cert", name),
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  CertSecretName(dot, name, &cert),
						DefaultMode: ptr.To[int32](0o440),
					},
				},
			})
		}

		adminTLS := values.Listeners.Admin.TLS
		cert := values.TLS.Certs[adminTLS.Cert]
		if adminTLS.RequireClientAuth {
			secretName := fmt.Sprintf("%s-client", Fullname(dot))
			if cert.ClientSecretRef != nil {
				secretName = cert.ClientSecretRef.Name
			}

			volumes = append(volumes, corev1.Volume{
				Name: "mtls-client",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  secretName,
						DefaultMode: ptr.To[int32](0o440),
					},
				},
			})
		}
	}

	if sasl := values.Auth.SASL; sasl.Enabled && sasl.SecretRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "users",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: sasl.SecretRef,
				},
			},
		})
	}

	return volumes
}

// return correct secretName to use based if secretRef exists
func CertSecretName(dot *helmette.Dot, certName string, cert *TLSCert) string {
	if cert.SecretRef != nil {
		return cert.SecretRef.Name
	}
	return fmt.Sprintf("%s-%s-cert", Fullname(dot), certName)
}

// PodSecurityContext returns a subset of [corev1.PodSecurityContext] for the
// redpanda Statefulset. It is also used as the default PodSecurityContext.
func PodSecurityContext(dot *helmette.Dot) *corev1.PodSecurityContext {
	values := helmette.Unwrap[Values](dot.Values)

	sc := ptr.Deref(values.Statefulset.PodSecurityContext, values.Statefulset.SecurityContext)

	return &corev1.PodSecurityContext{
		FSGroup:             sc.FSGroup,
		FSGroupChangePolicy: sc.FSGroupChangePolicy,
	}
}

// ContainerSecurityContext returns a subset of [corev1.SecurityContext] for
// the redpanda Statefulset. It is also used as the default
// ContainerSecurityContext.
func ContainerSecurityContext(dot *helmette.Dot) corev1.SecurityContext {
	values := helmette.Unwrap[Values](dot.Values)

	sc := ptr.Deref(values.Statefulset.PodSecurityContext, values.Statefulset.SecurityContext)

	return corev1.SecurityContext{
		RunAsUser:                sc.RunAsUser,
		RunAsGroup:               coalesce([]*int64{sc.RunAsGroup, sc.FSGroup}),
		AllowPrivilegeEscalation: coalesce([]*bool{sc.AllowPrivilegeEscalation, sc.AllowPriviledgeEscalation}),
		RunAsNonRoot:             sc.RunAsNonRoot,
	}
}

//nolint:stylecheck
func RedpandaAtLeast_22_2_0(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_2_0)
}

//nolint:stylecheck
func RedpandaAtLeast_22_3_0(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_3_0)
}

//nolint:stylecheck
func RedpandaAtLeast_23_1_1(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_1_1)
}

//nolint:stylecheck
func RedpandaAtLeast_23_1_2(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_1_2)
}

//nolint:stylecheck
func RedpandaAtLeast_22_3_atleast_22_3_13(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_3_atleast_22_3_13)
}

//nolint:stylecheck
func RedpandaAtLeast_22_2_atleast_22_2_10(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_2_atleast_22_2_10)
}

//nolint:stylecheck
func RedpandaAtLeast_23_2_1(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_2_1)
}

//nolint:stylecheck
func RedpandaAtLeast_23_3_0(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_3_0)
}

func redpandaAtLeast(dot *helmette.Dot, constraint string) bool {
	version := strings.TrimPrefix(Tag(dot), "v")

	result, err := helmette.SemverCompare(constraint, version)
	if err != nil {
		panic(err)
	}
	return result
}

func cleanForK8s(in string) string {
	return strings.TrimSuffix(helmette.Trunc(63, in), "-")
}

func cleanForK8sWithSuffix(s, suffix string) string {
	lengthToTruncate := (len(s) + len(suffix)) - 63
	if lengthToTruncate > 0 {
		s = helmette.Trunc(lengthToTruncate, s)
	}
	return fmt.Sprintf("%s-%s", s, suffix)
}

func RedpandaSMP(dot *helmette.Dot) int64 {
	values := helmette.Unwrap[Values](dot.Values)

	coresInMillies := values.Resources.CPU.Cores.MilliValue()

	if coresInMillies < 1000 {
		return 1
	}

	return values.Resources.CPU.Cores.Value()
}

// coalesce returns the first non-nil pointer. This is distinct from helmette's
// Coalesce which returns the first non-EMPTY pointer.
// It accepts a slice as variadic methods are not currently supported in
// gotohelm.
func coalesce[T any](values []*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// StrategicMergePatch is a half-baked implementation of Kubernetes' strategic
// merge patch. It's closer to a merge patch with smart handling of lists
// that's tailored to the values permitted by [PodTemplate].
func StrategicMergePatch(overrides PodTemplate, original corev1.PodTemplateSpec) corev1.PodTemplateSpec {
	// Divergences from an actual SMP:
	// - No support for Directives
	// - List merging by key is handled on a case by case basis.
	// - Can't "unset" optional values in the original due to there being no
	//   difference between *T being explicitly nil or not yet.

	overrideSpec := overrides.Spec
	if overrideSpec == nil {
		overrideSpec = &applycorev1.PodSpecApplyConfiguration{}
	}

	merged := helmette.MergeTo[corev1.PodTemplateSpec](
		applycorev1.PodTemplateSpecApplyConfiguration{
			ObjectMetaApplyConfiguration: &applymetav1.ObjectMetaApplyConfiguration{
				Labels:      overrides.Labels,
				Annotations: overrides.Annotations,
			},
			Spec: overrideSpec,
		},
		original,
	)

	merged.Spec.InitContainers = mergeSliceBy(
		original.Spec.InitContainers,
		overrideSpec.InitContainers,
		"name",
		mergeContainer,
	)

	merged.Spec.Containers = mergeSliceBy(
		original.Spec.Containers,
		overrideSpec.Containers,
		"name",
		mergeContainer,
	)

	merged.Spec.Volumes = mergeSliceBy(
		original.Spec.Volumes,
		overrideSpec.Volumes,
		"name",
		mergeVolume,
	)

	// Due to quirks in go's JSON marshalling and some default values in the
	// chart, GoHelmEquivalence can fail with meaningless diffs of null vs
	// empty slice/map. This defaulting ensures we are in fact equivalent at
	// all times but a functionally not required.
	if merged.ObjectMeta.Labels == nil {
		merged.ObjectMeta.Labels = map[string]string{}
	}

	if merged.ObjectMeta.Annotations == nil {
		merged.ObjectMeta.Annotations = map[string]string{}
	}

	if merged.Spec.NodeSelector == nil {
		merged.Spec.NodeSelector = map[string]string{}
	}

	if merged.Spec.Tolerations == nil {
		merged.Spec.Tolerations = []corev1.Toleration{}
	}

	return merged
}

func mergeSliceBy[Original any, Overrides any](
	original []Original,
	override []Overrides,
	mergeKey string,
	mergeFunc func(Original, Overrides) Original,
) []Original {
	originalKeys := map[string]bool{}
	overrideByKey := map[string]Overrides{}

	for _, el := range override {
		key, ok := helmette.Get[string](el, mergeKey)
		if !ok {
			continue
		}
		overrideByKey[key] = el
	}

	// Follow the ordering of original, merging in overrides as needed.
	var merged []Original
	for _, el := range original {
		// Cheating a bit here. We know that "original" types will always have
		// the key we're looking for.
		key, _ := helmette.Get[string](el, mergeKey)
		originalKeys[key] = true

		if elOverride, ok := overrideByKey[key]; ok {
			merged = append(merged, mergeFunc(el, elOverride))
		} else {
			merged = append(merged, el)
		}
	}

	// Append any non-merged overrides.
	for _, el := range override {
		key, ok := helmette.Get[string](el, mergeKey)
		if !ok {
			continue
		}

		if _, ok := originalKeys[key]; ok {
			continue
		}

		merged = append(merged, helmette.MergeTo[Original](el))
	}

	return merged
}

func mergeEnvVar(original corev1.EnvVar, overrides applycorev1.EnvVarApplyConfiguration) corev1.EnvVar {
	// If there's a case of having an env overridden, don't merge. Just accept
	// the override as merging could generate an env with multiple sources.
	return helmette.MergeTo[corev1.EnvVar](overrides)
}

func mergeVolume(original corev1.Volume, override applycorev1.VolumeApplyConfiguration) corev1.Volume {
	return helmette.MergeTo[corev1.Volume](override, original)
}

func mergeVolumeMount(original corev1.VolumeMount, override applycorev1.VolumeMountApplyConfiguration) corev1.VolumeMount {
	return helmette.MergeTo[corev1.VolumeMount](override, original)
}

func mergeContainer(original corev1.Container, override applycorev1.ContainerApplyConfiguration) corev1.Container {
	merged := helmette.MergeTo[corev1.Container](override, original)
	merged.Env = mergeSliceBy(original.Env, override.Env, "name", mergeEnvVar)
	merged.VolumeMounts = mergeSliceBy(original.VolumeMounts, override.VolumeMounts, "name", mergeVolumeMount)
	return merged
}
