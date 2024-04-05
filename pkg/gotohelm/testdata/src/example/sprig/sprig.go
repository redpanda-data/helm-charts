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
		"default": default_(),
		"keys":    keys(),
		"empty":   empty(),
		"strings": stringsFunctions(),
		"unset":   unset(),
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
