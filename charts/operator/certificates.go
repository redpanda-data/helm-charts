// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_certificates.go.tpl
package operator

import (
	"fmt"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func Certificate(dot *helmette.Dot) *certv1.Certificate {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Webhook.Enabled {
		return nil
	}

	return &certv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "redpanda-serving-cert",
			Namespace:   dot.Release.Namespace,
			Labels:      Labels(dot),
			Annotations: values.Annotations,
		},
		Spec: certv1.CertificateSpec{
			DNSNames: []string{
				fmt.Sprintf("%s-webhook-service.%s.svc", RedpandaOperatorName(dot), dot.Release.Namespace),
				fmt.Sprintf("%s-webhook-service.%s.svc.%s", RedpandaOperatorName(dot), dot.Release.Namespace, values.ClusterDomain),
			},
			IssuerRef: cmmeta.ObjectReference{
				Kind: "Issuer",
				Name: cleanForK8sWithSuffix(Fullname(dot), "selfsigned-issuer"),
			},
			SecretName: values.WebhookSecretName,
			PrivateKey: &certv1.CertificatePrivateKey{
				// There is an issue with gotohelm when RotationPolicyNever is used.
				// The conversion from constant string to helm template is failing.
				//
				// panic: interface conversion: types.Type is *types.Basic, not *types.Struct [recovered]
				RotationPolicy: "Never",
				// RotationPolicy: certv1.RotationPolicyNever,
			},
		},
	}
}

func Issuer(dot *helmette.Dot) *certv1.Issuer {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Webhook.Enabled {
		return nil
	}

	return &certv1.Issuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Issuer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cleanForK8sWithSuffix(Fullname(dot), "selfsigned-issuer"),
			Namespace:   dot.Release.Namespace,
			Labels:      Labels(dot),
			Annotations: values.Annotations,
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}
}
