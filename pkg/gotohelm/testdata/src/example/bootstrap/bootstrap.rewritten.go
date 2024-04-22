//go:build rewrites
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
	tmp_tuple_1 := helmette.Compact2(helmette.DictTest[string, any](d, key))
	ok := tmp_tuple_1.T2
	value := tmp_tuple_1.T1
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
