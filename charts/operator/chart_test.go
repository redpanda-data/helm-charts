// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package operator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"

	"github.com/redpanda-data/redpanda-operator/pkg/helm"
	"github.com/redpanda-data/redpanda-operator/pkg/kube"
	"github.com/redpanda-data/redpanda-operator/pkg/testutil"
)

func TestMain(m *testing.M) {
	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	// This is performed in TestMain as there may be many tests that run the
	// redpanda helm chart.
	out, err := exec.Command("helm", "repo", "add", "prometheus", "https://prometheus-community.github.io/helm-charts").CombinedOutput()
	if err != nil {
		log.Fatalf("failed to run helm repo add: %s", out)
	}

	out, err = exec.Command("helm", "dep", "build", ".").CombinedOutput()
	if err != nil {
		log.Fatalf("failed to run helm dep build: %s", out)
	}

	os.Exit(m.Run())
}

func TestHelmKustomizeEquivalence(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	kustomization, err := os.ReadFile("testdata/kustomization.yaml")
	require.NoError(t, err)
	require.Containsf(t, string(kustomization), ChartMeta().AppVersion, "kustomization.yaml should reference the current appVersion: %s", chartMeta.AppVersion)

	values := PartialValues{FullnameOverride: ptr.To("redpanda"), RBAC: &PartialRBAC{CreateAdditionalControllerCRs: ptr.To(true)}}

	rendered, err := client.Template(ctx, ".", helm.TemplateOptions{
		Name:      "redpanda",
		Namespace: "",
		Values:    values,
	})
	require.NoError(t, err)

	fSys := filesys.MakeFsOnDisk()
	buffy := new(bytes.Buffer)
	cmd := build.NewCmdBuild(
		fSys, build.MakeHelp(konfig.ProgramName, "build"), buffy)
	require.NoError(t, cmd.RunE(cmd, []string{"testdata"}))

	helmObjs, err := kube.DecodeYAML(rendered, Scheme)
	require.NoError(t, err)

	require.NoError(t, apiextensionsv1.AddToScheme(Scheme))
	kustomizeObjs, err := kube.DecodeYAML(buffy.Bytes(), Scheme)
	require.NoError(t, err)

	helmClusterRoleRules, helmRoleRules := ExtractRules(helmObjs)
	kClusterRoleRules, kRoleRules := ExtractRules(kustomizeObjs)

	assert.JSONEq(t, jsonStr(helmRoleRules), jsonStr(kRoleRules), "difference in Roles\n--- Helm / Missing from Kustomize\n+++ Kustomize / Missing from Helm")
	assert.JSONEq(t, jsonStr(helmClusterRoleRules), jsonStr(kClusterRoleRules), "difference in ClusterRoles\n--- Helm / Missing from Kustomize\n+++ Kustomize / Missing from Helm")
}

func jsonStr(in any) string {
	out, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	return string(out)
}

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

func CalculateRoleRules(rules []rbacv1.PolicyRule) map[string]map[string]struct{} {
	flattened := map[string]map[string]struct{}{}
	for _, rule := range rules {
		for _, api := range rule.APIGroups {
			for _, res := range rule.Resources {
				key := fmt.Sprintf("%s#%s", api, res)

				if _, ok := flattened[key]; !ok {
					flattened[key] = map[string]struct{}{}
				}

				for _, verb := range rule.Verbs {
					flattened[key][verb] = struct{}{}
				}
			}
		}
	}
	return flattened
}

func ExtractRules(objs []kube.Object) (map[string]map[string]struct{}, map[string]map[string]struct{}) {
	var rules []rbacv1.PolicyRule
	var clusterRules []rbacv1.PolicyRule
	for _, o := range objs {
		switch obj := o.(type) {
		case *rbacv1.Role:
			rules = append(rules, obj.Rules...)
		case *rbacv1.ClusterRole:
			clusterRules = append(clusterRules, obj.Rules...)
		}
	}
	return CalculateRoleRules(clusterRules), CalculateRoleRules(rules)
}
