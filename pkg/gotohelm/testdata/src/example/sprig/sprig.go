package sprig

import (
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
		"empty":   empty(),
		"default": default_(),
	}
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
