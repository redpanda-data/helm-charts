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
// +gotohelm:filename=_helpers.go.tpl
package redpanda

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	redpanda_22_2_0               = ">=22.2.0-0 || <0.0.1-0"
	redpanda_22_3_0               = ">=22.3.0-0 || <0.0.1-0"
	redpanda_23_1_1               = ">=23.1.1-0 || <0.0.1-0"
	redpanda_23_1_2               = ">=23.1.2-0 || <0.0.1-0"
	redpanda_22_3_atleast_22_3_13 = ">=22.3.13-0,<22.4"
	redpanda_22_2_atleast_22_2_10 = ">=22.2.10-0,<22.3"
	redpanda_23_2_1               = ">=23.2.1-0 || <0.0.1-0"
	redpanda_23_3_0               = ">=23.3.0-0 || <0.0.1-0"
)

// Create chart name and version as used by the chart label.
func Chart(dot *helmette.Dot) string {
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
		"helm.sh/chart":                Chart(dot),
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
		required := helmette.Dig(dot.Values.AsMap(), false, listener, "tls", "requireClientAuth")
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
			Name:      "config",
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
			mounts = append(mounts, corev1.VolumeMount{
				Name:      fmt.Sprintf("redpanda-%s-cert", name),
				MountPath: fmt.Sprintf("/etc/tls/certs/%s", name),
			})
		}

		if ClientAuthRequired(dot) {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      "mtls-client",
				MountPath: fmt.Sprintf("/etc/tls/certs/%s-client", Fullname(dot)),
			})
		}
	}

	return mounts
}

func DefaultVolumes(dot *helmette.Dot) []corev1.Volume {
	return append([]corev1.Volume{
		{
			Name: "config",
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

		if ClientAuthRequired(dot) {
			volumes = append(volumes, corev1.Volume{
				Name: "mtls-client",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  fmt.Sprintf("%s-client", Fullname(dot)),
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
func ContainerSecurityContext(dot *helmette.Dot) *corev1.SecurityContext {
	values := helmette.Unwrap[Values](dot.Values)

	sc := ptr.Deref(values.Statefulset.PodSecurityContext, values.Statefulset.SecurityContext)

	return &corev1.SecurityContext{
		RunAsUser:                sc.RunAsUser,
		RunAsGroup:               helmette.Coalesce(sc.RunAsGroup, sc.FSGroup),
		AllowPrivilegeEscalation: sc.AllowPriviledgeEscalation,
		RunAsNonRoot:             sc.RunAsNonRoot,
	}
}

func RedpandaAtLeast_22_2_0(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_2_0)
}

func RedpandaAtLeast_22_3_0(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_3_0)
}

func RedpandaAtLeast_23_1_1(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_1_1)
}

func RedpandaAtLeast_23_1_2(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_1_2)
}

func RedpandaAtLeast_22_3_atleast_22_3_13(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_3_atleast_22_3_13)
}

func RedpandaAtLeast_22_2_atleast_22_2_10(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_22_2_atleast_22_2_10)
}

func RedpandaAtLeast_23_2_1(dot *helmette.Dot) bool {
	return redpandaAtLeast(dot, redpanda_23_2_1)
}

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

func RedpandaSMP(dot *helmette.Dot) int64 {
	values := helmette.Unwrap[Values](dot.Values)

	coresInMillies := values.Resources.CPU.Cores.MilliValue()

	if coresInMillies < 1000 {
		return 1
	}

	return values.Resources.CPU.Cores.Value()
}
