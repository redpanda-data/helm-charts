package redpanda_test

import (
	"maps"
	"os"
	"path/filepath"
	"regexp"
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
	"sigs.k8s.io/yaml"
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

// preTranspilerChartVersion is the latest release of the Redpanda helm chart prior to the introduction of
// ConfigMap go base implementation. It's used to verify that translated code is functionally equivalent.
const preTranspilerChartVersion = "redpanda-5.8.8.tgz"

func TestConfigMap(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Downloading Redpanda helm chart release is required as client.Template
	// function does not pass HELM_CONFIG_HOME, that prevents from downloading specific
	// Redpanda helm chart version from public helm repository.
	require.NoError(t, client.DownloadFile(ctx, "https://github.com/redpanda-data/helm-charts/releases/download/redpanda-5.8.8/redpanda-5.8.8.tgz", preTranspilerChartVersion))

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		t.Run(v.Name(), func(t *testing.T) {
			t.Parallel()

			// First generate latest released Redpanda charts manifests. From ConfigMap bootstrap,
			// redpanda node configuration and RPK profile.
			manifests, err := client.Template(ctx, filepath.Join(client.GetConfigHome(), preTranspilerChartVersion), helm.TemplateOptions{
				Name:       "redpanda",
				ValuesFile: "./ci/" + v.Name(),
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

			oldConf, err := extractRedpandaConfigsFromConfigMap(manifests)
			require.NoError(t, err)

			// Now helm template will generate Redpanda configuration from local definition
			manifests, err = client.Template(ctx, ".", helm.TemplateOptions{
				Name:       "redpanda",
				ValuesFile: "./ci/" + v.Name(),
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

			newConf, err := extractRedpandaConfigsFromConfigMap(manifests)
			require.NoError(t, err)

			// Overprovisioned till Redpanda chart version 5.8.8 should be
			// set to `true` when Statefulset CPU request value is bellow 1000 mili cores.
			// Function `redpanda-smp`, that should overwrite `overprovisioned` flag,
			// was not called before setting `overprovisioned` flag.
			// redpanda-smp template - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_helpers.tpl#L187
			// redpanda-smp template invocation - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_configmap.tpl#L610
			// overprovisioned flag - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_configmap.tpl#L607
			rex := regexp.MustCompile("\"overprovisioned\":(true|false)")
			newConf.redpanda = rex.ReplaceAllString(newConf.redpanda, "\"overprovisioned\":false")

			require.JSONEq(t, oldConf.redpanda, newConf.redpanda)
			require.JSONEq(t, oldConf.bootstrap, newConf.bootstrap)
			require.JSONEq(t, oldConf.rpkProfile, newConf.rpkProfile)
		})
	}
}

type configmapRepresentation struct {
	redpanda   string
	bootstrap  string
	rpkProfile string
}

// extractRedpandaConfigsFromConfigMap is parsing all manifests (resources)
// created by helm template execution. Redpanda helm chart creates 3 distinct
// files in ConfigMap: redpanda.yaml (node, tunable and cluster configuration),
// bootstrap.yaml (only cluster configuration) and profile (external connectivity rpk profile).
func extractRedpandaConfigsFromConfigMap(manifests []byte) (*configmapRepresentation, error) {
	var result configmapRepresentation
	objs, err := kube.DecodeYAML(manifests, redpanda.Scheme)
	if err != nil {
		return nil, err
	}

	for _, obj := range objs {
		switch obj := obj.(type) {
		case *corev1.ConfigMap:
			switch obj.Name {
			case "redpanda":
				r := obj.Data["redpanda.yaml"]
				jsonR, err := yaml.YAMLToJSON([]byte(r))
				if err != nil {
					return nil, err
				}
				result.redpanda = string(jsonR)

				b := obj.Data["bootstrap.yaml"]
				jsonB, err := yaml.YAMLToJSON([]byte(b))
				if err != nil {
					return nil, err
				}
				result.bootstrap = string(jsonB)
			case "redpanda-rpk":
				p := obj.Data["profile"]
				jsonP, err := yaml.YAMLToJSON([]byte(p))
				if err != nil {
					return nil, err
				}
				result.rpkProfile = string(jsonP)
			}
		}
	}

	return &result, nil
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
