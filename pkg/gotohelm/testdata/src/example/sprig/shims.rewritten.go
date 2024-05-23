//go:build rewrites
package sprig

import "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"

func numericTestInputs(dot *helmette.Dot) []any {
	return []any{
		"",
		int(0),
		float64(1),
		[]int{},
		map[string]any{},
		dot.Values["numeric"],
	}
}

func asNumeric(dot *helmette.Dot) any {
	// Inputs here are intentionally setup in a strange way. We need to test
	// going across function boundaries, having specifically typed inputs
	// within the same function, and doing the same for .Values.
	inputs := numericTestInputs(dot)
	inputs = append(inputs, int(10), 1.5, dot.Values["numeric"])

	outputs := []any{}
	for _, in := range inputs {
		tmp_tuple_1 := helmette.Compact2(helmette.AsNumeric(in))
		isNumeric := tmp_tuple_1.T2
		value := tmp_tuple_1.T1

		outputs = append(outputs, []any{in, value, isNumeric})
	}

	return outputs
}

func asIntegral(dot *helmette.Dot) any {
	// Inputs here are intentionally setup in a strange way. We need to test
	// going across function boundaries, having specifically typed inputs
	// within the same function, and doing the same for .Values.
	inputs := numericTestInputs(dot)
	inputs = append(inputs, int(10), 1.5, dot.Values["numeric"])

	outputs := []any{}
	for _, in := range inputs {
		tmp_tuple_2 := helmette.Compact2(helmette.AsIntegral[int](in))
		isIntegral := tmp_tuple_2.T2
		value := tmp_tuple_2.T1

		outputs = append(outputs, []any{in, value, isIntegral})
	}

	return outputs
}
