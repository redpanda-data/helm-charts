package kube

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type (
	Object    = client.Object
	ObjectKey = client.ObjectKey
)

// FromEnv returns a [Ctl] for the default context in $KUBECONFIG.
func FromEnv() (*Ctl, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return &Ctl{
		config: config,
		client: c,
	}, nil
}

func FromConfig(cfg Config) (*Ctl, error) {
	rest, err := ConfigToRest(cfg)
	if err != nil {
		return nil, err
	}
	return FromRESTConfig(rest)
}

func FromRESTConfig(cfg *RESTConfig) (*Ctl, error) {
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	return &Ctl{config: cfg, client: c}, nil
}

// Ctl is a Kubernetes client inspired by the shape of the `kubectl` CLI with a
// focus on being ergonomic.
type Ctl struct {
	config *rest.Config
	client client.Client
}

// RestConfig returns a deep copy of the [rest.Config] used by this [Ctl].
func (c *Ctl) RestConfig() *rest.Config {
	return rest.CopyConfig(c.config)
}

// Get fetches the latest state of an object into `obj` from Kubernetes.
// Usage:
//
//	var pod corev1.Pod
//	ctl.Get(ctx, kube.ObjectKey{Namespace: "", Name:""}, &pod)
func (c *Ctl) Get(ctx context.Context, key ObjectKey, obj Object) error {
	if err := c.client.Get(ctx, key, obj); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Ctl) Create(ctx context.Context, obj Object) error {
	if err := c.client.Create(ctx, obj); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Ctl) Delete(ctx context.Context, obj Object) error {
	if err := c.client.Delete(ctx, obj); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type ExecOptions struct {
	Container string
	Command   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
}

// Exec runs `kubectl exec` on the given Pod in the style of [exec.Command].
func (c *Ctl) Exec(ctx context.Context, pod *corev1.Pod, opts ExecOptions) error {
	if opts.Container == "" {
		opts.Container = pod.Spec.Containers[0].Name
	}

	// Apparently, nothing in the k8s SDK, except exec'ing, uses RESTClientFor.
	// RESTClientFor checks for GroupVersion and NegotiatedSerializer which are
	// never set by the config loading tool chain.
	// The .APIPath setting was a random shot in the dark that happened to work...
	// Pulled from https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/cmd/util/kubectl_match_version.go#L115
	cfg := c.RestConfig()
	cfg.APIPath = "/api"
	cfg.GroupVersion = &schema.GroupVersion{Version: "v1"}
	cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	// Inspired by https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/cmd/exec/exec.go#L388
	req := restClient.Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		TTY:       false,
	}, runtime.NewParameterCodec(c.client.Scheme()))

	// TODO(chrisseto): SPDY is reported to be deprecated but
	// NewWebSocketExecutor doesn't appear to work in our version of KinD.
	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	// exec, err := remotecommand.NewWebSocketExecutor(c.config, "GET", req.URL().String())
	if err != nil {
		return errors.WithStack(err)
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stderr: opts.Stderr,
		Stdout: opts.Stdout,
		Stdin:  opts.Stdin,
	})
}

func (c *Ctl) PortForward(ctx context.Context, pod *corev1.Pod, out, errOut io.Writer) ([]portforward.ForwardedPort, func(), error) {
	// Apparently, nothing in the k8s SDK, except exec'ing, uses RESTClientFor.
	// RESTClientFor checks for GroupVersion and NegotiatedSerializer which are
	// never set by the config loading tool chain.
	// The .APIPath setting was a random shot in the dark that happened to work...
	// Pulled from https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/cmd/util/kubectl_match_version.go#L115
	cfg := c.RestConfig()
	cfg.APIPath = "/api"
	cfg.GroupVersion = &schema.GroupVersion{Version: "v1"}
	cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Inspired by https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/cmd/portforward/portforward.go#L410-L416
	req := restClient.Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	var ports []string
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			// port forward and spdy does not handle UDP connection correctly
			//
			// Reference
			// https://github.com/kubernetes/kubernetes/issues/47862
			// https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/cmd/portforward/portforward.go#L273-L290
			if port.Protocol != corev1.ProtocolTCP {
				continue
			}

			ports = append(ports, fmt.Sprintf(":%d", port.ContainerPort))
		}
	}

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			fmt.Fprintf(errOut, "failed while forwaring ports: %v\n", err)
		}
	}()

	select {
	case <-fw.Ready:
	case <-ctx.Done():
	}

	p, err := fw.GetPorts()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return p, func() {
		if stopChan != nil {
			close(stopChan)
		}
	}, nil
}
