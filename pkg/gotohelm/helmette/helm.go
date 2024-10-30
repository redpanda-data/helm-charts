package helmette

import (
	"bytes"
	"context"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Dot is a representation of the "global" context or `.` in the execution
// of a helm template.
// See also: https://github.com/helm/helm/blob/3764b483b385a12e7d3765bff38eced840362049/pkg/chartutil/values.go#L137-L166
type Dot struct {
	Values    Values
	Release   Release
	Chart     Chart
	Subcharts map[string]*Dot
	// Capabilities

	// KubeConfig is a hacked in value to allow `Lookup` to not rely on global
	// values. It's a kube.Config to support JSON marshalling and allow easy
	// transport into the `go run` test runner.
	// WARNING: DO NOT USE OR REFERENCE IN HELM CHARTS. IT WILL NOT WORK.
	KubeConfig kube.Config
}

type Release struct {
	Name      string
	Namespace string
	Service   string
	IsUpgrade bool
	IsInstall bool
	// Revision
}

type Chart struct {
	Name       string
	Version    string
	AppVersion string
}

type Values = chartutil.Values

// https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-tpl-function
// +gotohelm:builtin=tpl
func Tpl(tpl string, context any) string {
	var b bytes.Buffer

	f := sprig.TxtFuncMap()
	extra := template.FuncMap{
		// Not yet implemented in sprig.go
		//"toToml":        toTOML,
		//"fromYamlArray": fromYAMLArray,
		//"fromJsonArray": fromJSONArray,

		"toYaml":   ToYaml,
		"fromYaml": FromYaml,
		"toJson":   ToJSON,
		"fromJson": FromJSON,

		// Not yet implemented in gotohelm
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },

		"lookup": func(string, interface{}) string { return "not implemented" },
	}

	for k, v := range extra {
		f[k] = v
	}

	tmpl := template.Must(template.New("").Funcs(f).Parse(tpl))
	if err := tmpl.Execute(&b, context); err != nil {
		panic(err)
	}
	return b.String()
}

// Lookup is a wrapper around helm's builtin lookup function that instead
// returns `nil, false` if the lookup fails instead of an empty dictionary.
// See: https://github.com/helm/helm/blob/e24e31f6cc122405ae25069f5b3960036c202c46/pkg/engine/lookup_func.go#L60-L97
func Lookup[T any, PT kube.AddrofObject[T]](dot *Dot, namespace, name string) (obj *T, found bool) {
	obj, found, err := SafeLookup[T, PT](dot, namespace, name)
	if err != nil {
		panic(err)
	}

	return obj, found
}

// SafeLookup is a wrapper around helm's builtin lookup function. It acts
// exactly like Lookup except it returns any errors that may have occurred
// in the underlying lookup operations.
func SafeLookup[T any, PT kube.AddrofObject[T]](dot *Dot, namespace, name string) (*T, bool, error) {
	ctl, err := kube.FromConfig(dot.KubeConfig)
	if err != nil {
		return nil, false, err
	}

	obj, err := kube.Get[T, PT](context.Background(), ctl, kube.ObjectKey{Namespace: namespace, Name: name})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, false, err
		}

		return nil, false, nil
	}

	return obj, true, nil
}
