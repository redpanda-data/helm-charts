package redpanda

import (
	"os"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TieredStorageStatic(t *testing.T) Values {
	if os.Getenv("REDPANDA_LICENSE") == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	return Values{
		Enterprise: &Enterprise{
			License: os.Getenv("REDPANDA_LICENSE"),
		},
		Storage: &Storage{
			Tiered: &Tiered{
				Config: map[string]any{
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

func TieredStorageSecretRefs(t *testing.T, secret *corev1.Secret) Values {
	if os.Getenv("REDPANDA_LICENSE") == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	return Values{
		Enterprise: &Enterprise{
			License: os.Getenv("REDPANDA_LICENSE"),
		},
		Storage: &Storage{
			Tiered: &Tiered{
				CredentialsSecretRef: &TieredStorageCredentials{
					AccessKey: &SecretRef{Name: secret.Name, Key: "access"},
					SecretKey: &SecretRef{Name: secret.Name, Key: "secret"},
				},
				Config: map[string]any{
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

	env := helmtest.Setup(t).Namespaced(t)

	t.Run("tiered-storage-secrets", func(t *testing.T) {
		ctx := testutil.Context(t)

		credsSecret, err := kube.Create(ctx, env.Ctl(), TieredStorageSecret(env.Namespace()))
		require.NoError(t, err)

		rpRelease := env.Install(".", helm.InstallOptions{})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		config, err := rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, false, config["cloud_storage_enabled"])

		rpRelease = env.Upgrade(".", rpRelease, helm.UpgradeOptions{Values: TieredStorageStatic(t)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "static-access-key", config["cloud_storage_access_key"])

		rpRelease = env.Upgrade(".", rpRelease, helm.UpgradeOptions{Values: TieredStorageSecretRefs(t, credsSecret)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "from-secret-access-key", config["cloud_storage_access_key"])
	})
}
