package lint

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
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

func TestConnectArtifactHubImages(t *testing.T) {
	const connectRepo = "docker.redpanda.com/redpandadata/connect"

	chartBytes, err := os.ReadFile("../../charts/connect/Chart.yaml")
	require.NoError(t, err)

	var chart ChartYAML
	require.NoError(t, yaml.Unmarshal(chartBytes, &chart))

	assert.Contains(
		t,
		chart.Annotations["artifacthub.io/images"],
		fmt.Sprintf("%s:%s", connectRepo, chart.AppVersion),
		"artifacthub.io/images should be in sync with .appVersion",
	)
}

func TestRedpandaControllersTag(t *testing.T) {
	chartBytes, err := os.ReadFile("../../charts/operator/Chart.yaml")
	require.NoError(t, err)

	valuesYAML, err := os.ReadFile("../../charts/redpanda/values.yaml")
	require.NoError(t, err)

	var chart map[string]any
	require.NoError(t, yaml.Unmarshal(chartBytes, &chart))

	var values redpanda.Values
	require.NoError(t, yaml.Unmarshal(valuesYAML, &values))

	require.Equal(
		t,
		chart["appVersion"].(string),
		string(values.Statefulset.SideCars.Controllers.Image.Tag),
		"the redpanda chart's values.yaml's controllers tag should be equal to the operator chart's appVersion",
	)
}

func TestOperatorKustomizationTag(t *testing.T) {
	chartBytes, err := os.ReadFile("../../charts/operator/Chart.yaml")
	require.NoError(t, err)

	var chart map[string]any
	require.NoError(t, yaml.Unmarshal(chartBytes, &chart))

	kustomizationBytes, err := os.ReadFile("../../charts/operator/testdata/kustomization.yaml")
	require.NoError(t, err)

	var kustomization map[string]any
	require.NoError(t, yaml.Unmarshal(kustomizationBytes, &kustomization))

	for _, addr := range kustomization["resources"].([]any) {
		require.Contains(
			t,
			addr,
			chart["appVersion"].(string),
			"testdata kustomization address tag should be equal to the operator chart's appVersion",
		)
	}
}
