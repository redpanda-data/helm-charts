package tlsgeneration

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/cockroachdb/errors"
)

func ClientServerCertificate(chartReleaseName, chartReleaseNamespace string) (ca, serverPublic, serverPrivate, clientPublic, clientPrivate []byte, err error) {
	now := time.Now()

	rootCASubject := pkix.Name{
		CommonName:   "test.example.com",
		Organization: []string{"Σ Acme Co"},
		Country:      []string{"US"},
	}
	root := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      rootCASubject,
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),

		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  true,

		DNSNames:       []string{"test.example.com"},
		EmailAddresses: []string{"gopher@golang.org"},
	}

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, root, root, priv.Public(), priv)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	root, err = x509.ParseCertificate(derBytes)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	ca = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	commonName := fmt.Sprintf("%s.%s.svc.cluster.local", chartReleaseName, chartReleaseNamespace)
	shortTestName := fmt.Sprintf("%s.%s", chartReleaseName, chartReleaseNamespace)
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Σ Acme Co"},
			Country:      []string{"US"},
		},
		Issuer:    rootCASubject,
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(time.Hour),

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  false,

		DNSNames: []string{
			shortTestName,
			commonName,
			fmt.Sprintf("%s.", commonName),
			fmt.Sprintf("*.%s", commonName),
			fmt.Sprintf("*.%s.", commonName),
		},
		EmailAddresses: []string{"gopher@golang.org"},
	}

	privServer, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	derServerBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, root, privServer.Public(), priv)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privServer)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	serverPublic = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derServerBytes})
	serverPrivate = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "redpanda-client",
			Organization: []string{"Σ Acme Co"},
			Country:      []string{"US"},
		},
		Issuer:    rootCASubject,
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(time.Hour),

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  false,

		DNSNames:       []string{"redpanda-client"},
		EmailAddresses: []string{"gopher@golang.org"},
	}

	privClient, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	derClientBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, root, privClient.Public(), priv)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	privBytes, err = x509.MarshalPKCS8PrivateKey(privClient)

	clientPublic = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derClientBytes})
	clientPrivate = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return
}
