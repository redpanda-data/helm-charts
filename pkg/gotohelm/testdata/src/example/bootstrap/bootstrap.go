package bootstrap

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

type TypeSpec struct {
	ExpectedType string
	DefaultValue any
}

func hydrate(in any) any {
	return in
}

func mustget(d map[string]any, key string) any {
	value, ok := d[key]
	if !ok {
		panic(fmt.Sprintf("missing key %q", key))
	}
	return value
}

func zeroof(kind string) any {
	if kind == "int" {
		return 0
	} else if kind == "string" {
		return ""
	} else if kind == "slice" {
		return []any{} // TODO is this technically correct?
	} else {
		panic(fmt.Sprintf("unhandled kind %q", kind))
	}
}

func typetest(kind string, value any) []any {
	if helmette.KindOf(value) == kind {
		return []any{value, true}
	}
	return []any{zeroof(kind), false}
}
