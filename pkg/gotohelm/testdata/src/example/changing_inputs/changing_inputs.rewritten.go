//go:build rewrites
package changing_inputs

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func ChangingInputs(dot *helmette.Dot) map[string]any {
	for k, v := range dot.Values {
		tmp_tuple_1 := helmette.Compact2(helmette.TypeTest[string](v))
		ok_1 := tmp_tuple_1.T2
		if ok_1 {
			dot.Values[k] = "change that"
		}
	}
	return nil
}
