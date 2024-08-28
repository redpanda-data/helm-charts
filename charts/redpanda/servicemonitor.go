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
// +gotohelm:filename=_servicemonitor.go.tpl
package redpanda

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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
		EnableHttp2: values.Monitoring.EnableHttp2,
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
