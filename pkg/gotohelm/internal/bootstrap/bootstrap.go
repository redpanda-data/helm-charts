// Welcome to the magical bootstrap package. This package/file generates the
// _shims.tpl file included in all gotohelm outputs. A task in Taskfile.yaml is
// used to copy the generated file into the gotohelm package. In the future, it
// might be easier to transpile this file on the fly.
//
// Because this file sets up basic utilities and bridges between go and
// templating there are restricts on what may be used.
//
//   - go primitives without direct template support (switches, multi-value
//     returns, type assertions, etc) may not be used.
//   - Only go builtins with direct template support (fmt.Sprintf, etc) may be
//     called/imported.
//   - sprig functions must have a binding declared in sprig.go.
//
// +gotohelm:filename=_shims.tpl
// +gotohelm:namespace=_shims
//
//lint:file-ignore U1000 Ignore all unused code, it's exported into gotohelm templates
package bootstrap

import (
	"fmt"
	"time"
)

const (
	// For reference: https://physics.nist.gov/cuu/Units/binary.html
	milli = 0.001
	kilo  = 1000
	mega  = kilo * kilo
	giga  = kilo * kilo * kilo
	terra = kilo * kilo * kilo * kilo
	peta  = kilo * kilo * kilo * kilo * kilo

	kibi = 1024
	mebi = kibi * kibi
	gibi = kibi * kibi * kibi
	tebi = kibi * kibi * kibi * kibi
	pebi = kibi * kibi * kibi * kibi * kibi
)

// typeatest is the implementation of the go syntax `_, _ := m.(t)`.
func typetest(typ string, value, zero any) []any {
	if TypeIs(typ, value) {
		return []any{value, true}
	}
	return []any{zero, false}
}

// typeassertion is the implementation of the go syntax `_ := m.(t)`.
func typeassertion(typ string, value any) any {
	if !TypeIs(typ, value) {
		panic(fmt.Sprintf("expected type of %q got: %T", typ, value))
	}
	return value
}

// dicttest is the implementation of the go syntax `_, _ := m[k]`.
func dicttest(m map[string]any, key string, zero any) []any {
	if HasKey(m, key) {
		return []any{m[key], true}
	}
	return []any{zero, false}
}

// compact is the implementation of `helmette.CompactN`.
// It's a strange and hacky way of handling multi-value returns.
func compact(args []any) map[string]any {
	out := map[string]any{}
	for i, e := range args {
		out[fmt.Sprintf("T%d", 1+i)] = e
	}
	return out
}

// deref is the implementation of the go syntax `*variable`.
func deref(ptr any) any {
	if ptr == nil {
		panic("nil dereference")
	}
	return ptr
}

// +gotohelm:name=len
func _len(m any) int {
	// Handle empty/nil maps and lists as sprig does not.
	if m == nil {
		return 0
	}
	return Len(m)
}

// re-implementation of k8s.io/utils/ptr.Deref.
func ptr_Deref(ptr, def any) any {
	if ptr != nil {
		return ptr
	}
	return def
}

// re-implementation of k8s.io/utils/ptr.Equal.
func ptr_Equal(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	return a == b
}

// wrapper around helm's lookup.
func lookup(apiVersion, kind, namespace, name string) (map[string]any, bool) {
	result := Lookup(apiVersion, kind, namespace, name)
	// Helm's builtin lookup returns an empty dict for some godforsaken
	// reason. We return nil, false similar to how a map look up works
	// (sanely).

	// Helm recommends using `(empty result)` to test if there is a value or
	// not.
	if Empty(result) {
		return nil, false
	}
	return result, true
}

func asnumeric(value any) (any, bool) {
	if TypeIs("float64", value) {
		return value, true
	}

	if TypeIs("int64", value) {
		return value, true
	}

	if TypeIs("int", value) {
		return value, true
	}

	return 0, false
}

func asintegral(value any) (any, bool) {
	if TypeIs("int64", value) || TypeIs("int", value) {
		return value, true
	}

	if TypeIs("float64", value) && Floor(value) == value {
		return value, true
	}

	return 0, false
}

func parseResource(repr any) (float64, float64) {
	if TypeIs("float64", repr) {
		return Float64(repr), 1
	}

	if !TypeIs("string", repr) {
		panic(fmt.Sprintf("invalid Quantity expected string or float64 got: %T (%v)", repr, repr))
	}

	// TODO add an upper bound on the total amount of digits?
	if !RegexMatch(`^[0-9]+(\.[0-9]{0,6})?(k|m|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$`, repr) {
		// TODO write a longer message about support for quantities.
		// NB: Negative values are intentionally not supported.
		panic(fmt.Sprintf("invalid Quantity: %q", repr))
	}

	// Type cast would work but that relies on bootstrap to work so use sprig
	// functions.
	reprStr := ToString(repr)

	unit := RegexFind("(k|m|M|G|T|P|Ki|Mi|Gi|Ti|Pi)$", repr)

	numeric := Float64(Substr(0, len(reprStr)-len(unit), reprStr))

	// No switch statements, so index into a map instead.
	scale, ok := map[string]float64{
		"":   1,
		"m":  milli,
		"k":  kilo,
		"M":  mega,
		"G":  giga,
		"T":  terra,
		"P":  peta,
		"Ki": kibi,
		"Mi": mebi,
		"Gi": gibi,
		"Ti": tebi,
		"Pi": pebi,
	}[unit]
	if !ok {
		panic(fmt.Sprintf("unknown unit: %q", unit))
	}

	return numeric, scale
}

// pseudo implementation of k8s.io/apimachinery/pkg/api/resource.MustParse.
func resource_MustParse(repr any) any {
	numeric, scale := parseResource(repr)

	// No support for switches or maps with non-string keys.
	// So we fake a map[float64]string with two slices and a for loop.
	strs := []string{"", "m", "k", "M", "G", "T", "P", "Ki", "Mi", "Gi", "Ti", "Pi"}
	scales := []float64{1.0, milli, kilo, mega, giga, terra, peta, kibi, mebi, gibi, tebi, pebi}

	idx := -1
	for i, s := range scales {
		if float64(s) == float64(scale) {
			// Ideally this would be an early return but https://github.com/redpanda-data/helm-charts/issues/1331
			idx = i
			break
		}
	}

	if idx == -1 {
		panic(fmt.Sprintf("unknown scale: %v", scale))
	}

	// NB: ToString is used here because it prints out float64's in a reasonable format.
	// As far as I can tell, go's Sprintf can't print floats without trailing
	// zero or truncating precision.
	return fmt.Sprintf("%s%s", ToString(numeric), strs[idx])
}

func resource_Value(repr any) int64 {
	numeric, scale := parseResource(repr)
	return Int64(Ceil(numeric * scale))
}

func resource_MilliValue(repr any) int64 {
	numeric, scale := parseResource(repr)
	return Int64(Ceil(numeric * 1000 * scale))
}

// time_ParseDuration re-implements [time.ParseDuration] but only down to
// second precision. The returned value is represented as nano's.
func time_ParseDuration(repr string) int64 {
	// TODO: loading bootstrap might be faster if we could avoid importing
	// time?
	unitMap := map[string]int64{
		"s": int64(time.Second),
		"m": int64(time.Minute),
		"h": int64(time.Hour),
	}

	original := repr
	value := int64(0)

	// Special cases to match time.ParseDuration.
	if repr == "" {
		panic(fmt.Sprintf("invalid Duration: %q", original))
	}

	if repr == "0" {
		return 0
	}

	// This is written strangely as `for true` loops don't work in gotohelm.
	// Realistically we won't need any more loops than len(unitMap).
	for range []int{0, 0, 0} {
		if repr == "" {
			break
		}

		n := RegexFind(`^\d+`, repr)
		if n == "" {
			panic(fmt.Sprintf("invalid Duration: %q", original))
		}

		repr = repr[len(n):]

		unit := RegexFind(`^(h|m|s)`, repr)
		if unit == "" {
			panic(fmt.Sprintf("invalid Duration: %q", original))
		}

		repr = repr[len(unit):]

		value = value + (Int64(n) * unitMap[unit])
	}

	return value
}

// time_Duration_String re-implements [time.Duration.String] but only down to
// second precision.
func time_Duration_String(dur int64) string {
	return Duration(dur / int64(time.Second))
}
