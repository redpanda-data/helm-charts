package helmette

import (
	"time"

	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type Tuple2[T1, T2 any] struct {
	T1 T1
	T2 T2
}

func Compact2[T1, T2 any](t1, t2 any) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{}
}

// Unwrap "unwraps" .Values into a golang struct.
// DANGER: Unwrap performs no defaulting or validation. At the helm level, this
// is transpiled into .Values.AsMap.
// Callers are responsible for verifying that T is appropriately validated by
// the charts values.json.schema.
func Unwrap[T any](from Values) T {
	// TODO might be beneficial to have the helm side of this merge values with
	// a zero value of the struct?
	var out T
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &out,
	})
	if err != nil {
		panic(err)
	}

	if err := decoder.Decode(from.AsMap()); err != nil {
		panic(err)
	}
	return out
}
