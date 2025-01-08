// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.go.tpl
package connectors

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
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
