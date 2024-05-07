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
// +gotohelm:filename=servicemonitor.go.tpl
package redpanda

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceMonitor(dot *helmette.Dot) *monitoringv1.ServiceMonitor {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Monitoring.Enabled {
		return nil
	}

	monitorLabels := helmette.Merge(FullLabels(dot), values.Monitoring.Labels)

	matchLabels := map[string]string{
		"monitoring.redpanda.com/enabled": "true",
		"app.kubernetes.io/name":          Name(dot),
		"app.kubernetes.io/instance":      dot.Release.Name,
	}

	tlsConfig := values.Monitoring.TLSConfig

	// tslConfig is not let to be nil because of prexisting logic to
	// disable verify tls (scheme is always hptts) if not defined.
	if tlsConfig != nil {
		tlsConfig.SafeTLSConfig.InsecureSkipVerify = false
	} else {
		tlsConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				InsecureSkipVerify: true,
			},
		}
	}

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
			Labels:    monitorLabels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Interval:    values.Monitoring.ScrapeInterval,
					Path:        "/public_metrics",
					Port:        "admin",
					EnableHttp2: values.Monitoring.EnableHttp2,
					Scheme:      "https",
					TLSConfig:   tlsConfig,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}
}
