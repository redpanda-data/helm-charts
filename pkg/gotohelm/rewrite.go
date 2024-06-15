package gotohelm

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/cockroachdb/errors"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type astRewrite func(*packages.Package, *ast.File) (_ *ast.File, changed bool)

const (
	shimsPkg     = "helmette"
	shimsPkgPath = "github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

// NB: Order is very important here.
var rewrites = []astRewrite{
	hoistIfs,
	rewriteMultiValueSyntaxToHelpers,
	rewriteMultiValueReturns,
}

// LoadPackages is a wrapper around [packages.Load] that performs a handful of
// AST rewrites followed by a second invocation of [packages.Load] to
// appropriately populate the AST.
// AST rewriting is done to keep the transpilation process to be as simple as
// possible. Any unsuported or non-trivially supported expressions/statements
// will be rewritten to supported equivalents instead.
// If need be, the rewritten files can also be dumped to disk and have assertions made
func LoadPackages(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
	// Ensure we're getting all the values we need (which is pretty much everything...)
	cfg.Mode |= packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
		packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo |
		packages.NeedDeps | packages.NeedModule

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return pkgs, err
	}

	if cfg.Overlay == nil {
		cfg.Overlay = map[string][]byte{}
	}

	for _, pkg := range pkgs {
		var errs []error
		for i := range pkg.Errors {
			e := pkg.Errors[i]
			errs = append(errs, e)
		}

		for i := range pkg.TypeErrors {
			e := pkg.Errors[i]
			errs = append(errs, e)
		}

		if len(errs) > 0 {
			return nil, errors.Wrapf(errors.Join(errs...), "package %s", pkg.Name)
		}

		for _, parsed := range pkg.Syntax {
			filename := pkg.Fset.File(parsed.Pos()).Name()

			var changed bool
			for _, rewrite := range rewrites {
				var didChange bool
				parsed, didChange = rewrite(pkg, parsed)
				changed = changed || didChange
			}

			if !changed {
				continue
			}

			var buf bytes.Buffer
			if err := format.Node(&buf, pkg.Fset, parsed); err != nil {
				return nil, err
			}

			cfg.Overlay[filename] = buf.Bytes()
		}
	}

	pkgs, err = packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		var errs []error
		for _, e := range pkg.Errors {
			errs = append(errs, e)
		}

		for _, e := range pkg.TypeErrors {
			errs = append(errs, e)
		}

		if len(errs) > 0 {
			return nil, errors.Wrapf(errors.Join(errs...), "package %s", pkg.Name)
		}
	}

	return pkgs, nil
}

// typeToNode returns an [ast.Expr] representing the provided type.
func typeToNode(pkg *packages.Package, typ types.Type) ast.Expr {
	qualifier := func(p *types.Package) string {
		if p.Path() == pkg.PkgPath {
			return ""
		}

		// Technically this could break in the case of having multiple files
		// with different import aliases.
		for _, obj := range pkg.TypesInfo.Defs {
			if name, ok := obj.(*types.PkgName); ok && p.Path() == name.Imported().Path() {
				return name.Name()
			}
		}

		// If no package name was found in Defs, there's no import alias.
		// Fallback to p.Name().
		return p.Name()
	}

	// This should only happen if a rewrite forgot to update TypesInfo or
	// someone called TypeOf on `_`.
	if typ == nil {
		panic("nil type")
	}

	s := types.TypeString(typ, qualifier)

	expr, err := parser.ParseExpr(s)
	if err != nil {
		panic(fmt.Sprintf("pkg errors (%s) with type (%s): %v", pkg.Name, s, err))
	}

	return expr
}

// rewriteMultiValueReturns rewrites instances of multi-value returns into an
// equivalent set of statements that utilizes a tuple followed by unpacking it.
//
//	x, y := f(a)
//
//	mvr := Compact2(f(a))
//	x := mvr.First
//	y := mvr.Second
func rewriteMultiValueReturns(pkg *packages.Package, f *ast.File) (*ast.File, bool) {
	fset := pkg.Fset
	info := pkg.TypesInfo

	var count int
	f = astutil.Apply(f, func(c *astutil.Cursor) bool {
		assignment, ok := c.Node().(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assignment.Lhs) < 2 || len(assignment.Rhs) != 1 {
			return true
		}

		count++
		mvr := ast.NewIdent(fmt.Sprintf("tmp_tuple_%d", count))

		// TODO might be nicer to call c.InsertAfter in reverse order because
		// unpacking ends up looking "backwards".
		unpacked := 0
		var typeArgs []ast.Expr

		rhsTypes := info.TypeOf(assignment.Rhs[0]).(*types.Tuple)

		for i, v := range assignment.Lhs {
			typeArgs = append(typeArgs, typeToNode(pkg, rhsTypes.At(i).Type()))

			// Skip over any blackhole assignments.
			if ident, ok := v.(*ast.Ident); ok && ident.Name == "_" {
				continue
			}

			unpacked++

			c.InsertAfter(&ast.AssignStmt{
				Lhs: []ast.Expr{v},
				Tok: assignment.Tok,
				Rhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   mvr,
						Sel: ast.NewIdent(fmt.Sprintf("T%d", i+1)),
					},
				},
			})
		}

		tok := assignment.Tok

		if unpacked == 0 {
			mvr = &ast.Ident{Name: "_"}
			tok = token.ASSIGN
		}

		c.Replace(&ast.AssignStmt{
			Lhs: []ast.Expr{mvr},
			Tok: tok,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent(shimsPkg),
						Sel: ast.NewIdent(fmt.Sprintf("Compact%d", len(typeArgs))),
					},
					// TODO(chrisseto): This is commented out for the worse
					// possible reason. It seems that format.Node is _slightly_
					// non-deterministic in the case of long lines. The easiest
					// way to work around that for now is to not include
					// explicit type hints to CompactN as go seems to be able
					// to infer most cases.
					// Fun: &ast.IndexListExpr{
					// 	X: &ast.SelectorExpr{
					// 		X:   ast.NewIdent(shimsPkg),
					// 		Sel: ast.NewIdent(fmt.Sprintf("Compact%d", len(typeArgs))),
					// 	},
					// 	Indices: typeArgs,
					// },
					Args: assignment.Rhs,
				},
			},
		})

		return true
	}, nil).(*ast.File)

	if count > 0 {
		_ = astutil.AddImport(fset, f, shimsPkgPath)
	}

	return f, count > 0
}

// rewriteMultiValueSyntaxToHelpers rewrites instances of multi-value return
// syntax, such as dictionary tests and type tests into equivalent function
// invocations.
//
//	t, ok := x.(type)
//
//	t, ok := DictTest[keytype, valuetype](m, k)
func rewriteMultiValueSyntaxToHelpers(pkg *packages.Package, f *ast.File) (*ast.File, bool) {
	count := 0
	fset := pkg.Fset

	replace := func(c *astutil.Cursor, replacement ast.Expr) {
		// Increment count so we know when replacements have occurred.
		count++

		// Populate .Types for our replacement so downstream rewrites can
		// depend on .TypeInfo without having to reparse the entire package.
		pkg.TypesInfo.Types[replacement] = types.TypeAndValue{
			Type: pkg.TypesInfo.TypeOf(c.Node().(ast.Expr)),
		}

		// Actually replace the node.
		c.Replace(replacement)
	}

	f = astutil.Apply(f, func(c *astutil.Cursor) bool {
		assignment, ok := c.Parent().(*ast.AssignStmt)
		if !ok {
			return true
		}

		if len(assignment.Lhs) != 2 || len(assignment.Rhs) != 1 {
			return true
		}

		if assignment.Rhs[0] != c.Node() {
			return true
		}

		switch node := c.Node().(type) {
		case *ast.IndexExpr:
			// x, ok := m[key] -> x, ok := DictTest[K, V](y, key)
			typ := pkg.TypesInfo.TypeOf(node.X).Underlying().(*types.Map)

			replace(c, &ast.CallExpr{
				Fun: &ast.IndexListExpr{
					X: &ast.SelectorExpr{X: ast.NewIdent(shimsPkg), Sel: ast.NewIdent("DictTest")},
					Indices: []ast.Expr{
						typeToNode(pkg, typ.Key()),
						typeToNode(pkg, typ.Elem()),
					},
				},
				Args: []ast.Expr{node.X, node.Index},
			})

		case *ast.TypeAssertExpr:
			// x, ok := y.(type) -> x, ok := TypeTest[type](y)
			replace(c, &ast.CallExpr{
				Fun: &ast.IndexExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent(shimsPkg),
						Sel: ast.NewIdent("TypeTest"),
					},
					Index: node.Type,
				},
				Args: []ast.Expr{node.X},
			})
		}

		return true
	}, nil).(*ast.File)

	if count > 0 {
		_ = astutil.AddImport(fset, f, shimsPkgPath)
	}

	return f, count > 0
}

// hoistIfs "hoists" all assignments within an if else chain to be above said
// chain. It munges the variable names to ensure that variable shadowing
// doesn't become an issues.
// NOTE: All assignments within if-else chains MUST expect to be called as if
// hoisting nullifies the capabilities of short-circuiting.
//
//	if x, ok := m[k1]; ok {
//	} y, ok := m[k2]; ok {
//	}
//
// Will get rewritten to:
//
//	x, ok_1 := m[k1]
//	y, ok_2 := m[k2]
//
//	if ok_1 {
//	} else if ok_2 {
//	}
func hoistIfs(pkg *packages.Package, f *ast.File) (*ast.File, bool) {
	count := 0
	info := pkg.TypesInfo
	renames := map[*ast.Object]*ast.Ident{}

	return astutil.Apply(f, func(c *astutil.Cursor) bool {
		node, ok := c.Node().(*ast.IfStmt)
		if !ok || node.Init == nil {
			return true
		}

		for _, v := range node.Init.(*ast.AssignStmt).Lhs {
			old := v.(*ast.Ident)
			if old.Name == "_" {
				continue
			}

			count++
			new := ast.NewIdent(fmt.Sprintf("%s_%d", old.Name, count))
			new.Obj = old.Obj

			renames[old.Obj] = new

			info.Defs[new] = info.Defs[old]
			info.Instances[new] = info.Instances[old]
		}

		return true
	}, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.Ident:
			if rename, ok := renames[node.Obj]; ok {
				c.Replace(rename)
			}

		case *ast.IfStmt:
			// Don't process if-else statements as c.InsertBefore will panic.
			// Instead, we loop through the first if and hoist all child
			// assignments.
			if _, ok := c.Parent().(*ast.IfStmt); ok {
				return true
			}

			for n := node; n != nil; {
				if n.Init != nil {
					c.InsertBefore(n.Init)
					n.Init = nil
				}

				n, _ = n.Else.(*ast.IfStmt)
			}
		}

		return true
	}).(*ast.File), count > 0
}
