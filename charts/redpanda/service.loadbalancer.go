// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_service.loadbalancer.go.tpl
package redpanda

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func LoadBalancerServices(dot *helmette.Dot) []*corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	// This is technically a divergence from previous behavior but this matches
	// the NodePort's check and is more reasonable.
	if !values.External.Enabled || !values.External.Service.Enabled {
		return nil
	}

	if values.External.Type != corev1.ServiceTypeLoadBalancer {
		return nil
	}

	externalDNS := ptr.Deref(values.External.ExternalDNS, Enableable{})

	labels := FullLabels(dot)

	// This typo is intentionally being preserved for backwards compat
	// https://github.com/redpanda-data/helm-charts/blob/2baa77b99a71a993e639a7138deaf4543727c8a1/charts/redpanda/templates/service.loadbalancer.yaml#L33
	labels["repdanda.com/type"] = "loadbalancer"

	selector := StatefulSetPodLabelsSelector(dot)

	var services []*corev1.Service
	replicas := values.Statefulset.Replicas // TODO fix me once the transpiler is fixed.
	for i := int32(0); i < replicas; i++ {
		podname := fmt.Sprintf("%s-%d", Fullname(dot), i)

		// NB: A range loop is used here as its the most terse way to handle
		// nil maps in gotohelm.
		annotations := map[string]string{}
		for k, v := range values.External.Annotations {
			annotations[k] = v
		}

		if externalDNS.Enabled {
			prefix := podname
			if len(values.External.Addresses) > 0 {
				if len(values.External.Addresses) == 1 {
					prefix = values.External.Addresses[0]
				} else {
					prefix = values.External.Addresses[i]
				}
			}

			address := fmt.Sprintf("%s.%s", prefix, helmette.Tpl(*values.External.Domain, dot))

			annotations["external-dns.alpha.kubernetes.io/hostname"] = address
		}

		// NB: A range loop is used here as its the most terse way to handle
		// nil maps in gotohelm.
		podSelector := map[string]string{}
		for k, v := range selector {
			podSelector[k] = v
		}

		podSelector["statefulset.kubernetes.io/pod-name"] = podname

		var ports []corev1.ServicePort
		for name, listener := range values.Listeners.Admin.External {
			if !ptr.Deref(listener.Enabled, values.External.Enabled) {
				continue
			}

			fallbackPorts := append(listener.AdvertisedPorts, values.Listeners.Admin.Port)

			ports = append(ports, corev1.ServicePort{
				Name:       fmt.Sprintf("admin-%s", name),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(listener.Port),
				Port:       ptr.Deref(listener.NodePort, fallbackPorts[0]),
			})
		}

		for name, listener := range values.Listeners.Kafka.External {
			if !ptr.Deref(listener.Enabled, values.External.Enabled) {
				continue
			}

			fallbackPorts := append(listener.AdvertisedPorts, listener.Port)

			ports = append(ports, corev1.ServicePort{
				Name:       fmt.Sprintf("kafka-%s", name),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(listener.Port),
				Port:       ptr.Deref(listener.NodePort, fallbackPorts[0]),
			})
		}

		for name, listener := range values.Listeners.HTTP.External {
			if !ptr.Deref(listener.Enabled, values.External.Enabled) {
				continue
			}

			fallbackPorts := append(listener.AdvertisedPorts, listener.Port)

			ports = append(ports, corev1.ServicePort{
				Name:       fmt.Sprintf("http-%s", name),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(listener.Port),
				Port:       ptr.Deref(listener.NodePort, fallbackPorts[0]),
			})
		}

		for name, listener := range values.Listeners.SchemaRegistry.External {
			if !ptr.Deref(listener.Enabled, values.External.Enabled) {
				continue
			}

			fallbackPorts := append(listener.AdvertisedPorts, listener.Port)

			ports = append(ports, corev1.ServicePort{
				Name:       fmt.Sprintf("schema-%s", name),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(listener.Port),
				Port:       ptr.Deref(listener.NodePort, fallbackPorts[0]),
			})
		}

		svc := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("lb-%s", podname),
				Namespace:   dot.Release.Namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyLocal,
				LoadBalancerSourceRanges: values.External.SourceRanges,
				Ports:                    ports,
				PublishNotReadyAddresses: true,
				Selector:                 podSelector,
				SessionAffinity:          corev1.ServiceAffinityNone,
				Type:                     corev1.ServiceTypeLoadBalancer,
			},
		}

		services = append(services, svc)
	}

	return services
}
