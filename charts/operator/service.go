// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.go.tpl
package operator

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
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
