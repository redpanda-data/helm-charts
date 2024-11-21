//go:build !gotohelm

package bootstrap

import "github.com/Masterminds/sprig/v3"

// This file contains "bindings" to the sprig helpers utilized by bootstrap.go
// that run the actual sprig functions directly. This allows us to write unit
// tests for bootstrap functions and verify their behavior in go.

func TypeIs(string, any) bool {
	panic("not implemented")
}

func HasKey(map[string]any, string) bool {
	panic("not implemented")
}

func Get(map[string]any, string) any {
	panic("not implemented")
}

func Len(any) int {
	panic("not implemented")
}

func Lookup(apiVersion, kind, namespace, name string) map[string]any {
	panic("not implemented")
}

func Empty(any) bool {
	panic("not implemented")
}

func Mulf(any, any) float64 {
	panic("not implemented")
}

func Floor(any) float64 {
	panic("not implemented")
}

func Ceil(any) string {
	panic("not implemented")
}

func Int(any) int {
	panic("not implemented")
}

func Int64(in any) int64 {
	return sprig.FuncMap()["int64"].(func(any) int64)(in)
}

func Float64(any) float64 {
	panic("not implemented")
}

func RegexMatch(string, any) bool {
	panic("not implemented")
}

func Substr(int, int, any) string {
	panic("not implemented")
}

func ToString(any) string {
	panic("not implemented")
}

func Duration(d int64) string {
	return sprig.FuncMap()["duration"].(func(any) string)(d)
}

func RegexFind(pattern string, in any) string {
	return sprig.FuncMap()["regexFind"].(func(string, string) string)(pattern, in.(string))
}
