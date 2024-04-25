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
