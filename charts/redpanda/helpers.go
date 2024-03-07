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
	return cleanForK8s(fmt.Sprintf("%s", dot.Release.Name))
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

	// TODO Shouldn't our defaults take precedence over user provided labels?
	return helmette.Merge(defaults, labels)
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

	if values.TLS.Enabled != nil && *values.TLS.Enabled {
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

func cleanForK8s(in string) string {
	return strings.TrimSuffix(helmette.Trunc(63, in), "-")
}
