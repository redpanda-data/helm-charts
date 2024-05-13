package helmette

import (
	"context"

	"github.com/redpanda-data/helm-charts/pkg/kube"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Dot is a representation of the "global" context or `.` in the execution
// of a helm template.
// See also: https://github.com/helm/helm/blob/3764b483b385a12e7d3765bff38eced840362049/pkg/chartutil/values.go#L137-L166
type Dot struct {
	Values  Values
	Release Release
	Chart   Chart
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

type Values map[string]any

func (v Values) AsMap() map[string]any {
	if v == nil {
		return map[string]any{}
	}
	return v
}

// https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-tpl-function
// +gotohelm:builtin=tpl
func Tpl(tpl string, context any) string {
	panic("not yet implemented in Go")
}

// Lookup is a wrapper around helm's builtin lookup function that instead
// returns `nil, false` if the lookup fails instead of an empty dictionary.
// See: https://github.com/helm/helm/blob/e24e31f6cc122405ae25069f5b3960036c202c46/pkg/engine/lookup_func.go#L60-L97
func Lookup[T any, PT kube.AddrofObject[T]](dot *Dot, namespace, name string) (obj *T, found bool) {
	ctl, err := kube.FromConfig(dot.KubeConfig)
	if err != nil {
		panic(err)
	}

	obj, err = kube.Get[T, PT](context.Background(), ctl, kube.ObjectKey{Namespace: namespace, Name: name})
	if err != nil {
		if !errors.IsNotFound(err) {
			panic(err)
		}

		return nil, false
	}

	return obj, true
}
