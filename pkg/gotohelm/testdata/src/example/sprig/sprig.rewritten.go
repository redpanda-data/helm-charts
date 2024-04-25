//go:build rewrites
package sprig

import (
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

type AStruct struct {
	Value int
}

// Sprig runs a variety of values through various sprig functions. Assertions
// are no performed within this code, we're merely testing that the functions
// in helmette return the same values as the transpiled versions.
func Sprig() map[string]any {
	return map[string]any{
		"concat":  concat(),
		"default": default_(),
		"keys":    keys(),
		"empty":   empty(),
		"strings": stringsFunctions(),
		"unset":   unset(),
		"regex":   regex(),
		"atoi":    atoi(),
		"float":   float(),
		"len":     lenTest(),
	}
}

func lenTest() []int {
	mapWithKeys := map[string]string{
		"test": "test",
	}
	initializedMap := map[string]string{}
	return []int{
		helmette.Len(mapWithKeys),
		helmette.Len(initializedMap),
	}
}

func float() []float64 {
	tmp_tuple_1 := helmette.Compact2(helmette.Float64("3.2"))
	f := tmp_tuple_1.T1
	tmp_tuple_2 := helmette.Compact2(helmette.Float64("3"))
	integer := tmp_tuple_2.T1
	tmp_tuple_3 := helmette.Compact2(helmette.Float64("abc"))
	err := tmp_tuple_3.T2
	invalidInput := tmp_tuple_3.T1
	errorHappen := 0.3
	if err != nil {
		// The error will never happen in go template engine. That's why sprig is swallowing/omitting any error
		//errorHappen = 1.3
	}
	return []float64{
		f,
		integer,
		invalidInput,
		errorHappen,
	}
}

func regex() []bool {
	return []bool{
		helmette.MustRegexMatch(`^\d+(k|M|G|T|P|E|Ki|Mi|Gi|Ti|Pi|Ei)?$`, "2.5Gi"),
		helmette.RegexMatch(`^\d+(k|M|G|T|P|E|Ki|Mi|Gi|Ti|Pi|Ei)?$`, "2.5Gi"),
		helmette.RegexMatch(`^\d+(k|M|G|T|P|E|Ki|Mi|Gi|Ti|Pi|Ei)?$`, "25Gi"),
	}
}

func atoi() []int {
	tmp_tuple_4 := helmette.Compact2(helmette.Atoi("234"))
	positive := tmp_tuple_4.T1
	tmp_tuple_5 := helmette.Compact2(helmette.Atoi("-23"))
	negative := tmp_tuple_5.T1
	tmp_tuple_6 := helmette.Compact2(helmette.Atoi("paokwdpo"))
	err := tmp_tuple_6.T2
	invalidInput := tmp_tuple_6.T1
	errorHappen := 0
	if err != nil {
		// The error will never happen in go template engine. That's why sprig is swallowing/omitting any error
		//errorHappen = 1
	}
	return []int{
		positive,
		negative,
		errorHappen,
		invalidInput,
	}
}

func stringsFunctions() []string {
	return []string{
		helmette.Lower("hello WORLD"),
		helmette.Upper("hello WORLD"),
		strings.ToLower("hello WORLD"),
		strings.ToUpper("hello WORLD"),
	}
}

func keys() [][]string {
	// .Keys is non-deterministic, must sort to ensure tests always pass.
	keys := helmette.Keys(map[string]int{"0": 0, "1": 1})
	helmette.SortAlpha(keys)

	return [][]string{
		keys,
		helmette.Keys(map[string]int{}),
	}
}

func concat() [][]int {
	return [][]int{
		helmette.Concat([]int{1, 2}, []int{3, 4}),
		helmette.Concat([]int{1, 2}, []int{3, 4}, []int{5, 6}),
		append([]int{1, 2}, []int{3, 4}...),
		append([]int{1, 2}, 3, 4),
	}
}

func unset() []map[string]int {
	m1 := map[string]int{"0": 0, "1": 1, "2": 2}
	m2 := map[string]int{"0": 0, "1": 1, "2": 2}
	m3 := map[string]int{"0": 0, "1": 1, "2": 2}
	m4 := map[string]int{"0": 0, "1": 1, "2": 2}

	delete(m2, "0")

	helmette.Unset(m3, "2")

	delete(m3, "1")
	helmette.Unset(m3, "2")

	return []map[string]int{m1, m2, m3, m4}
}

func default_() []any {
	defaultStr := "DEFAULT"
	defaultInt := 1234
	defaultStrSlice := []string{defaultStr}

	return []any{
		helmette.Default("", defaultStr),
		helmette.Default("value", defaultStr),
		helmette.Default(nil, defaultStrSlice),
		helmette.Default([]string{}, defaultStrSlice),
		helmette.Default(0, defaultInt),
		helmette.Default(1, defaultInt),
	}
}

func empty() []bool {
	return []bool{
		helmette.Empty(nil),
		helmette.Empty([]string{}),
		helmette.Empty([]string{""}),
		helmette.Empty(map[string]any{}),
		helmette.Empty(map[string]any{"key": nil}),
		helmette.Empty(1),
		helmette.Empty(0),
		helmette.Empty(false),
		helmette.Empty(true),
		helmette.Empty(""),
		helmette.Empty("hello"),
		helmette.Empty(AStruct{}),
		helmette.Empty(AStruct{Value: 0}),
		helmette.Empty(AStruct{Value: 1}),
	}
}
