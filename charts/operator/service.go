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
// +gotohelm:filename=_service.go.tpl
package operator

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func WebhookService(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	if !(values.Webhook.Enabled && values.Scope == Cluster) {
		return nil
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-webhook-service", RedpandaOperatorName(dot)),
			Namespace:   dot.Release.Namespace,
			Labels:      Labels(dot),
			Annotations: values.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: SelectorLabels(dot),
			Ports: []corev1.ServicePort{
				{
					Port:       int32(443),
					TargetPort: intstr.FromInt32(9443),
				},
			},
		},
	}
}

func RedpandaOperatorName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.NameOverride != "" {
		return cleanForK8s(values.NameOverride)
	}

	return cleanForK8s(dot.Chart.Name)
}

func MetricsService(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cleanForK8sWithSuffix(Fullname(dot), "metrics-service"),
			Namespace:   dot.Release.Namespace,
			Labels:      Labels(dot),
			Annotations: values.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: SelectorLabels(dot),
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       int32(8443),
					TargetPort: intstr.FromString("https"),
				},
			},
		},
	}
}
