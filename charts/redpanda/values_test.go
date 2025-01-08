// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package redpanda

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestListeners_TrustStoreVolumes(t *testing.T) {
	// Closures for more terse definitions.
	cmKeyRef := func(name, key string) *corev1.ConfigMapKeySelector {
		return &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
			Key: key,
		}
	}

	sKeyRef := func(name, key string) *corev1.SecretKeySelector {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
			Key: key,
		}
	}

	// Common TLS used by all cases.
	tls := TLS{
		Enabled: true,
		Certs: TLSCertMap{
			"disabled": TLSCert{Enabled: ptr.To(false)},
			"enabled":  TLSCert{Enabled: ptr.To(true)},
		},
	}

	cases := []struct {
		Name      string
		Listeners Listeners
		Out       *corev1.Volume
	}{
		{Name: "zeros"},
		{
			Name: "all unique secrets",
			Listeners: Listeners{
				Admin: AdminListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
					},
					External: ExternalListeners[AdminExternal]{
						"admin-1": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-2", "KEY-2")},
							},
						},
					},
				},
				Kafka: KafkaListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-3", "KEY-3")},
					},
					External: ExternalListeners[KafkaExternal]{
						"kafka-1": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-4", "KEY-4")},
							},
						},
					},
				},
				HTTP: HTTPListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-5", "KEY-5")},
					},
					External: ExternalListeners[HTTPExternal]{
						"http-1": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-6", "KEY-6")},
							},
						},
					},
				},
			},
			Out: &corev1.Volume{
				Name: "truststores",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "secrets/SECRET-1-KEY-1"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-2"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-2", Path: "secrets/SECRET-2-KEY-2"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-3"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-3", Path: "secrets/SECRET-3-KEY-3"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-4"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-4", Path: "secrets/SECRET-4-KEY-4"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-5"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-5", Path: "secrets/SECRET-5-KEY-5"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-6"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-6", Path: "secrets/SECRET-6-KEY-6"},
								},
							}},
						},
					},
				},
			},
		},
		{
			Name: "all unique configmaps",
			Listeners: Listeners{
				Admin: AdminListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
					},
					External: ExternalListeners[AdminExternal]{
						"admin-1": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-2", "KEY-2")},
							},
						},
					},
				},
				Kafka: KafkaListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-3", "KEY-3")},
					},
					External: ExternalListeners[KafkaExternal]{
						"kafka-1": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-4", "KEY-4")},
							},
						},
					},
				},
				HTTP: HTTPListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-5", "KEY-5")},
					},
					External: ExternalListeners[HTTPExternal]{
						"http-1": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-6", "KEY-6")},
							},
						},
					},
				},
			},
			Out: &corev1.Volume{
				Name: "truststores",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "configmaps/CM-1-KEY-1"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-2"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-2", Path: "configmaps/CM-2-KEY-2"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-3"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-3", Path: "configmaps/CM-3-KEY-3"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-4"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-4", Path: "configmaps/CM-4-KEY-4"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-5"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-5", Path: "configmaps/CM-5-KEY-5"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-6"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-6", Path: "configmaps/CM-6-KEY-6"},
								},
							}},
						},
					},
				},
			},
		},
		{
			Name: "all duplicate secrets",
			Listeners: Listeners{
				Admin: AdminListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
					},
					External: ExternalListeners[AdminExternal]{
						"admin-1": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
							},
						},
					},
				},
				Kafka: KafkaListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
					},
					External: ExternalListeners[KafkaExternal]{
						"kafka-1": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
							},
						},
					},
				},
				HTTP: HTTPListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
					},
					External: ExternalListeners[HTTPExternal]{
						"http-1": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
							},
						},
					},
				},
			},
			Out: &corev1.Volume{
				Name: "truststores",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "secrets/SECRET-1-KEY-1"},
								},
							}},
						},
					},
				},
			},
		},
		{
			Name: "all duplicate configmaps",
			Listeners: Listeners{
				Admin: AdminListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
					},
					External: ExternalListeners[AdminExternal]{
						"admin-1": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
							},
						},
					},
				},
				Kafka: KafkaListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
					},
					External: ExternalListeners[KafkaExternal]{
						"kafka-1": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
							},
						},
					},
				},
				HTTP: HTTPListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
					},
					External: ExternalListeners[HTTPExternal]{
						"http-1": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
							},
						},
					},
				},
			},
			Out: &corev1.Volume{
				Name: "truststores",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "configmaps/CM-1-KEY-1"},
								},
							}},
						},
					},
				},
			},
		},
		{
			Name: "mixture",
			Listeners: Listeners{
				Admin: AdminListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
					},
					External: ExternalListeners[AdminExternal]{
						"admin-1": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-1")},
							},
						},
						"admin-2": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-2")},
							},
						},
						"admin-3": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("disabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-3")},
							},
						},
						"admin-4": AdminExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-2", "KEY-1")},
							},
						},
					},
				},
				Kafka: KafkaListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
					},
					External: ExternalListeners[KafkaExternal]{
						"kafka-1": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
							},
						},
						"kafka-2": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-2")},
							},
						},
						"kafka-3": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("disabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-3")},
							},
						},
						"kafka-4": KafkaExternal{
							Port: 9999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-2", "KEY-1")},
							},
						},
					},
				},
				HTTP: HTTPListeners{
					TLS: InternalTLS{
						Cert:       "enabled",
						TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-2", "KEY-2")},
					},
					External: ExternalListeners[HTTPExternal]{
						"http-1": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{SecretKeyRef: sKeyRef("SECRET-1", "KEY-1")},
							},
						},
						"http-2": HTTPExternal{
							Port: 999,
							TLS: &ExternalTLS{
								Cert:       ptr.To("enabled"),
								TrustStore: &TrustStore{ConfigMapKeyRef: cmKeyRef("CM-1", "KEY-2")},
							},
						},
					},
				},
			},
			Out: &corev1.Volume{
				Name: "truststores",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "configmaps/CM-1-KEY-1"},
									{Key: "KEY-2", Path: "configmaps/CM-1-KEY-2"},
									{Key: "KEY-3", Path: "configmaps/CM-1-KEY-3"},
								},
							}},
							{ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "CM-2"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "configmaps/CM-2-KEY-1"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-1"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "secrets/SECRET-1-KEY-1"},
									{Key: "KEY-2", Path: "secrets/SECRET-1-KEY-2"},
									{Key: "KEY-3", Path: "secrets/SECRET-1-KEY-3"},
								},
							}},
							{Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{Name: "SECRET-2"},
								Items: []corev1.KeyToPath{
									{Key: "KEY-1", Path: "secrets/SECRET-2-KEY-1"},
									{Key: "KEY-2", Path: "secrets/SECRET-2-KEY-2"},
								},
							}},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			vol := tc.Listeners.TrustStoreVolume(&tls)
			require.Equal(t, tc.Out, vol)
		})
	}
}

func TestTieredStorageConfigCreds(t *testing.T) {
	cases := []struct {
		Name     string
		Config   TieredStorageConfig
		Creds    TieredStorageCredentials
		Expected []corev1.EnvVar
	}{
		{
			Name: "azure-secrets",
			Config: TieredStorageConfig{
				"cloud_storage_enabled":               true,
				"cloud_storage_azure_container":       "fake-azure-container",
				"cloud_storage_azure_storage_account": "fake-storage-account",
			},
			Creds: TieredStorageCredentials{
				AccessKey: &SecretRef{},
				SecretKey: &SecretRef{
					Key:  "some-key",
					Name: "some-secret",
				},
			},
			Expected: []corev1.EnvVar{{
				Name: "REDPANDA_CLOUD_STORAGE_AZURE_SHARED_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "some-secret"},
						Key:                  "some-key",
					},
				},
			}},
		},
		{
			Name:   "standard-secrets",
			Config: TieredStorageConfig{},
			Creds: TieredStorageCredentials{
				AccessKey: &SecretRef{Name: "access-secret", Key: "access-key"},
				SecretKey: &SecretRef{Name: "secret-secret", Key: "secret-key"},
			},
			Expected: []corev1.EnvVar{{
				Name: "REDPANDA_CLOUD_STORAGE_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "access-secret"},
						Key:                  "access-key",
					},
				},
			}, {
				Name: "REDPANDA_CLOUD_STORAGE_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "secret-secret"},
						Key:                  "secret-key",
					},
				},
			}},
		},
		{
			Name: "explicit-precedence",
			Config: TieredStorageConfig{
				"cloud_storage_access_key":            "ACCESS_KEY",
				"cloud_storage_azure_shared_key":      "AZURE_SHARED",
				"cloud_storage_azure_container":       "fake-azure-container",
				"cloud_storage_azure_storage_account": "fake-storage-account",
			},
			Creds: TieredStorageCredentials{
				AccessKey: &SecretRef{Name: "access-secret", Key: "access-key"},
				SecretKey: &SecretRef{Name: "secret-secret", Key: "secret-key"},
			},
		},
		{
			Name: "explicit-precedence-azure",
			Config: TieredStorageConfig{
				"cloud_storage_access_key": "ACCESS_KEY",
				"cloud_storage_secret_key": "SECRET_KEY",
			},
			Creds: TieredStorageCredentials{
				AccessKey: &SecretRef{Name: "access-secret", Key: "access-key"},
				SecretKey: &SecretRef{Name: "secret-secret", Key: "secret-key"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			envvars := tc.Creds.AsEnvVars(tc.Config)
			clusterConfig := tc.Config.Translate(&tc.Creds)

			require.EqualValues(t, tc.Expected, envvars)

			// Assert that any envvars have corresponding placeholders at the
			// expected keys in the config. See also: [BootstrapFile].
			for _, envar := range envvars {
				key := strings.ToLower(envar.Name[len("REDPANDA_"):])
				require.Equal(t, "$"+envar.Name, clusterConfig[key])
			}
		})
	}
}
