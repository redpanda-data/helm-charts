package gotohelm

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
	"k8s.io/client-go/kubernetes/scheme"
)

var directiveRE = regexp.MustCompile(`\+gotohelm:([\w\.-]+)=([\w\.-]+)`)

type Unsupported struct {
	Node ast.Node
	Msg  string
	Fset *token.FileSet
}

func (u *Unsupported) Error() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "unsupported ast.Node: %T\n", u.Node)
	fmt.Fprintf(&b, "%s\n", u.Msg)
	fmt.Fprintf(&b, "%s\n\t", u.Fset.PositionFor(u.Node.Pos(), false).String())
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

	t := &Transpiler{
		Package:   pkg,
		Fset:      pkg.Fset,
		TypesInfo: pkg.TypesInfo,
		Files:     pkg.Syntax,

		packages: mkPkgTree(pkg),
		builtins: map[string]string{
			"fmt.Sprintf":                "printf",
			"golang.org/x/exp/maps.Keys": "keys",
			"maps.Keys":                  "keys",
			"math.Floor":                 "floor",
			"sort.Strings":               "sortAlpha",
			"strings.ToLower":            "lower",
			"strings.ToUpper":            "upper",
		},
	}

	return t.Transpile(), nil
}

type Transpiler struct {
	Package   *packages.Package
	Fset      *token.FileSet
	Files     []*ast.File
	TypesInfo *types.Info

	// builtins is a pre-populated cache of function id (fmt.Printf,
	// github.com/my/pkg.Function) to an equivalent go template / sprig
	// builtin. Functions may add a +gotohelm:builtin=blah directive to declare
	// their builtin equivalent.
	builtins map[string]string
	packages map[string]*packages.Package
}

func (t *Transpiler) Transpile() *Chart {
	var chart Chart
	for _, f := range t.Files {
		if transpiled := t.transpileFile(f); transpiled != nil {
			chart.Files = append(chart.Files, transpiled)
		}
	}

	// Finally, include the shims file with all transpiled charts.
	// NB: When the bootstrap package is transpiled shims is nil.
	chart.Files = append(chart.Files, shims)

	return &chart
}

func (t *Transpiler) transpileFile(f *ast.File) *File {
	path := t.Fset.File(f.Pos()).Name()
	source := filepath.Base(path)
	name := source[:len(source)-3] + ".yaml"

	isTestFile := strings.HasSuffix(name, "_test.go")
	if isTestFile || name == "main.go" {
		return nil
	}

	fileDirectives := parseDirectives(f.Doc.Text())
	if _, ok := fileDirectives["filename"]; ok {
		name = fileDirectives["filename"]
	}

	if _, ok := fileDirectives["ignore"]; ok {
		return nil
	}

	namespace := fileDirectives["namespace"]
	if namespace == "" {
		namespace = t.Package.Name
	}

	var funcs []*Func
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcDirectives := parseDirectives(fn.Doc.Text())
		if v, ok := funcDirectives["skip"]; ok && v == "true" {
			continue
		}

		// To not clash with the same method name in the same package
		// which can be declared in multiple struct the package and function
		// name could be separated with the name of the struct type.
		// package example
		//
		// func FunExample() {} => {{- define "example.FunExample" -}}
		//
		// func (e *Example) MethodExample() {} => {{- define "example.Example.MethodExample" -}}
		if _, ok := funcDirectives["name"]; !ok {
			funName := fn.Name.String()
			if fn.Recv != nil {
				funName = fmt.Sprintf("%s.%s", baseTypeName(fn.Recv.List[0].Type), funName)
			}
			funcDirectives["name"] = funName
		}

		if _, ok := funcDirectives["sprig"]; ok {
			continue
		}

		var params []Node
		if fn.Recv != nil {
			for _, param := range fn.Recv.List {
				for _, name := range param.Names {
					params = append(params, t.transpileExpr(name))
				}
			}
		}

		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				params = append(params, t.transpileExpr(name))
			}
		}

		var statements []Node
		for _, stmt := range fn.Body.List {
			statements = append(statements, t.transpileStatement(stmt))
		}

		// TODO add a source field here? Ideally with a line number.
		funcs = append(funcs, &Func{
			Name:       funcDirectives["name"],
			Namespace:  namespace,
			Params:     params,
			Statements: statements,
		})
	}

	return &File{
		Name:   name,
		Source: source,
		Funcs:  funcs,
	}
}

// baseTypeName returns the type name of the Expression
// Reference
// https://github.com/golang/go/blob/beaf7f3282c2548267d3c894417cc4ecacc5d575/src/go/doc/reader.go#L123-L145
func baseTypeName(x ast.Expr) (name string) {
	switch t := x.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return baseTypeName(t.X)
	case *ast.IndexListExpr:
		return baseTypeName(t.X)
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			// only possible for qualified type names;
			// assume type is imported
			return t.Sel.Name
		}
	case *ast.ParenExpr:
		return baseTypeName(t.X)
	case *ast.StarExpr:
		return baseTypeName(t.X)
	}
	return ""
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
		if len(stmt.Results) == 1 {
			return &Return{Expr: t.transpileExpr(stmt.Results[0])}
		}

		var results []Node
		for _, r := range stmt.Results {
			results = append(results, t.transpileExpr(r))
		}
		return &Return{Expr: &BuiltInCall{FuncName: "list", Arguments: results}}

	case *ast.AssignStmt:
		if len(stmt.Lhs) != len(stmt.Rhs) {
			break
		}

		if len(stmt.Lhs) == len(stmt.Rhs) && len(stmt.Lhs) > 1 {
			var stmts []Node
			for i := 0; i < len(stmt.Lhs); i++ {
				stmts = append(stmts, t.transpileStatement(&ast.AssignStmt{
					Lhs: []ast.Expr{stmt.Lhs[i]},
					Tok: stmt.Tok,
					Rhs: []ast.Expr{stmt.Rhs[i]},
				}))
			}
			return &Block{Statements: stmts}
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
	case *ast.ForStmt:
		var start, stop Node
		if b, ok := stmt.Cond.(*ast.BinaryExpr); ok {
			switch b.Op {
			case token.LSS, token.LEQ:
				if _, ok := b.X.(*ast.SelectorExpr); ok {
					start = t.transpileExpr(b.X)
				} else {
					switch declaration := b.X.(*ast.Ident).Obj.Decl.(type) {
					case *ast.AssignStmt:
						start = t.transpileExpr(declaration.Rhs[0])
					case *ast.Field:
						start = t.transpileExpr(declaration.Names[0])
					}
				}

				if _, ok := b.Y.(*ast.SelectorExpr); ok {
					stop = t.transpileExpr(b.Y)
				} else {
					switch declaration := b.Y.(*ast.Ident).Obj.Decl.(type) {
					case *ast.AssignStmt:
						stop = t.transpileExpr(declaration.Rhs[0])
					case *ast.Field:
						stop = t.transpileExpr(declaration.Names[0])
					}
				}
			case token.GTR, token.GEQ:
				if _, ok := b.X.(*ast.SelectorExpr); ok {
					stop = t.transpileExpr(b.X)
				} else {
					switch declaration := b.X.(*ast.Ident).Obj.Decl.(type) {
					case *ast.AssignStmt:
						stop = t.transpileExpr(declaration.Rhs[0])
					case *ast.Field:
						stop = t.transpileExpr(declaration.Names[0])
					}
				}

				if _, ok := b.Y.(*ast.SelectorExpr); ok {
					start = t.transpileExpr(b.Y)
				} else {
					switch declaration := b.Y.(*ast.Ident).Obj.Decl.(type) {
					case *ast.AssignStmt:
						start = t.transpileExpr(declaration.Rhs[0])
					case *ast.Field:
						start = t.transpileExpr(declaration.Names[0])
					}
				}
			default:
				panic(&Unsupported{
					Node: stmt,
					Fset: t.Fset,
					Msg:  fmt.Sprintf("%T of %s is not supported in for condition", b, b.Op),
				})
			}
		}

		var step Node
		switch p := stmt.Post.(type) {
		case *ast.AssignStmt:
			if b, ok := p.Rhs[0].(*ast.BasicLit); ok && p.Tok == token.SUB_ASSIGN {
				b.Value = fmt.Sprintf("-%s", b.Value)
				// switch start with stop expression as step is decreasing
				step = start
				start = stop
				stop = step

				step = t.transpileExpr(b)
			} else {
				step = t.transpileExpr(b)
			}
		case *ast.IncDecStmt:
			switch p.Tok {
			case token.INC:
				step = &Literal{Value: "1"}
			case token.DEC:
				step = &Literal{Value: "-1"}
			}
		default:
			panic(&Unsupported{
				Node: stmt,
				Fset: t.Fset,
				Msg:  "unhandled ast.ForStmt",
			})
		}

		if stop == nil || start == nil || step == nil {
			panic(&Unsupported{
				Node: stmt,
				Fset: t.Fset,
				Msg:  fmt.Sprintf("start: %v; stop: %v; step: %v", start, stop, step),
			})
		}
		return &Range{
			Key:   &Ident{Name: "_"},
			Value: &Literal{Value: fmt.Sprintf("$%s", stmt.Init.(*ast.AssignStmt).Lhs[0].(*ast.Ident).Name)},
			Over: &UntilStep{
				Start: start,
				Stop:  stop,
				Step:  step,
			},
			Body: t.transpileStatement(stmt.Body),
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
		return t.maybeCast(&Literal{Value: n.Value}, t.typeOf(n))

	case *ast.ParenExpr:
		return &ParenExpr{Expr: t.transpileExpr(n.X)}

	case *ast.StarExpr:
		// TODO this should be wrapped in something like "Assert not nil"
		return t.transpileExpr(n.X)

	case *ast.SliceExpr:
		target := t.transpileExpr(n.X)
		low := t.transpileExpr(n.Low)
		high := t.transpileExpr(n.High)
		max := t.transpileExpr(n.Max)

		// If low isn't specified it defaults to zero
		if low == nil {
			low = &Literal{Value: "0"}
		}

		// The builtin `slice` function from go would work great here but sprig
		// overwrites it for some reason with a worse version.
		if t.isString(n.X) {
			// NB: Triple slicing a string (""[1:2:3]) isn't valid. No need to
			// check .Max or .Slice3.

			// Empty slicing a string (""[:]) is effectively a noop
			if low == nil && high == nil {
				return target
			}

			// Sprig's substring will run [start:] if end is < 0.
			if high == nil {
				high = &Literal{Value: "-1"}
			}

			return &BuiltInCall{FuncName: "substr", Arguments: []Node{low, high, target}}
		}

		args := []Node{target, low}
		if high != nil {
			args = append(args, high)
		}
		if n.Slice3 && n.Max != nil {
			args = append(args, max)
		}
		return &BuiltInCall{FuncName: "mustSlice", Arguments: args}

	case *ast.CompositeLit:

		// TODO: Need to handle implementors of json.Marshaler.
		// TODO: Need to filter out zero value fields that are explicitly
		// provided.

		typ := t.typeOf(n)
		if p, ok := typ.(*types.Pointer); ok {
			typ = p.Elem()
		}

		switch underlying := typ.Underlying().(type) {
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
			if !types.AssignableTo(underlying.Key(), types.Typ[types.String]) {
				panic(fmt.Sprintf("map keys must be string. Got %#v", underlying.Key()))
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
			fields := t.getFields(underlying)
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
				if field.JSONOmit() {
					continue
				}

				if field.JSONInline() {
					embedded = append(embedded, t.transpileExpr(value))
					continue
				}

				d.KeysValues = append(d.KeysValues, &KeyValue{
					Key:   strconv.Quote(field.JSONName()),
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
		switch obj := t.TypesInfo.ObjectOf(n).(type) {
		case *types.Const:
			return t.transpileConst(obj)

		case *types.Nil:
			return &Nil{}

		case *types.Var:
			return &Ident{Name: obj.Name()}

		// Unclear how often this check is correct. true, false, and _ won't
		// have an Obj. AST rewriting can also result in .Obj being nil.
		case nil:
			if n.Name == "_" {
				return &Ident{Name: n.Name}
			}
			return &Literal{Value: n.Name}

		default:
			panic(&Unsupported{
				Node: n,
				Fset: t.Fset,
				Msg:  "Unsupported *ast.Ident",
			})
		}

	case *ast.SelectorExpr:
		switch obj := t.TypesInfo.ObjectOf(n.Sel).(type) {
		case *types.Const:
			return t.transpileConst(obj)

		case *types.Func:
			// TODO this needs better documentation
			// And probably needs a more aggressive check.
			return &Selector{
				Expr:  t.transpileExpr(n.X),
				Field: n.Sel.Name,
			}

		case *types.Var:
			// If our selector is a variable, we're probably accessing a field
			// on a struct.
			typ := t.typeOf(n.X)
			if p, ok := typ.(*types.Pointer); ok {
				typ = p.Elem()
			}

			// pod.Metadata.Name = "foo" -> (pod.name)
			for _, f := range t.getFields(typ.Underlying().(*types.Struct)) {
				if f.Field.Name() == n.Sel.Name {
					return t.maybeCast(&Selector{
						Expr:    t.transpileExpr(n.X),
						Field:   f.JSONName(),
						Inlined: f.JSONInline(),
					}, t.typeOf(n))
				}
			}
		}

		panic(&Unsupported{
			Node: n,
			Fset: t.Fset,
			Msg:  fmt.Sprintf("%T", t.TypesInfo.ObjectOf(n.Sel)),
		})

	case *ast.BinaryExpr:
		untyped := [3]string{"_", n.Op.String(), "_"}
		typed := [3]string{t.typeOf(n.X).String(), n.Op.String(), t.typeOf(n.Y).String()}

		f := func(op string) func(a, b Node) Node {
			return func(a, b Node) Node {
				return &BuiltInCall{FuncName: op, Arguments: []Node{a, b}}
			}
		}

		wrapWithCast := func(op, cast string) func(a, b Node) Node {
			return func(a, b Node) Node {
				return &Cast{To: cast, X: &BuiltInCall{FuncName: op, Arguments: []Node{a, b}}}
			}
		}

		// Poor man's pattern matching :[
		mapping := map[[3]string]func(a, b Node) Node{
			{"_", token.EQL.String(), "_"}:  f("eq"),
			{"_", token.NEQ.String(), "_"}:  f("ne"),
			{"_", token.LAND.String(), "_"}: f("and"),
			{"_", token.LOR.String(), "_"}:  f("or"),
			{"_", token.GTR.String(), "_"}:  f("gt"),
			{"_", token.LSS.String(), "_"}:  f("lt"),
			{"_", token.GEQ.String(), "_"}:  f("ge"),
			{"_", token.LEQ.String(), "_"}:  f("le"),

			{"float32", token.ADD.String(), "float32"}: wrapWithCast("addf", "float64"),
			{"float32", token.MUL.String(), "float32"}: wrapWithCast("mulf", "float64"),
			{"float32", token.QUO.String(), "float32"}: wrapWithCast("divf", "float32"),
			{"float32", token.SUB.String(), "float32"}: wrapWithCast("subf", "float64"),

			{"float64", token.ADD.String(), "float64"}: wrapWithCast("addf", "float64"),
			{"float64", token.MUL.String(), "float64"}: wrapWithCast("mulf", "float64"),
			{"float64", token.QUO.String(), "float64"}: wrapWithCast("divf", "float64"),
			{"float64", token.SUB.String(), "float64"}: wrapWithCast("subf", "float64"),

			{"int", token.ADD.String(), "int"}: wrapWithCast("add", "int"),
			{"int", token.MUL.String(), "int"}: wrapWithCast("mul", "int"),
			{"int", token.QUO.String(), "int"}: wrapWithCast("div", "int"),
			{"int", token.REM.String(), "int"}: wrapWithCast("mod", "int"),
			{"int", token.SUB.String(), "int"}: wrapWithCast("sub", "int"),

			{"int32", token.ADD.String(), "int32"}: wrapWithCast("add", "int"),
			{"int32", token.MUL.String(), "int32"}: wrapWithCast("mul", "int"),
			{"int32", token.QUO.String(), "int32"}: wrapWithCast("div", "int"),
			{"int32", token.REM.String(), "int32"}: wrapWithCast("mod", "int"),
			{"int32", token.SUB.String(), "int32"}: wrapWithCast("sub", "int"),

			{"int64", token.ADD.String(), "int64"}: wrapWithCast("add", "int64"),
			{"int64", token.MUL.String(), "int64"}: wrapWithCast("mul", "int64"),
			{"int64", token.QUO.String(), "int64"}: wrapWithCast("div", "int64"),
			{"int64", token.REM.String(), "int64"}: wrapWithCast("mod", "int64"),
			{"int64", token.SUB.String(), "int64"}: wrapWithCast("sub", "int64"),

			{"untyped int", token.ADD.String(), "untyped int"}: f("add"),
			{"untyped int", token.MUL.String(), "untyped int"}: f("mul"),
			{"untyped int", token.QUO.String(), "untyped int"}: f("div"),
			{"untyped int", token.REM.String(), "untyped int"}: f("mod"),
			{"untyped int", token.SUB.String(), "untyped int"}: f("sub"),

			{"untyped float", token.ADD.String(), "untyped float"}: f("addf"),
			{"untyped float", token.MUL.String(), "untyped float"}: f("mulf"),
			{"untyped float", token.QUO.String(), "untyped float"}: f("divf"),
			{"untyped float", token.SUB.String(), "untyped float"}: f("subf"),
		}

		// Typed versions take precedence.
		if funcName, ok := mapping[typed]; ok {
			return funcName(t.transpileExpr(n.X), t.transpileExpr(n.Y))
		}

		// Fallback to "wild cards" (_).
		if funcName, ok := mapping[untyped]; ok {
			return funcName(t.transpileExpr(n.X), t.transpileExpr(n.Y))
		}

		panic(&Unsupported{
			Node: n,
			Fset: t.Fset,
			Msg:  fmt.Sprintf(`No matching %T signature for %v or %v`, n, typed, untyped),
		})

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
		case token.SUB:
			if i, ok := n.X.(*ast.BasicLit); ok {
				return &Literal{Value: fmt.Sprintf("-%s", i.Value)}
			}
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
		typ := t.typeOf(n.Type)

		if basic, ok := typ.(*types.Basic); ok && (basic.Info()&types.IsNumeric != 0) {
			panic(&Unsupported{
				Node: n,
				Fset: t.Fset,
				Msg:  "type assertions on numeric types are unreliable due to JSON casting all numbers to float64's. Instead use `helmette.IsNumeric` or `helmette.AsIntegral`",
			})
		}

		return &Call{
			FuncName: "_shims.typeassertion",
			Arguments: []Node{
				t.transpileTypeRepr(typ),
				t.transpileExpr(n.X),
			},
		}
	}

	var b bytes.Buffer
	if err := format.Node(&b, t.Fset, n); err != nil {
		panic(err)
	}
	panic(fmt.Sprintf("unhandled Expr %T\n%s", n, b.String()))
}

// mkPkgTree "flattens" a loaded [packages.Package] and its dependencies into a
// map keyed by path.
func mkPkgTree(root *packages.Package) map[string]*packages.Package {
	tree := map[string]*packages.Package{}
	toVisit := []*packages.Package{root}

	// The naive approach here is crazy slow so instead we do a memomized
	// implementation.
	var pkg *packages.Package
	for len(toVisit) > 0 {
		pkg, toVisit = toVisit[0], toVisit[1:]

		if _, ok := tree[pkg.PkgPath]; ok {
			continue
		}

		tree[pkg.PkgPath] = pkg

		for _, imported := range pkg.Imports {
			toVisit = append(toVisit, imported)
		}
	}

	return tree
}

func (t *Transpiler) transpileCallExpr(n *ast.CallExpr) Node {
	var args []Node
	for _, arg := range n.Args {
		args = append(args, t.transpileExpr(arg))
	}

	callee := typeutil.Callee(t.TypesInfo, n)

	// go builtins
	if callee == nil || callee.Pkg() == nil {
		switch n.Fun.(*ast.Ident).Name {
		case "append":
			if len(args) > 2 {
				return &BuiltInCall{FuncName: "concat", Arguments: []Node{
					args[0],
					&BuiltInCall{FuncName: "list", Arguments: args[1:]},
				}}
			}
			if n.Ellipsis.IsValid() {
				return &BuiltInCall{FuncName: "concat", Arguments: args}
			}
			return &BuiltInCall{FuncName: "mustAppend", Arguments: args}
		case "int", "int32":
			return &Cast{X: args[0], To: "int"}
		case "int64":
			return &Cast{X: args[0], To: "int64"}
		case "float64":
			return &Cast{X: args[0], To: "float64"}
		case "any":
			return args[0]
		case "panic":
			return &BuiltInCall{FuncName: "fail", Arguments: args}
		case "string":
			x := t.typeOf(n.Args[0])
			if x.String() == "byte" {
				return &BuiltInCall{
					FuncName:  "printf",
					Arguments: append([]Node{&Literal{Value: "\"%c\""}}, args...),
				}
			}
			return &BuiltInCall{FuncName: "toString", Arguments: args}
		case "len":
			return t.maybeCast(&Call{FuncName: "_shims.len", Arguments: args}, types.Typ[types.Int])
		case "delete":
			return &BuiltInCall{FuncName: "unset", Arguments: args}
		default:
			panic(fmt.Sprintf("unsupport golang builtin %q", n.Fun.(*ast.Ident).Name))
		}
	}

	id := callee.Pkg().Path() + "." + callee.Name()
	signature := t.typeOf(n.Fun).(*types.Signature)

	// Before checking anything else, search for a +gotohelm:builtin=X
	// directive. If we find such a directive, we'll emit a BuiltInCall node
	// with the contents of the directive. The results are cached in t.builtins
	// as an optimization.
	if _, ok := t.builtins[id]; !ok {
		t.builtins[id] = ""
		pkg := t.packages[callee.Pkg().Path()]

		if fnDecl := findNearest[*ast.FuncDecl](pkg, callee.Pos()); fnDecl != nil {
			directives := parseDirectives(fnDecl.Doc.Text())
			t.builtins[id] = directives["builtin"]
		}
	}

	if builtin := t.builtins[id]; builtin != "" {
		if signature.Results().Len() < 2 {
			return &BuiltInCall{FuncName: builtin, Arguments: args}
		}

		// Special case, if the return signature is (T, error). We'll
		// automagically wrap the builtin invocation with (list CALLEXPR nil)
		// so it looks like this function returns an error similar to its go
		// counter part. In reality, there's no error handling in templates as
		// the template execution will be halted whenever a helper returns a
		// non-nil error.
		if named, ok := signature.Results().At(1).Type().(*types.Named); ok && named.Obj().Pkg() == nil && named.Obj().Name() == "error" {
			return &BuiltInCall{
				FuncName: "list",
				Arguments: []Node{
					&BuiltInCall{FuncName: builtin, Arguments: args},
					&Literal{Value: "nil"},
				},
			}
		}

		panic(&Unsupported{
			Fset: t.Fset,
			Node: n,
			Msg:  fmt.Sprintf("unsupported usage of builtin directive for signature: %v", signature),
		})
	}

	// Call to function within the same package. A-Okay. It's transpiled. NB:
	// This is intentionally after the builtins check to support our bootstrap
	// package's builtin bindings.
	if callee.Pkg().Path() == t.Package.PkgPath {
		// Method call.
		if r := callee.Type().(*types.Signature).Recv(); r != nil {
			typ := r.Type()

			mutable := false
			switch t := typ.(type) {
			case *types.Pointer:
				typ = t.Elem()
				mutable = true
			}

			if baseTypeName, ok := typ.(*types.Named); ok {
				var receiverArg Node

				// When receiver is a pointer then dictionary can be passed as is.
				// When receiver is not a pointer then dictionary is a deep copied.
				receiverArg = &BuiltInCall{FuncName: "deepCopy", Arguments: []Node{t.transpileExpr(n.Fun.(*ast.SelectorExpr).X)}}
				if mutable {
					receiverArg = t.transpileExpr(n.Fun.(*ast.SelectorExpr).X)
				}

				return &Call{
					FuncName: fmt.Sprintf("%s.%s.%s", t.Package.Name, baseTypeName.Obj().Name(), callee.Name()),
					// Method calls come in as a "top level" CallExpr where .Fun is the
					// selector up to that call. e.g. `Foo.Bar.Baz()` will be a `CallExpr`.
					// It's `.Fun` is a `SelectorExpr` where `.X` is `Foo.Bar`, the receiver,
					// and `.Sel` is `Baz`, the method name.
					Arguments: append([]Node{receiverArg}, args...),
				}
			}
			panic(&Unsupported{Fset: t.Fset, Node: n, Msg: "method calls with not pointer type with named type"})
		}

		// TODO need to support the namespace directive here when it's used
		// outside of the bootstrap package.
		call := &Call{FuncName: fmt.Sprintf("%s.%s", t.Package.Name, callee.Name()), Arguments: args}

		// If there's only a single return value, we'll possibly want to wrap
		// the value in a cast for safety. If there are multiple return values,
		// any casting will be handled by the transpilation of selector
		// expressions.
		if signature.Results().Len() == 1 {
			return t.maybeCast(call, signature.Results().At(0).Type())
		}
		return call
	}

	// Finally, we fall to calls to any other functions. We'll handle any
	// special cases that require a bit of extra fiddling to make work falling
	// back to a not supported message.

	switch id {
	case "sort.Strings":
		return &BuiltInCall{FuncName: "sortAlpha", Arguments: args}
	case "strings.TrimSuffix":
		return &BuiltInCall{FuncName: "trimSuffix", Arguments: []Node{args[1], args[0]}}
	case "strings.TrimPrefix":
		return &BuiltInCall{FuncName: "trimPrefix", Arguments: []Node{args[1], args[0]}}
	case "strings.ReplaceAll":
		return &BuiltInCall{FuncName: "replace", Arguments: []Node{args[1], args[2], args[0]}}
	case "k8s.io/apimachinery/pkg/util/intstr.FromInt32", "k8s.io/apimachinery/pkg/util/intstr.FromInt", "k8s.io/apimachinery/pkg/util/intstr.FromString":
		return args[0]
	case "k8s.io/utils/ptr.Deref":
		return t.maybeCast(&Call{FuncName: "_shims.ptr_Deref", Arguments: args}, signature.Results().At(0).Type())
	case "k8s.io/utils/ptr.To":
		return args[0]
	case "k8s.io/utils/ptr.Equal":
		return &Call{FuncName: "_shims.ptr_Equal", Arguments: args}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.MustDuration":
		return args[0]
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Dig":
		return &BuiltInCall{FuncName: "dig", Arguments: append(args[2:], args[1], args[0])}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Unwrap":
		return &Selector{Expr: args[0], Field: "AsMap"}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Compact2", "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Compact3":
		return &Call{FuncName: "_shims.compact", Arguments: args}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.AsIntegral":
		return &Call{FuncName: "_shims.asintegral", Arguments: args}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.AsNumeric":
		return &Call{FuncName: "_shims.asnumeric", Arguments: args}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.DictTest":
		valueType := callee.(*types.Func).Type().(*types.Signature).TypeParams().At(1)
		return &Call{FuncName: "_shims.dicttest", Arguments: append(args, t.zeroOf(valueType))}
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.AsMap":
		return t.transpileExpr(n.Fun)
	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Merge":
		dict := DictLiteral{}
		return &BuiltInCall{FuncName: "merge", Arguments: append([]Node{&dict}, args...)}
	case "gopkg.in/yaml.v3.Marshal":
		return &BuiltInCall{
			FuncName: "list",
			Arguments: []Node{
				&BuiltInCall{FuncName: "toYaml", Arguments: args},
				&Literal{Value: "nil"},
			},
		}

	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.Lookup":
		// Super ugly but it's fairly safe to assume that the return type of
		// Lookup will always be a pointer as only pointers implement
		// kube.Object.
		// Type params are difficult to work with so its easiest to extract the
		// return value (Which it a generic in Lookup) of the "instance" of the
		// function signature.
		k8sType := signature.Results().At(0).Type().(*types.Pointer).Elem().(*types.Named).Obj()

		// Step through the client set's Scheme to automatically infer the
		// APIVersion and Kind of objects. We don't want any accidental typos
		// or mistyping to occur.
		for gvk, typ := range scheme.Scheme.AllKnownTypes() {
			if typ.PkgPath() == k8sType.Pkg().Path() && typ.Name() == k8sType.Name() {
				apiVersion, kind := gvk.ToAPIVersionAndKind()

				// Inject the apiVersion and kind as arguments and snip `dot`
				// from the arguments list.
				args := append([]Node{NewLiteral(apiVersion), NewLiteral(kind)}, args[1:]...)

				return &Call{FuncName: "_shims.lookup", Arguments: args}
			}
		}

		// If we couldn't find the object in the scheme, panic. It's probably
		// due to the usage of a 3rd party resource. If you hit this, just
		// inject a Scheme into the transpiler instead of relying on the kube
		// client's builtin scheme.
		panic(fmt.Sprintf("unrecognized type: %v", k8sType))

	case "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette.TypeTest":
		typ := signature.Results().At(0).Type()
		if basic, ok := typ.(*types.Basic); ok {
			if basic.Info()&types.IsNumeric != 0 {
				panic(&Unsupported{
					Fset: t.Fset,
					Node: n,
					Msg:  "type checks on numeric types are unreliable due to JSON casting all numbers to float64's. Instead use `helmette.IsNumeric` or `helmette.AsIntegral`",
				})
			}
		}
		return &Call{FuncName: "_shims.typetest", Arguments: []Node{t.transpileTypeRepr(typ), args[0], t.zeroOf(typ)}}
	default:
		panic(fmt.Sprintf("unsupported function %q", id))
	}
}

func (t *Transpiler) transpileTypeRepr(typ types.Type) Node {
	// NB: Ideally, we'd just use typ.String(). Sadly, we can't as typ.String()
	// will return `any` but we need to match the result of fmt.Sprintf("%T")
	// which returns `interface {}`.
	switch typ := typ.(type) {
	case *types.Pointer:
		return &BuiltInCall{FuncName: "printf", Arguments: []Node{
			NewLiteral("*%s"),
			t.transpileTypeRepr(typ.Elem()),
		}}
	case *types.Array:
		return &BuiltInCall{FuncName: "printf", Arguments: []Node{
			NewLiteral(fmt.Sprintf("[%d]%%s", typ.Len())),
			t.transpileTypeRepr(typ.Elem()),
		}}
	case *types.Slice:
		return &BuiltInCall{FuncName: "printf", Arguments: []Node{
			NewLiteral("[]%s"),
			t.transpileTypeRepr(typ.Elem()),
		}}
	case *types.Map:
		return &BuiltInCall{FuncName: "printf", Arguments: []Node{
			NewLiteral("map[%s]%s"),
			t.transpileTypeRepr(typ.Key()),
			t.transpileTypeRepr(typ.Elem()),
		}}
	case *types.Basic:
		return NewLiteral(typ.String())
	case *types.Interface:
		if typ.Empty() {
			return NewLiteral("interface {}")
		}
	}
	panic(fmt.Sprintf("unsupported type: %v", typ))
}

func (t *Transpiler) transpileConst(c *types.Const) Node {
	// We could include definitions to constants and then reference
	// them. For now, it's easier to turn constants into their
	// definitions.

	if c.Val().Kind() != constant.Float {
		return t.maybeCast(&Literal{Value: c.Val().ExactString()}, c.Type())
	}

	// Floats are a bit weird. Go may store them as a quotient in some cases to
	// have an exact version of the value. .ExactString() will return the
	// quotient form (e.g. 0.1 is "1/10"). .String() will return a possibly
	// truncated value. The only other option is to get the value as a float64.
	// The second return value here is reporting if this is an exact
	// representation. It will be false for values like e, pi, and 0.1. That is
	// to say, there's not a reasonable way to handle it returning false. Given
	// that go's float64 values have the exact same problem, we're going to
	// ignore it for now and hope it's not a terrible mistake.
	as64, _ := constant.Float64Val(c.Val())
	return t.maybeCast(&Literal{Value: strconv.FormatFloat(as64, 'f', -1, 64)}, c.Type())
}

func (t *Transpiler) isString(e ast.Expr) bool {
	return types.AssignableTo(t.TypesInfo.TypeOf(e), types.Typ[types.String])
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
	case "k8s.io/apimachinery/pkg/util/intstr.IntOrString":
		// IntOrString's zero value appears to marshal to a 0 though it's
		// unclear how correct this is.
		return &Literal{Value: "0"}
	}

	switch underlying := typ.Underlying().(type) {
	case *types.Basic:
		switch underlying.Info() {
		case types.IsString:
			return &Literal{Value: `""`}
		case types.IsInteger, types.IsUnsigned | types.IsInteger:
			return &Literal{Value: "0"}
		case types.IsFloat:
			return &BuiltInCall{FuncName: "float64", Arguments: []Node{&Literal{Value: "0"}}}
		case types.IsBoolean:
			return &Literal{Value: "false"}
		default:
			panic(fmt.Sprintf("unsupported Basic type: %#v", typ))
		}

	case *types.Pointer, *types.Map, *types.Interface, *types.Slice:
		return &Nil{}

	case *types.Struct:
		var out DictLiteral

		// Skip fields that json Marshalling would itself skip.
		for _, field := range t.getFields(underlying) {
			if field.JSONOmit() || !field.IncludeInZero() {
				continue
			}

			if field.JSONInline() {
				continue
			}

			out.KeysValues = append(out.KeysValues, &KeyValue{
				Key:   strconv.Quote(field.JSONName()),
				Value: t.zeroOf(field.Field.Type()),
			})
		}
		return &out

	default:
		panic(fmt.Sprintf("unsupported type: %#v", typ))
	}
}

// getFields returns a _flattened_ list (embedded structs) of structFields for
// the given struct type.
func (t *Transpiler) getFields(root *types.Struct) []structField {
	_, rootSpec := t.getStructType(root)

	// Would be nice to have a tuple type but it's a bit too verbose for my
	// test.
	typs := []*types.Struct{root}
	specs := []*ast.StructType{rootSpec}

	var fields []structField
	for len(typs) > 0 && len(specs) > 0 {
		s := typs[0]
		spec := specs[0]

		typs = typs[1:]
		specs = specs[1:]

		for i, astField := range spec.Fields.List {
			field := structField{
				Field:      s.Field(i),
				Tag:        parseTag(s.Tag(i)),
				Definition: astField,
			}

			// If we encounter a JSON inlined field (See JSONInline for
			// details), merge the embedded struct into our list of fields to
			// support direct access thereof, just list go.
			if field.JSONInline() {
				embeddedType := field.Field.Type().(*types.Named).Underlying().(*types.Struct)
				_, embeddedSpec := t.getStructType(embeddedType)
				typs = append(typs, embeddedType)
				specs = append(specs, embeddedSpec)
			}

			fields = append(fields, field)
		}
	}

	return fields
}

// maybeCast may wrap the provided [Node] with a [Cast] to the provided
// [types.Type] if it's possible that text/template or sprig would misinterpret
// the value.
// For example: go can infer that `1` should be a float64 in some situations.
// text/template would require an explicit cast.
func (t *Transpiler) maybeCast(n Node, to types.Type) Node {
	// TODO: This can probably be optimized to not cast as frequently but
	// should otherwise perform just fine.
	if basic, ok := to.(*types.Basic); ok {
		switch basic.Kind() {
		case types.Int, types.Int32, types.UntypedInt:
			return &Cast{X: n, To: "int"}
		case types.Int64:
			return &Cast{X: n, To: "int64"}
		case types.Float64, types.UntypedFloat:
			// As a special case, floating point literals just need to contain
			// a decimal point (.) to be interpreted correctly.
			if lit, ok := n.(*Literal); ok {
				if !strings.Contains(lit.Value, ".") {
					lit.Value += ".0"
				}
				return lit
			}
			return &Cast{X: n, To: "float64"}
		}
	}
	return n
}

// getTypeSpec returns the [ast.StructType] for the given named type and the
// [packages.Package] that contains the definition.
func (t *Transpiler) getStructType(typ *types.Struct) (*packages.Package, *ast.StructType) {
	if typ.NumFields() == 0 {
		panic("unhandled")
	}

	pack := t.Package.Imports[typ.Field(0).Pkg().Path()]
	if pack == nil {
		pack = t.Package
	}

	// This is quite strange, struct
	spec := findNearest[*ast.StructType](pack, typ.Field(0).Pos())

	if spec == nil {
		panic(fmt.Sprintf("failed to resolve TypeSpec: %#v", typ))
	}

	return pack, spec
}

// omitemptyRespected return true if the `omitempty` JSON tag would be
// respected by [[json.Marshal]] for the given type.
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
	Field      *types.Var
	Tag        jsonTag
	Definition *ast.Field
}

func (f *structField) JSONName() string {
	if f.Tag.Name != "" && f.Tag.Name != "-" {
		return f.Tag.Name
	}
	return f.Field.Name()
}

// KubernetesOptional returns true if this field's comment contains any of
// Kubernetes' optional annotations.
func (f *structField) KubernetesOptional() bool {
	optional, _ := regexp.MatchString(`\+optional`, f.Definition.Doc.Text())
	return optional
}

// JSONOmit returns true if json.Marshal would omit this field. This is
// determined by checking if the field isn't exported or has the `json:"-"`
// tag.
func (f *structField) JSONOmit() bool {
	return f.Tag.Name == "-" || !f.Field.Exported()
}

// JSONInline returns true if this field should be merged with the JSON of it's
// parent rather than being placed within its own key.
func (f *structField) JSONInline() bool {
	// TODO(chrisseto) Should this respect the nonstandard ",inline" tag?
	return f.Field.Embedded() && f.Tag.Name == ""
}

// IncludeInZero returns true if this field would be included in the output
// [[json.Marshal]]'d called with a zero value of this field's parent struct.
func (f *structField) IncludeInZero() bool {
	// TODO(chrisseto): We can start producing more human readable/ergonomic
	// manifests if we process Kubernetes' +optional annotation. This however
	// breaks a lot of our tests as golang's json.Marshal does not respect
	// those annotations. It may be possible to fix by using one of Kubernetes'
	// marshallers?
	if f.JSONOmit() {
		return false
	}
	if f.Tag.OmitEmpty && omitemptyRespected(f.Field.Type()) {
		return false
	}
	return true
}

func parseDirectives(in string) map[string]string {
	match := directiveRE.FindAllStringSubmatch(in, -1)

	out := map[string]string{}
	for _, m := range match {
		out[m[1]] = m[2]
	}
	return out
}

// findNearest finds the nearest [ast.Node] to the given position. This allows
// finding the defining [ast.Node] from type instances or other such objects.
func findNearest[T ast.Node](pkg *packages.Package, pos token.Pos) T {
	// NB: It seems that pkg.Syntax is NOT ordered by position and therefore
	// can't be binary searched.
	var file *ast.File
	for _, f := range pkg.Syntax {
		if f.FileStart < pos && f.FileEnd > pos {
			file = f
			break
		}
	}

	if file == nil {
		panic(errors.Newf("pos %d not located in pkg: %v", pos, pkg))
	}

	var result T
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil || n.Pos() > pos || n.End() < pos {
			return false
		}

		if asT, ok := n.(T); ok {
			result = asT
		}

		return true
	})

	return result
}
