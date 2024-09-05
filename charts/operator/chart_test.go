package operator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	fuzz "github.com/google/gofuzz"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	require.NoError(t, client.RepoAdd(ctx, "prometheus", "https://prometheus-community.github.io/helm-charts"))
	require.NoError(t, client.DependencyBuild(ctx, "."), "failed to refresh helm dependencies")

	casesArchive, err := txtar.ParseFile("testdata/template-cases.txtar")
	require.NoError(t, err)

	generatedCasesArchive, err := txtar.ParseFile("testdata/template-cases-generated.txtar")
	require.NoError(t, err)

	goldens := testutil.NewTxTar(t, "testdata/template-cases.golden.txtar")

	for _, tc := range append(casesArchive.Files, generatedCasesArchive.Files...) {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			var values PartialValues
			require.NoError(t, yaml.Unmarshal(tc.Data, &values))

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
				Name:   "operator",
				Values: values,
				Set:    []string{},
			})
			require.NoError(t, err)
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
		func(p *corev1.PullPolicy, c fuzz.Continue) {
			policies := []corev1.PullPolicy{
				corev1.PullAlways,
				corev1.PullNever,
				corev1.PullIfNotPresent,
			}

			*p = policies[c.Intn(len(policies))]
		},
		func(r corev1.ResourceList, c fuzz.Continue) {
			r[corev1.ResourceCPU] = resource.MustParse(strconv.Itoa(c.Intn(1000)))
			r[corev1.ResourceMemory] = resource.MustParse(strconv.Itoa(c.Intn(1000)))
		},
		func(p *corev1.Probe, c fuzz.Continue) {
			p.InitialDelaySeconds = int32(c.Intn(1000))
			p.PeriodSeconds = int32(c.Intn(1000))
			p.TimeoutSeconds = int32(c.Intn(1000))
			p.SuccessThreshold = int32(c.Intn(1000))
			p.FailureThreshold = int32(c.Intn(1000))
			p.TerminationGracePeriodSeconds = ptr.To(int64(c.Intn(1000)))
		},
		func(p *corev1.PodFSGroupChangePolicy, c fuzz.Continue) {
			policies := []corev1.PodFSGroupChangePolicy{
				corev1.FSGroupChangeOnRootMismatch,
				corev1.FSGroupChangeAlways,
			}

			*p = policies[c.Intn(len(policies))]
		},
		func(s *intstr.IntOrString, c fuzz.Continue) {
			*s = intstr.FromInt32(c.Int31())
		},
		func(s *corev1.ResourceName, c fuzz.Continue) { asciiStrs((*string)(s), c) },
		func(_ *any, c fuzz.Continue) {},
		func(_ *[]corev1.ResourceClaim, c fuzz.Continue) {},
		func(_ *[]metav1.ManagedFieldsEntry, c fuzz.Continue) {},
	)

	schema, err := jsonschema.CompileString("", string(ValuesSchemaJSON))
	require.NoError(t, err)

	files := make([]txtar.File, 0, 100)
	for _, scope := range []OperatorScope{Namespace, Cluster} {
		nilChance := float64(0.8)
		for i := 0; i < 50; i++ {
			// Every 5 iterations, decrease nil chance to ensure that we're biased
			// towards exploring most cases.
			if i%5 == 0 && nilChance > .1 {
				nilChance -= .1
			}

			var values PartialValues
			fuzzer.NilChance(nilChance).Fuzz(&values)
			// Special case as fuzzer does not assign correctly scope
			values.Scope = &scope
			if scope == Cluster {
				values.Webhook = &PartialWebhook{Enabled: ptr.To(true)}
			} else {
				values.Webhook = &PartialWebhook{Enabled: ptr.To(false)}
			}
			makeSureTagIsNotEmptyString(values, fuzzer)

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

			index := i
			if scope == Cluster {
				index += 50
			}

			files = append(files, txtar.File{
				Name: fmt.Sprintf("case-%03d", index),
				Data: out,
			})
		}
	}

	archive := txtar.Format(&txtar.Archive{
		Comment: []byte(fmt.Sprintf(`Generated by %s`, t.Name())),
		Files:   files,
	})

	require.NoError(t, os.WriteFile("testdata/template-cases-generated.txtar", archive, 0o644))
}

func makeSureTagIsNotEmptyString(values PartialValues, fuzzer *fuzz.Fuzzer) {
	if values.Image != nil && values.Image.Tag != nil && len(*values.Image.Tag) == 0 {
		t := values.Image.Tag
		for len(*t) == 0 {
			fuzzer.Fuzz(t)
		}
	}
	if values.KubeRBACProxy != nil && values.KubeRBACProxy.Image != nil && values.KubeRBACProxy.Image.Tag != nil && len(*values.KubeRBACProxy.Image.Tag) == 0 {
		t := values.KubeRBACProxy.Image.Tag
		for len(*t) == 0 {
			fuzzer.Fuzz(t)
		}
	}
	if values.Configurator != nil && values.Configurator.Tag != nil && len(*values.Configurator.Tag) == 0 {
		t := values.Configurator.Tag
		for len(*t) == 0 {
			fuzzer.Fuzz(t)
		}
	}
}

// preTranspilerChartVersion is the latest release of the Operator helm chart prior to the introduction of
// ConfigMap go base implementation. It's used to verify that translated code is functionally equivalent.
const preTranspilerChartVersion = "0.4.28"

// TestChartDifferences can be removed if in the next operator chart version values definition changes or any resource.
// That test only validates clean transition to gotohelm definition of the operator helm chart.
func TestChartDifferences(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Downloading Operator helm chart release is required as client.Template
	// function does not pass HELM_CONFIG_HOME, that prevents from downloading specific
	// Operator helm chart version from public helm repository.
	require.NoError(t, client.DownloadFile(ctx,
		fmt.Sprintf("https://github.com/redpanda-data/helm-charts/releases/download/operator-%s/operator-%s.tgz", preTranspilerChartVersion, preTranspilerChartVersion),
		fmt.Sprintf("operator-%s.tgz", preTranspilerChartVersion)))

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		t.Run(v.Name(), func(t *testing.T) {
			t.Parallel()

			// First generate latest released Redpanda charts manifests. From ConfigMap bootstrap,
			// redpanda node configuration and RPK profile.
			manifests, err := client.Template(ctx,
				filepath.Join(client.GetConfigHome(), fmt.Sprintf("operator-%s.tgz", preTranspilerChartVersion)),
				helm.TemplateOptions{
					Name:       "operator",
					ValuesFile: "./ci/" + v.Name(),
					Set:        []string{},
				})
			require.NoError(t, err)

			oldOperator, err := convertToMap(manifests)
			require.NoError(t, err)

			// Now helm template will generate Redpanda configuration from local definition
			manifests, err = client.Template(ctx, ".", helm.TemplateOptions{
				Name:       "operator",
				ValuesFile: "./ci/" + v.Name(),
				Set:        []string{},
			})
			require.NoError(t, err)

			operator, err := convertToMap(manifests)
			require.NoError(t, err)

			for key, val := range oldOperator {
				require.Equal(t, val, operator[key])
				delete(oldOperator, key)
				delete(operator, key)
			}

			require.Len(t, oldOperator, 0)
			require.Len(t, operator, 0)
		})
	}
}

func convertToMap(manifests []byte) (map[string]string, error) {
	objs, err := kube.DecodeYAML(manifests, Scheme)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for _, obj := range objs {
		key := fmt.Sprintf("%s, %s", obj.GetObjectKind().GroupVersionKind().String(), obj.GetName())
		if _, exist := result[key]; exist {
			panic("duplicate key " + key)
		}

		labels := obj.GetLabels()
		delete(labels, "app.kubernetes.io/version")
		delete(labels, "helm.sh/chart")
		obj.SetLabels(labels)

		// Previous operator configuration was malformed as `{{.values.config}}` was dictionary
		// which should be translated by `toYaml` function
		if cfg, ok := obj.(*corev1.ConfigMap); ok && obj.GetName() == "operator-config" {
			cfg.Data = map[string]string{}
			obj = kube.Object(cfg)
		}

		// Due to operator helm chart bump the Deployment needs to remove few properites
		if dep, ok := obj.(*appsv1.Deployment); ok && obj.GetName() == "operator" {
			dep.Spec.Template.Spec.Containers[0].Image = "REDACTED_DUE_TO_CONTAINER_TAG_MISS_MATCH"
			dep.Spec.Template.Spec.Containers[0].Args[3] = "REDACTED_DUE_TO_CONTAINER_TAG_MISS_MATCH"
			obj = kube.Object(dep)
		}

		// In previous operator templates namespace was omitted in multiple places
		obj.SetNamespace("")

		b, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}

		result[key] = string(b)
	}

	return result, nil
}
