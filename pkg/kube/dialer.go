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
	"sync/atomic"
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

type refCountedConnection struct {
	httpstream.Connection
	references int
}

// PodDialer is a basic port-forwarding dialer that doesn't start
// any local listeners, but returns a net.Conn directly.
type PodDialer struct {
	config        *rest.Config
	clusterDomain string
	requestID     int

	connections map[types.NamespacedName]*refCountedConnection
	mutex       sync.RWMutex
}

// NewPodDialer create a PodDialer.
func NewPodDialer(config *rest.Config) *PodDialer {
	return &PodDialer{
		config:        config,
		clusterDomain: defaultClusterDomain,
		connections:   make(map[types.NamespacedName]*refCountedConnection),
	}
}

// WithClusterDomain overrides the domain used for FQDN parsing.
func (p *PodDialer) WithClusterDomain(domain string) *PodDialer {
	p.clusterDomain = domain
	return p
}

// cleanupConnection should be called via a function callback
// any time that one of its underlying streams is closed
// when it's called, it decrements the reference counted connection
// pruning it from our connection map when its references hit 0.
func (p *PodDialer) cleanupConnection(pod types.NamespacedName) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	connection, ok := p.connections[pod]
	if !ok {
		return
	}

	connection.references--

	if connection.references == 0 {
		connection.Close()

		delete(p.connections, pod)
	}
}

// DialContext dials the given pod's service-based DNS address and returns a
// net.Conn that can be used to reach the pod directly. It uses the passed in
// context to close the underlying connection when
func (p *PodDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
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

	conn, err := p.connectionForPodLocked(pod)
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
		// close off the error stream that's been opened
		errorStream.Reset()

		return nil, err
	}

	conn.references++
	onClose := func() {
		p.cleanupConnection(pod)
	}

	return wrapConn(ctx, onClose, network, address, dataStream, errorStream), nil
}

// parseDNS attempts to determine the intended pod to target, currently the
// following formats are supported for resolution of "service-based" hostnames:
//  1. [pod-name].[service-name].[namespace-name].svc.[cluster-domain]
//  2. [pod-name].[service-name].[namespace-name].svc
//  3. [pod-name].[service-name].[namespace-name]
//
// If no cluster-domain is supplied, the dialer's configured domain is assumed.
//
// No validation is done to ensure that the pod is actually referenced by the
// given service or that the DNS record exists in Kubernetes, instead this assumes
// that things are set up properly in Kubernetes such that the FQDN passed here
// matches Kubernetes' own DNS records and a pod within the cluster would be able
// to resolve the same FQDN to the pod that we do.
//
// These assumptions allow us to use this dialer to issues requests *as if* we are
// in the Kubernetes network even from outside of it (i.e. in tests that attempt to
// connect to a pod at a given hostname).
//
// The implementation of parsing for shorter, non-`svc` suffixed domains does not
// follow the typical service DNS scheme. Rather it allows for the following custom
// pod direct-dialing strategy:
//
//  4. [pod-name].[namespace-name]
//  5. [pod-name]
//
// If no namespace-name is supplied, the default namespace is assumed.
func (p *PodDialer) parseDNS(fqdn string) (types.NamespacedName, int, error) {
	var pod types.NamespacedName

	addressPort := strings.Split(fqdn, ":")
	if len(addressPort) != 2 {
		return pod, 0, ErrNoPort
	}

	port, err := strconv.Atoi(addressPort[1])
	if err != nil {
		return pod, 0, ErrNoPort
	}

	fqdn = addressPort[0]

	isServiceDNS := true

	if strings.Count(fqdn, ".") < 2 {
		// we have a direct pod DNS address
		isServiceDNS = false
	} else {
		// we have a service-based DNS address
		fqdn = strings.TrimSuffix(fqdn, "."+p.clusterDomain)
		fqdn = strings.TrimSuffix(fqdn, "."+svcLabel)
	}

	labels := strings.Split(fqdn, ".")

	// since we only dial pods we require 2 labels
	// (assuming we're trying to dial a pod in the
	// default namespace) or 3 (for a pod outside
	// of the default namespace)
	switch len(labels) {
	case 1:
		if !isServiceDNS {
			pod.Namespace = defaultNamespace
		} else {
			return pod, 0, ErrInvalidPodFQDN
		}
	case 2:
		if isServiceDNS {
			pod.Namespace = defaultNamespace
		} else {
			pod.Namespace = labels[1]
		}
	case 3:
		pod.Namespace = labels[2]
	default:
		return pod, 0, ErrInvalidPodFQDN
	}

	pod.Name = labels[0]

	return pod, port, nil
}

func (p *PodDialer) connectionForPodLocked(pod types.NamespacedName) (*refCountedConnection, error) {
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
		if conn != nil {
			conn.Close()
		}

		return nil, fmt.Errorf("unable to negotiate protocol: client supports %q, server returned %q", portForwardProtocolV1Name, protocol)
	}

	refCountedConn := &refCountedConnection{Connection: conn}

	p.connections[pod] = refCountedConn
	return refCountedConn, nil
}

type conn struct {
	dataStream  httpstream.Stream
	errorStream httpstream.Stream
	network     string
	remote      string
	onClose     func()

	errCh  chan error
	stopCh chan error
	closed atomic.Bool
}

var _ net.Conn = (*conn)(nil)

func wrapConn(ctx context.Context, onClose func(), network, remote string, s, err httpstream.Stream) *conn {
	c := &conn{
		dataStream:  s,
		errorStream: err,
		network:     network,
		remote:      remote,
		onClose:     onClose,
		errCh:       make(chan error, 1),
		stopCh:      make(chan error),
	}

	go c.pollErrors()
	go c.checkCancelation(ctx)

	return c
}

func (c *conn) checkCancelation(ctx context.Context) {
	select {
	case <-ctx.Done():
		c.writeError(ctx.Err())
		c.Close()
	case <-c.stopCh:
	}
}

func (c *conn) pollErrors() {
	defer c.Close()

	data, err := io.ReadAll(c.errorStream)
	if err != nil {
		c.writeError(err)
		return
	}

	if len(data) != 0 {
		c.writeError(fmt.Errorf("received error message from error stream: %s", string(data)))
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
	if c.closed.Load() {
		return net.ErrClosed
	}

	select {
	case err := <-c.errCh:
		return err
	default:
		return nil
	}
}

func (c *conn) Read(data []byte) (int, error) {
	if err := c.checkError(); err != nil {
		return 0, err
	}

	n, err := c.dataStream.Read(data)

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

	n, err := c.dataStream.Write(b)

	// prioritize any sort of checks propagated on
	// the error stream
	if err := c.checkError(); err != nil {
		return n, err
	}
	return n, err
}

func (c *conn) Close() error {
	// make Close idempotent since we may close off
	// the stream when a context is canceled but also
	// may have had Close called manually
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	// call our onClose cleanup handler
	defer c.onClose()

	// signal to any underlying goroutines that we are
	// stopping
	defer close(c.stopCh)

	// closing the underlying connection should cause
	// our error stream reading routine to stop
	c.errorStream.Reset()
	closeErr := c.dataStream.Reset()

	// prioritize any sort of checks propagated on
	// the error stream
	if err := c.checkError(); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			return err
		}
	}
	return closeErr
}

func (c *conn) SetDeadline(t time.Time) error {
	if conn, ok := c.dataStream.(net.Conn); ok {
		return conn.SetDeadline(t)
	}
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	if conn, ok := c.dataStream.(net.Conn); ok {
		return conn.SetReadDeadline(t)
	}
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	if conn, ok := c.dataStream.(net.Conn); ok {
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
