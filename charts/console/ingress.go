// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_ingress.go.tpl
package console

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
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
