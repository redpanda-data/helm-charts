// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.internal.go.tpl
package redpanda

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func MonitoringEnabledLabel(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)
	return map[string]string{
		// no gotohelm support for strconv.FormatBool
		"monitoring.redpanda.com/enabled": fmt.Sprintf("%t", values.Monitoring.Enabled),
	}
}

func ServiceInternal(dot *helmette.Dot) *corev1.Service {
	// This service is only used to create the DNS enteries for each pod in
	// the stateful set and allow the serviceMonitor to target the pods.
	// This service should not be used by any client application.
	values := helmette.Unwrap[Values](dot.Values)
	ports := []corev1.ServicePort{}

	ports = append(ports, corev1.ServicePort{
		Name:        "admin",
		Protocol:    "TCP",
		AppProtocol: values.Listeners.Admin.AppProtocol,
		Port:        values.Listeners.Admin.Port,
		TargetPort:  intstr.FromInt32(values.Listeners.Admin.Port),
	})

	if values.Listeners.HTTP.Enabled {
		ports = append(ports, corev1.ServicePort{
			Name:       "http",
			Protocol:   "TCP",
			Port:       values.Listeners.HTTP.Port,
			TargetPort: intstr.FromInt32(values.Listeners.HTTP.Port),
		})
	}
	ports = append(ports, corev1.ServicePort{
		Name:       "kafka",
		Protocol:   "TCP",
		Port:       values.Listeners.Kafka.Port,
		TargetPort: intstr.FromInt32(values.Listeners.Kafka.Port),
	})
	ports = append(ports, corev1.ServicePort{
		Name:       "rpc",
		Protocol:   "TCP",
		Port:       values.Listeners.RPC.Port,
		TargetPort: intstr.FromInt32(values.Listeners.RPC.Port),
	})
	if values.Listeners.SchemaRegistry.Enabled {
		ports = append(ports, corev1.ServicePort{
			Name:       "schemaregistry",
			Protocol:   "TCP",
			Port:       values.Listeners.SchemaRegistry.Port,
			TargetPort: intstr.FromInt32(values.Listeners.SchemaRegistry.Port),
		})
	}

	annotations := map[string]string{}
	if values.Service != nil {
		annotations = values.Service.Internal.Annotations
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceName(dot),
			Namespace:   dot.Release.Namespace,
			Labels:      helmette.Merge(FullLabels(dot), MonitoringEnabledLabel(dot)),
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:                     corev1.ServiceTypeClusterIP,
			PublishNotReadyAddresses: true,
			ClusterIP:                corev1.ClusterIPNone,
			Selector:                 StatefulSetPodLabelsSelector(dot),
			Ports:                    ports,
		},
	}
}
