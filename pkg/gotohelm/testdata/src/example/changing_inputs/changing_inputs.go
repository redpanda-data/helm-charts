package changing_inputs

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func ChangingInputs(dot *helmette.Dot) map[string]any {
	for k, v := range dot.Values {
		if _, ok := v.(string); ok {
			dot.Values[k] = "change that"
		}
	}
	return nil
}
