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
package console

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
