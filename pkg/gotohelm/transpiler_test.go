package gotohelm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/cockroachdb/errors"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"golang.org/x/tools/go/packages"
)

type TestSpec struct {
	Unsupported bool
	// ValuesChanged when `false`, the default, will cause the test to assert that
	// .Values has not be mutated by the chart. Set to `true` to disable.
	ValuesChanged bool
	Values        []map[string]any
}

var testSpecs = map[string]TestSpec{
	"a":   {},
	"b":   {},
	"k8s": {},
	"labels": {
		Values: []map[string]any{
			{"commonLabels": map[string]any{"test": "test"}},
			{"commonLabels": map[string]any{"helm.sh/chart": "overwrite"}},
			{},
			{"commonLabels": map[string]any{}},
		},
	},
	"syntax": {},
	"sprig":  {},
	"changing_inputs": {
		Values: []map[string]any{
			{"int": 8, "boolean": true, "string": "testing-testing"},
		},
		ValuesChanged: true,
	},
	"directives": {},
	"mutability": {},
	"inputs": {
		Values: []map[string]any{
			{"foo": 1, "bar": "baz", "nested": map[string]any{"quux": true}},
			{"foo": []any{}, "bar": "baz", "nested": map[string]any{"quux": "hello"}},
			{"foo": []any{}, "bar": "baz", "nested": map[string]any{"quux": 1}},
			{"foo": []any{}, "bar": "baz", "nested": map[string]any{"quux": []string{"1", "2"}}},
		},
	},
	"flowcontrol": {
		Values: []map[string]any{
			{"ints": []int{}, "boolean": true, "oneToFour": 1},
			{"ints": []int{}, "boolean": false, "oneToFour": 2},
			{"ints": []int{1, 2, 3}, "boolean": false, "oneToFour": 3},
			{"ints": []int{1, 2, 3}, "boolean": false, "oneToFour": 4},
		},
	},
	"typing": {
		Values: []map[string]any{
			{"t": int(1)},
			{"t": float64(1)},
			{"t": true},
			{"t": "a string"},
		},
	},
}

func TestTranspile(t *testing.T) {
	td, err := filepath.Abs("testdata")
	require.NoError(t, err)

	pkgs, err := LoadPackages(&packages.Config{
		Dir:   td + "/src/example",
		Tests: true,
		Env: append(
			os.Environ(),
			"GOPATH="+td,
			"GO111MODULE=on",
		),
	}, "./...")
	require.NoError(t, err)

	// Ensure there are no compile errors before proceeding.
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			require.NoErrorf(t, err, "failed to compile %q", pkg.Name)
		}
	}

	ctx := testutil.Context(t)
	runner := NewGoRunner(td)

	go func() {
		require.NoError(t, runner.Run(ctx))
	}()

	for _, pkg := range pkgs {
		pkg := pkg
		t.Run(pkg.Name, func(t *testing.T) {
			spec, ok := testSpecs[pkg.Name]
			if !ok {
				t.Skipf("no test spec for %q", pkg.Name)
			}

			if spec.Unsupported {
				t.Skipf("%q is not currently supported", pkg.Name)
			}

			chart, err := Transpile(pkg)
			require.NoError(t, err)

			for _, f := range chart.Files {
				var actual bytes.Buffer
				f.Write(&actual)

				output := filepath.Join(td, "src", "example", pkg.Name, f.Name)
				testutil.AssertGolden(t, testutil.Text, output, actual.Bytes())
			}

			// Ensure syntactic validity of generated values.
			var tpl *template.Template
			funcs := sprig.FuncMap()
			funcs["include"] = func(template string, args ...any) (string, error) {
				if len(args) > 1 {
					return "", fmt.Errorf("include accepts either 0 or 1 arguments. got: %d", len(args))
				}

				args = append(args, nil)

				var b bytes.Buffer
				if err := tpl.ExecuteTemplate(&b, template, args[0]); err != nil {
					return "", err
				}
				t.Logf("%q(%#v) -> %s", template, args[0], b.String())
				return b.String(), nil
			}
			tpl, err = template.New(pkg.Name).Funcs(funcs).ParseGlob(filepath.Join(td, "src", "example", pkg.Name, "*.yaml"))
			require.NoError(t, err)
			tpl, err = tpl.ParseFiles(filepath.Join(td, "src", "example", pkg.Name, "_shims.tpl"))
			require.NoError(t, err)

			// If .Values isn't explicitly specified, default to an empty object.
			if spec.Values == nil {
				spec.Values = append(spec.Values, map[string]any{})
			}

			for i, values := range spec.Values {
				values := values

				t.Run(fmt.Sprintf("Values%d", i), func(t *testing.T) {
					t.Logf("using values: %#v", values)

					dot := helmette.Dot{
						Values: values,
						Chart: helmette.Chart{
							Name:    pkg.Name,
							Version: "1.2.3",
						},
						Release: helmette.Release{
							Name:      "release-name",
							Namespace: "release-namespace",
						},
					}

					// MUST round trip values through JSON marshalling to
					// ensure that types between go/helm/templates are the same.
					// Numbers should always be integers :[ (TODO: Can Yaml
					// technically encode the difference between ints and
					// floats?)
					dotJSON, err := json.Marshal(dot)
					require.NoError(t, err)
					require.NoError(t, json.Unmarshal(dotJSON, &dot))
					var clonedDot helmette.Dot
					require.NoError(t, json.Unmarshal(dotJSON, &clonedDot))

					actualJSON := map[string]any{}
					for _, tpl := range tpl.Templates() {
						spl := strings.Split(tpl.Name(), ".")
						if len(spl) != 2 || !unicode.IsUpper(rune(spl[1][0])) {
							continue
						}

						var b bytes.Buffer
						require.NoError(t, tpl.Execute(&b, map[string]any{"a": []any{dot}}))
						if spec.ValuesChanged {
							require.NotEqual(t, clonedDot, dot)
						} else {
							require.Equal(t, clonedDot, dot)
						}

						var x map[string]any
						require.NoError(t, json.Unmarshal(b.Bytes(), &x))
						actualJSON[spl[1]] = x["r"] // HACK
					}

					gocodeJSON, err := runner.Render(ctx, &dot)
					require.NoError(t, err)

					goPretty, err := json.MarshalIndent(gocodeJSON, "", "\t")
					require.NoError(t, err)

					tplPretty, err := json.MarshalIndent(actualJSON, "", "\t")
					require.NoError(t, err)

					t.Logf("go code output:\n%s", goPretty)
					t.Logf("template output:\n%s", tplPretty)

					require.Equal(t, gocodeJSON, actualJSON, "Divergence between Go code and generated template")
				})
			}
		})
	}
}

type GoRunner struct {
	inputCh  chan *helmette.Dot
	outputCh chan map[string]any
	cmd      *exec.Cmd
}

func NewGoRunner(root string) *GoRunner {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = filepath.Join(root, "src", "example")
	cmd.Env = append(
		os.Environ(),
		"GOPATH="+root,
		"GO111MODULE=on",
	)

	return &GoRunner{
		cmd:      cmd,
		inputCh:  make(chan *helmette.Dot),
		outputCh: make(chan map[string]any),
	}
}

func (g *GoRunner) Render(ctx context.Context, dot *helmette.Dot) (map[string]any, error) {
	select {
	case g.inputCh <- dot:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-g.outputCh:
		if err, ok := res["err"]; ok && err != nil {
			return nil, fmt.Errorf("%#v", err)
		}
		if m, ok := res["result"].(map[string]any); ok {
			return m, nil
		}
		return nil, fmt.Errorf("unexpected return %#v", res)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (g *GoRunner) Run(ctx context.Context) error {
	defer close(g.inputCh)
	defer close(g.outputCh)

	stdin, err := g.cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := g.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := g.cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := g.cmd.Start(); err != nil {
		return err
	}

	enc := json.NewEncoder(stdin)
	dec := json.NewDecoder(stdout)
	for {
		var in *helmette.Dot

		select {
		case in = <-g.inputCh:
		case <-ctx.Done():
			return ctx.Err()
		}

		if err := enc.Encode(in); err != nil {
			stderrout, _ := io.ReadAll(stderr)
			return errors.Wrapf(err, "stderr: %s", stderrout)
		}

		var out map[string]any
		if err := dec.Decode(&out); err != nil {
			stderrout, _ := io.ReadAll(stderr)
			return errors.Wrapf(err, "stderr: %s", stderrout)
		}

		select {
		case g.outputCh <- out:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
