package helmette

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// TypeTest is an equivalent of `val, ok := x.(type)` that is exercised as a
// function call rather than a special form of syntax.
// See also: "_shims.typetest".
func TypeTest[T any](val any) (T, bool) {
	asT, ok := val.(T)
	return asT, ok
}

// TypeAssertion is an equivalent of `x.(type)` that is exercised as a function
// call rather than a special form of syntax.
// See also: "_shims.typeassertion".
func TypeAssertion[T any](val any) T {
	return val.(T)
}

// DictTest is an equivalent of `val, ok := map[key]` that is exercised as a
// function call rather than a special form of syntax.
// See also: "_shims.dicttest".
// func DictTest[K comparable, V any](m map[K]V, key K) TestResult[V] {
func DictTest[K comparable, V any](m map[K]V, key K) (V, bool) {
	val, ok := m[key]
	return val, ok
}

func MustDuration(duration string) *metav1.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}
	return &metav1.Duration{Duration: d}
}

// AsNumeric attempts to interpret the provided value as a helm friendly
// "numeric" (float64). It should be used in place of type assertions to
// numeric types.
// Due to helm's, sprig's, and gotohelm's use of untyped JSON marshalling all
// numeric values are cast to float64s. To ensure that gocode and helm code
// function the same way, AsNumeric must be used.
func AsNumeric(value any) (float64, bool) {
	switch value := value.(type) {
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case float32:
		return float64(value), true
	case float64:
		return value, true
	}

	return 0, false
}

// HelmNumber is a union type of valid numeric primitives within the context of
// helm templates.
type HelmNumber interface {
	~float64 | ~int | ~int64
}

// AsIntegral is a helm-specific replacement for type assertions/tests of
// numeric types. The combination of helm, text/template, and sprig prevent us from
// being able to reasonably distinguish between the various types of numeric
// types. Instead we must rely on loose heuristic to determine if a value could
// reasonably be interpreted as anything other than a float64.
func AsIntegral[T HelmNumber](value any) (T, bool) {
	switch value := value.(type) {
	case int:
		return T(value), true
	case int32:
		return T(value), true
	case int64:
		return T(value), true
	case float32:
		if math.Floor(float64(value)) == float64(value) {
			return T(value), true
		}
	case float64:
		if math.Floor(value) == value {
			return T(value), true
		}
	}
	return 0, false
}

// Unwrap "unwraps" .Values into a golang struct.
// DANGER: Unwrap performs no defaulting or validation. At the helm level, this
// is transpiled into .Values.AsMap.
// Callers are responsible for verifying that T is appropriately validated by
// the charts values.schema.json.
func Unwrap[T any](from Values) T {
	// TODO might be beneficial to have the helm side of this merge values with
	// a zero value of the struct?
	var out T
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &out,
		DecodeHook: mapstructure.DecodeHookFuncType(func(from, to reflect.Type, val interface{}) (interface{}, error) {
			switch reflect.New(to).Interface().(type) {
			case *resource.Quantity:
				return valuesutil.UnmarshalInto[*resource.Quantity](val)
			}
			return val, nil
		}),
	})
	if err != nil {
		panic(errors.WithStack(err))
	}

	if err := decoder.Decode(from.AsMap()); err != nil {
		panic(errors.WithStack(err))
	}

	return out
}

// UnmarshalInto [valuesutil.UnmarshalInto] without an error return for use to
// the gotohelm world.
//
// It may be used to "convert" untyped values into values of type T provided
// that their JSON representations are the same. For example, an any type
// holding a known struct value that can't be asserted via a type assertion.
//
// DANGER: In helm, no validation or default is done. This function effectively
// transpiles to `return value`. Use with care.
func UnmarshalInto[T any](value any) T {
	t, err := valuesutil.UnmarshalInto[T](value)
	if err != nil {
		panic(err)
	}
	return t
}

// UnmarshalYaml fills in the type requested
// +gotohelm:builtin=fromYamlArray
func UnmarshalYamlArray[T any](repr string) []T {
	var output []T
	if err := yaml.Unmarshal([]byte(repr), &output); err != nil {
		panic(fmt.Errorf("cannot unmarshal yaml: %w", err))
	}
	return output
}

type Tuple2[T1, T2 any] struct {
	T1 T1
	T2 T2
}

type Tuple3[T1, T2, T3 any] struct {
	T1 T1
	T2 T2
	T3 T3
}

func Compact2[T1, T2 any](t1 T1, t2 T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{T1: t1, T2: t2}
}

func Compact3[T1, T2, T3 any](t1 T1, t2 T2, t3 T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{T1: t1, T2: t2, T3: t3}
}
