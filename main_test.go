package main

import (
	"context"
	"strings"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func EnsureCertManager(ctx context.Context, client *helm.Client) error {
	releases, err := client.List(ctx)
	if err != nil {
		return err
	}

	for _, release := range releases {
		if strings.HasPrefix(release.Chart, "cert-manager-") && release.Status == "deployed" {
			return nil
		}
	}

	if err := client.RepoAdd(ctx, "jetstack", "https://charts.jetstack.io"); err != nil {
		return err
	}

	return client.Install(ctx, "jetstack/cert-manager", helm.InstallOptions{
		Name:            "cert-manager",
		Version:         "v1.14.2",
		Namespace:       "cert-manager",
		CreateNamespace: true,
		Values: map[string]any{
			"installCRDs": true,
		},
	})
}

func TestRedpandaChart(t *testing.T) {
	ctx := testutil.Context(t)

	ctl, err := kube.FromEnv()
	require.NoError(t, err)

	client, err := helm.New(helm.Options{
		KubeConfig: ctl.RestConfig(),
		ConfigHome: t.TempDir(),
	})
	require.NoError(t, err)

	require.NoError(t, EnsureCertManager(ctx, client))

	// TODO(chrisseto): This is a bit kludgey as we're relying on the directory
	// that `go test` is being run from.
	require.NoError(t, client.Install(ctx, "./charts/redpanda", helm.InstallOptions{
		Namespace:       "redpanda",
		CreateNamespace: true,
		// TODO(chrisseto): redpanda won't successfully install on a default
		// kind cluster due to resource limitations.
		NoWait:        true,
		NoWaitForJobs: true,
	}))
}
