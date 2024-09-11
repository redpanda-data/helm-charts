package lint

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const tagURL = "https://github.com/redpanda-data/helm-charts/releases/tag/"

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
			strings.Contains(string(changelog), releaseHeader),
			"CHANGELOG.md is missing the release header for %s\nDid you forget to add it?\n%s",
			chartName,
			releaseHeader,
		)
	}
}
