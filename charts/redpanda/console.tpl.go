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
// +gotohelm:filename=_console.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/console/backend/pkg/config"
	"github.com/redpanda-data/helm-charts/charts/console"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// consoleChartIntegration plumbs redpanda connection information into the console subchart.
// It does this by calculating Kafka, Schema registry, Redpanda Admin API configuration
// from Redpanda chart values.
func consoleChartIntegration(dot *helmette.Dot) []kube.Object {
	values := helmette.UnmarshalInto[Values](dot.Values)

	if !ptr.Deref(values.Console.Enabled, true) {
		return nil
	}

	consoleDot := dot.Subcharts["console"]
	loadedValues := consoleDot.Values

	consoleValue := helmette.UnmarshalInto[console.Values](consoleDot.Values)
	// Pass the same Redpanda License to Console
	if license := GetLicenseLiteral(dot); license != "" && !ptr.Deref(values.Console.Secret.Create, false) {
		consoleValue.Secret.Create = true
		consoleValue.Secret.Enterprise = console.EnterpriseSecrets{License: ptr.To(license)}
	}

	// Create console configuration based on Redpanda helm chart values.
	if !ptr.Deref(values.Console.ConfigMap.Create, false) {
		consoleValue.ConfigMap.Create = true
		consoleValue.Console.Config = ConsoleConfig(dot)
	}

	if !ptr.Deref(values.Console.Deployment.Create, false) {
		consoleValue.Deployment.Create = true

		// Adopt Console entry point to use SASL user in Kafka,
		// Schema Registry and Redpanda Admin API connection
		if values.Auth.IsSASLEnabled() {
			command := []string{
				"sh",
				"-c",
				"set -e; IFS=':' read -r KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM < <(grep \"\" $(find /mnt/users/* -print));" +
					fmt.Sprintf(" KAFKA_SASL_MECHANISM=${KAFKA_SASL_MECHANISM:-%s};", SASLMechanism(dot)) +
					" export KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM;" +
					" export KAFKA_SCHEMAREGISTRY_USERNAME=$KAFKA_SASL_USERNAME;" +
					" export KAFKA_SCHEMAREGISTRY_PASSWORD=$KAFKA_SASL_PASSWORD;" +
					" export REDPANDA_ADMINAPI_USERNAME=$KAFKA_SASL_USERNAME;" +
					" export REDPANDA_ADMINAPI_PASSWORD=$KAFKA_SASL_PASSWORD;" +
					" /app/console $@",
				" --",
			}
			consoleValue.Deployment.Command = command
		}

		// Create License reference for Console
		if secret := GetLicenseSecretReference(dot); secret != nil {
			consoleValue.Enterprise = console.Enterprise{
				LicenseSecretRef: console.SecretKeyRef{
					Name: secret.Name,
					Key:  secret.Key,
				},
			}
		}

		consoleValue.ExtraVolumes = consoleTLSVolumes(dot)
		consoleValue.ExtraVolumeMounts = consoleTLSVolumesMounts(dot)

		consoleDot.Values = helmette.UnmarshalInto[helmette.Values](consoleValue)
		cfg := console.ConfigMap(consoleDot)
		if consoleValue.PodAnnotations == nil {
			consoleValue.PodAnnotations = map[string]string{}
		}
		consoleValue.PodAnnotations["checksum-redpanda-chart/config"] = helmette.Sha256Sum(helmette.ToYaml(cfg))

	}

	consoleDot.Values = helmette.UnmarshalInto[helmette.Values](consoleValue)

	manifests := []kube.Object{
		console.Secret(consoleDot),
		console.ConfigMap(consoleDot),
		console.Deployment(consoleDot),
	}

	consoleDot.Values = loadedValues

	// NB: This slice may contain nil interfaces!
	// Filtering happens elsewhere, don't call this function directly if you
	// can avoid it.
	return manifests
}

func consoleTLSVolumesMounts(dot *helmette.Dot) []corev1.VolumeMount {
	values := helmette.Unwrap[Values](dot.Values)

	mounts := []corev1.VolumeMount{}

	if sasl := values.Auth.SASL; sasl.Enabled && sasl.SecretRef != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-users", Fullname(dot)),
			MountPath: "/mnt/users",
			ReadOnly:  true,
		})
	}

	if len(values.Listeners.TrustStores(&values.TLS)) > 0 {
		mounts = append(
			mounts,
			corev1.VolumeMount{Name: "truststores", MountPath: TrustStoreMountPath, ReadOnly: true},
		)
	}

	visitedCert := map[string]bool{}
	for _, tlsCfg := range []InternalTLS{
		values.Listeners.Kafka.TLS,
		values.Listeners.SchemaRegistry.TLS,
		values.Listeners.Admin.TLS,
	} {
		_, visited := visitedCert[tlsCfg.Cert]
		if !tlsCfg.IsEnabled(&values.TLS) || visited {
			continue
		}
		visitedCert[tlsCfg.Cert] = true

		mounts = append(mounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("redpanda-%s-cert", tlsCfg.Cert),
			MountPath: fmt.Sprintf("/etc/tls/certs/%s", tlsCfg.Cert),
		})
	}

	return mounts
}

func consoleTLSVolumes(dot *helmette.Dot) []corev1.Volume {
	values := helmette.Unwrap[Values](dot.Values)

	volumes := []corev1.Volume{}

	if sasl := values.Auth.SASL; sasl.Enabled && sasl.SecretRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%s-users", Fullname(dot)),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: values.Auth.SASL.SecretRef,
				},
			},
		})
	}

	if vol := values.Listeners.TrustStoreVolume(&values.TLS); vol != nil {
		volumes = append(volumes, *vol)
	}

	visitedCert := map[string]bool{}
	for _, tlsCfg := range []InternalTLS{
		values.Listeners.Kafka.TLS,
		values.Listeners.SchemaRegistry.TLS,
		values.Listeners.Admin.TLS,
	} {
		_, visited := visitedCert[tlsCfg.Cert]
		if !tlsCfg.IsEnabled(&values.TLS) || visited {
			continue
		}
		visitedCert[tlsCfg.Cert] = true

		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("redpanda-%s-cert", tlsCfg.Cert),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: ptr.To[int32](0o420),
					SecretName:  CertSecretName(dot, tlsCfg.Cert, values.TLS.Certs.MustGet(tlsCfg.Cert)),
				},
			},
		})
	}

	return volumes
}

func ConsoleConfig(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	var schemaURLs []string
	if values.Listeners.SchemaRegistry.Enabled {
		schema := "http"
		if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
			schema = "https"
		}

		for i := int32(0); i < values.Statefulset.Replicas; i++ {
			schemaURLs = append(schemaURLs, fmt.Sprintf("%s://%s-%d.%s:%d", schema, Fullname(dot), i, InternalDomain(dot), values.Listeners.SchemaRegistry.Port))
		}
	}

	schema := "http"
	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		schema = "https"
	}

	c := map[string]any{
		"kafka": map[string]any{
			"brokers": BrokerList(dot, values.Statefulset.Replicas, values.Listeners.Kafka.Port),
			"sasl": map[string]any{
				"enabled": values.Auth.IsSASLEnabled(),
			},
			"tls": values.Listeners.Kafka.ConsoleTLS(&values.TLS),
			"schemaRegistry": map[string]any{
				"enabled": values.Listeners.SchemaRegistry.Enabled,
				"urls":    schemaURLs,
				"tls":     values.Listeners.SchemaRegistry.ConsoleTLS(&values.TLS),
			},
		},
		"redpanda": map[string]any{
			"adminApi": map[string]any{
				"enabled": true,
				"urls": []string{
					fmt.Sprintf("%s://%s:%d", schema, InternalDomain(dot), values.Listeners.Admin.Port),
				},
				"tls": values.Listeners.Admin.ConsoleTLS(&values.TLS),
			},
		},
	}

	if values.Connectors.Enabled {
		// TODO Do not cal Dig with dot.Values as restPort that is defined in connectors helm chart is not
		// available in this function.
		// TODO Find a way to call `(include "connectors.serviceName" $connectorsValues)` template defined
		// in connectors helm chart repo.

		port := helmette.Dig(dot.Values.AsMap(), 8083, "connectors", "connectors", "restPort")
		p, ok := helmette.AsIntegral[int](port)
		if !ok {
			return c
		}

		connectorsURL := fmt.Sprintf("http://%s.%s.svc.%s:%d",
			ConnectorsFullName(dot),
			dot.Release.Namespace,
			helmette.TrimSuffix(".", values.ClusterDomain),
			p)

		c["connect"] = config.Connect{
			Enabled: values.Connectors.Enabled,
			Clusters: []config.ConnectCluster{
				{
					Name: "connectors",
					URL:  connectorsURL,
					TLS: config.ConnectClusterTLS{
						Enabled:               false,
						CaFilepath:            "",
						CertFilepath:          "",
						KeyFilepath:           "",
						InsecureSkipTLSVerify: false,
					},
					Username: "",
					Password: "",
					Token:    "",
				},
			},
		}
	}

	if values.Console.Console == nil {
		values.Console.Console = &console.PartialConsole{
			Config: map[string]any{},
		}
	}

	return helmette.Merge(values.Console.Console.Config, c)
}

func ConnectorsFullName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if helmette.Dig(dot.Values.AsMap(), "", "connectors", "connectors", "fullnameOverwrite") != "" {
		return cleanForK8s(values.Connectors.Connectors.FullnameOverwrite)
	}

	return cleanForK8s(fmt.Sprintf("%s-connectors", dot.Release.Name))
}
