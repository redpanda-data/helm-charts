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
// # Directives
//
// Directives in the form of `+gotohelm:DIRECTIVE=VALUE` may be specified as go
// comments on the file, function, or type they modify.
//
//	// +gotohelm:filename=NAME        // Changes the name of the transpiled file to NAME.
//	// +gotohelm:ignore=true          // Skips transpilation of the annotated file, function, or type.
//	// +gotohelm:namespace=NAMESPACE  // Changes the namespace of transpiled package.
//	// +gotohelm:builtin=BUILTIN_FUNC // Replaces transpilation of the annotated function with `BUILTIN_FUNC`.
//
// # Interop
//
// Transpiled go functions can be invoked within existing templates using the
// following syntax:
//
//	((include NAMESPACE.NAME (dict "a" (list ARGS...))) | fromJson | get "r")
//
// # Limitations
//
//   - There is no "trap door" to fallback to raw templates
//   - Switch statements, in all forms, are not currently supported
//   - Type assertions don't work.
//   - Most forms of incompatibility are handled with panics and fmt.Sprintf.
//   - As all data is represented as JSON within helm. All numeric types must
//     be treated as float64s. [helmette.AsIntegral] may be used to approximate
//     integrals numbers.
//
// # Internals
//
// Functions are "implemented" by abusing the `include` builtin and are turned
// into `define` blocks. Return values are wrapped in a dictionary and
// marshalled to JSON. (Almost like Internet Explorer circa 2011).
// Function calls are then a pipeline of
//
//	(include NAME ARGS...) | fromJson | get RETURNKEY
//
// Structs are represented as their JSON representation. Any implementer of
// [json.Marshaller] or [json.Unmarshaller], such as [resource.Quantity] will
// need to be special cased within the transpiler itself.
//
// Certain types of go syntax are difficult to transpile. To preserve
// simplicity of the transpiler, more complicated pieces of syntax are
// re-written to equivalent syntax that can more easily be transpiled.
// Particularly, anything that returns multiple values.
//
//	value, ok := dict[key]
//
// Becomes
//
//	tmp_tuple_1 := helmette.DictTest(dict, key)
//	value := tmp_tuple_1.T1
//	ok := tmp_tuple_1.T2
package gotohelm
