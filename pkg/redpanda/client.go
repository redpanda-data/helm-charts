package redpanda

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrServerCertificateNotFound          = errors.New("kafka TLS certificate not found")
	ErrServerCertificatePublicKeyNotFound = errors.New("kafka TLS certificate does not contain a public key")

	ErrClientCertificateNotFound           = errors.New("client TLS certificate not found")
	ErrClientCertificatePublicKeyNotFound  = errors.New("client TLS certificate does not contain a public key")
	ErrClientCertificatePrivateKeyNotFound = errors.New("client TLS certificate does not contain a private key")

	ErrSASLSecretNotFound          = errors.New("users secret not found")
	ErrSASLSecretKeyNotFound       = errors.New("users secret key not found")
	ErrSASLSecretSuperuserNotFound = errors.New("users secret has no users")

	supportedSASLMechanisms = []string{
		"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512",
	}
)

func firstUser(data []byte) (user string, password string, mechanism string) {
	file := string(data)

	for _, line := range strings.Split(file, "\n") {
		tokens := strings.Split(line, ":")
		if len(tokens) != 3 {
			continue
		}

		if !slices.Contains(supportedSASLMechanisms, tokens[2]) {
			continue
		}

		user, password, mechanism = tokens[0], tokens[1], tokens[2]
		return
	}

	return
}

func saslOpt(user, password, mechanism string) kgo.Opt {
	var m sasl.Mechanism
	switch mechanism {
	case "PLAIN":
		m = plain.Auth{User: user, Pass: password}.AsMechanism()
	case "SCRAM-SHA-256", "SCRAM-SHA-512":
		scram := scram.Auth{User: user, Pass: password}

		switch mechanism {
		case "SCRAM-SHA-256":
			m = scram.AsSha256Mechanism()
		case "SCRAM-SHA-512":
			m = scram.AsSha512Mechanism()
		}
	default:
		panic(fmt.Sprintf("unhandled SASL mechanism: %s", mechanism))
	}

	return kgo.SASL(m)
}

// KafkaClient creates a client to talk to a Redpanda cluster based on its helm
// configuration over its internal listeners.
func KafkaClient(release helmette.Release, partial redpanda.PartialValues, opts ...kgo.Opt) (*kgo.Client, error) {
	dot, err := redpanda.Dot(release, partial)
	if err != nil {
		return nil, err
	}

	values := helmette.Unwrap[redpanda.Values](dot.Values)

	name := redpanda.Fullname(dot)
	namespace := dot.Release.Namespace
	serviceName := redpanda.ServiceName(dot)
	saslUsers := redpanda.SecretSASLUsers(dot)
	clientCertName := fmt.Sprintf("%s-client", name)
	kafkaRootCertName := fmt.Sprintf("%s-%s-root-certificate", name, values.Listeners.Kafka.TLS.Cert)

	brokers := []string{}
	for i := int32(0); i < values.Statefulset.Replicas; i++ {
		brokers = append(brokers, fmt.Sprintf("%s-%d.%s.%s.svc:%d", name, i, serviceName, namespace, values.Listeners.Kafka.Port))
	}

	opts = append(opts, kgo.SeedBrokers(brokers...))

	serverTLSError := func(err error) error {
		return fmt.Errorf("error fetching server root CA %s/%s: %w", namespace, kafkaRootCertName, err)
	}
	clientTLSError := func(err error) error {
		return fmt.Errorf("error fetching client certificate default/%s: %w", clientCertName, err)
	}
	saslError := func(err error) error {
		return fmt.Errorf("error fetching SASL authentication for %s/%s: %w", saslUsers.Namespace, saslUsers.Name, err)
	}

	if redpanda.TLSEnabled(dot) {
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

		serverCert, found := helmette.Lookup[corev1.Secret](dot, namespace, kafkaRootCertName)
		if !found {
			return nil, serverTLSError(ErrServerCertificateNotFound)
		}

		serverPublicKey, found := serverCert.Data[corev1.TLSCertKey]
		if !found {
			return nil, serverTLSError(ErrServerCertificatePublicKeyNotFound)
		}

		block, _ := pem.Decode(serverPublicKey)
		serverParsedCertificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, serverTLSError(fmt.Errorf("unable to parse public key %w", err))
		}
		pool := x509.NewCertPool()
		pool.AddCert(serverParsedCertificate)

		tlsConfig.RootCAs = pool

		if redpanda.ClientAuthRequired(dot) {
			clientCert, found := helmette.Lookup[corev1.Secret](dot, "default", clientCertName)
			if !found {
				return nil, clientTLSError(ErrServerCertificateNotFound)
			}

			clientPublicKey, found := clientCert.Data[corev1.TLSCertKey]
			if !found {
				return nil, clientTLSError(ErrClientCertificatePublicKeyNotFound)
			}

			clientPrivateKey, found := clientCert.Data[corev1.TLSPrivateKeyKey]
			if !found {
				return nil, clientTLSError(ErrClientCertificatePrivateKeyNotFound)
			}

			clientKey, err := tls.X509KeyPair(clientPublicKey, clientPrivateKey)
			if err != nil {
				return nil, clientTLSError(fmt.Errorf("unable to parse public and private key %w", err))
			}

			tlsConfig.Certificates = []tls.Certificate{clientKey}
		}

		opts = append(opts, kgo.DialTLSConfig(tlsConfig))
	}

	if saslUsers != nil {
		// read from the server since we're assuming all the resources
		// have already been created
		users, found := helmette.Lookup[corev1.Secret](dot, saslUsers.Namespace, saslUsers.Name)
		if !found {
			return nil, saslError(ErrSASLSecretNotFound)
		}

		data, found := users.Data["users.txt"]
		if !found {
			return nil, saslError(ErrSASLSecretKeyNotFound)
		}

		user, password, mechanism := firstUser(data)
		if user == "" {
			return nil, saslError(ErrSASLSecretSuperuserNotFound)
		}

		opts = append(opts, saslOpt(user, password, mechanism))
	}

	return kgo.NewClient(opts...)
}
