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
// +gotohelm:filename=_ingress.go.tpl
package console

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Ingress(dot *helmette.Dot) *networkingv1.Ingress {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Ingress.Enabled {
		return nil
	}

	var tls []networkingv1.IngressTLS
	for _, t := range values.Ingress.TLS {
		var hosts []string
		for _, host := range t.Hosts {
			hosts = append(hosts, helmette.Tpl(host, dot))
		}
		tls = append(tls, networkingv1.IngressTLS{
			SecretName: t.SecretName,
			Hosts:      hosts,
		})
	}

	var rules []networkingv1.IngressRule
	for _, host := range values.Ingress.Hosts {
		var paths []networkingv1.HTTPIngressPath
		for _, path := range host.Paths {
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     path.Path,
				PathType: path.PathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: Fullname(dot),
						Port: networkingv1.ServiceBackendPort{
							Number: values.Service.Port,
						},
					},
				},
			})
		}

		rules = append(rules, networkingv1.IngressRule{
			Host: helmette.Tpl(host.Host, dot),
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}

	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        Fullname(dot),
			Labels:      Labels(dot),
			Namespace:   dot.Release.Namespace,
			Annotations: values.Ingress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: values.Ingress.ClassName,
			TLS:              tls,
			Rules:            rules,
		},
	}
}
