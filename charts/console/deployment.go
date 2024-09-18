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
// +gotohelm:filename=_deployment.go.tpl
package console

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

// Console's HTTP server Port.
// The port is defined from the provided config but can be overridden
// by setting service.targetPort and if that is missing defaults to 8080.
func ContainerPort(dot *helmette.Dot) int32 {
	values := helmette.Unwrap[Values](dot.Values)

	listenPort := int32(8080)
	if values.Service.TargetPort != nil {
		listenPort = *values.Service.TargetPort
	}

	configListenPort := helmette.Dig(values.Console.Config, nil, "server", "listenPort")
	if asInt, ok := helmette.AsIntegral[int](configListenPort); ok {
		return int32(asInt)
	}

	return listenPort
}

func Deployment(dot *helmette.Dot) *appsv1.Deployment {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Deployment.Create {
		return nil
	}

	var replicas *int32
	if !values.Autoscaling.Enabled {
		replicas = ptr.To(values.ReplicaCount)
	}

	var initContainers []corev1.Container
	if values.InitContainers.ExtraInitContainers != nil {
		initContainers = helmette.UnmarshalYamlArray[corev1.Container](helmette.Tpl(*values.InitContainers.ExtraInitContainers, dot))
	}
	if initContainers == nil {
		initContainers = []corev1.Container{}
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "configs",
			MountPath: "/etc/console/configs",
			ReadOnly:  true,
		},
	}

	if values.Secret.Create {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "secrets",
			MountPath: "/etc/console/secrets",
			ReadOnly:  true,
		})
	}

	for _, mount := range values.SecretMounts {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      mount.Name,
			MountPath: mount.Path,
			SubPath:   ptr.Deref(mount.SubPath, ""),
		})
	}

	volumeMounts = append(volumeMounts, values.ExtraVolumeMounts...)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        Fullname(dot),
			Labels:      Labels(dot),
			Namespace:   dot.Release.Namespace,
			Annotations: values.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: SelectorLabels(dot),
			},
			Strategy: values.Strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: helmette.Merge(map[string]string{
						"checksum/config": helmette.Sha256Sum(helmette.ToYaml(ConfigMap(dot))),
					}, values.PodAnnotations),
					Labels: helmette.Merge(SelectorLabels(dot), values.PodLabels),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets:             values.ImagePullSecrets,
					ServiceAccountName:           ServiceAccountName(dot),
					AutomountServiceAccountToken: &values.AutomountServiceAccountToken,
					SecurityContext:              &values.PodSecurityContext,
					NodeSelector:                 values.NodeSelector,
					Affinity:                     &values.Affinity,
					TopologySpreadConstraints:    values.TopologySpreadConstraints,
					PriorityClassName:            values.PriorityClassName,
					Tolerations:                  values.Tolerations,
					Volumes:                      consolePodVolumes(dot),
					InitContainers:               initContainers,
					Containers: append([]corev1.Container{
						{
							Name:    dot.Chart.Name,
							Command: values.Deployment.Command,
							Args: append([]string{
								"--config.filepath=/etc/console/configs/config.yaml",
							}, values.Deployment.ExtraArgs...),
							SecurityContext: &values.SecurityContext,
							Image:           containerImage(dot),
							ImagePullPolicy: values.Image.PullPolicy,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: ContainerPort(dot),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: volumeMounts,
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: values.LivenessProbe.InitialDelaySeconds, // TODO what to do with this??
								PeriodSeconds:       values.LivenessProbe.PeriodSeconds,
								TimeoutSeconds:      values.LivenessProbe.TimeoutSeconds,
								SuccessThreshold:    values.LivenessProbe.SuccessThreshold,
								FailureThreshold:    values.LivenessProbe.FailureThreshold,
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/admin/health",
										Port: intstr.FromString("http"),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: values.ReadinessProbe.InitialDelaySeconds,
								PeriodSeconds:       values.ReadinessProbe.PeriodSeconds,
								TimeoutSeconds:      values.ReadinessProbe.TimeoutSeconds,
								SuccessThreshold:    values.ReadinessProbe.SuccessThreshold,
								FailureThreshold:    values.ReadinessProbe.FailureThreshold,
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/admin/health",
										Port: intstr.FromString("http"),
									},
								},
							},
							Resources: values.Resources,
							Env:       consoleContainerEnv(dot),
							EnvFrom:   values.ExtraEnvFrom,
						},
					}, values.ExtraContainers...),
				},
			},
		},
	}
}

// ConsoleImage
func containerImage(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	tag := dot.Chart.AppVersion
	if !helmette.Empty(values.Image.Tag) {
		tag = *values.Image.Tag
	}

	image := fmt.Sprintf("%s:%s", values.Image.Repository, tag)

	if !helmette.Empty(values.Image.Registry) {
		return fmt.Sprintf("%s/%s", values.Image.Registry, image)
	}

	return image
}

type PossibleEnvVar struct {
	Value  any
	EnvVar corev1.EnvVar
}

func consoleContainerEnv(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Secret.Create {
		vars := values.ExtraEnv

		if !helmette.Empty(values.Enterprise.LicenseSecretRef.Name) {
			vars = append(values.ExtraEnv, corev1.EnvVar{
				Name: "LICENSE",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: values.Enterprise.LicenseSecretRef.Name,
						},
						Key: helmette.Default("enterprise-license", values.Enterprise.LicenseSecretRef.Key),
					},
				},
			})
		}

		return vars
	}

	possibleVars := []PossibleEnvVar{
		{
			Value: values.Secret.Kafka.SASLPassword,
			EnvVar: corev1.EnvVar{
				Name: "KAFKA_SASL_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "kafka-sasl-password",
					},
				},
			},
		},
		{
			Value: values.Secret.Kafka.ProtobufGitBasicAuthPassword,
			EnvVar: corev1.EnvVar{
				Name: "KAFKA_PROTOBUF_GIT_BASICAUTH_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "kafka-protobuf-git-basicauth-password",
					},
				},
			},
		},
		{
			Value: values.Secret.Kafka.AWSMSKIAMSecretKey,
			EnvVar: corev1.EnvVar{
				Name: "KAFKA_SASL_AWSMSKIAM_SECRETKEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "kafka-sasl-aws-msk-iam-secret-key",
					},
				},
			},
		},
		{
			Value: values.Secret.Kafka.TLSCA,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_TLS_CAFILEPATH",
				Value: "/etc/console/secrets/kafka-tls-ca",
			},
		},
		{
			Value: values.Secret.Kafka.TLSCert,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_TLS_CERTFILEPATH",
				Value: "/etc/console/secrets/kafka-tls-cert",
			},
		},
		{
			Value: values.Secret.Kafka.TLSKey,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_TLS_KEYFILEPATH",
				Value: "/etc/console/secrets/kafka-tls-key",
			},
		},
		{
			Value: values.Secret.Kafka.SchemaRegistryTLSCA,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_SCHEMAREGISTRY_TLS_CAFILEPATH",
				Value: "/etc/console/secrets/kafka-schemaregistry-tls-ca",
			},
		},
		{
			Value: values.Secret.Kafka.SchemaRegistryTLSCert,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_SCHEMAREGISTRY_TLS_CERTFILEPATH",
				Value: "/etc/console/secrets/kafka-schemaregistry-tls-cert",
			},
		},
		{
			Value: values.Secret.Kafka.SchemaRegistryTLSKey,
			EnvVar: corev1.EnvVar{
				Name:  "KAFKA_SCHEMAREGISTRY_TLS_KEYFILEPATH",
				Value: "/etc/console/secrets/kafka-schemaregistry-tls-key",
			},
		},
		{
			Value: values.Secret.Kafka.SchemaRegistryPassword,
			EnvVar: corev1.EnvVar{
				Name: "KAFKA_SCHEMAREGISTRY_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "kafka-schema-registry-password",
					},
				},
			},
		},
		{
			Value: true,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_JWTSECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-jwt-secret",
					},
				},
			},
		},
		{
			Value: values.Secret.Login.Google.ClientSecret,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_GOOGLE_CLIENTSECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-google-oauth-client-secret",
					},
				},
			},
		},

		{
			Value: values.Secret.Login.Google.GroupsServiceAccount,
			EnvVar: corev1.EnvVar{
				Name:  "LOGIN_GOOGLE_DIRECTORY_SERVICEACCOUNTFILEPATH",
				Value: "/etc/console/secrets/login-google-groups-service-account.json",
			},
		},
		{
			Value: values.Secret.Login.Github.ClientSecret,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_GITHUB_CLIENTSECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-github-oauth-client-secret",
					},
				},
			},
		},
		{
			Value: values.Secret.Login.Github.PersonalAccessToken,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_GITHUB_DIRECTORY_PERSONALACCESSTOKEN",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-github-personal-access-token",
					},
				},
			},
		},
		{
			Value: values.Secret.Login.Okta.ClientSecret,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_OKTA_CLIENTSECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-okta-client-secret",
					},
				},
			},
		},
		{
			Value: values.Secret.Login.Okta.DirectoryAPIToken,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_OKTA_DIRECTORY_APITOKEN",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-okta-directory-api-token",
					},
				},
			},
		},
		{
			Value: values.Secret.Login.OIDC.ClientSecret,
			EnvVar: corev1.EnvVar{
				Name: "LOGIN_OIDC_CLIENTSECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "login-oidc-client-secret",
					},
				},
			},
		},
		{
			Value: values.Secret.Enterprise.License,
			EnvVar: corev1.EnvVar{
				Name: "LICENSE",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "enterprise-license",
					},
				},
			},
		},
		{
			Value: values.Secret.Redpanda.AdminAPI.Password,
			EnvVar: corev1.EnvVar{
				Name: "REDPANDA_ADMINAPI_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: Fullname(dot),
						},
						Key: "redpanda-admin-api-password",
					},
				},
			},
		},
		{
			Value: values.Secret.Redpanda.AdminAPI.TLSCA,
			EnvVar: corev1.EnvVar{
				Name:  "REDPANDA_ADMINAPI_TLS_CAFILEPATH",
				Value: "/etc/console/secrets/redpanda-admin-api-tls-ca",
			},
		},
		{
			Value: values.Secret.Redpanda.AdminAPI.TLSKey,
			EnvVar: corev1.EnvVar{
				Name:  "REDPANDA_ADMINAPI_TLS_KEYFILEPATH",
				Value: "/etc/console/secrets/redpanda-admin-api-tls-key",
			},
		},
		{
			Value: values.Secret.Redpanda.AdminAPI.TLSCert,
			EnvVar: corev1.EnvVar{
				Name:  "REDPANDA_ADMINAPI_TLS_CERTFILEPATH",
				Value: "/etc/console/secrets/redpanda-admin-api-tls-cert",
			},
		},
	}

	vars := values.ExtraEnv
	for _, possible := range possibleVars {
		if !helmette.Empty(possible.Value) {
			vars = append(vars, possible.EnvVar)
		}
	}

	return vars
}

func consolePodVolumes(dot *helmette.Dot) []corev1.Volume {
	values := helmette.Unwrap[Values](dot.Values)

	volumes := []corev1.Volume{
		{
			Name: "configs",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: Fullname(dot),
					},
				},
			},
		},
	}

	if values.Secret.Create {
		volumes = append(volumes, corev1.Volume{
			Name: "secrets",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: Fullname(dot),
				},
			},
		})
	}

	for _, mount := range values.SecretMounts {
		volumes = append(volumes, corev1.Volume{
			Name: mount.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  mount.SecretName,
					DefaultMode: mount.DefaultMode,
				},
			},
		})
	}

	return append(volumes, values.ExtraVolumes...)
}
