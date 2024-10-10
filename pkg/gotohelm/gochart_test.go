package gotohelm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

var (
	//go:embed testdata/subchart/root/Chart.yaml
	rootYaml []byte

	//go:embed testdata/subchart/dep1/Chart.yaml
	dep1Yaml []byte

	//go:embed testdata/subchart/dep2/Chart.yaml
	dep2Yaml []byte

	//go:embed testdata/subchart/root/values.yaml
	rootDefaultValuesYAML []byte

	//go:embed testdata/subchart/dep1/values.yaml
	dep1DefaultValuesYAML []byte

	//go:embed testdata/subchart/dep2/values.yaml
	dep2DefaultValuesYAML []byte
)

func TestDependencyChainRender(t *testing.T) {
	dep2, err := Load(dep2Yaml, dep2DefaultValuesYAML, renderDep2)
	require.NoError(t, err)
	dep1, err := Load(dep1Yaml, dep1DefaultValuesYAML, renderDep1, dep2)
	require.NoError(t, err)
	root, err := Load(rootYaml, rootDefaultValuesYAML, renderRoot, dep1)
	require.NoError(t, err)

	inputValues := map[string]any{
		"change-me":      "changed",
		"value-addition": true,
		"dep1": map[string]any{
			"change-me":      "changed",
			"value-addition": true,
			"dep2": map[string]any{
				"change-me":      "changed",
				"value-addition": true,
			},
		},
	}

	val, err := root.LoadValues(inputValues)
	require.NoError(t, err)

	dep2Val := Dep2Values{
		ChangeMe:      "changed",
		DoNotChange:   "default",
		ValueAddition: true,
		RootAddition:  true,
		RootOverwrite: "root-overwrite",
		Dep1Addition:  true,
		Dep1Overwrite: "dep1-overwrite",
	}

	dep1Val := Dep1Values{
		ChangeMe:      "changed",
		DoNotChange:   "default",
		ValueAddition: true,
		RootAddition:  true,
		RootOverwrite: "root-overwrite",
		Dep2:          dep2Val,
	}

	rootVal := RootValues{
		ChangeMe:      "changed",
		DoNotChange:   "default",
		ValueAddition: true,
		Dep1:          dep1Val,
	}

	require.JSONEq(t, MustMarshalJSON(rootVal), MustMarshalJSON(val))

	objs, err := root.Render(kube.Config{}, helmette.Release{}, inputValues)
	require.NoError(t, err)

	require.Equal(t, []kube.Object{
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(rootVal),
			},
		},
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(dep1Val),
			},
		},
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(dep2Val),
			},
		},
	}, objs)
}

func MustMarshalJSON(x any) string {
	bs, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

type Dep2Values struct {
	ChangeMe      string `json:"change-me"`
	DoNotChange   string `json:"do-not-change"`
	ValueAddition bool   `json:"value-addition"`

	RootAddition  bool   `json:"root-addition"`
	RootOverwrite string `json:"root-overwrite"`

	Dep1Addition  bool   `json:"dep1-addition"`
	Dep1Overwrite string `json:"dep1-overwrite"`
}

type Dep1Values struct {
	ChangeMe      string `json:"change-me"`
	DoNotChange   string `json:"do-not-change"`
	ValueAddition bool   `json:"value-addition"`

	RootAddition  bool   `json:"root-addition"`
	RootOverwrite string `json:"root-overwrite"`

	Dep2 Dep2Values `json:"dep2"`
}

type RootValues struct {
	ChangeMe      string `json:"change-me"`
	DoNotChange   string `json:"do-not-change"`
	ValueAddition bool   `json:"value-addition"`

	Dep1 Dep1Values `json:"dep1"`
}

func renderDep2(dot *helmette.Dot) []kube.Object {
	values := helmette.Unwrap[Dep2Values](dot.Values)

	return []kube.Object{
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(values),
			},
		},
	}
}

func renderDep1(dot *helmette.Dot) []kube.Object {
	values := helmette.Unwrap[Dep1Values](dot.Values)

	return []kube.Object{
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(values),
			},
		},
	}
}

func renderRoot(dot *helmette.Dot) []kube.Object {
	values := helmette.Unwrap[RootValues](dot.Values)

	return []kube.Object{
		&corev1.ConfigMap{
			Data: map[string]string{
				"values": MustMarshalJSON(values),
			},
		},
	}
}
