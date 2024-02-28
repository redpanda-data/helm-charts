package inputs

import (
	"slices"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"golang.org/x/exp/maps"
)

type Nested struct {
	Quux any `json:"quux,omitempty"`
}

type Values struct {
	Foo    any    `json:"foo,omitempty"`
	Bar    string `json:"bar,omitempty"`
	Nested Nested `json:"nested,omitempty"`
}

func Inputs(dot *helmette.Dot) map[string]any {
	return map[string]any{
		"unwrap":    unwrap(dot),
		"echo":      echo(dot),
		"digCompat": digCompat(dot),
		"keys":      keys(dot),
	}
}

func unwrap(dot *helmette.Dot) Nested {
	return helmette.Unwrap[Values](dot.Values).Nested
}

func echo(globals *helmette.Dot) map[string]any {
	return globals.Values
}

func digCompat(dot *helmette.Dot) string {
	return helmette.Dig(dot.Values.AsMap(), "hello", "doesn't", "exist").(string)
}

func keys(globals *helmette.Dot) []string {
	// Get the keys in all possible ways but only return the stable ones.

	keys := []string{}
	for key := range globals.Values {
		keys = append(keys, key)
	}

	keys = maps.Keys(globals.Values)
	slices.Sort(keys)

	return keys
}
