package kube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

const (
	defaultClusterDomain = "cluster.local"
	defaultNamespace     = "default"
	svcLabel             = "svc"

	portForwardProtocolV1Name = "portforward.k8s.io"
)

var (
	ErrInvalidPodFQDN = errors.New("invalid pod FQDN")
	ErrNoPort         = errors.New("no port specified")

	negotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()
)

// PodDialer is a basic port-forwarding dialer that doesn't start
// any local listeners, but returns a net.Conn directly.
type PodDialer struct {
	config        *rest.Config
	clusterDomain string
	requestID     int

	connections map[types.NamespacedName]httpstream.Connection
	mutex       sync.RWMutex
}

// NewPodDialer create a PodDialer.
func NewPodDialer(config *rest.Config) *PodDialer {
	return &PodDialer{
		config:        config,
		clusterDomain: defaultClusterDomain,
		connections:   make(map[types.NamespacedName]httpstream.Connection),
	}
}

// WithClusterDomain overrides the domain used for FQDN parsing.
func (p *PodDialer) WithClusterDomain(domain string) *PodDialer {
	p.clusterDomain = domain
	return p
}

// Dial dials the given pod's service-based DNS address and returns a
// net.Conn that can be used to reach the pod directly.
func (p *PodDialer) Dial(network string, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("dialer only supports TCP-based networks: %w", net.UnknownNetworkError(network))
	}

	pod, port, err := p.parseDNS(address)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	conn, err := p.connectionForPod(pod)
	if err != nil {
		return nil, err
	}

	p.requestID++

	headers := http.Header{}
	headers.Set(corev1.PortHeader, strconv.Itoa(port))
	headers.Set(corev1.PortForwardRequestIDHeader, strconv.Itoa(p.requestID))

	headers.Set(corev1.StreamType, corev1.StreamTypeError)
	errorStream, err := conn.CreateStream(headers)
	if err != nil {
		return nil, err
	}

	headers.Set(corev1.StreamType, corev1.StreamTypeData)
	dataStream, err := conn.CreateStream(headers)
	if err != nil {
		return nil, err
	}

	return wrapConn(network, address, dataStream, errorStream), nil
}

// DialContext is a simple wrapper around Dial.
func (p *PodDialer) DialContext(_ context.Context, network string, address string) (net.Conn, error) {
	return p.Dial(network, address)
}

// parseDNS attempts to determine the intended pod to target, currently the
// following formats are supported:
//  1. [pod-name].[service-name].[namespace-name].svc.[cluster-domain]
//  2. [pod-name].[service-name].[namespace-name].svc
//  3. [pod-name].[service-name].[namespace-name]
//  4. [pod-name].[service-name]
//
// If no cluster-domain is supplied, the dialer's configured domain is assumed.
// If no namespace-name is supplied, the default namespace is assumed.
// No validation is done to ensure that the pod is actually referenced by the
// given service or that the DNS record exists in Kubernetes, instead this assumes
// that things are set up properly in Kubernetes such that the FQDN passed here
// matches Kubernetes' own DNS records and a pod within the cluster would be able
// to resolve the same FQDN to the pod that we do.
//
// These assumptions allow us to use this dialer to issues requests *as if* we are
// in the Kubernetes network even from outside of it (i.e. in tests that attempt to
// connect to a pod at a given DNS address).
func (p *PodDialer) parseDNS(fqdn string) (pod types.NamespacedName, port int, err error) {
	addressPort := strings.Split(fqdn, ":")
	if len(addressPort) != 2 {
		err = ErrNoPort
		return
	}

	port, err = strconv.Atoi(addressPort[1])
	if err != nil {
		return
	}

	fqdn = addressPort[0]
	fqdn = strings.TrimSuffix(fqdn, "."+p.clusterDomain)
	fqdn = strings.TrimSuffix(fqdn, "."+svcLabel)
	labels := strings.Split(fqdn, ".")

	// since we only dial pods we require 2 labels
	// (assuming we're trying to dial a pod in the
	// default namespace) or 3 (for a pod outside
	// of the default namespace)
	switch len(labels) {
	case 2:
		pod.Namespace = defaultNamespace
	case 3:
		pod.Namespace = labels[2]
	default:
		err = ErrInvalidPodFQDN
		return
	}

	// service - labels[1]
	pod.Name = labels[0]
	return
}

// mutex must be held while calling this
func (p *PodDialer) connectionForPod(pod types.NamespacedName) (httpstream.Connection, error) {
	if conn, ok := p.connections[pod]; ok {
		return conn, nil
	}

	transport, upgrader, err := spdy.RoundTripperFor(p.config)
	if err != nil {
		return nil, err
	}

	cfg := p.config
	cfg.APIPath = "/api"
	cfg.GroupVersion = &schema.GroupVersion{Version: "v1"}
	cfg.NegotiatedSerializer = negotiatedSerializer
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}

	req := restClient.Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	conn, protocol, err := dialer.Dial(portForwardProtocolV1Name)
	if err != nil {
		return nil, err
	}
	if protocol != portForwardProtocolV1Name {
		return nil, fmt.Errorf("unable to negotiate protocol: client supports %q, server returned %q", portForwardProtocolV1Name, protocol)
	}

	p.connections[pod] = conn
	return conn, nil
}

type conn struct {
	httpstream.Stream
	errorStream httpstream.Stream
	network     string
	remote      string

	errCh chan error
}

func wrapConn(network, remote string, s, err httpstream.Stream) *conn {
	c := &conn{
		Stream:      s,
		errorStream: err,
		network:     network,
		remote:      remote,
		errCh:       make(chan error, 1),
	}

	go c.pollErrors()
	return c
}

func (c *conn) pollErrors() {
	data, err := io.ReadAll(c.errorStream)
	if err != nil {
		c.writeError(err)
		return
	}

	if len(data) != 0 {
		c.writeError(errors.New(string(data)))
		return
	}
}

func (c *conn) writeError(err error) {
	select {
	case c.errCh <- err:
	default:
	}
}

func (c *conn) checkError() error {
	select {
	case err := <-c.errCh:
		return err
	default:
		return nil
	}
}

var _ net.Conn = (*conn)(nil)

func (c *conn) Read(data []byte) (int, error) {
	if err := c.checkError(); err != nil {
		return 0, err
	}

	n, err := c.Stream.Read(data)

	// prioritize any sort of checks propagated on
	// the error stream
	if err := c.checkError(); err != nil {
		return n, err
	}
	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	if err := c.checkError(); err != nil {
		return 0, err
	}

	n, err := c.Stream.Write(b)

	// prioritize any sort of checks propagated on
	// the error stream
	if err := c.checkError(); err != nil {
		return n, err
	}
	return n, err
}

func (c *conn) Close() error {
	closeErr := c.Reset()

	// prioritize any sort of checks propagated on
	// the error stream
	if err := c.checkError(); err != nil {
		return err
	}
	return closeErr
}

func (c *conn) SetDeadline(t time.Time) error {
	if conn, ok := c.Stream.(net.Conn); ok {
		return conn.SetDeadline(t)
	}
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	if conn, ok := c.Stream.(net.Conn); ok {
		return conn.SetReadDeadline(t)
	}
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	if conn, ok := c.Stream.(net.Conn); ok {
		return conn.SetWriteDeadline(t)
	}
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	return addr{c.network, "localhost:0"}
}

func (c *conn) RemoteAddr() net.Addr {
	return addr{c.network, c.remote}
}

type addr struct {
	Net  string
	Addr string
}

func (a addr) Network() string { return a.Net }
func (a addr) String() string  { return a.Addr }
