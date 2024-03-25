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
	"go/token"
	"go/types"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/cockroachdb/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

const (
	mode = packages.NeedTypes | packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo
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
		Dir:        cwd,
		Mode:       mode,
		BuildFlags: []string{"-tags=generate"},
	}, flag.Arg(0)))

	var buf bytes.Buffer
	if err := GeneratePartial(pkgs[0], *structFlag, &buf); err != nil {
		panic(err)
	}

	if *outFlag == "-" {
		fmt.Println(buf.String())
	} else {
		if err := os.WriteFile(*outFlag, buf.Bytes(), 0o644); err != nil {
			panic(err)
		}
	}
}

// PackageErrors returns any error reported by pkg during load or nil.
func PackageErrors(pkg *packages.Package) error {
	for _, err := range pkg.Errors {
		return err
	}

	for _, err := range pkg.TypeErrors {
		return err
	}

	return nil
}

func GeneratePartial(pkg *packages.Package, structName string, out io.Writer) error {
	root := pkg.Types.Scope().Lookup(structName)

	if root == nil {
		return errors.Newf("named struct not found in package %q: %q", pkg.Name, structName)
	}

	if !IsType[*types.Named](root.Type()) || !IsType[*types.Struct](root.Type().Underlying()) {
		return errors.Newf("named struct not found in package %q: %q", pkg.Name, structName)
	}

	names := FindAllNames(root.Type())
	nameMap := map[string]*types.Named{}

	for _, name := range names {
		nameMap[name.Obj().Name()] = name
	}

	var partials []ast.Node
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			// nil decls indicate an empty line, skip over them.
			if decl == nil {
				continue
			}

			// Skip over any non-type declaration.
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			// We now know that genDecl contains a type declaration. Traverse
			// the AST that comprises it and rewrite all fields to be nullable
			// have an omitempty json tag.
			partial := astutil.Apply(genDecl, func(c *astutil.Cursor) bool {
				switch node := c.Node().(type) {
				case *ast.Comment:
					// Remove comments
					c.Delete()
					return false

				case *ast.TypeSpec:
					// For any type spec, if it's in our list of named types,
					// rename it to "Partial<Name>".
					if _, ok := nameMap[node.Name.String()]; ok {
						original := node.Name.Name
						node.Name.Name = "Partial" + original
						node.Name.Obj.Name = "Partial" + original
						// TODO: Generate some nice comments. The trimming of
						// original comments will remove generated comments as
						// well.
						// node.Doc = &ast.CommentGroup{
						// 	List: []*ast.Comment{
						// 		{Text: fmt.Sprintf(`// %s is a generated "Partial" variant of [%s]`, node.Name.Name, original)},
						// 	},
						// }
						return true
					}
					// Or delete it if it's not in our list.
					c.Delete()
					return false

				case *ast.Ident:
					// Rewrite all identifiers to be a nullable version.
					switch parent := c.Parent().(type) {
					case *ast.StarExpr, *ast.ArrayType, *ast.IndexExpr, *ast.MapType:
						if _, ok := nameMap[node.Name]; ok {
							node.Name = "Partial" + node.Name
						}
						return false

					case *ast.Field:
						if parent.Type != node {
							return true
						}

						if _, ok := nameMap[node.Name]; ok {
							node.Name = "Partial" + node.Name
							c.Replace(&ast.StarExpr{X: node})
							return false
						}

						// If Obj is nil, this is a builtin type like int,
						// string, etc. We want these to become *int, *string.
						// "any" however is already nullable, so skip that.
						if node.Obj == nil && node.Name != "any" {
							c.Replace(&ast.StarExpr{X: node})
							return false
						}
						return true
					}

					return false

				case *ast.Field:
					if node.Tag == nil {
						node.Tag = &ast.BasicLit{Value: "``"}
					}
					node.Tag.Value = EnsureOmitEmpty(node.Tag.Value)
					return true

				default:
					return true
				}
			}, nil).(*ast.GenDecl)

			// If we've filtered out all the specs, skip over this declaration.
			if len(partial.Specs) == 0 {
				continue
			}

			partials = append(partials, partial)
		}
	}

	// Now that we have our partial structs, we need to generate the import
	// block for them. Because we're doing fancy re-writing to avoid some extra
	// work, we have to do a bit of extra work to make sure import aliases are preserved.
	//
	// Traverse the AST of our partial structs looking for identifiers that
	// refer to packages (PkgNames really) and store them into a set that we'll
	// later emit.
	imports := map[string]*types.Package{}
	for _, partial := range partials {
		ast.Inspect(partial, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				obj := pkg.TypesInfo.ObjectOf(n)
				if pkgName, ok := obj.(*types.PkgName); ok {
					imports[pkgName.Name()] = pkgName.Imported()
				}
			}
			return true
		})
	}

	// Printout the resultant set of structs to `out`. We could generate an
	// ast.File and print that but it's a bit finicky. Printf and then
	// formatting rewritten nodes is easier.
	fmt.Fprintf(out, "// !DO NOT EDIT! Generated by genpartial\n")
	fmt.Fprintf(out, "//\n")
	fmt.Fprintf(out, "//go:build !generate\n")
	fmt.Fprintf(out, "//+gotohelm:ignore=true\n")
	fmt.Fprintf(out, "package %s\n\n", pkg.Name)

	// Only print out imports if we have them. Would be nice to just lean on go
	// fmt for this but generating something for use with format.File is quite
	// difficult. It might be worth while to try out format.Source so we don't
	// have to worry about formatting or spacing while printing.
	if len(imports) > 0 {
		names := maps.Keys(imports)
		sort.Strings(names)

		fmt.Fprintf(out, "import (\n")
		for _, name := range names {
			pkg := imports[name]
			if pkg.Name() == name {
				fmt.Fprintf(out, "\t%q\n", pkg.Path())
			} else {
				fmt.Fprintf(out, "\t%s %q\n", name, pkg.Path())
			}
		}
		fmt.Fprintf(out, ")\n\n")
	}

	for i, d := range partials {
		if i > 0 {
			fmt.Fprintf(out, "\n\n")
		}
		format.Node(out, pkg.Fset, d)
	}
	fmt.Fprintf(out, "\n")

	return nil
}

// FindAllNames traverses the given type and returns a slice of all named types
// that are referenced from the "root" type.
func FindAllNames(root types.Type) []*types.Named {
	names := []*types.Named{}

	switch root := root.(type) {
	case *types.Pointer:
		names = append(names, FindAllNames(root.Elem())...)

	case *types.Slice:
		names = append(names, FindAllNames(root.Elem())...)

	case *types.Named:
		if _, ok := root.Underlying().(*types.Basic); ok {
			break
		}

		names = append(names, root)
		names = append(names, FindAllNames(root.Underlying())...)

		for i := 0; i < root.TypeArgs().Len(); i++ {
			arg := root.TypeArgs().At(i)
			if named, ok := arg.(*types.Named); ok {
				names = append(names, FindAllNames(named)...)
			}
		}

	case *types.Map:
		names = append(names, FindAllNames(root.Key())...)
		names = append(names, FindAllNames(root.Elem())...)

	case *types.Struct:
		for i := 0; i < root.NumFields(); i++ {
			field := root.Field(i)
			// TODO how to handle Embeds?
			names = append(names, FindAllNames(field.Type())...)
		}
	}

	return names
}

var jsonTagRE = regexp.MustCompile(`json:"([^,"]+)"`)

// EnsureOmitEmpty injects ,omitempty into existing json tags or adds one if
// not already present.
func EnsureOmitEmpty(tag string) string {
	if !jsonTagRE.MatchString(tag) {
		return tag[:len(tag)-1] + `json:",omitempty"` + "`"
	}
	return jsonTagRE.ReplaceAllString(tag, `json:"$1,omitempty"`)
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
