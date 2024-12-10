package console

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

// TestValues asserts that the chart's values.yaml file can be losslessly
// loaded into our type [Values] struct.
// NB: values.yaml should round trip through [Values], not [PartialValues], as
// [Values]'s omitempty tags are models after values.yaml.
func TestValues(t *testing.T) {
	var typedValues Values
	var unstructuredValues map[string]any

	require.NoError(t, yaml.Unmarshal(DefaultValuesYAML, &typedValues))
	require.NoError(t, yaml.Unmarshal(DefaultValuesYAML, &unstructuredValues))

	typedValuesJSON, err := json.Marshal(typedValues)
	require.NoError(t, err)

	unstructuredValuesJSON, err := json.Marshal(unstructuredValues)
	require.NoError(t, err)

	require.JSONEq(t, string(unstructuredValuesJSON), string(typedValuesJSON))
}

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	casesArchive, err := txtar.ParseFile("testdata/template-cases.txtar")
	require.NoError(t, err)

	generatedCasesArchive, err := txtar.ParseFile("testdata/template-cases-generated.txtar")
	require.NoError(t, err)

	goldens := testutil.NewTxTar(t, "testdata/template-cases.golden.txtar")

	for _, tc := range append(casesArchive.Files, generatedCasesArchive.Files...) {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			var values PartialValues
			require.NoError(t, yaml.Unmarshal(tc.Data, &values))

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
				Name:      "console",
				Namespace: "test-namespace",
				Values:    values,
				Set: []string{
					// jwtSecret defaults to a random string. Can't have that
					// in snapshot testing so set it to a static value.
					"secret.login.jwtSecret=SECRETKEY",
				},
			})
			require.NoError(t, err)

			objs, err := kube.DecodeYAML(out, Scheme)
			require.NoError(t, err)

			for _, obj := range objs {
				assert.Equalf(t, "test-namespace", obj.GetNamespace(), "%T %q did not have namespace correctly set", obj, obj.GetName())
			}

			goldens.AssertGolden(t, testutil.YAML, fmt.Sprintf("testdata/%s.yaml.golden", tc.Name), out)
		})
	}
}

// TestGenerateCases is not a test case (sorry) but a test case generator for
// the console chart.
func TestGenerateCases(t *testing.T) {
	// Nasty hack to avoid making a main function somewhere. Sorry not sorry.
	if !slices.Contains(os.Args, fmt.Sprintf("-test.run=%s", t.Name())) {
		t.Skipf("%s will only run if explicitly specified (-run %q)", t.Name(), t.Name())
	}

	// Makes strings easier to read.
	asciiStrs := func(s *string, c fuzz.Continue) {
		const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		var x []byte
		for i := 0; i < c.Intn(25); i++ {
			x = append(x, alphabet[c.Intn(len(alphabet))])
		}
		*s = string(x)
	}
	smallInts := func(s *int, c fuzz.Continue) {
		*s = c.Intn(501)
	}

	fuzzer := fuzz.New().NumElements(0, 3).SkipFieldsWithPattern(
		regexp.MustCompile("^(SELinuxOptions|WindowsOptions|SeccompProfile|TCPSocket|HTTPHeaders|VolumeSource)$"),
	).Funcs(
		asciiStrs,
		smallInts,
		func(t *corev1.ServiceType, c fuzz.Continue) {
			types := []corev1.ServiceType{
				corev1.ServiceTypeClusterIP,
				corev1.ServiceTypeExternalName,
				corev1.ServiceTypeNodePort,
				corev1.ServiceTypeLoadBalancer,
			}
			*t = types[c.Intn(len(types))]
		},
		func(s *corev1.ResourceName, c fuzz.Continue) { asciiStrs((*string)(s), c) },
		func(_ *any, c fuzz.Continue) {},
		func(_ *[]corev1.ResourceClaim, c fuzz.Continue) {},
		func(_ *[]metav1.ManagedFieldsEntry, c fuzz.Continue) {},
	)

	schema, err := jsonschema.CompileString("", string(ValuesSchemaJSON))
	require.NoError(t, err)

	nilChance := float64(0.8)

	files := make([]txtar.File, 0, 50)
	for i := 0; i < 50; i++ {
		// Every 5 iterations, decrease nil chance to ensure that we're biased
		// towards exploring most cases.
		if i%5 == 0 && nilChance > .1 {
			nilChance -= .1
		}

		var values PartialValues
		fuzzer.NilChance(nilChance).Fuzz(&values)

		out, err := yaml.Marshal(values)
		require.NoError(t, err)

		merged, err := helm.MergeYAMLValues(t.TempDir(), DefaultValuesYAML, out)
		require.NoError(t, err)

		// Ensure that our generated values comply with the schema set by the chart.
		if err := schema.Validate(merged); err != nil {
			t.Logf("Generated invalid values; trying again...\n%v", err)
			i--
			continue
		}

		files = append(files, txtar.File{
			Name: fmt.Sprintf("case-%03d", i),
			Data: out,
		})
	}

	archive := txtar.Format(&txtar.Archive{
		Comment: []byte(fmt.Sprintf(`Generated by %s`, t.Name())),
		Files:   files,
	})

	require.NoError(t, os.WriteFile("testdata/template-cases-generated.txtar", archive, 0o644))
}

func TestGoHelmEquivalence(t *testing.T) {
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	values := PartialValues{
		Tests: &PartialEnableable{
			Enabled: ptr.To(false),
		},
		Secret: &PartialSecretConfig{
			Login: &PartialLoginSecrets{
				JWTSecret: ptr.To("SECRET"),
			},
		},
		Ingress: &PartialIngressConfig{
			Enabled: ptr.To(true),
		},
	}

	goObjs, err := Chart.Render(kube.Config{}, helmette.Release{
		Name:      "gotohelm",
		Namespace: "mynamespace",
		Service:   "Helm",
	}, values)
	require.NoError(t, err)

	rendered, err := client.Template(context.Background(), ".", helm.TemplateOptions{
		Name:      "gotohelm",
		Namespace: "mynamespace",
		Values:    values,
	})
	require.NoError(t, err)

	helmObjs, err := kube.DecodeYAML(rendered, Scheme)
	require.NoError(t, err)

	slices.SortStableFunc(helmObjs, func(a, b kube.Object) int {
		aStr := fmt.Sprintf("%s/%s/%s", a.GetObjectKind().GroupVersionKind().String(), a.GetNamespace(), a.GetName())
		bStr := fmt.Sprintf("%s/%s/%s", b.GetObjectKind().GroupVersionKind().String(), b.GetNamespace(), b.GetName())
		return strings.Compare(aStr, bStr)
	})

	slices.SortStableFunc(goObjs, func(a, b kube.Object) int {
		aStr := fmt.Sprintf("%s/%s/%s", a.GetObjectKind().GroupVersionKind().String(), a.GetNamespace(), a.GetName())
		bStr := fmt.Sprintf("%s/%s/%s", b.GetObjectKind().GroupVersionKind().String(), b.GetNamespace(), b.GetName())
		return strings.Compare(aStr, bStr)
	})

	// resource.Quantity is a special object. To Ensure they compare correctly,
	// we'll round trip it through JSON so the internal representations will
	// match (assuming the values are actually equal).
	assert.Equal(t, len(helmObjs), len(goObjs))

	// Iterate and compare instead of a single comparison for better error
	// messages. Some divergences will fail an Equal check on slices but not
	// report which element(s) aren't equal.
	for i := range helmObjs {
		assert.Equal(t, helmObjs[i], goObjs[i])
	}
}
