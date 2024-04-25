// Welcome to the magical bootstrap package. This package/file generates the
// _shims.tpl file included in all gotohelm outputs. A task in Taskfile.yaml is
// used to copy the generated file into the gotohelm package. In the future, it
// might be easier to transpile this file on the fly.
//
// Because this file sets up basic utilities and bridges between go and
// templating there are restricts on what may be used.
//
//   - only sprig functions may be used from the `helmette` package.
//   - go primitives without direct template support (switches, multi-value
//     returns, type assertions, etc) may not be used.
package bootstrap

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func typetest(typ string, value, zero any) []any {
	if helmette.TypeIs(typ, value) {
		return []any{value, true}
	}
	return []any{zero, false}
}

func typeassertion(typ string, value any) any {
	if !helmette.TypeIs(typ, value) {
		panic(fmt.Sprintf("expected type of %q got: %T", typ, value))
	}
	return value
}

func dicttest(m map[string]any, key string, zero any) []any {
	if helmette.HasKey(m, key) {
		return []any{m[key], true}
	}
	return []any{zero, false}
}

func compact(args []any) map[string]any {
	out := map[string]any{}
	for i, e := range args {
		out[fmt.Sprintf("T%d", 1+i)] = e
	}
	return out
}

func deref(ptr any) any {
	if ptr == nil {
		panic("nil dereference")
	}
	return ptr
}

func len(m map[string]any) int {
	if m == nil {
		return 0
	}
	return helmette.Len(m)
}
