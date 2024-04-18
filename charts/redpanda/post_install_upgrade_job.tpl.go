// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +gotohelm:filename=_post-install-upgrade-job.go.tpl
package redpanda

import (
	"fmt"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

// RedpandaEnvironmentVariables returns environment variables assigned to Redpanda
// container.
func RedpandaEnvironmentVariables(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	tieredStorageConfig := values.Storage.Tiered.Config

	if values.Storage.TieredConfig != nil {
		if len(values.Storage.TieredConfig) > 0 {
			tieredStorageConfig = values.Storage.TieredConfig
		}
	}

	envars := []corev1.EnvVar{}

	if license := GetLicense(dot); license != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "REDPANDA_LICENSE",
			Value: license,
		})
	} else if secretReference := EnterpriseSecretNameReference(dot); secretReference != nil {
		envars = append(envars, corev1.EnvVar{
			Name: "REDPANDA_LICENSE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: secretReference,
			},
		})
	}

	if !IsTieredStorageEnabled(tieredStorageConfig) {
		return envars
	}

	if !helmette.Empty(tieredStorageConfig["cloud_storage_azure_container"]) && !helmette.Empty(tieredStorageConfig["cloud_storage_azure_storage_account"]) {
		if !helmette.Empty(tieredStorageConfig["cloud_storage_azure_shared_key"]) {
			envars = append(envars, corev1.EnvVar{
				Name:  "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				Value: helmette.MustToJSON(tieredStorageConfig["cloud_storage_azure_shared_key"]),
			})
		} else if values.Storage.Tiered.CredentialsSecretRef.SecretKey != nil &&
			!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.SecretKey.Name) &&
			!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.SecretKey.Key) {
			envars = append(envars, corev1.EnvVar{
				Name: "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: values.Storage.Tiered.CredentialsSecretRef.SecretKey.Name},
						Key:                  values.Storage.Tiered.CredentialsSecretRef.SecretKey.Key,
					},
				},
			})
		}
	} else {
		if !helmette.Empty(tieredStorageConfig["cloud_storage_secret_key"]) {
			envars = append(envars, corev1.EnvVar{
				Name:  "RPK_CLOUD_STORAGE_SECRET_KEY",
				Value: tieredStorageConfig["cloud_storage_secret_key"].(string),
			})
		} else if values.Storage.Tiered.CredentialsSecretRef.SecretKey != nil &&
			!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.SecretKey.Name) &&
			!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.SecretKey.Key) {
			envars = append(envars, corev1.EnvVar{
				Name: "RPK_CLOUD_STORAGE_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: values.Storage.Tiered.CredentialsSecretRef.SecretKey.Name},
						Key:                  values.Storage.Tiered.CredentialsSecretRef.SecretKey.Key,
					},
				},
			})
		}
	}

	if !helmette.Empty(tieredStorageConfig["cloud_storage_access_key"]) {
		envars = append(envars, corev1.EnvVar{
			Name:  "RPK_CLOUD_STORAGE_ACCESS_KEY",
			Value: tieredStorageConfig["cloud_storage_access_key"].(string),
		})
	} else if values.Storage.Tiered.CredentialsSecretRef.SecretKey != nil &&
		!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.AccessKey.Name) &&
		!helmette.Empty(values.Storage.Tiered.CredentialsSecretRef.AccessKey.Key) {
		envars = append(envars, corev1.EnvVar{
			Name: "RPK_CLOUD_STORAGE_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: values.Storage.Tiered.CredentialsSecretRef.AccessKey.Name},
					Key:                  values.Storage.Tiered.CredentialsSecretRef.AccessKey.Key,
				},
			},
		})
	}

	for k, v := range tieredStorageConfig {
		if k == "cloud_storage_access_key" || k == "cloud_storage_secret_key" || k == "cloud_storage_azure_shared_key" {
			continue
		}

		if v == nil || helmette.Empty(v) {
			continue
		}

		if k == "cloud_storage_cache_size" {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: helmette.ToJSON(helmette.Int64(helmette.Sitobytes(v))),
			})
			continue
		}

		envars = append(envars, corev1.EnvVar{
			Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
			Value: helmette.MustToJSON(v),
		})
	}

	return envars
}

func GetLicense(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Enterprise.License != "" {
		return values.Enterprise.License
	}

	return values.LicenseKey
}

func EnterpriseSecretNameReference(dot *helmette.Dot) *corev1.SecretKeySelector {
	values := helmette.Unwrap[Values](dot.Values)

	if !helmette.Empty(values.Enterprise.LicenseSecretRef) {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: values.Enterprise.LicenseSecretRef.Name,
			},
			Key: values.Enterprise.LicenseSecretRef.Key,
		}
	} else if !helmette.Empty(values.LicenseSecretRef) {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: values.LicenseSecretRef.SecretName,
			},
			Key: values.LicenseSecretRef.SecretKey,
		}
	}
	return nil
}

func IsTieredStorageEnabled(tieredStorageConfig TieredStorageConfig) bool {
	if b, ok := tieredStorageConfig["cloud_storage_enabled"]; ok && b.(bool) {
		return true
	}
	return false
}
