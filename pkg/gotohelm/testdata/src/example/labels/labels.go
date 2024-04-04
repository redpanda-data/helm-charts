package labels

import "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"

type Values struct {
	CommonLabels map[string]string `json:"commonLabels"`
}

// full helm labels + common labels
func FullLabels(dot *helmette.Dot) map[string]string {
	values := helmette.Unwrap[Values](dot.Values)

	commonLabels := map[string]string{}
	if values.CommonLabels != nil {
		commonLabels = values.CommonLabels
	}
	defaults := map[string]string{
		"helm.sh/chart":                "chart",
		"app.kubernetes.io/name":       "name",
		"app.kubernetes.io/instance":   dot.Release.Name,
		"app.kubernetes.io/managed-by": dot.Release.Service,
		"app.kubernetes.io/component":  "component",
	}

	// As Merge function would not only return the dictionary, but also mutate its first argument
	// the empty map is provided to not mutate user provided commonLabels
	//
	// https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/docs/dicts.md?plain=1#L125-L182
	return helmette.Merge(map[string]string{}, commonLabels, defaults)
}
