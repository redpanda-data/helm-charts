//go:build rewrites
package flowcontrol

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func FlowControl(dot *helmette.Dot) map[string]any {
	return map[string]any{
		"earlyReturn":    earlyReturn(dot),
		"ifElse":         ifElse(dot),
		"sliceRanges":    sliceRanges(dot),
		"mapRanges":      mapRanges(dot),
		"intBinaryExprs": intBinaryExprs(),
	}
}

func earlyReturn(dot *helmette.Dot) string {
	tmp_tuple_1 :=
		// This is trickily written on purpose.
		helmette.Compact2(helmette.DictTest[string, any](dot.Values, "boolean"))
	ok_2 := tmp_tuple_1.T2
	b_1 := tmp_tuple_1.T1
	if ok_2 && b_1.(bool) {
		return "Early Returns work!"
	}
	return "Should have returned early"
}

func ifElse(dot *helmette.Dot) string {
	tmp_tuple_2 := helmette.Compact2(helmette.AsIntegral[int](dot.Values["oneToFour"]))
	ok := tmp_tuple_2.T2
	oneToFour := tmp_tuple_2.T1
	if !ok {
		return "oneToFour not specified!"
	}

	if oneToFour == 1 {
		return "It's 1"
	} else if oneToFour == 2 {
		return "It's 2"
	} else if oneToFour == 3 {
		return "It's 3"
	} else {
		return "It's 4"
	}
	return "unreachable"
}

func sliceRanges(dot *helmette.Dot) []any {
	tmp_tuple_3 := helmette.Compact2(helmette.DictTest[string, any](dot.Values, "ints"))
	ok := tmp_tuple_3.T2
	intsAny := tmp_tuple_3.T1
	if !ok {
		intsAny = []any{}
	}

	ints := intsAny.([]any)

	sumOfIndexes := 0
	for i := range ints {
		sumOfIndexes = sumOfIndexes + i
	}

	continuesWork := true
	for range ints {
		continue
		continuesWork = false
	}

	breaksWork := true
	for range ints {
		break
		breaksWork = false
	}

	return []any{
		sumOfIndexes,
		continuesWork,
		breaksWork,
	}
}

func mapRanges(dot *helmette.Dot) []any {
	m := map[string]int{"1": 1, "2": 2, "3": 3}

	// NOTE: Ranges of maps are not technically equivalent. In go, they are
	// non-deterministic but range nodes with templates are deterministic.
	for k := range m {
		_ = k
	}

	sum := 0
	for _, v := range m {
		sum = sum + v
	}

	return []any{sum}
}

func intBinaryExprs() []int {
	x := 1
	y := 2
	z := 3

	// Not currently supported.
	// z += x
	// z -= y
	// z *= y
	// z /= y

	return []int{
		z,
		x - y,
		x + y,
		x / y,
		x * y,
	}
}
