package gotohelm

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	//go:embed testdata/subchart/root/Chart.yaml
	rootChart []byte

	//go:embed testdata/subchart/values-overwrite/Chart.yaml
	valuesOverwriteChart []byte

	//go:embed testdata/subchart/dependency-excluded-by-default/Chart.yaml
	depExcludedChart []byte

	//go:embed testdata/subchart/dependency-included-by-default/Chart.yaml
	depIncludedChart []byte

	//go:embed testdata/subchart/dependency/Chart.yaml
	depChart []byte

	//go:embed testdata/subchart/root/values.yaml
	rootDefaultValuesYAML []byte

	//go:embed testdata/subchart/values-overwrite/values.yaml
	valuesOverwriteDefaultValuesYAML []byte

	//go:embed testdata/subchart/dependency-excluded-by-default/values.yaml
	depExcludedDefaultValuesYAML []byte

	//go:embed testdata/subchart/dependency-included-by-default/values.yaml
	depIncludedDefaultValuesYAML []byte

	//go:embed testdata/subchart/dependency/values.yaml
	depDefaultValuesYAML []byte
)

func TestDependencyChainRender(t *testing.T) {
	// The chart dependency graph
	//            ┌───────────────┐   ┌──────────┐
	//     ┌─────►│valuesOverwrite├──►│Dependency│
	//     │      └───────────────┘   └──────────┘
	//     │
	// ┌───┴┐     ┌──────────┐        ┌──────────┐
	// │root┼────►│ExcludeDep├───────►│Dependency│
	// └───┬┘     └──────────┘        └──────────┘
	//     │
	//     │      ┌──────────┐        ┌───────────┐
	//     └─────►│IncludeDep├───────►│Dependency │
	//            └──────────┘        └───────────┘
	// Graph created by https://asciiflow.com/#/
	//
	// ExcludeDep - has condition that points to value which is false.
	// IncludeDep - has condition that points to value which is true.
	// valuesOverwrite - does not have any condition
	//
	// All charts are creating Config map which has one data, the rendered values
	dep, err := Load(depChart, depDefaultValuesYAML, renderConfigMap)
	require.NoError(t, err)
	valuesOverwrite, err := Load(valuesOverwriteChart, valuesOverwriteDefaultValuesYAML, renderConfigMap, dep)
	require.NoError(t, err)
	excludeDep, err := Load(depExcludedChart, depExcludedDefaultValuesYAML, renderConfigMap, dep)
	require.NoError(t, err)
	includeDep, err := Load(depIncludedChart, depIncludedDefaultValuesYAML, renderConfigMap, dep)
	require.NoError(t, err)
	root, err := Load(rootChart, rootDefaultValuesYAML, renderConfigMap, valuesOverwrite, excludeDep, includeDep)
	require.NoError(t, err)

	helmCli, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	for _, chartPath := range []string{
		"./testdata/subchart/dependency",
		"./testdata/subchart/dependency-excluded-by-default",
		"./testdata/subchart/dependency-included-by-default",
		"./testdata/subchart/values-overwrite",
		"./testdata/subchart/root",
	} {
		out, err := exec.Command("helm", "dep", "build", chartPath).CombinedOutput()
		if err != nil {
			require.NoErrorf(t, err, "failed to run helm dep build: %s", out)
		}
	}

	inputVal, err := os.ReadFile("testdata/subchart/root/input-val.yaml")
	require.NoError(t, err)

	inputValues := map[string]any{}

	err = yaml.Unmarshal(inputVal, &inputValues)
	require.NoError(t, err)

	expected, err := helmCli.Template(context.Background(), "testdata/subchart/root", helm.TemplateOptions{
		Name:      "subchart",
		Namespace: "test",
		Values:    inputValues,
		Version:   "0.0.1",
	})
	require.NoError(t, err)

	actual, err := root.Render(kube.Config{}, helmette.Release{}, inputValues)
	require.NoError(t, err)

	actualByte, err := convertToString(actual)
	require.NoError(t, err)

	actualDocuments, err := ytbx.LoadDocuments(actualByte)
	require.NoError(t, err)

	expectedDocuments, err := ytbx.LoadDocuments(expected)
	require.NoError(t, err)

	sorter := func(a, b *yamlv3.Node) int {
		aNode, err := ytbx.Grab(a, "data.values")
		require.NoError(t, err)
		bNode, err := ytbx.Grab(b, "data.values")
		require.NoError(t, err)
		return strings.Compare(aNode.Value, bNode.Value)
	}
	slices.SortStableFunc(actualDocuments, sorter)
	slices.SortStableFunc(expectedDocuments, sorter)

	report, err := dyff.CompareInputFiles(
		ytbx.InputFile{Documents: expectedDocuments},
		ytbx.InputFile{Documents: actualDocuments},
	)
	if err != nil {
		require.NoError(t, err)
	}

	if len(report.Diffs) > 0 {
		hr := dyff.HumanReport{Report: report, OmitHeader: true}

		var buf bytes.Buffer
		require.NoError(t, hr.WriteReport(&buf))

		require.Fail(t, buf.String())
	}
}

func convertToString(objs []kube.Object) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	for _, obj := range objs {
		fmt.Fprintf(b, "---\n%s\n", MustMarshalYAML(obj))
	}
	return b.Bytes(), nil
}

func MustMarshalYAML(x any) string {
	bs, err := yaml.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

func renderConfigMap(dot *helmette.Dot) []kube.Object {
	return []kube.Object{
		&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: dot.Chart.Name,
			},
			Data: map[string]string{
				"values": MustMarshalYAML(dot.Values),
			},
		},
	}
}
