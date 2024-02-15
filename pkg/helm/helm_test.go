package helm_test

import (
	"path"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestHelm(t *testing.T) {
	ctx := testutil.Context(t)

	configDir := path.Join(t.TempDir(), "helm-1")

	c, err := helm.New(helm.Options{ConfigHome: configDir})

	repos := []helm.Repo{
		{Name: "redpanda", URL: "https://charts.redpanda.com"},
		{Name: "jetstack", URL: "https://charts.jetstack.io"},
	}

	listedRepos, err := c.RepoList(ctx)
	require.NoError(t, err)
	require.Len(t, listedRepos, 0)

	for _, repo := range repos {
		require.NoError(t, c.RepoAdd(ctx, repo.Name, repo.URL))
	}

	listedRepos, err = c.RepoList(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, repos, listedRepos)

	charts, err := c.Search(ctx, "redpanda/redpanda")
	require.NoError(t, err)
	require.Len(t, charts, 1)
	require.Equal(t, "Redpanda is the real-time engine for modern apps.", charts[0].Description)
}
