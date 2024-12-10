package redpanda

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/pkg/sr"

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

// DialContextFunc is a function that acts as a dialer for the underlying Kafka client.
type DialContextFunc = func(ctx context.Context, network, host string) (net.Conn, error)

// AdminClient creates a client to talk to a Redpanda cluster admin API based on its helm
// configuration over its internal listeners.
func AdminClient(dot *helmette.Dot, dialer DialContextFunc) (*rpadmin.AdminAPI, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)
	name := redpanda.Fullname(dot)
	domain := redpanda.InternalDomain(dot)
	prefix := "http://"

	var tlsConfig *tls.Config
	var err error

	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		prefix = "https://"

		tlsConfig, err = tlsConfigFromDot(dot, values.Listeners.Admin.TLS)
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

	hosts := redpanda.ServerList(values.Statefulset.Replicas, prefix, name, domain, values.Listeners.Admin.Port)

	return rpadmin.NewAdminAPIWithDialer(hosts, auth, tlsConfig, dialer)
}

// SchemaRegistryClient creates a client to talk to a Redpanda cluster admin API based on its helm
// configuration over its internal listeners.
func SchemaRegistryClient(dot *helmette.Dot, dialer DialContextFunc, opts ...sr.ClientOpt) (*sr.Client, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)
	name := redpanda.Fullname(dot)
	domain := redpanda.InternalDomain(dot)
	prefix := "http://"

	// These transport values come from the TLS client options found here:
	// https://github.com/twmb/franz-go/blob/cea7aa5d803781e5f0162187795482ba1990c729/pkg/sr/clientopt.go#L48-L68
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		DialContext:           dialer,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if dialer == nil {
		transport.DialContext = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		prefix = "https://"

		tlsConfig, err := tlsConfigFromDot(dot, values.Listeners.SchemaRegistry.TLS)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}

	copts := []sr.ClientOpt{sr.HTTPClient(&http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	})}

	username, password, _, err := authFromDot(dot)
	if err != nil {
		return nil, err
	}

	if username != "" {
		copts = append(copts, sr.BasicAuth(username, password))
	}

	hosts := redpanda.ServerList(values.Statefulset.Replicas, prefix, name, domain, values.Listeners.SchemaRegistry.Port)
	copts = append(copts, sr.URLs(hosts...))

	// finally, override any calculated client opts with whatever was
	// passed in
	return sr.NewClient(append(copts, opts...)...)
}

// KafkaClient creates a client to talk to a Redpanda cluster based on its helm
// configuration over its internal listeners.
func KafkaClient(dot *helmette.Dot, dialer DialContextFunc, opts ...kgo.Opt) (*kgo.Client, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)
	name := redpanda.Fullname(dot)
	domain := redpanda.InternalDomain(dot)

	brokers := redpanda.ServerList(values.Statefulset.Replicas, "", name, domain, values.Listeners.Kafka.Port)

	opts = append(opts, kgo.SeedBrokers(brokers...))

	if values.Listeners.Kafka.TLS.IsEnabled(&values.TLS) {
		tlsConfig, err := tlsConfigFromDot(dot, values.Listeners.Kafka.TLS)
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
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	bootstrapUser := redpanda.SecretBootstrapUser(dot)

	if bootstrapUser != nil {
		// if we have any errors grabbing the credentials from the bootstrap user
		// then we'll just fallback to the superuser parsing code
		user, found, lookupErr := helmette.SafeLookup[corev1.Secret](dot, bootstrapUser.Namespace, bootstrapUser.Name)
		if lookupErr == nil && found {
			selector := values.Auth.SASL.BootstrapUser.SecretKeySelector(redpanda.Fullname(dot))
			mechanism := values.Auth.SASL.BootstrapUser.GetMechanism()
			if data, found := user.Data[selector.Key]; found {
				return values.Auth.SASL.BootstrapUser.Username(), string(data), mechanism, nil
			}
		}
	}

	saslUsers := redpanda.SecretSASLUsers(dot)
	saslUsersError := func(err error) error {
		return fmt.Errorf("error fetching SASL authentication for %s/%s: %w", saslUsers.Namespace, saslUsers.Name, err)
	}

	if saslUsers != nil {
		// read from the server since we're assuming all the resources
		// have already been created
		users, found, lookupErr := helmette.SafeLookup[corev1.Secret](dot, saslUsers.Namespace, saslUsers.Name)
		if lookupErr != nil {
			err = saslUsersError(lookupErr)
			return
		}

		if !found {
			err = saslUsersError(ErrSASLSecretNotFound)
			return
		}

		data, found := users.Data["users.txt"]
		if !found {
			err = saslUsersError(ErrSASLSecretKeyNotFound)
			return
		}

		username, password, mechanism = firstUser(data)
		if username == "" {
			err = saslUsersError(ErrSASLSecretSuperuserNotFound)
			return
		}
	}

	return
}

func certificatesFor(dot *helmette.Dot, cert string) (certSecret, certKey, clientSecret string) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	name := redpanda.Fullname(dot)

	// default to cert manager issued names and tls.crt which is
	// where cert-manager outputs the root CA
	certKey = corev1.TLSCertKey
	certSecret = fmt.Sprintf("%s-%s-root-certificate", name, cert)
	clientSecret = fmt.Sprintf("%s-client", name)

	if certificate, ok := values.TLS.Certs[cert]; ok {
		// if this references a non-enabled certificate, just return
		// the default cert-manager issued names
		if certificate.Enabled != nil && !*certificate.Enabled {
			return certSecret, certKey, clientSecret
		}

		if certificate.ClientSecretRef != nil {
			clientSecret = certificate.ClientSecretRef.Name
		}
		if certificate.SecretRef != nil {
			certSecret = certificate.SecretRef.Name
			if certificate.CAEnabled {
				certKey = "ca.crt"
			}
		}
	}
	return certSecret, certKey, clientSecret
}

func tlsConfigFromDot(dot *helmette.Dot, listener redpanda.InternalTLS) (*tls.Config, error) {
	namespace := dot.Release.Namespace
	serverName := redpanda.InternalDomain(dot)

	rootCertName, rootCertKey, clientCertName := certificatesFor(dot, listener.Cert)

	serverTLSError := func(err error) error {
		return fmt.Errorf("error fetching server root CA %s/%s: %w", namespace, rootCertName, err)
	}
	clientTLSError := func(err error) error {
		return fmt.Errorf("error fetching client certificate default/%s: %w", clientCertName, err)
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}

	serverCert, found, lookupErr := helmette.SafeLookup[corev1.Secret](dot, namespace, rootCertName)
	if lookupErr != nil {
		return nil, serverTLSError(lookupErr)
	}

	if !found {
		return nil, serverTLSError(ErrServerCertificateNotFound)
	}

	serverPublicKey, found := serverCert.Data[rootCertKey]
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

	if listener.RequireClientAuth {
		clientCert, found, lookupErr := helmette.SafeLookup[corev1.Secret](dot, namespace, clientCertName)
		if lookupErr != nil {
			return nil, clientTLSError(lookupErr)
		}

		if !found {
			return nil, clientTLSError(ErrServerCertificateNotFound)
		}

		// we always use tls.crt for client certs
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

func firstUser(data []byte) (user string, password string, mechanism string) {
	file := string(data)

	for _, line := range strings.Split(file, "\n") {
		tokens := strings.Split(line, ":")

		switch len(tokens) {
		case 2:
			return tokens[0], tokens[1], redpanda.DefaultSASLMechanism

		case 3:
			if !slices.Contains(supportedSASLMechanisms, tokens[2]) {
				continue
			}

			return tokens[0], tokens[1], tokens[2]

		default:
			continue
		}
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

func wrapTLSDialer(dialer DialContextFunc, config *tls.Config) DialContextFunc {
	return func(ctx context.Context, network, host string) (net.Conn, error) {
		conn, err := dialer(ctx, network, host)
		if err != nil {
			return nil, err
		}
		return tls.Client(conn, config), nil
	}
}
