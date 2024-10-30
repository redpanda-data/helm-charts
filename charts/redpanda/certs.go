// +gotohelm:filename=_certs.go.tpl
package redpanda

import (
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func ClientCerts(dot *helmette.Dot) []*certmanagerv1.Certificate {
	if !TLSEnabled(dot) {
		return []*certmanagerv1.Certificate{}
	}

	values := helmette.Unwrap[Values](dot.Values)

	fullname := Fullname(dot)
	service := ServiceName(dot)
	ns := dot.Release.Namespace
	domain := strings.TrimSuffix(values.ClusterDomain, ".")

	var certs []*certmanagerv1.Certificate
	for name, data := range values.TLS.Certs {
		if !helmette.Empty(data.SecretRef) || !ptr.Deref(data.Enabled, true) {
			continue
		}

		var names []string
		if data.IssuerRef == nil || ptr.Deref(data.ApplyInternalDNSNames, false) {
			names = append(names, fmt.Sprintf("%s-cluster.%s.%s.svc.%s", fullname, service, ns, domain))
			names = append(names, fmt.Sprintf("%s-cluster.%s.%s.svc", fullname, service, ns))
			names = append(names, fmt.Sprintf("%s-cluster.%s.%s", fullname, service, ns))
			names = append(names, fmt.Sprintf("*.%s-cluster.%s.%s.svc.%s", fullname, service, ns, domain))
			names = append(names, fmt.Sprintf("*.%s-cluster.%s.%s.svc", fullname, service, ns))
			names = append(names, fmt.Sprintf("*.%s-cluster.%s.%s", fullname, service, ns))
			names = append(names, fmt.Sprintf("%s.%s.svc.%s", service, ns, domain))
			names = append(names, fmt.Sprintf("%s.%s.svc", service, ns))
			names = append(names, fmt.Sprintf("%s.%s", service, ns))
			names = append(names, fmt.Sprintf("*.%s.%s.svc.%s", service, ns, domain))
			names = append(names, fmt.Sprintf("*.%s.%s.svc", service, ns))
			names = append(names, fmt.Sprintf("*.%s.%s", service, ns))
		}

		if values.External.Domain != nil {
			names = append(names, helmette.Tpl(*values.External.Domain, dot))
			names = append(names, fmt.Sprintf("*.%s", helmette.Tpl(*values.External.Domain, dot)))
		}

		duration := helmette.Default("43800h", data.Duration)
		issuerRef := ptr.Deref(data.IssuerRef, cmmeta.ObjectReference{
			Kind:  "Issuer",
			Group: "cert-manager.io",
			Name:  fmt.Sprintf("%s-%s-root-issuer", fullname, name),
		})

		certs = append(certs, &certmanagerv1.Certificate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cert-manager.io/v1",
				Kind:       "Certificate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-cert", fullname, name),
				Labels:    FullLabels(dot),
				Namespace: dot.Release.Namespace,
			},
			Spec: certmanagerv1.CertificateSpec{
				DNSNames:   names,
				Duration:   helmette.MustDuration(duration),
				IsCA:       false,
				IssuerRef:  issuerRef,
				SecretName: fmt.Sprintf("%s-%s-cert", fullname, name),
				PrivateKey: &certmanagerv1.CertificatePrivateKey{
					Algorithm: "ECDSA",
					Size:      256,
				},
			},
		})
	}

	name := values.Listeners.Kafka.TLS.Cert

	data, ok := values.TLS.Certs[name]
	if !ok {
		panic(fmt.Sprintf("Certificate %q referenced but not defined", name))
	}

	if !helmette.Empty(data.SecretRef) || !ClientAuthRequired(dot) {
		return certs
	}

	issuerRef := cmmeta.ObjectReference{
		Group: "cert-manager.io",
		Kind:  "Issuer",
		Name:  fmt.Sprintf("%s-%s-root-issuer", fullname, name),
	}

	if data.IssuerRef != nil {
		issuerRef = *data.IssuerRef
		issuerRef.Group = "cert-manager.io"
	}

	duration := helmette.Default("43800h", data.Duration)

	return append(certs, &certmanagerv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-client", fullname),
			Labels: FullLabels(dot),
		},
		Spec: certmanagerv1.CertificateSpec{
			CommonName: fmt.Sprintf("%s-client", fullname),
			Duration:   helmette.MustDuration(duration),
			IsCA:       false,
			SecretName: fmt.Sprintf("%s-client", fullname),
			PrivateKey: &certmanagerv1.CertificatePrivateKey{
				Algorithm: "ECDSA",
				Size:      256,
			},
			IssuerRef: issuerRef,
		},
	})
}
