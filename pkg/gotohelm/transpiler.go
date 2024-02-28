package gotohelm

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

var directiveRE = regexp.MustCompile(`\+gotohelm:([\w\.-]+)=([\w\.-]+)`)

// TODO need to ensure dict test returns the correct zero value...
// TODO _shims.compact is a little bit hacky. It might malfunction if a slice
// is one of the return values it's called with.
//
//go:embed shims.yaml
var shimsYAML string

type Unsupported struct {
	Node ast.Node
	Msg  string
	Fset *token.FileSet
}

func (u *Unsupported) Error() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "unsupported ast.Node: %T\n", u.Node)
	fmt.Fprintf(&b, "%s\n", u.Msg)
	if err := format.Node(&b, u.Fset, u.Node); err != nil {
		panic(err) // Oh the irony
	}
	return b.String()
}

type Chart struct {
	Files []*File
}

func Transpile(pkg *packages.Package) (_ *Chart, err error) {
	defer func() {
		switch v := recover().(type) {
		case nil:
		case *Unsupported:
			err = v
		default:
			panic(v)
		}
	}()

	// Ensure there are no errors in the package before we transpile it.
	for _, err := range pkg.TypeErrors {
		return nil, err
	}

	for _, err := range pkg.Errors {
		return nil, err
	}

	t := &Transpiler{
		Package:   pkg.Name,
		Fset:      pkg.Fset,
		TypesInfo: pkg.TypesInfo,
		Files:     pkg.Syntax,
	}

	return t.Transpile(), nil
}

type Transpiler struct {
	Package   string
	Fset      *token.FileSet
	Files     []*ast.File
	TypesInfo *types.Info
}

func (t *Transpiler) Transpile() *Chart {
	var chart Chart
	for _, f := range t.Files {
		path := t.Fset.File(f.Pos()).Name()
		name := filepath.Base(path)
		source := filepath.Base(path)
		name = source[:len(source)-3] + ".yaml"

		isTestFile := strings.HasSuffix(name, "_test.go")
		if isTestFile || name == "main.go" {
			continue
		}

		fileDirectives := parseDirectives(f.Doc.Text())
		if _, ok := fileDirectives["filename"]; ok {
			name = fileDirectives["filename"]
		}

		if _, ok := fileDirectives["ignore"]; ok {
			continue
		}

		var funcs []*Func
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}

			var params []Node
			for _, param := range fn.Type.Params.List {
				for _, name := range param.Names {
					params = append(params, t.transpileExpr(name))
				}
			}

			var statements []Node
			for _, stmt := range fn.Body.List {
				statements = append(statements, t.transpileStatement(stmt))
			}

			funcDirectives := parseDirectives(fn.Doc.Text())
			name := funcDirectives["name"]
			if name == "" {
				name = fn.Name.String()
			}

			// TODO add a source field here? Ideally with a line number.
			funcs = append(funcs, &Func{
				Name:       name,
				Namespace:  t.Package,
				Params:     params,
				Statements: statements,
			})
		}

		chart.Files = append(chart.Files, &File{
			Name:   name,
			Source: source,
			Funcs:  funcs,
		})
	}
	// TODO Do this better
	// Write out some basic shim functions to help us better match go's
	// behavior.
	chart.Files = append(chart.Files, &File{
		Source: "",
		Name:   "_shims.tpl",
		Header: shimsYAML,
	})
	return &chart
}

func (t *Transpiler) transpileStatement(stmt ast.Stmt) Node {
	switch stmt := stmt.(type) {
	case nil:
		return nil

	case *ast.DeclStmt:
		switch d := stmt.Decl.(type) {
		case *ast.GenDecl:
			if len(d.Specs) > 1 {
				// TODO could just return multiple statements.
				panic(&Unsupported{
					Node: d,
					Fset: t.Fset,
					Msg:  "declarations may only contain 1 spec",
				})
			}
			spec := d.Specs[0].(*ast.ValueSpec)

			if len(spec.Names) > 1 || len(spec.Values) > 1 {
				panic(&Unsupported{
					Node: d,
					Fset: t.Fset,
					Msg:  "specs may only contain 1 value",
				})
			}

			rhs := t.zeroOf(t.TypesInfo.TypeOf(spec.Names[0]))
			if len(spec.Values) > 0 {
				rhs = t.transpileExpr(spec.Values[0])
			}

			return &Assignment{
				LHS: t.transpileExpr(spec.Names[0]),
				New: true,
				RHS: rhs,
			}

		default:
			panic(fmt.Sprintf("unsupported declaration: %#v", d))
		}

	case *ast.BranchStmt:
		switch stmt.Tok {
		case token.BREAK:
			return &Statement{NoCapture: true, Expr: &Literal{Value: "break"}}

		case token.CONTINUE:
			return &Statement{NoCapture: true, Expr: &Literal{Value: "continue"}}
		}

	case *ast.ReturnStmt:
		if len(stmt.Results) != 1 {
			panic(&Unsupported{
				Node: stmt,
				Fset: t.Fset,
				Msg:  "returns must return exactly 1 value",
			})
		}

		return &Return{
			Expr: t.transpileExpr(stmt.Results[0]),
		}

	case *ast.AssignStmt:
		if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
			break
		}

		// +=, /=, *=, etc show up as assignments. They're not supported in
		// templates. We'll need to either rewrite the expression here or add
		// another AST rewrite.
		switch stmt.Tok {
		case token.ASSIGN, token.DEFINE:
		default:
			panic(&Unsupported{
				Node: stmt,
				Fset: t.Fset,
				Msg:  "Unsupported assignment token",
			})
		}

		// TODO could simplify this by performing a type switch on the
		// transpiled result of lhs.
		if _, ok := stmt.Lhs[0].(*ast.SelectorExpr); ok {
			selector := t.transpileExpr(stmt.Lhs[0]).(*Selector)

			return &Statement{
				Expr: &BuiltInCall{
					FuncName: "set",
					Arguments: []Node{
						selector.Expr,
						&Literal{Value: strconv.Quote(selector.Field)},
						t.transpileExpr(stmt.Rhs[0]),
					},
				},
			}
		}

		// TODO could simplify this by implementing an IndexExpr node and then
		// performing a type switch on the transpiled result of lhs.
		if idx, ok := stmt.Lhs[0].(*ast.IndexExpr); ok {
			return &Statement{
				Expr: &BuiltInCall{
					FuncName: "set",
					Arguments: []Node{
						t.transpileExpr(idx.X),
						t.transpileExpr(idx.Index),
						t.transpileExpr(stmt.Rhs[0]),
					},
				},
			}
		}

		rhs := t.transpileExpr(stmt.Rhs[0])
		lhs := t.transpileExpr(stmt.Lhs[0])

		return &Assignment{RHS: rhs, LHS: lhs, New: stmt.Tok.String() == ":="}

	case *ast.RangeStmt:
		return &Range{
			Key:   t.transpileExpr(stmt.Key),
			Value: t.transpileExpr(stmt.Value),
			Over:  t.transpileExpr(stmt.X),
			Body:  t.transpileStatement(stmt.Body),
		}

	case *ast.ExprStmt:
		return &Statement{
			Expr: t.transpileExpr(stmt.X),
		}

	case *ast.BlockStmt:
		var out []Node
		for _, s := range stmt.List {
			out = append(out, t.transpileStatement(s))
		}
		return &Block{Statements: out}

	case *ast.IfStmt:
		return &IfStmt{
			Init: t.transpileStatement(stmt.Init),
			Cond: t.transpileExpr(stmt.Cond),
			Body: t.transpileStatement(stmt.Body),
			Else: t.transpileStatement(stmt.Else),
		}
	}

	panic(&Unsupported{
		Node: stmt,
		Fset: t.Fset,
		Msg:  "unhandled ast.Stmt",
	})
}

func (t *Transpiler) transpileExpr(n ast.Expr) Node {
	switch n := n.(type) {
	case nil:
		return nil

	case *ast.BasicLit:
		return &Literal{Value: n.Value}

	case *ast.StarExpr:
		// TODO this should be wrapped in something like "Assert not nil"
		return t.transpileExpr(n.X)

	case *ast.CompositeLit:
		typ := t.typeOf(n)

		// TODO: Need to handle implementors of json.Marshaler.
		// TODO: Need to filter out zero value fields that are explicitly
		// provided.

		if p, ok := typ.(*types.Pointer); ok {
			typ = p.Elem()
		}

		switch typ := typ.Underlying().(type) {
		case *types.Slice:
			var elts []Node
			for _, el := range n.Elts {
				elts = append(elts, t.transpileExpr(el))
			}
			return &BuiltInCall{
				FuncName:  "list",
				Arguments: elts,
			}

		case *types.Map:
			if !types.AssignableTo(typ.Key(), types.Typ[types.String]) {
				panic(fmt.Sprintf("map keys must be string. Got %#v", typ.Key()))
			}

			var d DictLiteral
			for _, el := range n.Elts {
				d.KeysValues = append(d.KeysValues, &KeyValue{
					Key:   el.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value,
					Value: t.transpileExpr(el.(*ast.KeyValueExpr).Value),
				})
			}
			return &d

		case *types.Struct:
			zero := t.zeroOf(typ)
			fields := getFields(typ)
			fieldByName := map[string]*structField{}
			for _, f := range fields {
				f := f
				fieldByName[f.Field.Name()] = &f
			}

			var embedded []Node
			var d DictLiteral
			for _, el := range n.Elts {
				key := el.(*ast.KeyValueExpr).Key.(*ast.Ident).Name
				value := el.(*ast.KeyValueExpr).Value

				field := fieldByName[key]
				if field.Omit {
					continue
				}

				if field.Embedded {
					embedded = append(embedded, t.transpileExpr(value))
					continue
				}

				d.KeysValues = append(d.KeysValues, &KeyValue{
					Key:   strconv.Quote(field.JSONName),
					Value: t.transpileExpr(value),
				})
			}

			args := []Node{zero}
			args = append(args, embedded...)
			args = append(args, &d)

			return &BuiltInCall{
				FuncName:  "mustMergeOverwrite",
				Arguments: args,
			}

		default:
			panic(fmt.Sprintf("unsupported composite literal %#v", typ))
		}

	case *ast.CallExpr:
		return t.transpileCallExpr(n)

	case *ast.Ident:
		// Unclear how often this check is correct. true, false, and _ won't
		// have an Obj. AST rewriting can also result in .Obj being nil.
		if n.Obj == nil {
			if n.Name == "_" {
				return &Ident{Name: n.Name}
			}
			return &Literal{Value: n.Name}
		}

		return &Ident{Name: n.Name}

	case *ast.SelectorExpr:
		if s, ok := unwrapStruct(t.typeOf(n.X)); ok {
			for _, f := range getFields(s) {
				if f.Field.Name() == n.Sel.Name {
					return &Selector{
						Expr:  t.transpileExpr(n.X),
						Field: f.JSONName,
					}
				}
			}
		}

		// TODO when would this ever get hit?
		return &Selector{
			Expr:  t.transpileExpr(n.X),
			Field: n.Sel.Name,
		}

	case *ast.BinaryExpr:
		untyped := [3]string{"_", n.Op.String(), "_"}
		typed := [3]string{t.typeOf(n.X).String(), n.Op.String(), t.typeOf(n.Y).String()}

		// Poor man's pattern matching :[
		mapping := map[[3]string]string{
			{"_", token.EQL.String(), "_"}:             "eq",
			{"_", token.NEQ.String(), "_"}:             "ne",
			{"_", token.LAND.String(), "_"}:            "and",
			{"_", token.LOR.String(), "_"}:             "or",
			{"_", token.GTR.String(), "_"}:             "gt",
			{"_", token.LSS.String(), "_"}:             "lt",
			{"_", token.GEQ.String(), "_"}:             "gte",
			{"_", token.LEQ.String(), "_"}:             "lte",
			{"int", token.ADD.String(), "int"}:         "add",
			{"int", token.SUB.String(), "int"}:         "sub",
			{"int", token.MUL.String(), "int"}:         "mul",
			{"int", token.QUO.String(), "int"}:         "div",
			{"float32", token.QUO.String(), "float32"}: "divf",
			{"float64", token.QUO.String(), "float64"}: "divf",
		}

		// Typed versions take precedence.
		if funcName, ok := mapping[typed]; ok {
			return &BuiltInCall{
				FuncName:  funcName,
				Arguments: []Node{t.transpileExpr(n.X), t.transpileExpr(n.Y)},
			}
		}

		// Fallback to "wild cards" (_).
		if funcName, ok := mapping[untyped]; ok {
			return &BuiltInCall{
				FuncName:  funcName,
				Arguments: []Node{t.transpileExpr(n.X), t.transpileExpr(n.Y)},
			}
		}

		// TODO re-add suport for rewriting str + str into printf "%s%s". For
		// now its easier to just require writers to use printf
		// No support for easy string concatenation in helm/sprig/templates soooo. Printf.
		// if t.isString(n.Y) && t.isString(n.X) {
		// 	return &BuiltInCall{
		// 		FuncName: "printf",
		// 		Arguments: []Node{
		// 			&Literal{Value: `"%s%s"`},
		// 			t.transpileExpr(n.X),
		// 			t.transpileExpr(n.Y),
		// 		},
		// 	}
		// }

	case *ast.UnaryExpr:
		switch n.Op {
		case token.NOT:
			return &BuiltInCall{
				FuncName:  "not",
				Arguments: []Node{t.transpileExpr(n.X)},
			}
		case token.AND:
			// Can't take addresses in templates so just return the variable.
			return t.transpileExpr(n.X)
		}

	case *ast.IndexExpr:
		return &BuiltInCall{
			FuncName: "index",
			Arguments: []Node{
				t.transpileExpr(n.X),
				t.transpileExpr(n.Index),
			},
		}

	case *ast.TypeAssertExpr:
		// return &BuiltInCall{
		// 	FuncName: "_shims.typeassertion",
		// 	Arguments: []Node{
		// 		t.transpileExpr(n.Type),
		// 		t.transpileExpr(n.X),
		// 	},
		// }

		// TODO figure out how to support type switches. For now, hope for the
		// best and expect something to break if the type happens to be
		// incorrect.
		// Could potentially inject some "bootstrap" functions that would make this easier.
		// IE
		return t.transpileExpr(n.X)
	}

	var b bytes.Buffer
	if err := format.Node(&b, t.Fset, n); err != nil {
		panic(err)
	}
	panic(fmt.Sprintf("unhandled Expr %T\n%s", n, b.String()))
}

func (t *Transpiler) transpileCallExpr(n *ast.CallExpr) Node {
	var args []Node
	for _, arg := range n.Args {
		args = append(args, t.transpileExpr(arg))
	}

	callee := typeutil.Callee(t.TypesInfo, n)

	switch {
	// go builtins
	case callee == nil, callee.Pkg() == nil:
		switch n.Fun.(*ast.Ident).Name {
		case "append":
			// TODO: mustAppend isn't variadic. Need to either nest mustAppend calls or use concat.
			if len(args) > 2 {
				panic(&Unsupported{
					Node: n,
					Fset: t.Fset,
					Msg:  "appending multiple values at once is not currently supported",
				})
			}
			return &BuiltInCall{FuncName: "mustAppend", Arguments: args}
		case "int", "int32":
			return &BuiltInCall{FuncName: "int", Arguments: args}
		case "panic":
			return &BuiltInCall{FuncName: "fail", Arguments: args}
		case "string":
			return &BuiltInCall{FuncName: "toString", Arguments: args}
		default:
			panic(fmt.Sprintf("unsupport golang builtin %q", n.Fun.(*ast.Ident).Name))
		}

	// Method call.
	case callee.Type().(*types.Signature).Recv() != nil:
		if len(args) != 0 {
			panic(&Unsupported{Fset: t.Fset, Node: n, Msg: "method calls with arguments are not implemented"})
		}
		// Method calls come in as a "top level" CallExpr where .Fun is the
		// selector up to that call. IE all of `Foo.Bar.Baz()` will be "within"
		// the CallExpr. CallExpr.Fun will contain Foo.Bar.Baz. In the case of
		// zero argument methods, text/template will automatically call them.
		return t.transpileExpr(n.Fun)

	// Call to function within the same package. A-Okay. It's
	// transpiled.
	case callee.Pkg().Name() == t.Package:
		return &Call{FuncName: fmt.Sprintf("%s.%s", t.Package, callee.Name()), Arguments: args}
	}

	// Mapping of go functions to sprig/helm/template functions where arguments
	// are also the same.
	funcMapping := map[string]string{
		"fmt.Sprintf":           "printf",
		"helmette.Default":      "default",
		"helmette.Empty":        "empty",
		"helmette.FromJSON":     "fromJson",
		"helmette.Keys":         "keys",
		"helmette.KindIs":       "kindIs",
		"helmette.KindOf":       "kindOf",
		"helmette.Merge":        "merge",
		"helmette.MustFromJSON": "mustFromJson",
		"helmette.MustToJSON":   "mustToJson",
		"helmette.RegexMatch":   "regexMatch",
		"helmette.ToJSON":       "toJson",
		"helmette.Tpl":          "tpl",
		"helmette.Trunc":        "trunc",
		"maps.Keys":             "keys",
		"math.Floor":            "floor",
	}

	// Call to any other function.
	// This check's a bit... buggy
	name := callee.Pkg().Name() + "." + callee.Name()

	if tplFuncName, ok := funcMapping[name]; ok {
		return &BuiltInCall{FuncName: tplFuncName, Arguments: args}
	}

	// Mappings that are not 1:1 and require some argument fiddling to make
	// them match up as expected.
	switch name {
	case "slices.Sort":
		// TODO: This only works for strings :[
		return &BuiltInCall{FuncName: "sortAlpha", Arguments: args}
	case "strings.TrimSuffix":
		return &BuiltInCall{FuncName: "trimSuffix", Arguments: []Node{args[1], args[0]}}
	case "strings.ReplaceAll":
		return &BuiltInCall{FuncName: "replace", Arguments: []Node{args[1], args[2], args[0]}}
	case "intstr.FromInt32":
		return args[0]
	case "helmette.MustDuration":
		return args[0]
	case "helmette.Dig":
		return &BuiltInCall{FuncName: "dig", Arguments: append(args[2:], args[1], args[0])}
	case "helmette.Unwrap":
		return &Selector{Expr: args[0], Field: "AsMap"}
	case "helmette.Compact2":
		return &Call{FuncName: "_shims.compact", Arguments: args}
	case "helmette.DictTest":
		// TODO need to figure out how to get the generic argument here.
		// TODO revalidate arguments
		// TODO add in zerof
		return &Call{FuncName: "_shims.dicttest", Arguments: args}
	case "helmette.TypeTest":
		// TODO there's got to be a better way to get the type params....
		args = append([]Node{
			&Literal{
				Value: fmt.Sprintf("%q", n.Fun.(*ast.IndexExpr).Index.(*ast.Ident).Name),
			},
		}, args...)
		return &Call{FuncName: "_shims.typetest", Arguments: args}
	case "helmette.TypeAssertion":
		// TODO need to figure out how to get the generic argument here.
		// TODO revalidate arguments
		// TODO there's got to be a better way to get the type params....
		args = append([]Node{
			&Literal{
				Value: fmt.Sprintf("%q", n.Fun.(*ast.IndexExpr).Index.(*ast.Ident).Name),
			},
		}, args...)
		return &Call{FuncName: "_shims.typeassertion", Arguments: args}
	default:
		panic(fmt.Sprintf("unsupported function %s", name))
	}
}

func (t *Transpiler) isString(e ast.Expr) bool {
	return types.AssignableTo(t.TypesInfo.TypeOf(e), types.Typ[types.String])
}

func (t *Transpiler) isBasic(e ast.Expr, typ types.BasicKind) bool {
	if b, ok := t.typeOf(e).(*types.Basic); ok && b.Kind() == typ {
		return true
	}
	return false
}

func (t *Transpiler) typeOf(expr ast.Expr) types.Type {
	return t.TypesInfo.TypeOf(expr)
}

func (t *Transpiler) zeroOf(typ types.Type) Node {
	// TODO need to detect and reject or special case implementors of
	// json.Marshaler. Getting a handle to a that interface is... difficult.

	// Special cases.
	switch typ.String() {
	case "k8s.io/apimachinery/pkg/apis/meta/v1.Time":
		return &Nil{}
	}

	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		switch typ.Info() {
		case types.IsString:
			return &Literal{Value: `""`}
		case types.IsInteger, types.IsUnsigned | types.IsInteger:
			return &Literal{Value: "0"}
		case types.IsBoolean:
			return &Literal{Value: "false"}
		default:
			panic(fmt.Sprintf("unsupported Basic type: %#v", typ))
		}

	case *types.Pointer, *types.Map, *types.Interface, *types.Slice:
		return &Nil{}

	case *types.Struct:
		var embedded []Node
		var out DictLiteral

		// Skip fields that json Marshalling would itself skip.
		for _, field := range getFields(typ) {
			if field.Omit || (field.OmitEmpty && omitemptyRespected(field.Field.Type())) {
				continue
			}

			if field.Embedded {
				embedded = append(embedded, t.zeroOf(field.Field.Type()))
				continue
			}

			out.KeysValues = append(out.KeysValues, &KeyValue{
				Key:   strconv.Quote(field.JSONName),
				Value: t.zeroOf(field.Field.Type()),
			})
		}
		if len(embedded) < 1 {
			return &out
		}
		return &BuiltInCall{
			FuncName:  "mustMergeOverwrite",
			Arguments: append(embedded, &out),
		}

	default:
		panic(fmt.Sprintf("unsupported type: %#v", typ))
	}
}

func unwrapStruct(typ types.Type) (*types.Struct, bool) {
	if p, ok := typ.Underlying().(*types.Pointer); ok {
		typ = p.Elem()
	}

	if s, ok := typ.Underlying().(*types.Struct); ok {
		return s, true
	}

	return nil, false
}

func omitemptyRespected(typ types.Type) bool {
	switch typ.(type) {
	case *types.Basic, *types.Pointer, *types.Slice, *types.Map:
		return true
	case *types.Named:
		return omitemptyRespected(typ.Underlying())
	default:
		return false
	}
}

type jsonTag struct {
	Name      string
	Inline    bool
	OmitEmpty bool
}

func parseTag(tag string) jsonTag {
	match := regexp.MustCompile(`json:"([^"]+)"`).FindStringSubmatch(tag)
	if match == nil {
		return jsonTag{}
	}

	idx := strings.Index(match[1], ",")
	if idx == -1 {
		idx = len(match[1])
	}

	return jsonTag{
		Name:      match[1][:idx],
		Inline:    strings.Contains(match[1], "inline"),
		OmitEmpty: strings.Contains(match[1], "omitempty"),
	}
}

type structField struct {
	Field     *types.Var
	Embedded  bool
	Omit      bool
	OmitEmpty bool
	JSONName  string
}

func getFields(s *types.Struct) []structField {
	var fields []structField
	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		tag := parseTag(s.Tag(i))

		omit := tag.Name == "-" || !field.Exported()
		embedded := tag.Name == "" && field.Embedded()

		if tag.Name == "" && tag.Name != "-" {
			tag.Name = field.Name()
		}

		fields = append(fields, structField{
			Field:     field,
			Omit:      omit,
			OmitEmpty: tag.OmitEmpty,
			JSONName:  tag.Name,
			Embedded:  embedded,
		})

	}
	return fields
}

func parseDirectives(in string) map[string]string {
	match := directiveRE.FindAllStringSubmatch(in, -1)

	out := map[string]string{}
	for _, m := range match {
		out[m[1]] = m[2]
	}
	return out
}
