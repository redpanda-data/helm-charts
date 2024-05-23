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
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/kube/kubetest"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"golang.org/x/tools/go/packages"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// seedObjects is a slice of kubernetes objects that will be seeded into the
// testenv for the purpose of exercising helm's `lookup` function.
var seedObjects = []kube.Object{
	&corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "namespace",
		},
	},
	&corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "hello", Port: 123},
			},
		},
	},
}

type TestSpec struct {
	Unsupported bool
	// ValuesChanged when `false`, the default, will cause the test to assert that
	// .Values has not be mutated by the chart. Set to `true` to disable.
	ValuesChanged bool
	Values        []map[string]any
}

var testSpecs = map[string]TestSpec{
	"astrewrites": {},
	"bootstrap":   {},
	"directives":  {},
	"k8s":         {},
	"mutability":  {},
	"sprig":       {},
	"syntax":      {},
	"labels": {
		Values: []map[string]any{
			{"commonLabels": map[string]any{"test": "test"}},
			{"commonLabels": map[string]any{"helm.sh/chart": "overwrite"}},
			{},
			{"commonLabels": map[string]any{}},
		},
	},
	"changing_inputs": {
		Values: []map[string]any{
			{"int": 8, "boolean": true, "string": "testing-testing"},
		},
		ValuesChanged: true,
	},
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

	// Create and populate the test environment.
	ctl := kubetest.NewEnv(t)
	for _, obj := range seedObjects {
		require.NoError(t, ctl.Create(context.Background(), obj))
	}

	runner := NewGoRunner(t, td)

	for _, pkg := range pkgs {
		pkg := pkg
		t.Run(pkg.Name, func(t *testing.T) {
			ctx := testutil.Context(t)

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

			helmRunner, err := NewHelmRunner(pkg.Name, filepath.Join(td, "src", "example", pkg.Name), ctl.RestConfig(), t.Logf)
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
						KubeConfig: kube.RestToConfig(ctl.RestConfig()),
					}

					// MUST round trip values through JSON marshalling to
					// ensure that types between go/helm/templates are the same.
					// Numbers should always be integers :[ (TODO: Can Yaml
					// technically encode the difference between ints and
					// floats?)
					dot, err = valuesutil.RoundTripThrough[map[string]any](dot)
					require.NoError(t, err)

					clonedDot, err := valuesutil.RoundTripThrough[map[string]any](dot)
					require.NoError(t, err)

					actualJSON, err := helmRunner.Render(ctx, &dot)
					require.NoError(t, err)

					gocodeJSON, err := runner.Render(ctx, &dot)
					require.NoError(t, err)

					if spec.ValuesChanged {
						require.NotEqual(t, clonedDot, dot)
					} else {
						require.Equal(t, clonedDot, dot)
					}

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

type HelmRunner struct {
	tpl    *template.Template
	logf   func(string, ...any)
	client client.Client
}

func NewHelmRunner(chartName, chartDir string, cfg *kube.RESTConfig, logf func(string, ...any)) (*HelmRunner, error) {
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	runner := &HelmRunner{
		tpl:    template.New(chartName),
		logf:   logf,
		client: c,
	}

	funcs := sprig.FuncMap()
	funcs["include"] = runner.includeFn
	funcs["lookup"] = runner.lookupFn
	funcs["toYaml"] = helmette.ToYaml

	runner.tpl = runner.tpl.Funcs(funcs)

	logf("loading %q/*.yaml...", chartDir)

	runner.tpl, err = runner.tpl.ParseGlob(filepath.Join(chartDir, "*.yaml"))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	runner.tpl, err = runner.tpl.ParseGlob(filepath.Join(chartDir, "*.tpl"))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return runner, nil
}

func (r *HelmRunner) Render(ctx context.Context, dot *helmette.Dot) (map[string]any, error) {
	out := map[string]any{}
	for _, tpl := range r.tpl.Templates() {
		spl := strings.Split(tpl.Name(), ".")
		if len(spl) != 2 || !unicode.IsUpper(rune(spl[1][0])) {
			continue
		}

		r.logf("rendering %q...", spl[1])

		var b bytes.Buffer
		if err := tpl.Execute(&b, map[string]any{"a": []any{dot}}); err != nil {
			return nil, errors.WithStack(err)
		}

		var x map[string]any
		if err := json.NewDecoder(&b).Decode(&x); err != nil {
			return nil, errors.WithStack(err)
		}

		out[spl[1]] = x["r"] // HACK
	}

	return out, nil
}

func (r *HelmRunner) includeFn(template string, args ...any) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("include accepts either 0 or 1 arguments. got: %d", len(args))
	}

	args = append(args, nil)

	var b bytes.Buffer
	if err := r.tpl.ExecuteTemplate(&b, template, args[0]); err != nil {
		return "", err
	}
	r.logf("%q(%#v) -> %s", template, args[0], b.String())
	return b.String(), nil
}

func (r *HelmRunner) lookupFn(apiVersion, kind, namespace, name string) (map[string]any, error) {
	gvk := schema.FromAPIVersionAndKind(apiVersion, kind)
	key := client.ObjectKey{Namespace: namespace, Name: name}

	obj, err := scheme.Scheme.New(gvk)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := r.client.Get(context.Background(), key, obj.(client.Object)); err != nil {
		// Match the behavior of helm which is to return an empty dictionary if
		// the object is not found.
		return map[string]any{}, client.IgnoreNotFound(err)
	}

	// Convert into an unstructured object the fun way.
	return valuesutil.UnmarshalInto[map[string]any](obj)
}

type GoRunner struct {
	inputCh  chan *helmette.Dot
	outputCh chan map[string]any
	cmd      *exec.Cmd
}

func NewGoRunner(t *testing.T, root string) *GoRunner {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = filepath.Join(root, "src", "example")
	cmd.Env = append(
		os.Environ(),
		"GOPATH="+root,
		"GO111MODULE=on",
	)

	runner := &GoRunner{
		cmd:      cmd,
		inputCh:  make(chan *helmette.Dot),
		outputCh: make(chan map[string]any),
	}

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)

	go func() {
		defer cancel()
		errChan <- runner.run(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		require.NoError(t, <-errChan)
	})

	return runner
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
			return nil, fmt.Errorf("error from go code: %s", err)
		}
		if m, ok := res["result"].(map[string]any); ok {
			return m, nil
		}
		return nil, fmt.Errorf("unexpected return %#v", res)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (g *GoRunner) run(ctx context.Context) error {
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
			return nil
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
			return nil
		}
	}
}
