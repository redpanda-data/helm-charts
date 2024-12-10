package redpanda

import (
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestFirstUser(t *testing.T) {
	cases := []struct {
		In  string
		Out [3]string
	}{
		{
			In:  "hello:world:SCRAM-SHA-256",
			Out: [3]string{"hello", "world", "SCRAM-SHA-256"},
		},
		{
			In:  "name:password\n#Intentionally Blank\n",
			Out: [3]string{"name", "password", "SCRAM-SHA-512"},
		},
		{
			In:  "name:password:SCRAM-MD5-999",
			Out: [3]string{"", "", ""},
		},
	}

	for _, c := range cases {
		user, password, mechanism := firstUser([]byte(c.In))
		assert.Equal(t, [3]string{user, password, mechanism}, c.Out)
	}
}

func TestCertificates(t *testing.T) {
	cases := map[string]struct {
		Cert                   *redpanda.TLSCert
		CertificateName        string
		ExpectedRootCertName   string
		ExpectedRootCertKey    string
		ExpectedClientCertName string
	}{
		"default": {
			CertificateName:        "default",
			ExpectedRootCertName:   "redpanda-default-root-certificate",
			ExpectedRootCertKey:    "tls.crt",
			ExpectedClientCertName: "redpanda-client",
		},
		"default with non-enabled global cert": {
			Cert: &redpanda.TLSCert{
				Enabled: ptr.To(false),
				SecretRef: &v1.LocalObjectReference{
					Name: "some-cert",
				},
			},
			CertificateName:        "default",
			ExpectedRootCertName:   "redpanda-default-root-certificate",
			ExpectedRootCertKey:    "tls.crt",
			ExpectedClientCertName: "redpanda-client",
		},
		"certificate with secret ref": {
			Cert: &redpanda.TLSCert{
				SecretRef: &v1.LocalObjectReference{
					Name: "some-cert",
				},
			},
			CertificateName:        "default",
			ExpectedRootCertName:   "some-cert",
			ExpectedRootCertKey:    "tls.crt",
			ExpectedClientCertName: "redpanda-client",
		},
		"certificate with CA": {
			Cert: &redpanda.TLSCert{
				CAEnabled: true,
				SecretRef: &v1.LocalObjectReference{
					Name: "some-cert",
				},
			},
			CertificateName:        "default",
			ExpectedRootCertName:   "some-cert",
			ExpectedRootCertKey:    "ca.crt",
			ExpectedClientCertName: "redpanda-client",
		},
		"certificate with client certificate": {
			Cert: &redpanda.TLSCert{
				CAEnabled: true,
				SecretRef: &v1.LocalObjectReference{
					Name: "some-cert",
				},
				ClientSecretRef: &v1.LocalObjectReference{
					Name: "client-cert",
				},
			},
			CertificateName:        "default",
			ExpectedRootCertName:   "some-cert",
			ExpectedRootCertKey:    "ca.crt",
			ExpectedClientCertName: "client-cert",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			certMap := redpanda.TLSCertMap{}

			if c.Cert != nil {
				certMap[c.CertificateName] = *c.Cert
			}

			dot, err := redpanda.Chart.Dot(kube.Config{}, helmette.Release{
				Name:      "redpanda",
				Namespace: "redpanda",
				Service:   "Helm",
			}, redpanda.Values{
				TLS: redpanda.TLS{
					Certs: certMap,
				},
			})
			require.NoError(t, err)

			actualRootCertName, actualRootCertKey, actualClientCertName := certificatesFor(dot, c.CertificateName)
			require.Equal(t, c.ExpectedRootCertName, actualRootCertName)
			require.Equal(t, c.ExpectedRootCertKey, actualRootCertKey)
			require.Equal(t, c.ExpectedClientCertName, actualClientCertName)
		})
	}
}
