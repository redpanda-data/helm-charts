package redpanda

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestPostInstallUpgradeEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name            string
		values          Values
		expectedEnvVars []corev1.EnvVar
	}{
		{
			"empty-result",
			Values{Storage: Storage{Tiered: &Tiered{}}},
			[]corev1.EnvVar{},
		},
		{
			"only-literal-license",
			Values{
				Storage:    Storage{Tiered: &Tiered{}},
				Enterprise: Enterprise{License: "fake.license"},
			},
			[]corev1.EnvVar{{Name: "REDPANDA_LICENSE", Value: "fake.license"}},
		},
		{
			"only-deprecated-literal-license",
			Values{
				Storage:    Storage{Tiered: &Tiered{}},
				LicenseKey: "fake.license",
			},
			[]corev1.EnvVar{{Name: "REDPANDA_LICENSE", Value: "fake.license"}},
		},
		{
			name: "only-secret-ref-license",
			values: Values{
				Storage: Storage{Tiered: &Tiered{}},
				Enterprise: Enterprise{LicenseSecretRef: &struct {
					Key  string `json:"key"`
					Name string `json:"name"`
				}{
					Key:  "some-key",
					Name: "some-secret",
				}},
			},
			expectedEnvVars: []corev1.EnvVar{{Name: "REDPANDA_LICENSE", ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret"},
					Key:                  "some-key",
				},
			}}},
		},
		{
			name: "only-deprecated-secret-ref-license",
			values: Values{
				Storage: Storage{Tiered: &Tiered{}},
				LicenseSecretRef: &LicenseSecretRef{
					SecretName: "some-secret",
					SecretKey:  "some-key",
				},
			},
			expectedEnvVars: []corev1.EnvVar{{Name: "REDPANDA_LICENSE", ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret"},
					Key:                  "some-key",
				},
			}}},
		},
		{
			name: "azure-literal-shared-key",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":               true,
						"cloud_storage_azure_shared_key":      "fake-shared-key",
						"cloud_storage_azure_container":       "fake-azure-container",
						"cloud_storage_azure_storage_account": "fake-storage-account",
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_CONTAINER", Value: "fake-azure-container"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_STORAGE_ACCOUNT", Value: "fake-storage-account"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY", Value: "fake-shared-key"},
			},
		},
		{
			name: "azure-shared-key-via-credential-secret-reference",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":               true,
						"cloud_storage_azure_container":       "fake-azure-container",
						"cloud_storage_azure_storage_account": "fake-storage-account",
					},
					CredentialsSecretRef: TieredStorageCredentials{
						AccessKey: &SecretRef{},
						SecretKey: &SecretRef{
							Key:  "some-key",
							Name: "some-secret",
						},
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_CONTAINER", Value: "fake-azure-container"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_STORAGE_ACCOUNT", Value: "fake-storage-account"},
				{Name: "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret"},
						Key:                  "some-key",
					},
				}},
			},
		},
		{
			name: "azure-with-nil-values-configuration",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":               true,
						"cloud_storage_azure_shared_key":      "fake-shared-key",
						"cloud_storage_azure_container":       nil,
						"cloud_storage_azure_storage_account": nil,
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
			},
		},
		{
			name: "literal-cloud-storage-secret-key",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":    true,
						"cloud_storage_secret_key": "fake-secret-key",
						"cloud_storage_access_key": "fake-access-key",
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_ACCESS_KEY", Value: "fake-access-key"},
				{Name: "RPK_CLOUD_STORAGE_SECRET_KEY", Value: "fake-secret-key"},
			},
		},
		{
			name: "cloud-storage-storage-secret-key-via-credential-secret-reference",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled": true,
					},
					CredentialsSecretRef: TieredStorageCredentials{
						AccessKey: &SecretRef{
							Key:  "some-key-1",
							Name: "some-secret-1",
						},
						SecretKey: &SecretRef{
							Key:  "some-key-2",
							Name: "some-secret-2",
						},
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_ACCESS_KEY", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret-1"},
						Key:                  "some-key-1",
					},
				}},
				{Name: "RPK_CLOUD_STORAGE_SECRET_KEY", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret-2"},
						Key:                  "some-key-2",
					},
				}},
			},
		},
		{
			name: "multiple-types-in-tiered-storage-configuration",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":                         true,
						"cloud_storage_secret_key":                      "fake-secret-key",
						"cloud_storage_segment_max_upload_interval_sec": 1,
						"cloud_storage_cache_size":                      "20Gi",
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_SECRET_KEY", Value: "fake-secret-key"},
				{Name: "RPK_CLOUD_STORAGE_SEGMENT_MAX_UPLOAD_INTERVAL_SEC", Value: "1"},
				{Name: "RPK_CLOUD_STORAGE_CACHE_SIZE", Value: "21474836480"},
			},
		},
		{
			name: "multiple-types-in-deprecated-tiered-storage-configuration",
			values: Values{
				Storage: Storage{Tiered: &Tiered{}, TieredConfig: TieredStorageConfig{
					"cloud_storage_enabled":                         true,
					"cloud_storage_secret_key":                      "fake-secret-key",
					"cloud_storage_segment_max_upload_interval_sec": 1,
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
				{Name: "RPK_CLOUD_STORAGE_SECRET_KEY", Value: "fake-secret-key"},
				{Name: "RPK_CLOUD_STORAGE_SEGMENT_MAX_UPLOAD_INTERVAL_SEC", Value: "1"},
			},
		},
		{
			name: "nil-tiered-storage-config-value",
			values: Values{
				Storage: Storage{Tiered: &Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled": true,
						"invalid-configuration": nil,
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{
				{Name: "RPK_CLOUD_STORAGE_ENABLED", Value: "true"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.values)
			require.NoError(t, err)
			dot := helmette.Dot{}
			err = json.Unmarshal(b, &dot.Values)
			require.NoError(t, err)

			envVars := PostInstallUpgradeEnvironmentVariables(&dot)

			slices.SortFunc(envVars, compareEnvVars)
			slices.SortFunc(tc.expectedEnvVars, compareEnvVars)
			require.Equal(t, tc.expectedEnvVars, envVars)
		})
	}
}

func compareEnvVars(a, b corev1.EnvVar) int {
	if a.Name < b.Name {
		return -1
	} else {
		return 1
	}
}
