package helmette

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var (
	// TrimPrefix is the go equivalent of sprig's `trimPrefix`
	TrimPrefix = strings.TrimPrefix

	// SortAlpha is the go equivalent of sprig's `sortAlpha`
	SortAlpha = sort.Strings

	// SortAlpha is the go equivalent of text/templates's `printf`
	Printf = fmt.Sprintf
)

// KindOf is the go equivalent of sprig's `kindOf`.
func KindOf(v any) string {
	return reflect.TypeOf(v).Kind().String()
}

// KindIs is the go equivalent of sprig's `kindIs`.
func KindIs(kind string, v any) bool {
	return KindOf(v) == kind
}

// Keys is the go equivalent of sprig's `keys`.
func Keys[K comparable, V any](m map[K]V) []K {
	return nil
}

// Merge is a go equivalent of sprig's `merge`.
func Merge[K comparable, V any](dst, src map[K]V) map[K]V {
	maps.Copy(dst, src)
	return dst
}

// Dig is a go equivalent of sprig's `dig`.
func Dig(m map[string]any, fallback any, path ...string) any {
	val := any(m)

	for _, key := range path {
		var ok bool
		val, ok = val.(map[string]any)
		if !ok {
			return fallback
		}

		val, ok = val.(map[string]any)[key]
		if !ok {
			return fallback
		}
	}

	return val
}

// Trunc is a go equivalent of sprig's `trunc`.
func Trunc(length int, in string) string {
	if len(in) < length {
		return in
	}
	return in[:length]
}

// Default is a go equivalent of sprig's `default`.
func Default(default_, value any) any {
	if Empty(value) {
		return default_
	}
	return value
}

// RegexMatch is the go equivalent of sprig's `regexMatch`.
func RegexMatch(pattern, s string) bool {
	return regexp.MustCompile(pattern).MatchString(s)
}

// MustRegexMatch is the go equivalent of sprig's `mustRegexMatch`.
func MustRegexMatch(pattern, s string) {
	if !RegexMatch(pattern, s) {
		panic("did not match")
	}
}

// Coalesce is the go equivalent of sprig's `coalesce`.
func Coalesce(values ...any) any {
	for _, v := range values {
		if !Empty(v) {
			return v
		}
	}
	return nil
}

// Empty is the go equivalent of sprig's `empty`.
func Empty(value any) bool {
	truthy, ok := template.IsTrue(value)
	if !truthy || !ok {
		return true
	}
	return false
}

// Required is the go equivalent of sprig's `required`.
func Required(msg string, value any) {
	if Empty(value) {
		Fail(msg)
	}
}

// Fail is the go equivalent of sprig's `fail`.
func Fail(msg string) {
	panic(msg)
}

// ToJSON is the go equivalent of sprig's `toJson`.
func ToJSON(value any) string {
	marshalled, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(marshalled)
}

// MustToJSON is the go equivalent of sprig's `mustToJson`.
func MustToJSON(value any) string {
	marshalled, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

// FromJSON is the go equivalent of sprig's `fromJson`.
func FromJSON(data string) any {
	var out any
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return ""
	}
	return out
}

// MustFromJSON is the go equivalent of sprig's `mustFromJson`.
func MustFromJSON(data string) any {
	var out any
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		panic(err)
	}
	return out
}
