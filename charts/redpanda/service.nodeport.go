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
// +gotohelm:filename=service.nodeport.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NodePortService(dot *helmette.Dot) *corev1.Service {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.External.Enabled || values.External.Type != "NodePort" {
		return nil
	}

	if values.External.Service == nil || !values.External.Service.Enabled {
		return nil
	}

	// NB: As of writing, `mustAppend` appears to not work with nil values.
	// Hence ports is defined as an empty list rather than a zero list.
	ports := []corev1.ServicePort{}

	for name, listener := range values.Listeners.Admin.External {
		if listener.Enabled != nil && *listener.Enabled == false {
			continue
		}
		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("admin-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: listener.AdvertisedPorts[0],
		})
	}

	for name, listener := range values.Listeners.Kafka.External {
		if listener.Enabled != nil && *listener.Enabled == false {
			continue
		}
		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("kafka-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: listener.AdvertisedPorts[0],
		})
	}

	for name, listener := range values.Listeners.HTTP.External {
		if listener.Enabled != nil && *listener.Enabled == false {
			continue
		}
		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("http-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: listener.AdvertisedPorts[0],
		})
	}

	for name, listener := range values.Listeners.SchemaRegistry.External {
		if listener.Enabled != nil && *listener.Enabled == false {
			continue
		}
		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprintf("schema-%s", name),
			Protocol: corev1.ProtocolTCP,
			Port:     listener.Port,
			NodePort: listener.AdvertisedPorts[0],
		})
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
			Annotations: helmette.Default(map[string]string{}, values.External.Annotations).(map[string]string),
		},
		Spec: corev1.ServiceSpec{
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyLocal,
			Ports:                    ports,
			PublishNotReadyAddresses: true,
			Selector:                 StatefulSetPodLabelsSelector(dot, nil /* TODO this probably needs to be filled out */),
			SessionAffinity:          corev1.ServiceAffinityNone,
			Type:                     corev1.ServiceTypeNodePort,
		},
	}
}
