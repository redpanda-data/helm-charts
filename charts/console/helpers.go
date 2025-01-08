// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_helpers.go.tpl
package console

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
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
func ChartLabel(dot *helmette.Dot) string {
	chart := fmt.Sprintf("%s-%s", dot.Chart.Name, dot.Chart.Version)
	return cleanForK8s(strings.ReplaceAll(chart, "+", "_"))
}

// Common labels
func Labels(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	labels := map[string]string{
		"helm.sh/chart":                ChartLabel(dot),
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
