// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_servicemonitor.go.tpl
package redpanda

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func ServiceMonitor(dot *helmette.Dot) *monitoringv1.ServiceMonitor {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Monitoring.Enabled {
		return nil
	}

	endpoint := monitoringv1.Endpoint{
		Interval:    values.Monitoring.ScrapeInterval,
		Path:        "/public_metrics",
		Port:        "admin",
		EnableHttp2: values.Monitoring.EnableHTTP2,
		Scheme:      "http",
	}

	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) || values.Monitoring.TLSConfig != nil {
		endpoint.Scheme = "https"
		endpoint.TLSConfig = values.Monitoring.TLSConfig

		if endpoint.TLSConfig == nil {
			endpoint.TLSConfig = &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
				},
			}
		}
	}

	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.coreos.com/v1",
			Kind:       monitoringv1.ServiceMonitorsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
			Labels:    helmette.Merge(FullLabels(dot), values.Monitoring.Labels),
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{endpoint},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"monitoring.redpanda.com/enabled": "true",
					"app.kubernetes.io/name":          Name(dot),
					"app.kubernetes.io/instance":      dot.Release.Name,
				},
			},
		},
	}
}
