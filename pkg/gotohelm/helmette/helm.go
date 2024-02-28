package helmette

// Dot is a representation of the "global" context or `.` in the execution
// of a helm template.
// See also: https://github.com/helm/helm/blob/3764b483b385a12e7d3765bff38eced840362049/pkg/chartutil/values.go#L137-L166
type Dot struct {
	Values  Values
	Release Release
	Chart   Chart
	// Capabilities
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
func Tpl(tpl string, context any) string {
	panic("not yet implemented in Go")
}
