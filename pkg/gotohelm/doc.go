// package gotohelm implements a source to source compiler (transpiler) from go
// to helm templates.
//
// gotohelm relies on the go compiler to type check code and on go test to
// assert the correctness thereof. Doing so allows the transpiling process to
// generally "trust" the code that's being transpiled. After the initial
// parsing and type checking, a collection of AST rewrites are performed to
// convert various bits of go syntax into (mostly) equivalent but more easily
// transpilable syntax.
//
// gotohelm takes the approach of bootstrapping a rudimentary LISP-y
// programming language within helm templates using the available builtins.
//
// # Functions
// Functions are "implemented" by abusing the `include` builtin and are turned
// into `define` blocks. Return values are wrapped in a dictionary and
// marshalled to JSON. (Almost like Internet Explorer circa 2011).
// Function calls are then a pipline of `(include NAME ARGS...) | fromJson | get RETURNKEY`
//
// # Interop
// Transpiled go functions can be invoked within existing templates using the
// following syntax: `((include NAME (dict "a" (list ARGS...))) | fromJson | get "r")`
//
// # Limitations
//   - There is no "trap door" to fallback to raw templates
//   - Switch statements, in all forms, are not currently supported
//   - Code must deal with the "lowest common denominator" of .Values in the form
//     of map[string]any. Values coalescing has not yet been implemented.
//   - Type assertions don't work.
//   - Many helpers and bits of syntax are missing.
//   - Most forms of incompatibility are handled with panics and fmt.Sprintf.
package gotohelm
