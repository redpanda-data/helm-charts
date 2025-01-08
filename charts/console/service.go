// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.go.tpl
package console

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func Service(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	port := corev1.ServicePort{
		Name:     "http",
		Port:     int32(values.Service.Port),
		Protocol: corev1.ProtocolTCP,
	}

	if values.Service.TargetPort != nil {
		port.TargetPort = intstr.FromInt32(*values.Service.TargetPort)
	}

	if helmette.Contains("NodePort", string(values.Service.Type)) && values.Service.NodePort != nil {
		port.NodePort = *values.Service.NodePort
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        Fullname(dot),
			Namespace:   dot.Release.Namespace,
			Labels:      Labels(dot),
			Annotations: values.Service.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     values.Service.Type,
			Selector: SelectorLabels(dot),
			Ports:    []corev1.ServicePort{port},
		},
	}
}
