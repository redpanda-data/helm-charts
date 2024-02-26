package kube

import (
	"context"
	"io"

	"github.com/cockroachdb/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
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
