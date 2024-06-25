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
// +gotohelm:filename=cert-issuers.go.tpl
package redpanda

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CertIssuers(dot *helmette.Dot) []*certmanagerv1.Issuer {
	issuers, _ := certIssuersAndCAs(dot)
	return issuers
}

func RootCAs(dot *helmette.Dot) []*certmanagerv1.Certificate {
	_, cas := certIssuersAndCAs(dot)
	return cas
}

func certIssuersAndCAs(dot *helmette.Dot) ([]*certmanagerv1.Issuer, []*certmanagerv1.Certificate) {
	values := helmette.Unwrap[Values](dot.Values)

	var issuers []*certmanagerv1.Issuer
	var certs []*certmanagerv1.Certificate

	if !TLSEnabled(dot) {
		return issuers, certs
	}

	for name, data := range values.TLS.Certs {
		// If secretRef is defined, do not create any of these certificates.
		if data.SecretRef != nil {
			continue
		}

		// If issuerRef is defined, use the specified issuer for the certs
		// If it's not defined, create and use our own issuer.
		if data.IssuerRef == nil {
			// The self-signed issuer is used to create the self-signed CA
			issuers = append(issuers,
				&certmanagerv1.Issuer{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "cert-manager.io/v1",
						Kind:       "Issuer",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf(`%s-%s-selfsigned-issuer`, Fullname(dot), name),
						Namespace: dot.Release.Namespace,
						Labels:    FullLabels(dot),
					},
					Spec: certmanagerv1.IssuerSpec{
						IssuerConfig: certmanagerv1.IssuerConfig{
							SelfSigned: &certmanagerv1.SelfSignedIssuer{},
						},
					},
				},
			)
		}

		// This is the self-signed CA used to issue certs
		issuers = append(issuers,
			&certmanagerv1.Issuer{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "cert-manager.io/v1",
					Kind:       "Issuer",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf(`%s-%s-root-issuer`, Fullname(dot), name),
					Namespace: dot.Release.Namespace,
					Labels:    FullLabels(dot),
				},
				Spec: certmanagerv1.IssuerSpec{
					IssuerConfig: certmanagerv1.IssuerConfig{
						CA: &certmanagerv1.CAIssuer{
							SecretName: fmt.Sprintf(`%s-%s-root-certificate`, Fullname(dot), name),
						},
					},
				},
			},
		)

		// This is the root CA certificate
		certs = append(certs,
			&certmanagerv1.Certificate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "cert-manager.io/v1",
					Kind:       "Certificate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf(`%s-%s-root-certificate`, Fullname(dot), name),
					Namespace: dot.Release.Namespace,
					Labels:    FullLabels(dot),
				},
				Spec: certmanagerv1.CertificateSpec{
					Duration:   helmette.MustDuration(helmette.Default("43800h", data.Duration)),
					IsCA:       true,
					CommonName: fmt.Sprintf(`%s-%s-root-certificate`, Fullname(dot), name),
					SecretName: fmt.Sprintf(`%s-%s-root-certificate`, Fullname(dot), name),
					PrivateKey: &certmanagerv1.CertificatePrivateKey{
						Algorithm: "ECDSA",
						Size:      256,
					},
					IssuerRef: cmmetav1.ObjectReference{
						Name:  fmt.Sprintf(`%s-%s-selfsigned-issuer`, Fullname(dot), name),
						Kind:  "Issuer",
						Group: "cert-manager.io",
					},
				},
			},
		)
	}

	return issuers, certs
}
