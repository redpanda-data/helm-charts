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
// +gotohelm:filename=_pod-monitor.go.tpl
package connectors

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
