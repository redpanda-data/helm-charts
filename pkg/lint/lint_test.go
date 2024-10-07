package lint

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const tagURL = "https://github.com/redpanda-data/helm-charts/releases/tag/"

type ChartYAML struct {
	Version     string            `json:"version"`
	AppVersion  string            `json:"appVersion"`
	Annotations map[string]string `json:"annotations"`
}

func TestChartYAMLVersions(t *testing.T) {
	chartYAMLs, err := fs.Glob(os.DirFS("../.."), "charts/*/Chart.yaml")
	require.NoError(t, err)

	changelog, err := os.ReadFile("../../CHANGELOG.md")
	require.NoError(t, err)

	for _, chartYAML := range chartYAMLs {
		chartBytes, err := os.ReadFile("../../" + chartYAML)
		require.NoError(t, err)

		var chart map[string]any
		require.NoError(t, yaml.Unmarshal(chartBytes, &chart))

		chartName := chart["name"].(string)
		chartVersion := chart["version"].(string)

		releaseHeader := fmt.Sprintf("### [%s](%s%s-%s)", chartVersion, tagURL, chartName, chartVersion)

		// require.Contains is noisy with a large file. Fallback to
		// require.True for friendlier messages.
		assert.Truef(
			t,
			bytes.Contains(changelog, []byte(releaseHeader)),
			"CHANGELOG.md is missing the release header for %s %s\nDid you forget to add it?\n%s",
			chartName,
			chartVersion,
			releaseHeader,
		)
	}
}

func TestOperatorArtifactHubImages(t *testing.T) {
	const operatorRepo = "docker.redpanda.com/redpandadata/redpanda-operator"
	const configuratorRepo = "docker.redpanda.com/redpandadata/configurator"

	chartBytes, err := os.ReadFile("../../charts/operator/Chart.yaml")
	require.NoError(t, err)

	var chart ChartYAML
	require.NoError(t, yaml.Unmarshal(chartBytes, &chart))

	assert.Contains(
		t,
		chart.Annotations["artifacthub.io/images"],
		fmt.Sprintf("%s:%s", operatorRepo, chart.AppVersion),
		"artifacthub.io/images should be in sync with .appVersion",
	)

	assert.Contains(
		t,
		chart.Annotations["artifacthub.io/images"],
		fmt.Sprintf("%s:%s", configuratorRepo, chart.AppVersion),
		"artifacthub.io/images should be in sync with .appVersion",
	)
}
