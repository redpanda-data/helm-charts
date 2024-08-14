package redpanda

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/scram"

	"github.com/redpanda-data/common-go/rpadmin"

	corev1 "k8s.io/api/core/v1"
)

var (
	ErrServerCertificateNotFound          = errors.New("server TLS certificate not found")
	ErrServerCertificatePublicKeyNotFound = errors.New("server TLS certificate does not contain a public key")

	ErrClientCertificateNotFound           = errors.New("client TLS certificate not found")
	ErrClientCertificatePublicKeyNotFound  = errors.New("client TLS certificate does not contain a public key")
	ErrClientCertificatePrivateKeyNotFound = errors.New("client TLS certificate does not contain a private key")

	ErrSASLSecretNotFound          = errors.New("users secret not found")
	ErrSASLSecretKeyNotFound       = errors.New("users secret key not found")
	ErrSASLSecretSuperuserNotFound = errors.New("users secret has no users")

	supportedSASLMechanisms = []string{
		"SCRAM-SHA-256", "SCRAM-SHA-512",
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

// DialContextFunc is a function that acts as a dialer for the underlying Kafka client.
type DialContextFunc = func(ctx context.Context, network, host string) (net.Conn, error)

func wrapTLSDialer(dialer DialContextFunc, config *tls.Config) DialContextFunc {
	return func(ctx context.Context, network, host string) (net.Conn, error) {
		conn, err := dialer(ctx, network, host)
		if err != nil {
			return nil, err
		}
		return tls.Client(conn, config), nil
	}
}

// AdminClient creates a client to talk to a Redpanda cluster admin API based on its helm
// configuration over its internal listeners.
func AdminClient(config kube.Config, release helmette.Release, partial redpanda.PartialValues, dialer DialContextFunc) (*rpadmin.AdminAPI, error) {
	dot, err := redpanda.Dot(release, partial)
	if err != nil {
		return nil, err
	}
	dot.KubeConfig = config

	values := helmette.Unwrap[redpanda.Values](dot.Values)
	prefix := "http://"

	var tlsConfig *tls.Config

	if redpanda.TLSEnabled(dot) {
		prefix = "https://"

		tlsConfig, err = tlsConfigFromDot(dot, values.Listeners.Kafka.TLS.Cert)
		if err != nil {
			return nil, err
		}
	}

	var auth rpadmin.Auth
	username, password, _, err := authFromDot(dot)
	if err != nil {
		return nil, err
	}

	if username != "" {
		auth = &rpadmin.BasicAuth{
			Username: username,
			Password: password,
		}
	} else {
		auth = &rpadmin.NopAuth{}
	}

	hosts := urlsFromDot(dot, prefix, values.Statefulset.Replicas, values.Listeners.Admin.Port)
	return rpadmin.NewAdminAPIWithDialer(hosts, auth, tlsConfig, dialer)
}

// KafkaClient creates a client to talk to a Redpanda cluster based on its helm
// configuration over its internal listeners.
func KafkaClient(config kube.Config, release helmette.Release, partial redpanda.PartialValues, dialer DialContextFunc, opts ...kgo.Opt) (*kgo.Client, error) {
	dot, err := redpanda.Dot(release, partial)
	if err != nil {
		return nil, err
	}
	dot.KubeConfig = config

	values := helmette.Unwrap[redpanda.Values](dot.Values)
	brokers := urlsFromDot(dot, "", values.Statefulset.Replicas, values.Listeners.Kafka.Port)

	opts = append(opts, kgo.SeedBrokers(brokers...))

	if redpanda.TLSEnabled(dot) {
		tlsConfig, err := tlsConfigFromDot(dot, values.Listeners.Kafka.TLS.Cert)
		if err != nil {
			return nil, err
		}

		// we can only specify one of DialTLSConfig or Dialer
		if dialer == nil {
			opts = append(opts, kgo.DialTLSConfig(tlsConfig))
		} else {
			opts = append(opts, kgo.Dialer(wrapTLSDialer(dialer, tlsConfig)))
		}
	} else if dialer != nil {
		opts = append(opts, kgo.Dialer(dialer))
	}

	username, password, mechanism, err := authFromDot(dot)
	if err != nil {
		return nil, err
	}

	if username != "" {
		opts = append(opts, saslOpt(username, password, mechanism))
	}

	return kgo.NewClient(opts...)
}

func authFromDot(dot *helmette.Dot) (username string, password string, mechanism string, err error) {
	saslUsers := redpanda.SecretSASLUsers(dot)

	saslError := func(err error) error {
		return fmt.Errorf("error fetching SASL authentication for %s/%s: %w", saslUsers.Namespace, saslUsers.Name, err)
	}

	if saslUsers != nil {
		// read from the server since we're assuming all the resources
		// have already been created
		users, found := helmette.Lookup[corev1.Secret](dot, saslUsers.Namespace, saslUsers.Name)
		if !found {
			err = saslError(ErrSASLSecretNotFound)
			return
		}

		data, found := users.Data["users.txt"]
		if !found {
			err = saslError(ErrSASLSecretKeyNotFound)
			return
		}

		username, password, mechanism = firstUser(data)
		if username == "" {
			err = saslError(ErrSASLSecretSuperuserNotFound)
			return
		}
	}

	return
}

func urlsFromDot(dot *helmette.Dot, prefix string, replicas int32, port int32) []string {
	name := redpanda.Fullname(dot)
	namespace := dot.Release.Namespace
	serviceName := redpanda.ServiceName(dot)

	urls := []string{}
	for i := int32(0); i < replicas; i++ {
		urls = append(urls, fmt.Sprintf("%s%s-%d.%s.%s.svc:%d", prefix, name, i, serviceName, namespace, port))
	}

	return urls
}

func tlsConfigFromDot(dot *helmette.Dot, cert string) (*tls.Config, error) {
	name := redpanda.Fullname(dot)
	namespace := dot.Release.Namespace
	serviceName := redpanda.ServiceName(dot)
	clientCertName := fmt.Sprintf("%s-client", name)
	rootCertName := fmt.Sprintf("%s-%s-root-certificate", name, cert)
	serverName := fmt.Sprintf("%s.%s.svc", serviceName, namespace)

	serverTLSError := func(err error) error {
		return fmt.Errorf("error fetching server root CA %s/%s: %w", namespace, rootCertName, err)
	}
	clientTLSError := func(err error) error {
		return fmt.Errorf("error fetching client certificate default/%s: %w", clientCertName, err)
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}

	serverCert, found := helmette.Lookup[corev1.Secret](dot, namespace, rootCertName)
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

	return tlsConfig, nil
}
