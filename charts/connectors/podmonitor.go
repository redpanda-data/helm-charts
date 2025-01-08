// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_pod-monitor.go.tpl
package connectors

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func PodMonitor(dot *helmette.Dot) *monitoringv1.PodMonitor {
	values := helmette.Unwrap[Values](dot.Values)

	// TODO Add check for .Capabilities.APIVersions.Has "monitoring.coreos.com/v1"
	if !values.Monitoring.Enabled {
		return nil
	}

	return &monitoringv1.PodMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.coreos.com/v1",
			Kind:       "PodMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        Fullname(dot),
			Labels:      values.Monitoring.Labels,
			Annotations: values.Monitoring.Annotations,
		},
		Spec: monitoringv1.PodMonitorSpec{
			NamespaceSelector: values.Monitoring.NamespaceSelector,
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path: "/",
					Port: "prometheus",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: PodLabels(dot),
			},
		},
	}
}
