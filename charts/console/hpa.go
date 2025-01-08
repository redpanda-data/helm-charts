// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_hpa.go.tpl
package console

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func HorizontalPodAutoscaler(dot *helmette.Dot) *autoscalingv2.HorizontalPodAutoscaler {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Autoscaling.Enabled {
		return nil
	}

	metrics := []autoscalingv2.MetricSpec{}

	if values.Autoscaling.TargetCPUUtilizationPercentage != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: "Resource",
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: values.Autoscaling.TargetCPUUtilizationPercentage,
				},
			},
		})
	}

	if values.Autoscaling.TargetMemoryUtilizationPercentage != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: "Resource",
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: values.Autoscaling.TargetMemoryUtilizationPercentage,
				},
			},
		})
	}

	return &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    Labels(dot),
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       Fullname(dot),
			},
			MinReplicas: ptr.To(values.Autoscaling.MinReplicas),
			MaxReplicas: values.Autoscaling.MaxReplicas,
			Metrics:     metrics,
		},
	}
}
