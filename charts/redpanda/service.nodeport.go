// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.nodeport.go.tpl
package redpanda

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func NodePortService(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.External.Enabled || !values.External.Service.Enabled {
		return nil
	}

	if values.External.Type != corev1.ServiceTypeNodePort {
		return nil
	}

	var ports []corev1.ServicePort
	for name, listener := range values.Listeners.Admin.External {
		if !listener.IsEnabled() {
			continue
		}

		nodePort := listener.Port
		if len(listener.AdvertisedPorts) > 0 {
			nodePort = listener.AdvertisedPorts[0]
		}

		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("admin-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: nodePort,
		})
	}

	for name, listener := range values.Listeners.Kafka.External {
		if !listener.IsEnabled() {
			continue
		}

		nodePort := listener.Port
		if len(listener.AdvertisedPorts) > 0 {
			nodePort = listener.AdvertisedPorts[0]
		}

		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("kafka-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: nodePort,
		})
	}

	for name, listener := range values.Listeners.HTTP.External {
		if !listener.IsEnabled() {
			continue
		}

		nodePort := listener.Port
		if len(listener.AdvertisedPorts) > 0 {
			nodePort = listener.AdvertisedPorts[0]
		}

		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("http-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: nodePort,
		})
	}

	for name, listener := range values.Listeners.SchemaRegistry.External {
		if !listener.IsEnabled() {
			continue
		}

		nodePort := listener.Port
		if len(listener.AdvertisedPorts) > 0 {
			nodePort = listener.AdvertisedPorts[0]
		}

		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("schema-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: nodePort,
		})
	}

	annotations := values.External.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-external", ServiceName(dot)),
			Namespace:   dot.Release.Namespace,
			Labels:      FullLabels(dot),
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyLocal,
			Ports:                    ports,
			PublishNotReadyAddresses: true,
			Selector:                 StatefulSetPodLabelsSelector(dot),
			SessionAffinity:          corev1.ServiceAffinityNone,
			Type:                     corev1.ServiceTypeNodePort,
		},
	}
}
