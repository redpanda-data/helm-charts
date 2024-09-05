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
package operator

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

// Expand the name of the chart.
func Name(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	name := helmette.Default(dot.Chart.Name, values.NameOverride)
	return cleanForK8s(name)
}

// Create a default fully qualified app name.
// We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
// If release name contains chart name it will be used as a full name.
func Fullname(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.FullnameOverride != "" {
		return cleanForK8s(values.FullnameOverride)
	}

	name := helmette.Default(dot.Chart.Name, values.NameOverride)

	if helmette.Contains(name, dot.Release.Name) {
		return cleanForK8s(dot.Release.Name)
	}

	return cleanForK8s(fmt.Sprintf("%s-%s", dot.Release.Name, name))
}

// Create chart name and version as used by the chart label.
func Chart(dot *helmette.Dot) string {
	chart := fmt.Sprintf("%s-%s", dot.Chart.Name, dot.Chart.Version)
	return cleanForK8s(strings.ReplaceAll(chart, "+", "_"))
}

// Common labels
func Labels(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	labels := map[string]string{
		"helm.sh/chart":                Chart(dot),
		"app.kubernetes.io/managed-by": dot.Release.Service,
	}

	if dot.Chart.AppVersion != "" {
		labels["app.kubernetes.io/version"] = dot.Chart.AppVersion
	}

	return helmette.Merge(labels, SelectorLabels(dot), values.CommonLabels)
}

func SelectorLabels(dot *helmette.Dot) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     Name(dot),
		"app.kubernetes.io/instance": dot.Release.Name,
	}
}

func cleanForK8s(s string) string {
	return helmette.TrimSuffix("-", helmette.Trunc(63, s))
}

func cleanForK8sWithSuffix(s, suffix string) string {
	lengthToTruncate := (len(s) + len(suffix)) - 63
	if lengthToTruncate > 0 {
		s = helmette.Trunc(lengthToTruncate, s)
	}
	return fmt.Sprintf("%s-%s", s, suffix)
}

// StrategicMergePatch is a half-baked implementation of Kubernetes' strategic
// merge patch. It's closer to a merge patch with smart handling of lists
// that's tailored to the values permitted by [PodTemplate].
func StrategicMergePatch(overrides *corev1.PodTemplateSpec, original corev1.PodTemplateSpec) corev1.PodTemplateSpec {
	// TODO(chrisseto): I'd like to march this towards being a more general
	// solution but getting go & helm to work correctly is going to take some
	// critical thinking.
	// - Pushing everything into a single MergeTo call won't work without VERY
	// careful handling as `merge` is quite sensitive to the inclusion of `nil`
	// values.
	// - Full support of SMP (e.i. directive keys) would require a custom data
	// type or just accepting JSON/YAML strings.
	// - Potentially some careful handling of generics and `get` could be used
	// to make a mostly generic SMP implementation.
	// - Or just use real SMP in go and inject static metadata into helm to
	// have a minimal recursive solution.

	if overrides.ObjectMeta.Labels != nil {
		original.ObjectMeta.Labels = helmette.MergeTo[map[string]string](
			overrides.ObjectMeta.Labels,
			helmette.Default(map[string]string{}, original.ObjectMeta.Labels),
		)
	}

	if overrides.ObjectMeta.Annotations != nil {
		original.ObjectMeta.Annotations = helmette.MergeTo[map[string]string](
			overrides.ObjectMeta.Annotations,
			helmette.Default(map[string]string{}, original.ObjectMeta.Annotations),
		)
	}

	if overrides.Spec.SecurityContext != nil {
		original.Spec.SecurityContext = helmette.MergeTo[*corev1.PodSecurityContext](
			overrides.Spec.SecurityContext,
			helmette.Default(&corev1.PodSecurityContext{}, original.Spec.SecurityContext),
		)
	}

	if !helmette.Empty(overrides.Spec.AutomountServiceAccountToken) {
		original.Spec.AutomountServiceAccountToken = overrides.Spec.AutomountServiceAccountToken
	}

	if overrides.Spec.ImagePullSecrets != nil && len(overrides.Spec.ImagePullSecrets) > 0 {
		original.Spec.ImagePullSecrets = overrides.Spec.ImagePullSecrets
	}

	if !helmette.Empty(overrides.Spec.ServiceAccountName) {
		original.Spec.ServiceAccountName = overrides.Spec.ServiceAccountName
	}

	if !helmette.Empty(overrides.Spec.NodeSelector) {
		original.Spec.NodeSelector = helmette.MergeTo[map[string]string](
			overrides.Spec.NodeSelector,
			helmette.Default(map[string]string{}, original.Spec.NodeSelector),
		)
	}

	if overrides.Spec.Affinity != nil {
		original.Spec.Affinity = helmette.MergeTo[*corev1.Affinity](
			overrides.Spec.Affinity,
			helmette.Default(&corev1.Affinity{}, original.Spec.Affinity),
		)
	}

	if overrides.Spec.TopologySpreadConstraints != nil && len(overrides.Spec.TopologySpreadConstraints) > 0 {
		original.Spec.TopologySpreadConstraints = overrides.Spec.TopologySpreadConstraints
	}

	if overrides.Spec.Volumes != nil && len(overrides.Spec.Volumes) > 0 {
		original.Spec.Volumes = overrides.Spec.Volumes
	}

	overrideContainers := map[string]*corev1.Container{}
	for i := range overrides.Spec.Containers {
		container := &overrides.Spec.Containers[i]
		overrideContainers[string(container.Name)] = container
	}

	if !helmette.Empty(overrides.Spec.RestartPolicy) {
		original.Spec.RestartPolicy = overrides.Spec.RestartPolicy
	}

	if overrides.Spec.TerminationGracePeriodSeconds != nil {
		original.Spec.TerminationGracePeriodSeconds = overrides.Spec.TerminationGracePeriodSeconds
	}

	if overrides.Spec.ActiveDeadlineSeconds != nil {
		original.Spec.ActiveDeadlineSeconds = overrides.Spec.ActiveDeadlineSeconds
	}

	if !helmette.Empty(overrides.Spec.DNSPolicy) {
		original.Spec.DNSPolicy = overrides.Spec.DNSPolicy
	}

	if !helmette.Empty(overrides.Spec.NodeName) {
		original.Spec.NodeName = overrides.Spec.NodeName
	}

	if !helmette.Empty(overrides.Spec.HostNetwork) {
		original.Spec.HostNetwork = overrides.Spec.HostNetwork
	}

	if !helmette.Empty(overrides.Spec.HostPID) {
		original.Spec.HostPID = overrides.Spec.HostPID
	}

	if !helmette.Empty(overrides.Spec.HostIPC) {
		original.Spec.HostIPC = overrides.Spec.HostIPC
	}

	if !helmette.Empty(overrides.Spec.ShareProcessNamespace) {
		original.Spec.ShareProcessNamespace = overrides.Spec.ShareProcessNamespace
	}

	if !helmette.Empty(overrides.Spec.Hostname) {
		original.Spec.Hostname = overrides.Spec.Hostname
	}

	if !helmette.Empty(overrides.Spec.Subdomain) {
		original.Spec.Subdomain = overrides.Spec.Subdomain
	}

	if !helmette.Empty(overrides.Spec.SchedulerName) {
		original.Spec.SchedulerName = overrides.Spec.SchedulerName
	}

	if overrides.Spec.Tolerations != nil && len(overrides.Spec.Tolerations) > 0 {
		original.Spec.Tolerations = overrides.Spec.Tolerations
	}

	if overrides.Spec.HostAliases != nil && len(overrides.Spec.HostAliases) > 0 {
		original.Spec.HostAliases = overrides.Spec.HostAliases
	}

	if !helmette.Empty(overrides.Spec.PriorityClassName) {
		original.Spec.PriorityClassName = overrides.Spec.PriorityClassName
	}

	if !helmette.Empty(overrides.Spec.Priority) {
		original.Spec.Priority = overrides.Spec.Priority
	}

	if overrides.Spec.DNSConfig != nil {
		original.Spec.DNSConfig = helmette.MergeTo[*corev1.PodDNSConfig](
			overrides.Spec.DNSConfig,
			helmette.Default(&corev1.PodDNSConfig{}, original.Spec.DNSConfig),
		)
	}

	if overrides.Spec.ReadinessGates != nil && len(overrides.Spec.ReadinessGates) > 0 {
		original.Spec.ReadinessGates = overrides.Spec.ReadinessGates
	}

	if !helmette.Empty(overrides.Spec.RuntimeClassName) {
		original.Spec.RuntimeClassName = overrides.Spec.RuntimeClassName
	}

	if !helmette.Empty(overrides.Spec.EnableServiceLinks) {
		original.Spec.EnableServiceLinks = overrides.Spec.EnableServiceLinks
	}

	if overrides.Spec.PreemptionPolicy != nil {
		original.Spec.PreemptionPolicy = overrides.Spec.PreemptionPolicy
	}

	// TODO(Rafal) gotohelm does not process maps with different key than string
	// Currently Overhead is of type ResourceList map[ResourceName]resource.Quantity
	//if overrides.Spec.Overhead != nil {
	//	original.Spec.Overhead = helmette.MergeTo[corev1.ResourceList](
	//		overrides.Spec.Overhead,
	//		helmette.Default(corev1.ResourceList{}, original.Spec.Overhead),
	//	)
	//}

	if overrides.Spec.SetHostnameAsFQDN != nil {
		original.Spec.SetHostnameAsFQDN = overrides.Spec.SetHostnameAsFQDN
	}

	if overrides.Spec.HostUsers != nil {
		original.Spec.HostUsers = overrides.Spec.HostUsers
	}

	if overrides.Spec.SchedulingGates != nil && len(overrides.Spec.SchedulingGates) > 0 {
		original.Spec.SchedulingGates = overrides.Spec.SchedulingGates
	}

	if overrides.Spec.ResourceClaims != nil && len(overrides.Spec.ResourceClaims) > 0 {
		original.Spec.ResourceClaims = overrides.Spec.ResourceClaims
	}

	var merged []corev1.Container
	for _, container := range original.Spec.Containers {
		if override, ok := overrideContainers[container.Name]; ok {
			// TODO(chrisseto): Actually implement this as a strategic merge patch.
			// EnvVar's are "last in wins" so there's not too much of a need to fully
			// implement a patch for this usecase.
			env := append(container.Env, override.Env...)
			container = helmette.MergeTo[corev1.Container](override, container)
			container.Env = env
		}

		// TODO(chrisseto): There's a minor divergence in gotohelm that'll be tedious to fix.
		// In go: append(nil, nil) -> nil
		// In helm: append(nil, nil) -> []T{}
		// Work around for now by setting Env to []T{} if it's nil.
		if container.Env == nil {
			container.Env = []corev1.EnvVar{}
		}

		merged = append(merged, container)
	}

	original.Spec.Containers = merged

	overrideContainers = map[string]*corev1.Container{}
	for i := range overrides.Spec.InitContainers {
		container := &overrides.Spec.InitContainers[i]
		overrideContainers[string(container.Name)] = container
	}

	merged = []corev1.Container{}
	for _, container := range original.Spec.InitContainers {
		if override, ok := overrideContainers[container.Name]; ok {
			// TODO(chrisseto): Actually implement this as a strategic merge patch.
			// EnvVar's are "last in wins" so there's not too much of a need to fully
			// implement a patch for this usecase.
			env := append(container.Env, override.Env...)
			container = helmette.MergeTo[corev1.Container](override, container)
			container.Env = env
		}

		// TODO(chrisseto): There's a minor divergence in gotohelm that'll be tedious to fix.
		// In go: append(nil, nil) -> nil
		// In helm: append(nil, nil) -> []T{}
		// Work around for now by setting Env to []T{} if it's nil.
		if container.Env == nil {
			container.Env = []corev1.EnvVar{}
		}

		merged = append(merged, container)
	}

	original.Spec.InitContainers = merged

	overrideEphemeralContainers := map[string]*corev1.EphemeralContainer{}
	for i := range overrides.Spec.EphemeralContainers {
		container := &overrides.Spec.EphemeralContainers[i]
		overrideEphemeralContainers[string(container.Name)] = container
	}

	var mergedEphemeralContainers []corev1.EphemeralContainer
	for _, container := range original.Spec.EphemeralContainers {
		if override, ok := overrideEphemeralContainers[container.Name]; ok {
			// TODO(chrisseto): Actually implement this as a strategic merge patch.
			// EnvVar's are "last in wins" so there's not too much of a need to fully
			// implement a patch for this usecase.
			env := append(container.Env, override.Env...)
			container = helmette.MergeTo[corev1.EphemeralContainer](override, container)
			container.Env = env
		}

		// TODO(chrisseto): There's a minor divergence in gotohelm that'll be tedious to fix.
		// In go: append(nil, nil) -> nil
		// In helm: append(nil, nil) -> []T{}
		// Work around for now by setting Env to []T{} if it's nil.
		if container.Env == nil {
			container.Env = []corev1.EnvVar{}
		}

		mergedEphemeralContainers = append(mergedEphemeralContainers, container)
	}

	original.Spec.EphemeralContainers = mergedEphemeralContainers

	return original
}
