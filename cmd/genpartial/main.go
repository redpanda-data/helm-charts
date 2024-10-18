// main executes the "genpartial" program which loads up a go package and
// recursively generates a "partial" variant of an specified struct.
//
// If you've ever worked with the AWS SDK, you've worked with "partial" structs
// before. They are any struct where every field is nullable and the json tag
// specifies "omitempty".
//
// genpartial allows us to write structs in ergonomic go where fields that must
// always exist are presented as values rather than pointers. In cases we were
// need to marshal a partial value back to json or only specify a subset of
// values (IE helm values), use the generated partial.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/tools/go/packages"
	gofumpt "mvdan.cc/gofumpt/format"
)

const (
	mode = packages.NeedTypes | packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports
)

func main() {
	cwd, _ := os.Getwd()

	outFlag := flag.String("out", "-", "The file to output to or `-` for stdout")
	structFlag := flag.String("struct", "Values", "The struct name to generate a partial for")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: genpartial <pkg>\n")
		fmt.Printf("Example: genpartial -struct Values ./charts/redpanda\n")
		os.Exit(1)
	}

	pkgs := Must(packages.Load(&packages.Config{
		Dir:  cwd,
		Mode: mode,
	}, flag.Arg(0)))

	var buf bytes.Buffer
	if err := GeneratePartial(pkgs[0], *structFlag, &buf); err != nil {
		panic(err)
	}

	if *outFlag == "-" {
		fmt.Println(buf.Bytes())
	} else {
		if err := os.WriteFile(*outFlag, buf.Bytes(), 0o644); err != nil {
			panic(err)
		}
	}
}

type Generator struct {
	pkg   *packages.Package
	cache map[types.Type]ast.Expr
}

func (g *Generator) Generate(t types.Type) []ast.Node {
	toPartialize := FindAllNames(g.pkg.Types, t)

	var out []ast.Node
	for _, named := range toPartialize {
		// For any types that we've identified as wanting to partialize,
		// generate a new anonymous struct from the underlying struct of the
		// named type.
		// This allows the partialization algorithm to be much more sane and
		// terse. Partialization of named types is a game of deciding if the
		// reference needs to be a pointer or changed to a newly generated
		// type. Partialization of (anonymous) structs, is generation of a new
		// struct type.
		partialized := g.partialize(named.Underlying())

		var params *ast.FieldList
		if named.TypeParams().Len() > 0 {
			params = &ast.FieldList{List: make([]*ast.Field, named.TypeParams().Len())}
			for i := 0; i < named.TypeParams().Len(); i++ {
				param := named.TypeParams().At(i)
				params.List[i] = &ast.Field{
					Names: []*ast.Ident{{Name: param.Obj().Name()}},
					Type:  g.typeToNode(param.Constraint()).(ast.Expr),
				}
			}
		}

		out = append(out, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name:       &ast.Ident{Name: "Partial" + named.Obj().Name()},
						TypeParams: params,
						Type:       g.typeToNode(partialized).(ast.Expr),
					},
				},
			},
		})
	}

	return out
}

func (g *Generator) qualifier(p *types.Package) string {
	if g.pkg.Types == p {
		return "" // same package; unqualified
	}

	// Technically this could break in the case of having multiple files
	// with different import aliases.
	for _, obj := range g.pkg.TypesInfo.Defs {
		if name, ok := obj.(*types.PkgName); ok && p.Path() == name.Imported().Path() {
			return name.Name()
		}
	}

	// If no package name was found in Defs, there's no import alias.
	// Fallback to p.Name().
	return p.Name()
}

func (g *Generator) typeToNode(t types.Type) ast.Node {
	str := types.TypeString(t, g.qualifier)
	node, err := parser.ParseExpr(str)
	if err != nil {
		panic(fmt.Errorf("%s\n%v", str, err))
	}
	return node
}

func (g *Generator) partialize(t types.Type) types.Type {
	// TODO cache me.

	switch t := t.(type) {
	case *types.Basic, *types.Interface:
		return t
	case *types.Pointer:
		return types.NewPointer(g.partialize(t.Elem()))
	case *types.Map:
		return types.NewMap(t.Key(), g.partialize(t.Elem()))
	case *types.Slice:
		return types.NewSlice(g.partialize(t.Elem()))
	case *types.Struct:
		return g.partializeStruct(t)
	case *types.Named:
		return g.partializeNamed(t)
	case *types.TypeParam:
		return t // TODO this isn't super easy to fully support without a lot of additional information......
	default:
		panic(fmt.Sprintf("Unhandled: %T", t))
	}
}

func (g *Generator) partializeStruct(t *types.Struct) *types.Struct {
	tags := make([]string, t.NumFields())
	fields := make([]*types.Var, t.NumFields())
	for i := 0; i < t.NumFields(); i++ {
		field := t.Field(i)

		partialized := g.partialize(field.Type())
		switch partialized.Underlying().(type) {
		case *types.Basic, *types.Struct:
			partialized = types.NewPointer(partialized)
		}

		// TODO Docs injection would be nice but given that we're crawling the
		// type tree, that's going to be quite difficult. Could probably stash
		// away a map of types to comments and then inject that into the ast
		// after parsing?
		// Or just implement our own type printer.
		tags[i] = EnsureOmitEmpty(t.Tag(i))
		fields[i] = types.NewVar(0, g.pkg.Types, field.Name(), partialized)
	}

	return types.NewStruct(fields, tags)
}

func (g *Generator) partializeNamed(t *types.Named) types.Type {
	// This check isn't going to be correct in the long run but it's intention
	// boils down to "Have we generated a Partialized version of this named
	// type?"
	// NB: This check MUST match the check in FindAllNames.
	isPartialized := t.Obj().Pkg() == g.pkg.Types && !IsType[*types.Basic](t.Underlying())

	if !isPartialized {
		// If we haven't partialized this type, there's nothing we can do. Noop.
		return t
	}

	// If this is a partialized type, we just need to make a NamedTyped with
	// any type params that reference the partial name. The Underlying aspect
	// of this named type is ignored so we pass in the existing underlying type
	// as nil isn't acceptable.

	var args []types.Type
	for i := 0; i < t.TypeArgs().Len(); i++ {
		args = append(args, g.partialize(t.TypeArgs().At(i)))
	}

	params := make([]*types.TypeParam, t.TypeParams().Len())
	for i := 0; i < t.TypeParams().Len(); i++ {
		param := t.TypeParams().At(i)
		// Might need to clone the typename here
		params[i] = types.NewTypeParam(param.Obj(), param.Constraint())
	}

	named := types.NewNamed(types.NewTypeName(0, g.pkg.Types, "Partial"+t.Obj().Name(), t.Underlying()), t.Underlying(), nil)
	if len(args) < 1 {
		return named
	}
	named.SetTypeParams(params)
	result, err := types.Instantiate(nil, named, args, true)
	if err != nil {
		panic(err)
	}
	return result
}

func GeneratePartial(pkg *packages.Package, structName string, out io.Writer) error {
	root := pkg.Types.Scope().Lookup(structName)

	if root == nil {
		return errors.Newf("named struct not found in package %q: %q", pkg.Name, structName)
	}

	if !IsType[*types.Named](root.Type()) || !IsType[*types.Struct](root.Type().Underlying()) {
		return errors.Newf("named struct not found in package %q: %q", pkg.Name, structName)
	}

	gen := Generator{pkg: pkg, cache: map[types.Type]ast.Expr{}}

	partials := gen.Generate(root.Type())

	// Map of import aliases to actual package for all imports in the
	// originally parsed package.
	// One of the more frustrating aspects of Go's AST/Type system is dealing
	// with imports and their aliases. The best way to get them is to crawl the
	// AST for import specs and then manually resolve them.
	originalImports := map[string]*types.Package{}
	for _, f := range pkg.Syntax {
		for _, imp := range f.Imports {
			// For some reason, imports can be nil.
			if imp == nil {
				continue
			}

			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				panic(err)
			}

			imported := pkg.Imports[path]
			name := imported.Name

			// if an alias is specified, use it.
			if imp.Name != nil {
				name = imp.Name.Name
			}

			originalImports[name] = imported.Types
		}
	}

	// Now that we have our partial structs, we need to generate the import
	// block for them. We'll crawl the AST of the structs and find references
	// to external packages. This method could possibly lead to conflicts as
	// we're just looking for [ast.SelectorExpr]'s
	imports := map[string]*types.Package{}
	for _, partial := range partials {
		ast.Inspect(partial, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.SelectorExpr:
				parent, ok := n.X.(*ast.Ident)
				if !ok {
					return true
				}

				if pkg, ok := originalImports[parent.Name]; ok {
					imports[parent.Name] = pkg
				}
			}
			return true
		})
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "//go:build !generate\n\n")
	fmt.Fprintf(&buf, "// +gotohelm:ignore=true\n")
	fmt.Fprintf(&buf, "//\n")
	// This line must match `^// Code generated .* DO NOT EDIT\.$`. See https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
	fmt.Fprintf(&buf, "// Code generated by genpartial DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "package %s\n\n", pkg.Name)

	// Only print out imports if we have them. We lean on source.Format later
	// to align and sort them.
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for name, pkg := range imports {
			if pkg.Name() == name {
				fmt.Fprintf(&buf, "\t%q\n", pkg.Path())
			} else {
				fmt.Fprintf(&buf, "\t%s %q\n", name, pkg.Path())
			}
		}
		fmt.Fprintf(&buf, ")\n\n")
	}

	for i, d := range partials {
		if i > 0 {
			fmt.Fprintf(&buf, "\n\n")
		}
		format.Node(&buf, token.NewFileSet(), d)
	}

	formatted, err := gofumpt.Source(buf.Bytes(), gofumpt.Options{})
	if err != nil {
		return err
	}

	_, err = out.Write(formatted)
	return err
}

// FindAllNames traverses the given type and returns a slice of all non-Basic
// named types that are referenced from the "root" type.
func FindAllNames(pkg *types.Package, root types.Type) []*types.Named {
	names := []*types.Named{}
	seen := map[types.Type]struct{}{}
	toTraverse := []types.Type{}

	push := func(t types.Type) {
		if _, ok := seen[t]; ok {
			return
		}

		// Partialize all named types within this the provided package that are
		// not aliases for Basic types.
		// This could be "more efficient" by avoiding partialization of named
		// types that don't require changes but that's much more error prone
		// and makes working with partialized types a bit strange.
		if named, ok := t.(*types.Named); ok && named.Obj().Pkg() == pkg && named.Origin() == named {
			switch named.Underlying().(type) {
			case *types.Basic:
			default:
				names = append(names, named)
			}
		}

		seen[t] = struct{}{}
		toTraverse = append(toTraverse, t)
	}

	push(root)

	for len(toTraverse) > 0 {
		current := toTraverse[0]
		toTraverse = toTraverse[1:]

		push(current.Underlying())

		switch current := current.(type) {
		case *types.Basic, *types.Interface, *types.TypeParam:
			continue

		case *types.Pointer:
			push(current.Elem())

		case *types.Slice:
			push(current.Elem())

		case *types.Map:
			push(current.Key())
			push(current.Elem())

		case *types.Struct:
			for i := 0; i < current.NumFields(); i++ {
				push(current.Field(i).Type())
			}
		case *types.Named:
			push(current.Origin())
			for i := 0; i < current.TypeArgs().Len(); i++ {
				push(current.TypeArgs().At(i))
			}

		default:
			panic(fmt.Sprintf("unhandled: %T", current))
		}
	}

	return names
}

var jsonTagRE = regexp.MustCompile(`json:"([^"]+)"`)

// EnsureOmitEmpty injects ,omitempty into existing json tags or adds one if
// not already present.
func EnsureOmitEmpty(tag string) string {
	if !strings.Contains(tag, `json:"`) {
		return strings.TrimLeft(tag+` json:",omitempty"`, " ")
	}

	return jsonTagRE.ReplaceAllStringFunc(tag, func(s string) string {
		if strings.Contains(s, ",omitempty") {
			return s
		}
		return s[:len(s)-1] + `,omitempty"`
	})
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func IsType[T types.Type](typ types.Type) bool {
	_, ok := typ.(T)
	return ok
}
