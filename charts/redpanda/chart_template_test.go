// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package redpanda_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	jsoniter "github.com/json-iterator/go"
	"github.com/redpanda-data/redpanda/src/go/rpk/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	"github.com/redpanda-data/redpanda-operator/charts/redpanda"
	"github.com/redpanda-data/redpanda-operator/pkg/helm"
	"github.com/redpanda-data/redpanda-operator/pkg/kube"
	"github.com/redpanda-data/redpanda-operator/pkg/testutil"
	"github.com/redpanda-data/redpanda-operator/pkg/valuesutil"
)

func TestMain(m *testing.M) {
	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	// This is performed in TestMain as there may be many tests that run the
	// redpanda helm chart.
	out, err := exec.Command("helm", "repo", "add", "redpanda", "https://charts.redpanda.com").CombinedOutput()
	if err != nil {
		log.Fatalf("failed to run helm repo add: %s", out)
	}

	out, err = exec.Command("helm", "dep", "build", ".").CombinedOutput()
	if err != nil {
		log.Fatalf("failed to run helm dep build: %s", out)
	}

	os.Exit(m.Run())
}

// TestTemplateHelm310 is a smoke test that the redpanda helm chart (with
// default values) can successfully be executed against helm 3.10.3 which
// happens to ship with many installations of argocd and flux.
//
// Thus far issues have been with:
// - text/template's `eq` implementation in go 1.18.x (used to compile helm 3.10.3)
// - helm's `tpl` helper needing .Template.BasePath to be present
//
// Example Issues:
// - https://github.com/redpanda-data/helm-charts/issues/1531
// - https://github.com/redpanda-data/helm-charts/issues/1454
// - https://redpandadata.slack.com/archives/C01H6JRQX1S/p1728322201042639 (Kinda)
func TestTemplateHelm310(t *testing.T) {
	cmd := exec.Command("helm-3.10.3", "template", ".", "--generate-name")
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "helm-3.10.3 template failed:\n%s", out)
}

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	archive, err := txtar.ParseFile("testdata/template-cases.txtar")
	require.NoError(t, err)

	goldens := testutil.NewTxTar(t, "testdata/template-cases.golden.txtar")

	cases := archive.Files
	cases = append(cases, VersionGoldenTestsCases(t)...)
	cases = append(cases, CIGoldenTestCases(t)...)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			// To make it easy to add tests and assertions on various sets of
			// data, we add markers in YAML comments in the form of:
			// # ASSERT-<NAME> ["OPTIONAL", "PARAMS", "AS", "JSON"]
			// Assertions are run in the order they are specified. A good default
			// if you've got no other options is ASSERT-NO-ERROR.
			assertions := regexp.MustCompile(`(?m)^# (ASSERT-\S+) *(.+)?$`).FindAllSubmatch(tc.Data, -1)
			require.NotEmpty(t, assertions, "no ASSERT- markers found. All cases must have at least 1 marker.")

			var values map[string]any
			require.NoError(t, yaml.Unmarshal(tc.Data, &values), "input values are invalid YAML")

			out, renderErr := client.Template(ctx, ".", helm.TemplateOptions{
				Name:   "redpanda",
				Values: values,
				Set: []string{
					// Tests utilize some non-deterministic helpers (rng). We don't
					// really care about the stability of their output, so globally
					// disable them.
					"tests.enabled=false",
					// jwtSecret defaults to a random string. Can't have that
					// in snapshot testing so set it to a static value.
					"console.secret.login.jwtSecret=SECRETKEY",
					// the bootstrap user password has the same issues as jwtSecret
					"auth.sasl.bootstrapUser.password=changeme",
				},
			})

			objs, decodeErr := kube.DecodeYAML(out, redpanda.Scheme)

			for _, assertion := range assertions {
				name := string(assertion[1])

				var params []json.RawMessage
				if len(assertion[2]) > 0 {
					require.NoError(t, json.Unmarshal(assertion[2], &params))
				}

				// All assertions, with the exception of a few, imply
				// "ASSERT-NO-ERROR" as they need to operate on valid output.
				switch name {
				case `ASSERT-ERROR-CONTAINS`, `ASSERT-GOLDEN`:
				default:
					require.NoError(t, decodeErr)
					require.NoError(t, renderErr)
				}

				switch name {
				case `ASSERT-NO-ERROR`:
					AssertValidTypes(t, objs)

				case `ASSERT-ERROR-CONTAINS`:
					var errFragment string
					require.NoError(t, json.Unmarshal(params[0], &errFragment))
					require.ErrorContains(t, renderErr, errFragment)

				case `ASSERT-GOLDEN`:
					if renderErr == nil {
						goldens.AssertGolden(t, testutil.YAML, fmt.Sprintf("testdata/%s.yaml.golden", t.Name()), out)
					} else {
						// Trailing new lines are added by the txtar format if
						// they're not already present. Add one here otherwise
						// we'll see failures.
						goldens.AssertGolden(t, testutil.Text, fmt.Sprintf("testdata/%s.yaml.golden", t.Name()), []byte(renderErr.Error()+"\n"))
					}

				case `ASSERT-TRUST-STORES`:
					AssertTrustStores(t, out, params)

				case `ASSERT-NO-CERTIFICATES`:
					AssertNoCertficates(t, objs, out)

				case `ASSERT-FIELD-EQUALS`:
					AssertFieldEquals(t, params, objs)

				case `ASSERT-FIELD-CONTAINS`:
					AssertFieldContains(t, params, objs)

				case `ASSERT-FIELD-NOT-CONTAINS`:
					AssertFieldNotContains(t, params, objs)

				case `ASSERT-VALID-RPK-CONFIGURATION`:
					AssertValidRPKConfiguration(t, objs)

				case `ASSERT-STATEFULSET-VOLUME-MOUNTS-VERIFICATION`:
					AssertStatefulSetVolumeMountsVerification(t, objs)

				case `ASSERT-STATEFULSET-ALL-VOLUMES-ARE-USED`:
					AssertStatefulsetAllVolumesAreUsed(t, objs)

				case `ASSERT-SUPER-USERS-ARE-VALID`:
					AssertSuperUsersAreValid(t, objs)

				default:
					t.Fatalf("unknown assertion marker: %q\nFull Line: %s", name, assertion[0])
				}
			}
		})
	}
}

func AssertSuperUsersAreValid(t *testing.T, objs []kube.Object) {
	for _, obj := range objs {
		secret, ok := obj.(*corev1.Secret)
		if !ok || !strings.Contains(secret.Name, "users") {
			continue
		}
		users, ok := secret.StringData["users.txt"]
		require.True(t, ok)
		usersArray := strings.Split(string(users), "\n")
		for _, user := range usersArray {
			line := strings.Split(user, ":")
			// Username
			require.True(t, len(line[0]) > 0)
			// Password
			require.True(t, len(line[1]) > 0)

			// Sometimes mechanism can be missing
			if len(line) > 2 {
				// Machanism
				switch line[2] {
				case "SCRAM-SHA-256":
				case "SCRAM-SHA-512":
				default:
					t.Fatalf("mechanism is not recognized: %s", line[2])
				}
			}
		}
	}
}

func AssertStatefulSetVolumeMountsVerification(t *testing.T, objs []kube.Object) {
	for _, obj := range objs {
		sts, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			continue
		}

		volumes := map[string]struct{}{}
		for _, v := range sts.Spec.Template.Spec.Containers {
			for _, m := range v.VolumeMounts {
				volumes[m.Name] = struct{}{}
			}
		}

		for _, v := range sts.Spec.Template.Spec.InitContainers {
			for _, m := range v.VolumeMounts {
				volumes[m.Name] = struct{}{}
			}
		}

		for _, v := range sts.Spec.Template.Spec.Volumes {
			delete(volumes, v.Name)
		}
		require.Len(t, volumes, 0)
	}
}

func AssertStatefulsetAllVolumesAreUsed(t *testing.T, objs []kube.Object) {
	for _, obj := range objs {
		sts, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			continue
		}

		volumes := map[string]struct{}{}
		for _, v := range sts.Spec.Template.Spec.Containers {
			for _, m := range v.VolumeMounts {
				volumes[m.Name] = struct{}{}
			}
		}

		for _, v := range sts.Spec.Template.Spec.InitContainers {
			for _, m := range v.VolumeMounts {
				volumes[m.Name] = struct{}{}
			}
		}

		for _, v := range sts.Spec.Template.Spec.Volumes {
			if _, ok := volumes[v.Name]; !ok {
				t.Fatalf("missing volume %s", v.Name)
			}
		}
	}
}

func CIGoldenTestCases(t *testing.T) []txtar.File {
	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	cases := make([]txtar.File, len(values))
	for i, f := range values {
		data, err := os.ReadFile("./ci/" + f.Name())
		require.NoError(t, err)

		cases[i] = txtar.File{
			Name: f.Name(),
			Data: append([]byte("# ASSERT-NO-ERROR\n# ASSERT-GOLDEN\n# ASSERT-STATEFULSET-ALL-VOLUMES-ARE-USED\n"), data...),
		}
	}
	return cases
}

func VersionGoldenTestsCases(t *testing.T) []txtar.File {
	// A collection of versions that should trigger all the gates guarded by
	// "redpanda-atleast-*" helpers.
	versions := []struct {
		Image  redpanda.PartialImage
		ErrMsg *string
	}{
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.1.0"))},
			ErrMsg: ptr.To("no longer supported"),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.2.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.3.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.3.14"))},
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.4.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.1"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.2"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.3"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.2.1"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.3.0"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
		},
		{
			Image: redpanda.PartialImage{Repository: ptr.To("somecustomrepo"), Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
		},
		{
			Image: redpanda.PartialImage{Repository: ptr.To("somecustomrepo"), Tag: ptr.To(redpanda.ImageTag("v23.2.8"))},
		},
	}

	// A collection of features that are protected by the various above version
	// gates.
	permutations := []redpanda.PartialValues{
		{
			Config: &redpanda.PartialConfig{
				Tunable: redpanda.PartialTunableConfig{
					"log_segment_size_min":  100,
					"log_segment_size_max":  99999,
					"kafka_batch_max_bytes": 7777,
				},
			},
		},
		{
			Enterprise: &redpanda.PartialEnterprise{License: ptr.To("ATOTALLYVALIDLICENSE")},
		},
		{
			RackAwareness: &redpanda.PartialRackAwareness{
				Enabled:        ptr.To(true),
				NodeAnnotation: ptr.To("topology-label"),
			},
		},
	}

	var cases []txtar.File
	for _, version := range versions {
		version := version
		for i, perm := range permutations {
			values, err := valuesutil.UnmarshalInto[redpanda.PartialValues](perm)
			require.NoError(t, err)

			values.Image = &version.Image

			name := fmt.Sprintf("%s-%s-%d", ptr.Deref(version.Image.Repository, "default"), *version.Image.Tag, i)

			header := []byte("# ASSERT-NO-ERROR\n# ASSERT-GOLDEN\n# ASSERT-STATEFULSET-ALL-VOLUMES-ARE-USED\n")
			if version.ErrMsg != nil {
				header = []byte(fmt.Sprintf("# ASSERT-ERROR-CONTAINS [%q]\n# ASSERT-GOLDEN\n", *version.ErrMsg))
			}

			data, err := yaml.Marshal(values)
			require.NoError(t, err)

			cases = append(cases, txtar.File{
				Name: name,
				Data: append(header, data...),
			})
		}
	}
	return cases
}

func AssertTrustStores(t *testing.T, manifests []byte, params []json.RawMessage) {
	var listener string
	var expected map[string]string

	require.NoError(t, json.Unmarshal(params[0], &listener))
	require.NoError(t, json.Unmarshal(params[1], &expected))

	cm, _, err := getConfigMaps(manifests)
	require.NoError(t, err)

	redpandaYAML, err := yaml.YAMLToJSON([]byte(cm.Data["redpanda.yaml"]))
	require.NoError(t, err)

	tlsConfigs := map[string]jsoniter.Any{
		"kafka":           jsoniter.Get(redpandaYAML, "redpanda", "kafka_api_tls"),
		"admin":           jsoniter.Get(redpandaYAML, "redpanda", "admin_api_tls"),
		"http":            jsoniter.Get(redpandaYAML, "pandaproxy", "pandaproxy_api_tls"),
		"schema_registry": jsoniter.Get(redpandaYAML, "schema_registry", "schema_registry_api_tls"),
	}

	actual := map[string]map[string]string{}
	for name, cfg := range tlsConfigs {
		m := map[string]string{}
		for i := 0; i < cfg.Size(); i++ {
			name := cfg.Get(i, "name").ToString()
			truststore := cfg.Get(i, "truststore_file").ToString()
			m[name] = truststore
		}
		actual[name] = m
	}

	assert.Equal(t, expected, actual[listener])
}

func AssertNoCertficates(t *testing.T, objs []kube.Object, manifests []byte) {
	for _, obj := range objs {
		_, ok := obj.(*certmanagerv1.Certificate)
		// The -root-certificate is always created right now, ignore that
		// one.
		if ok && strings.HasSuffix(obj.GetName(), "-root-certificate") {
			continue
		}
		require.Falsef(t, ok, "Found unexpected Certificate %q", obj.GetName())
	}

	require.NotContains(t, manifests, []byte(certmanagerv1.CertificateKind))
}

func AssertValidRPKConfiguration(t *testing.T, objs []kube.Object) {
	for _, obj := range objs {
		cm, ok := obj.(*corev1.ConfigMap)
		if !(ok && obj.GetName() == "redpanda") {
			continue
		}
		rpCfg, exist := cm.Data["redpanda.yaml"]
		require.True(t, exist, "redpanda.yaml not found")

		var cfg config.RedpandaYaml
		require.NoError(t, yaml.Unmarshal([]byte(rpCfg), &cfg))
	}
}

func AssertValidTypes(t *testing.T, objs []kube.Object) {
	allowableTypes := map[reflect.Type]struct{}{}
	for _, t := range redpanda.Types() {
		allowableTypes[reflect.TypeOf(t)] = struct{}{}
	}

	for _, obj := range objs {
		annos := obj.GetAnnotations()
		if annos == nil {
			annos = map[string]string{}
		}

		// Skip over tests.
		if annos["helm.sh/hook"] == "test" {
			continue
		}

		_, ok := allowableTypes[reflect.TypeOf(obj)]
		require.True(t, ok, "%T is not an allowable type. Did you forget to update `redpanda.Types`?", obj)

		gvk, err := apiutil.GVKForObject(obj, redpanda.Scheme)
		require.NoError(t, err)

		require.Equal(t, gvk, obj.GetObjectKind().GroupVersionKind())
	}
}

func AssertFieldEquals(t *testing.T, params []json.RawMessage, objs []kube.Object) {
	var gvk string
	var key string
	var fieldPath string
	fieldValue := params[3] // No need to unmarshal this one.

	require.NoError(t, json.Unmarshal(params[0], &gvk))
	require.NoError(t, json.Unmarshal(params[1], &key))
	require.NoError(t, json.Unmarshal(params[2], &fieldPath))

	execJSONPath(t, objs, gvk, key, fieldPath, func(result any) {
		actual, err := json.Marshal(result)
		require.NoError(t, err)

		require.JSONEq(t, string(fieldValue), string(actual), "%q", fieldPath)
	})
}

func AssertFieldContains(t *testing.T, params []json.RawMessage, objs []kube.Object) {
	var gvk string
	var key string
	var fieldPath string
	var contained any

	require.NoError(t, json.Unmarshal(params[0], &gvk))
	require.NoError(t, json.Unmarshal(params[1], &key))
	require.NoError(t, json.Unmarshal(params[2], &fieldPath))
	require.NoError(t, json.Unmarshal(params[3], &contained))

	execJSONPath(t, objs, gvk, key, fieldPath, func(result any) {
		assert.Contains(t, result, contained)
	})
}

func AssertFieldNotContains(t *testing.T, params []json.RawMessage, objs []kube.Object) {
	var gvk string
	var key string
	var fieldPath string
	var contained any

	require.NoError(t, json.Unmarshal(params[0], &gvk))
	require.NoError(t, json.Unmarshal(params[1], &key))
	require.NoError(t, json.Unmarshal(params[2], &fieldPath))
	require.NoError(t, json.Unmarshal(params[3], &contained))

	execJSONPath(t, objs, gvk, key, fieldPath, func(result any) {
		assert.NotContains(t, result, contained)
	})
}

func execJSONPath(t *testing.T, objs []kube.Object, gvk, key, jsonPath string, fn func(any)) {
	for _, obj := range objs {
		kind := obj.GetObjectKind().GroupVersionKind().Kind
		groupVersion := obj.GetObjectKind().GroupVersionKind().GroupVersion().String()

		if groupVersion+"/"+kind != gvk {
			continue
		}

		if obj.GetNamespace()+"/"+obj.GetName() != key {
			continue
		}

		// See https://kubernetes.io/docs/reference/kubectl/jsonpath/
		path := jsonpath.New("").AllowMissingKeys(true)
		require.NoError(t, path.Parse(jsonPath))

		results, err := path.FindResults(obj)
		require.NoError(t, err)

		for _, result := range results {
			fn(result[0].Interface())
		}

		return
	}

	t.Fatalf("object %q of kind %q not found", gvk, key)
}
