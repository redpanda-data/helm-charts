package redpanda_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	lock, err := helm.GetChartLock("Chart.lock")
	require.NoError(t, err)

	// Chart deps are kept within ./charts as a tgz archive. Helm dep build
	// will ensure that ./charts is in sync with Chart.lock.
	_, err = exec.CommandContext(ctx, "helm", "dep", "build").CombinedOutput()
	require.NoError(t, err, "failed to refresh helm dependencies")

	newLock, err := helm.GetChartLock("Chart.lock")
	require.NoError(t, err)

	// Comparison between Chart.lock before and after `helm dep build` execution is required to prevent
	// circular dependency between Chart.lock update, which is included in release [Auto commit], and
	// next unnecessary release pipeline execution due to change in `generated` field. Allowed change
	// in Chart.lock is only when dependencies are different. Dependencies might change only when
	// Redpanda chart dependencies got new release.
	if lock.Digest == newLock.Digest {
		err = helm.UpdateChartLock(lock, "Chart.lock")
		require.NoError(t, err)
	}

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		v := v
		t.Run(v.Name(), func(t *testing.T) {
			t.Parallel()

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
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

			testutil.AssertGolden(t, testutil.YAML, "./testdata/"+v.Name()+".golden", out)
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
