package quantity

import "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"

func SiToBytes(dot *helmette.Dot) string {
	return helmette.Sitobytes(dot.Values["q"])
}
