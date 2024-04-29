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

// PostInstallUpgradeEnvironmentVariables returns environment variables assigned to Redpanda
// container.
func PostInstallUpgradeEnvironmentVariables(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	envars := []corev1.EnvVar{}

	if license := GetLicenseLiteral(dot); license != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "REDPANDA_LICENSE",
			Value: license,
		})
	} else if secretReference := GetLicenseSecretReference(dot); secretReference != nil {
		envars = append(envars, corev1.EnvVar{
			Name: "REDPANDA_LICENSE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: secretReference,
			},
		})
	}

	if !values.Storage.IsTieredStorageEnabled() {
		return envars
	}

	tieredStorageConfig := values.Storage.GetTieredStorageConfig()

	ac, azureContainerExists := tieredStorageConfig["cloud_storage_azure_container"]
	asa, azureStorageAccountExists := tieredStorageConfig["cloud_storage_azure_storage_account"]
	if azureContainerExists && ac != nil && azureStorageAccountExists && asa != nil {
		envars = append(envars, addAzureSharedKey(tieredStorageConfig, values)...)
	} else {
		envars = append(envars, addCloudStorageSecretKey(tieredStorageConfig, values)...)
	}

	envars = append(envars, addCloudStorageAccessKey(tieredStorageConfig, values)...)

	for k, v := range tieredStorageConfig {
		if k == "cloud_storage_access_key" || k == "cloud_storage_secret_key" || k == "cloud_storage_azure_shared_key" {
			continue
		}

		if v == nil || helmette.Empty(v) {
			continue
		}

		// cloud_storage_cache_size can be represented as Resource.Quantity that why value can be converted
		// from value with SI suffix to bytes number.
		if asStr, isStr := v.(string); k == "cloud_storage_cache_size" && isStr && asStr != "" {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: helmette.ToJSON(SIToBytes(v.(string))),
			})
			continue
		}

		if str, ok := v.(string); ok {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: str,
			})
		} else {
			envars = append(envars, corev1.EnvVar{
				Name:  fmt.Sprintf("RPK_%s", helmette.Upper(k)),
				Value: helmette.MustToJSON(v),
			})
		}
	}

	return envars
}

func addCloudStorageAccessKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	if v, ok := tieredStorageConfig["cloud_storage_access_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_ACCESS_KEY",
				Value: v.(string),
			},
		}
		// TODO change this to the following representation when struct function transpilation would work
		//} else if values.Storage.Tiered.CredentialsSecretRef.IsAccessKeyReferenceValid() {
	} else if ak := values.Storage.Tiered.CredentialsSecretRef.AccessKey; ak != nil &&
		!helmette.Empty(ak.Name) &&
		!helmette.Empty(ak.Key) {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: ak.Name},
						Key:                  ak.Key,
					},
				},
			},
		}
	}
	return []corev1.EnvVar{}
}

func addCloudStorageSecretKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	if v, ok := tieredStorageConfig["cloud_storage_secret_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_SECRET_KEY",
				Value: v.(string),
			},
		}
		// TODO change this to the following representation when struct function transpilation would work
		//} else if values.Storage.Tiered.CredentialsSecretRef.IsSecretKeyReferenceValid() {
	} else if sk := values.Storage.Tiered.CredentialsSecretRef.SecretKey; sk != nil &&
		!helmette.Empty(sk.Name) &&
		!helmette.Empty(sk.Key) {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: sk.Name},
						Key:                  sk.Key,
					},
				},
			},
		}
	}
	return []corev1.EnvVar{}
}

func addAzureSharedKey(tieredStorageConfig TieredStorageConfig, values Values) []corev1.EnvVar {
	// Preference Tiered Storage Config over credential secret reference
	if v, ok := tieredStorageConfig["cloud_storage_azure_shared_key"]; ok && v != "" {
		return []corev1.EnvVar{
			{
				Name:  "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				Value: v.(string),
			},
		}
		// TODO change this to the following representation when struct function transpilation would work
		//} else if values.Storage.Tiered.CredentialsSecretRef.IsSecretKeyReferenceValid() {
	} else if sk := values.Storage.Tiered.CredentialsSecretRef.SecretKey; sk != nil &&
		!helmette.Empty(sk.Name) &&
		!helmette.Empty(sk.Key) {
		return []corev1.EnvVar{
			{
				Name: "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: sk.Name},
						Key:                  sk.Key,
					},
				},
			},
		}
	}

	return []corev1.EnvVar{}
}

func GetLicenseLiteral(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Enterprise.License != "" {
		return values.Enterprise.License
	}

	// Deprecated licenseKey fallback if Enterprise.License is not set
	return values.LicenseKey
}

func GetLicenseSecretReference(dot *helmette.Dot) *corev1.SecretKeySelector {
	values := helmette.Unwrap[Values](dot.Values)

	if !helmette.Empty(values.Enterprise.LicenseSecretRef) {
		return &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: values.Enterprise.LicenseSecretRef.Name,
			},
			Key: values.Enterprise.LicenseSecretRef.Key,
		}
		// Deprecated licenseSecretRef fallback if Enterprise.LicenseSecretRef is not set
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
