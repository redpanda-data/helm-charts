package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/packages"
)

const (
	mode = packages.NeedTypes | packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps
)

func main() {
	pkgs := Must(packages.Load(&packages.Config{
		Mode: mode,
	}, "./charts/redpanda"))

	pkg := pkgs[0]

	root := pkg.Types.Scope().Lookup("Values").(*types.TypeName).Type()

	s := NewSchemaer(pkg)
	fmt.Printf("%s\n", Must(json.MarshalIndent(s.Schema(root), "", "\t")))
}

type Schemaer struct {
	Pkgs map[string]*packages.Package
}

func NewSchemaer(pkg *packages.Package) *Schemaer {
	seen := map[string]bool{}
	pkgs := map[string]*packages.Package{}

	q := []*packages.Package{pkg}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]

		if _, ok := seen[p.PkgPath]; ok {
			continue
		}

		seen[p.PkgPath] = true

		pkgs[p.PkgPath] = p
		maps.Copy(pkgs, p.Imports)
		q = append(q, maps.Values(p.Imports)...)
	}

	return &Schemaer{
		Pkgs: pkgs,
	}
}

func BasicKindToType(kind types.BasicKind) string {
	switch kind {
	case types.String:
		return "string"
	case types.Int, types.Int8, types.Int32, types.Int64:
		return "integer"
	case types.Uint, types.Uint8, types.Uint32, types.Uint64:
		return "integer"
	case types.Bool:
		return "boolean"
	default:
		panic(kind)
	}
}

func IsRequired(t types.Type) bool {
	switch t.(type) {
	case *types.Pointer, *types.Map, *types.Slice:
		return false
	case *types.Basic:
		return true
	case *types.Named, *types.Struct:
		return true
	case *types.Interface:
		return false
	default:
		panic(t)
	}
}

func (s *Schemaer) Schema(t types.Type) *jsonschema.Schema {
	// TODO should probably do some caching

	// fmt.Printf("%#v\n", t)

	switch t := t.(type) {
	case *types.Basic:
		return &jsonschema.Schema{
			Type: BasicKindToType(t.Kind()),
		}

	case *types.Named:
		schema := s.Schema(t.Underlying())

		// TODO this isn't going to work all the time. DST might be a better
		// option?
		pkg := s.Pkgs[t.Obj().Pkg().Path()]
		spec := FindNearest[*ast.GenDecl](pkg, t.Obj().Pos())

		markers := ParseMarkers(spec.Doc.Text())
		if markers.Enum {
			var values []*types.Const
			for _, obj := range pkg.TypesInfo.Defs {
				con, ok := obj.(*types.Const)
				if !ok || con.Type() != t {
					continue
				}
				values = append(values, con)
			}

			// Sort for stability.
			slices.SortFunc(values, func(a, b *types.Const) int {
				return strings.Compare(b.Name(), a.Name())
			})

			// TODO could probably collapse this into a single loop.
			for _, con := range values {
				switch con.Val().Kind() {
				case constant.String:
					out, _ := strconv.Unquote(con.Val().ExactString())
					schema.Enum = append(schema.Enum, out)
				default:
					panic(fmt.Sprintf("unsupported constant type: %#v\n", con.Val()))
				}
			}
		}

		return schema

	case *types.Pointer:
		// TODO??
		return s.Schema(t.Elem())

	case *types.Slice:
		return &jsonschema.Schema{Type: "array", Items: s.Schema(t.Elem())}
		// TODO.
		// return &jsonschema.Schema{
		// 	OneOf: []*jsonschema.Schema{
		// 		{Type: "null"},
		// 		{Type: "array", Items: s.Schema(t.Elem())},
		// 	},
		// }

	case *types.Interface:
		if t.Empty() {
			return jsonschema.TrueSchema
		}
		// Might be able to make this a oneOf?
		panic("unsupported")

	case *types.Map:
		// Should probably assert that keys are strings?
		return &jsonschema.Schema{
			Type:                 "object",
			AdditionalProperties: s.Schema(t.Elem()),
		}

	case *types.Struct:
		var required []string
		props := orderedmap.New[string, *jsonschema.Schema]()
		for i := 0; i < t.NumFields(); i++ {
			// TODO respect JSON tags.
			// TODO handle embeds.
			field := t.Field(i)

			if !field.Exported() {
				continue
			}

			tag := ParseTags(t.Tag(i))
			typeSchema := s.Schema(field.Type())

			astField := FindNearest[*ast.Field](s.Pkgs[field.Pkg().Path()], field.Pos())
			markers := ParseMarkers(astField.Doc.Text())

			typeSchema.Pattern = markers.Pattern

			if field.Embedded() {
				for pair := typeSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
					props.Set(pair.Key, pair.Value)
				}
				continue
			}

			name := field.Name()
			if tag.Name != "" {
				name = tag.Name
			}

			if IsRequired(field.Type()) {
				required = append(required, name)
			}

			props.Set(name, typeSchema)
		}

		return &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			Required:   required,
		}

	default:
		panic(fmt.Sprintf("%T", t))
	}
}

func ZeroOf(t types.Type) any {
	return nil
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

type Markers struct {
	Pattern  string
	Required *bool
	Enum     bool
}

func ParseMarkers(comment string) Markers {
	matches := regexp.MustCompile(`\+([^=+\s]+)(=(.+))?\n`).FindAllStringSubmatch(comment, -1)
	if matches == nil {
		return Markers{}
	}

	var m Markers
	for _, submatch := range matches {
		switch submatch[1] {
		case "kubebuilder:validation:Pattern":
			m.Pattern = submatch[3]
		case "enum":
			m.Enum = true
		case "optional":
			t := true
			m.Required = &t
		// default:
		// 	fmt.Printf("unhandled marked: %#v\n", submatch)
		}
	}

	return m
}

type JSONTag struct {
	Name      string
	OmitEmpty bool
	Inline    bool
}

func ParseTags(tag string) JSONTag {
	match := regexp.MustCompile(`json:"([^"]*)"`).FindStringSubmatch(tag)
	if match == nil {
		return JSONTag{}
	}

	idx := strings.Index(match[1], ",")
	if idx == -1 {
		return JSONTag{Name: match[1]}
	}

	return JSONTag{
		Name:      match[1][:idx],
		OmitEmpty: strings.Contains(match[1][idx:], "omitempty"),
		Inline:    strings.Contains(match[1][idx:], "inline"),
	}
}

func FindNearest[T ast.Node](pkg *packages.Package, pos token.Pos) T {
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

	var result *T
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil || n.Pos() > pos || n.End() < pos {
			return false
		}

		if asT, ok := n.(T); ok {
			result = &asT
		}

		return true
	})

	if result != nil {
		return *result
	}

	var zero T
	panic(fmt.Sprintf("No %T near %d", zero, pos))
	return zero
}
