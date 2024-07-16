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
package console

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
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
