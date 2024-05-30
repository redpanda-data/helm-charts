//go:build rewrites
// +gotohelm:namespace=_directives
package directives

func Directives() bool {
	// Calling Noop does nothing but asserts that it's referenced correctly
	// with the correct namespacing.
	Noop()
	return true
}
