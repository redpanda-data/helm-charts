package helmtest

import (
	"context"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Env represents a test environment consisting of a Kubernetes cluster,
// helm client, and cert-manager installation.
//
// Env's are expected to be [Setup] infrequently. Ideally, once per test
// package and then reused via [Env.Namespaced] and [testing.T.Run].
type Env struct {
	helm *helm.Client
	ctl  *kube.Ctl
}

// Setup creates a new [Env] using whatever cluster is available in KUBECONFIG.
func Setup(t *testing.T) *Env {
	ctl, err := kube.FromEnv()
	require.NoError(t, err)

	client, err := helm.New(helm.Options{
		KubeConfig: ctl.RestConfig(),
		ConfigHome: testutil.TempDir(t),
	})
	require.NoError(t, err)

	require.NoError(t, ensureCertManager(context.Background(), client))

	return &Env{ctl: ctl, helm: client}
}

func (e *Env) Ctl() *kube.Ctl {
	return e.ctl
}

// Namespaced creates a sandboxed Kubernetes namespace that will be cleaned up
// at the end of `t`. Usage:
//
//	t.Run('subtest', func(t *testing.T) {
//		env := env.Namespaced(t)
//		// Testing....
//	})
func (e *Env) Namespaced(t *testing.T) *NamespacedEnv {
	ns := tempNamespace(t, context.Background(), e.ctl)
	t.Logf("using namespace %q", ns.Name)
	return &NamespacedEnv{t: t, env: e, namespace: ns}
}

// NamespacedEnv is effectively an [Env] that is bound to a specific Kubernetes
// Namespace.
// It exposes convenience methods for common operations without the need to set
// .Namespace or check errors.
type NamespacedEnv struct {
	t         *testing.T
	env       *Env
	namespace *corev1.Namespace
}

func (e *NamespacedEnv) Namespace() string {
	return e.namespace.Name
}

func (e *NamespacedEnv) Ctl() *kube.Ctl {
	return e.env.Ctl()
}

func (e *NamespacedEnv) Install(chart string, opts helm.InstallOptions) helm.Release {
	require.Zero(e.t, opts.Name, ".Name may not be specified")
	require.Zero(e.t, opts.Namespace, ".Namespace may not be specified")
	require.False(e.t, opts.CreateNamespace, ".CreateNamespace may not be specified")

	opts.Namespace = e.namespace.Name

	release, err := e.env.helm.Install(context.Background(), chart, opts)
	require.NoError(e.t, err)
	return release
}

func (e *NamespacedEnv) Test(release helm.Release) {
	require.NoError(e.t, e.env.helm.Test(context.Background(), release))
}

func (e *NamespacedEnv) Upgrade(chart string, release helm.Release, opts helm.UpgradeOptions) helm.Release {
	require.Zero(e.t, opts.Namespace, ".Namespace may not be specified")
	require.False(e.t, opts.CreateNamespace, ".CreateNamespace may not be specified")

	opts.Namespace = e.namespace.Name

	release, err := e.env.helm.Upgrade(context.Background(), release.Name, chart, opts)
	require.NoError(e.t, err)
	return release
}

func tempNamespace(t *testing.T, ctx context.Context, ctl *kube.Ctl) *corev1.Namespace {
	ns, err := kube.Create(ctx, ctl, corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
		},
		Spec: corev1.NamespaceSpec{},
	})
	require.NoError(t, err)

	testutil.MaybeCleanup(t, func() {
		require.NoError(t, kube.Delete[corev1.Namespace](ctx, ctl, kube.AsKey(ns)))
	})

	return ns
}

func ensureCertManager(ctx context.Context, client *helm.Client) error {
	if err := client.RepoAdd(ctx, "jetstack", "https://charts.jetstack.io"); err != nil {
		return err
	}

	_, err := client.Upgrade(ctx, "cert-manager", "jetstack/cert-manager", helm.UpgradeOptions{
		Install:         true,
		Version:         "v1.14.2",
		Namespace:       "cert-manager",
		CreateNamespace: true,
		Values: map[string]any{
			"installCRDs": true,
		},
	})
	return err
}
