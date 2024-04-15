package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"strings"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestGenerateSchema(t *testing.T) {
	for _, spec := range []string{
		"k8s.io/api/core/v1.PodFSGroupChangePolicy",
		"k8s.io/api/core/v1.PodSpec",
		"k8s.io/api/core/v1.TopologySpreadConstraint",
	} {
		idx := strings.LastIndex(spec, ".")
		pkgPath := spec[:idx]
		typeName := spec[idx+1:]

		// TODO: Could speed this up by loading all specs at once.
		pkgs, err := packages.Load(&packages.Config{
			Mode: mode,
		}, pkgPath)
		require.NoError(t, err)

		pkg := pkgs[0]

		root := pkg.Types.Scope().Lookup(typeName).(*types.TypeName).Type()

		s := NewSchemaer(pkg)

		out, err := json.MarshalIndent(s.Schema(root), "", "\t")
		require.NoError(t, err)

		testutil.AssertGolden(t, testutil.JSON, fmt.Sprintf("testdata/%s.schema.json", strings.ToLower(typeName)), out)
	}
}
