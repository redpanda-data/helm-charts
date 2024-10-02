// +gotohelm:ignore=true
package bootstrap

// This file contains "bindings" to the sprig helpers utilized by bootstrap.go.
//
// Importing helmette causes a lot of difficulty and slows down the
// transpilation process ever so slighly due to importing MANY additional
// libraries.
//
// Instead any used sprig functions are defined here without a body as we don't
// expect to actually run this go code.

// +gotohelm:builtin=typeIs
func TypeIs(string, any) bool {
	panic("not implemented")
}

// +gotohelm:builtin=hasKey
func HasKey(map[string]any, string) bool {
	panic("not implemented")
}

// +gotohelm:builtin=len
func Len(any) int {
	panic("not implemented")
}

// +gotohelm:builtin=lookup
func Lookup(apiVersion, kind, namespace, name string) map[string]any {
	panic("not implemented")
}

// +gotohelm:builtin=empty
func Empty(any) bool {
	panic("not implemented")
}

// +gotohelm:builtin=mulf
func Mulf(any, any) float64 {
	panic("not implemented")
}

// +gotohelm:builtin=floor
func Floor(any) float64 {
	panic("not implemented")
}

// +gotohelm:builtin=ceil
func Ceil(any) string {
	panic("not implemented")
}

// +gotohelm:builtin=int
func Int(any) int {
	panic("not implemented")
}

// +gotohelm:builtin=int64
func Int64(any) int64 {
	panic("not implemented")
}

// +gotohelm:builtin=float64
func Float64(any) float64 {
	panic("not implemented")
}

// +gotohelm:builtin=regexMatch
func RegexMatch(string, any) bool {
	panic("not implemented")
}

// +gotohelm:builtin=regexFind
func RegexFind(string, any) string {
	panic("not implemented")
}

// +gotohelm:builtin=substr
func Substr(int, int, any) string {
	panic("not implemented")
}

// +gotohelm:builtin=toString
func ToString(any) string {
	panic("not implemented")
}
