package redpanda_test

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TieredStorageStatic(t *testing.T) redpanda.PartialValues {
	license := os.Getenv("REDPANDA_LICENSE")
	if license == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	return redpanda.PartialValues{
		Config: &redpanda.PartialConfig{
			Node: &redpanda.PartialNodeConfig{
				"developer_mode": true,
			},
		},
		Enterprise: &redpanda.PartialEnterprise{
			License: &license,
		},
		Storage: &redpanda.PartialStorage{
			Tiered: &redpanda.PartialTiered{
				Config: &redpanda.PartialTieredStorageConfig{
					"cloud_storage_enabled":    true,
					"cloud_storage_region":     "static-region",
					"cloud_storage_bucket":     "static-bucket",
					"cloud_storage_access_key": "static-access-key",
					"cloud_storage_secret_key": "static-secret-key",
				},
			},
		},
	}
}

func TieredStorageSecret(namespace string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "tiered-storage-",
			Namespace:    namespace,
		},
		Data: map[string][]byte{
			"access": []byte("from-secret-access-key"),
			"secret": []byte("from-secret-secret-key"),
		},
	}
}

func TieredStorageSecretRefs(t *testing.T, secret *corev1.Secret) redpanda.PartialValues {
	license := os.Getenv("REDPANDA_LICENSE")
	if license == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	access := "access"
	secretKey := "secret"
	return redpanda.PartialValues{
		Config: &redpanda.PartialConfig{
			Node: &redpanda.PartialNodeConfig{
				"developer_mode": true,
			},
		},
		Enterprise: &redpanda.PartialEnterprise{
			License: &license,
		},
		Storage: &redpanda.PartialStorage{
			Tiered: &redpanda.PartialTiered{
				CredentialsSecretRef: &redpanda.PartialTieredStorageCredentials{
					AccessKey: &redpanda.PartialSecretRef{Name: &secret.Name, Key: &access},
					SecretKey: &redpanda.PartialSecretRef{Name: &secret.Name, Key: &secretKey},
				},
				Config: &redpanda.PartialTieredStorageConfig{
					"cloud_storage_enabled": true,
					"cloud_storage_region":  "a-region",
					"cloud_storage_bucket":  "a-bucket",
				},
			},
		},
	}
}

type TemplateTestCase struct {
	Name         string
	Values       any
	ValuesFile   string
	AssertGolden func(*testing.T, []byte)
}

func CITestCases(t *testing.T) []TemplateTestCase {
	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	cases := make([]TemplateTestCase, len(values))
	for i, f := range values {
		name := f.Name()
		cases[i] = TemplateTestCase{
			Name:       name,
			ValuesFile: "./ci/" + name,
			AssertGolden: func(t *testing.T, b []byte) {
				testutil.AssertGolden(t, testutil.YAML, path.Join("testdata", "ci", name+".golden"), b)
			},
		}
	}
	return cases
}

func VersionTestsCases(t *testing.T) []TemplateTestCase {
	// A collection of versions that should trigger all the gates guarded by
	// "redpanda-atleast-*" helpers.
	versions := []redpanda.PartialImage{
		{Tag: ptr.To(redpanda.ImageTag("v22.2.0"))},
		{Tag: ptr.To(redpanda.ImageTag("v22.3.0"))},
		{Tag: ptr.To(redpanda.ImageTag("v22.3.14"))},
		{Tag: ptr.To(redpanda.ImageTag("v22.4.0"))},
		{Tag: ptr.To(redpanda.ImageTag("v23.1.1"))},
		{Tag: ptr.To(redpanda.ImageTag("v23.1.2"))},
		{Tag: ptr.To(redpanda.ImageTag("v23.1.3"))},
		{Tag: ptr.To(redpanda.ImageTag("v23.2.1"))},
		{Tag: ptr.To(redpanda.ImageTag("v23.3.0"))},
		{Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
		{Repository: ptr.To("somecustomrepo"), Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
	}

	// A collection of features that are protected by the various above version
	// gates.
	permutations := []redpanda.PartialValues{
		{
			Config: &redpanda.PartialConfig{
				Tunable: &redpanda.PartialTunableConfig{
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

	var cases []TemplateTestCase
	for _, version := range versions {
		for i, perm := range permutations {
			values, err := valuesutil.UnmarshalInto[redpanda.PartialValues](perm)
			require.NoError(t, err)

			values.Image = &version

			name := fmt.Sprintf("%s-%s-%d", ptr.Deref(version.Repository, "default"), *version.Tag, i)

			cases = append(cases, TemplateTestCase{
				Name:   name,
				Values: values,
				AssertGolden: func(t *testing.T, b []byte) {
					testutil.AssertGolden(t, testutil.YAML, path.Join("testdata", "versions", name+".yaml.golden"), b)
				},
			})
		}
	}
	return cases
}

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	require.NoError(t, client.RepoAdd(ctx, "redpanda", "https://charts.redpanda.com"))
	require.NoError(t, client.DependencyBuild(ctx, "."), "failed to refresh helm dependencies")

	cases := CITestCases(t)
	cases = append(cases, VersionTestsCases(t)...)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
				Name:       "redpanda",
				Values:     tc.Values,
				ValuesFile: tc.ValuesFile,
				Set: []string{
					// Tests utilize some non-deterministic helpers (rng). We don't
					// really care about the stability of their output, so globally
					// disable them.
					"tests.enabled=false",
					// jwtSecret defaults to a random string. Can't have that
					// in snapshot testing so set it to a static value.
					"console.secret.login.jwtSecret=SECRETKEY",
				},
			})
			require.NoError(t, err)

			tc.AssertGolden(t, out)

			// kube-lint template file
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			inputYaml := bytes.NewBuffer(out)

			cmd := exec.CommandContext(ctx, "kube-linter", "lint", "-", "--format", "json")
			cmd.Stdin = inputYaml
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			errKubeLinter := cmd.Run()
			if errKubeLinter != nil && len(stderr.String()) > 0 {
				t.Logf("kube-linter error(s) found for %q: \n%s\nstderr:\n%s", tc.Name, stdout.String(), stderr.String())
			} else if errKubeLinter != nil {
				t.Logf("kube-linter error(s) found for %q: \n%s", tc.Name, errKubeLinter)
			}
			// TODO: remove comment below and the logging above once we agree to linter
			// require.NoError(t, errKubeLinter)
		})
	}
}

func TestChart(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skipping log running test...")
	}

	redpandaChart := "."

	env := helmtest.Setup(t).Namespaced(t)

	t.Run("tiered-storage-secrets", func(t *testing.T) {
		ctx := testutil.Context(t)

		credsSecret, err := kube.Create(ctx, env.Ctl(), TieredStorageSecret(env.Namespace()))
		require.NoError(t, err)

		rpRelease := env.Install(redpandaChart, helm.InstallOptions{
			Values: redpanda.PartialValues{
				Config: &redpanda.PartialConfig{
					Node: &redpanda.PartialNodeConfig{
						"developer_mode": true,
					},
				},
			},
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		config, err := rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, false, config["cloud_storage_enabled"])

		rpRelease = env.Upgrade(redpandaChart, rpRelease, helm.UpgradeOptions{Values: TieredStorageStatic(t)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "static-access-key", config["cloud_storage_access_key"])

		rpRelease = env.Upgrade(redpandaChart, rpRelease, helm.UpgradeOptions{Values: TieredStorageSecretRefs(t, credsSecret)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "from-secret-access-key", config["cloud_storage_access_key"])
	})
}

func TestLabels(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	for _, labels := range []map[string]string{
		{"foo": "bar"},
		{"baz": "1", "quux": "2"},
		// TODO: Add a test for asserting the behavior of adding a commonLabel
		// overriding a builtin value (app.kubernetes.io/name) once the
		// expected behavior is decided.
	} {
		values := &redpanda.PartialValues{
			CommonLabels: labels,
		}

		helmValues, err := valuesutil.UnmarshalInto[helmette.Values](values)
		require.NoError(t, err)

		dot := &helmette.Dot{
			Values: helmValues,
			Chart:  redpanda.ChartMeta(),
			Release: helmette.Release{
				Name:      "redpanda",
				Namespace: "redpanda",
				Service:   "Helm",
			},
		}

		manifests, err := client.Template(ctx, ".", helm.TemplateOptions{
			Name:      dot.Release.Name,
			Namespace: dot.Release.Namespace,
			// This guarantee does not currently extend to console.
			Set: []string{"console.enabled=false"},
			// Nor does it extend to tests.
			SkipTests: true,
			Values:    values,
		})
		require.NoError(t, err)

		objs, err := kube.DecodeYAML(manifests, redpanda.Scheme)
		require.NoError(t, err)

		expectedLabels := redpanda.FullLabels(dot)
		require.Subset(t, expectedLabels, values.CommonLabels, "FullLabels does not contain CommonLabels")

		for _, obj := range objs {
			// Assert that CommonLabels is included on all top level objects.
			require.Subset(t, obj.GetLabels(), expectedLabels, "%T %q", obj, obj.GetName())

			// For other objects (replication controllers) we want to assert
			// that common labels are also included on whatever object (Pod)
			// they generate/contain a template of.
			switch obj := obj.(type) {
			case *appsv1.StatefulSet:
				expectedLabels := maps.Clone(expectedLabels)
				expectedLabels["app.kubernetes.io/component"] += "-statefulset"
				require.Subset(t, obj.Spec.Template.GetLabels(), expectedLabels, "%T/%s's %T", obj, obj.Name, obj.Spec.Template)
			}
		}
	}
}
