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
package connectors

import (
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func Service(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	ports := []corev1.ServicePort{
		{
			Name:       "rest-api",
			Port:       values.Connectors.RestPort,
			TargetPort: intstr.FromInt32(values.Connectors.RestPort),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	for _, port := range values.Service.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.Port),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceName(dot),
			// TODO this isn't 100% correct as users could have previously
			// added: `annotations: {}` as the value for annotations to get
			// them to render correctly.
			Labels: helmette.Merge(
				FullLabels(dot),
				values.Service.Annotations,
			),
		},
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{
				corev1.IPv4Protocol,
			},
			IPFamilyPolicy:  ptr.To(corev1.IPFamilyPolicySingleStack),
			Ports:           ports,
			Selector:        PodLabels(dot),
			SessionAffinity: corev1.ServiceAffinityNone,
			Type:            corev1.ServiceTypeClusterIP,
		},
	}
}
