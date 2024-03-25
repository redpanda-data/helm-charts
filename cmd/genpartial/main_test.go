package main

import (
	"bytes"
	alias1 "os"
	alias2 "os"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

type (
	IntAlias          int
	MapStructAlias    map[string]int
	MapGeneric[T any] map[string]T
)

type ExampleStruct struct {
	// Generics
	A1 MapGeneric[int]
	A2 MapGeneric[NestedStruct]
	A3 MapGeneric[*NestedStruct]
	A4 MapGeneric[IntAlias]

	// BasicTypes
	B1 int
	B2 *int

	// Inline structs
	C1 struct {
		Any any
		Int int
	}
	C2 *struct{}

	// Structs
	D1 NestedStruct
	D2 *NestedStruct

	// Slices
	E1 []any
	E2 []int
	E3 []*int

	// Tags
	F1 []*int `json:"L"`
	F2 string `yaml:"M"`
	F3 IntAlias

	// Struct from another package
	G1 bytes.Buffer
	G2 alias1.File
	G3 alias2.FileMode
}

type NestedStruct struct {
	Map map[string]string
}

func TestGenerateParital(t *testing.T) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:       mode,
		BuildFlags: []string{"-tags=generate"},
		Tests:      true,
	}, ".")
	require.NoError(t, err)

	// Loading with tests is weird but it let's us load up the example struct
	// seen above.
	require.Len(t, pkgs, 3)
	pkg := pkgs[1]
	require.Equal(t, "main", pkg.Name)

	require.NoError(t, PackageErrors(pkg))

	require.EqualError(t, GeneratePartial(pkg, "Values", nil), `named struct not found in package "main": "Values"`)

	var buf bytes.Buffer
	require.NoError(t, GeneratePartial(pkg, "ExampleStruct", &buf))
	testutil.AssertGolden(t, testutil.Text, "./testdata/partial.go", buf.Bytes())
}
