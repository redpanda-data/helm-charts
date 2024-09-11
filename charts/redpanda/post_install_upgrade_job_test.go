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
			Values{Storage: Storage{Tiered: Tiered{}}},
			[]corev1.EnvVar{},
		},
		{
			"only-literal-license",
			Values{
				Storage:    Storage{Tiered: Tiered{}},
				Enterprise: Enterprise{License: "fake.license"},
			},
			[]corev1.EnvVar{{Name: "REDPANDA_LICENSE", Value: "fake.license"}},
		},
		{
			"only-deprecated-literal-license",
			Values{
				Storage:    Storage{Tiered: Tiered{}},
				LicenseKey: "fake.license",
			},
			[]corev1.EnvVar{{Name: "REDPANDA_LICENSE", Value: "fake.license"}},
		},
		{
			name: "only-secret-ref-license",
			values: Values{
				Storage: Storage{Tiered: Tiered{}},
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
				Storage: Storage{Tiered: Tiered{}},
				LicenseSecretRef: &LicenseSecretRef{
					SecretName: "some-secret",
					SecretKey:  "some-key",
				},
			},
			expectedEnvVars: []corev1.EnvVar{
				{
					Name: "REDPANDA_LICENSE",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret"},
							Key:                  "some-key",
						},
					},
				},
			},
		},
		{
			name: "azure-literal-shared-key",
			values: Values{
				Storage: Storage{Tiered: Tiered{
					Config: TieredStorageConfig{
						"cloud_storage_enabled":               true,
						"cloud_storage_azure_shared_key":      "fake-shared-key",
						"cloud_storage_azure_container":       "fake-azure-container",
						"cloud_storage_azure_storage_account": "fake-storage-account",
					},
				}},
			},
			expectedEnvVars: []corev1.EnvVar{},
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
