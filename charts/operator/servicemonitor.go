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
package operator

import (
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func ServiceMonitor(dot *helmette.Dot) *monitorv1.ServiceMonitor {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Monitoring.Enabled {
		return nil
	}

	return &monitorv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMonitor",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cleanForK8sWithSuffix(Fullname(dot), "metrics-monitor"),
			Labels:      Labels(dot),
			Namespace:   dot.Release.Namespace,
			Annotations: values.Annotations,
		},
		Spec: monitorv1.ServiceMonitorSpec{
			Endpoints: []monitorv1.Endpoint{
				{
					Port:   "https",
					Path:   "/metrics",
					Scheme: "https",
					TLSConfig: &monitorv1.TLSConfig{
						SafeTLSConfig: monitorv1.SafeTLSConfig{
							InsecureSkipVerify: ptr.To(true),
						},
					},
					BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
				},
			},
			NamespaceSelector: monitorv1.NamespaceSelector{
				MatchNames: []string{dot.Release.Namespace},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: Labels(dot),
			},
		},
	}
}
