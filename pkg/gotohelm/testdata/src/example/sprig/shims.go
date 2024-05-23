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
		value, isNumeric := helmette.AsNumeric(in)

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
		value, isIntegral := helmette.AsIntegral[int](in)

		outputs = append(outputs, []any{in, value, isIntegral})
	}

	return outputs
}
