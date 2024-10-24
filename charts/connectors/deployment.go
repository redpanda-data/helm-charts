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
package connectors

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func Deployment(dot *helmette.Dot) *appsv1.Deployment {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Deployment.Create {
		return nil
	}

	var topologySpreadConstraints []corev1.TopologySpreadConstraint
	for _, spread := range values.Deployment.TopologySpreadConstraints {
		topologySpreadConstraints = append(topologySpreadConstraints, corev1.TopologySpreadConstraint{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: PodLabels(dot),
			},
			MaxSkew:           spread.MaxSkew,
			TopologyKey:       spread.TopologyKey,
			WhenUnsatisfiable: spread.WhenUnsatisfiable,
		})
	}

	ports := []corev1.ContainerPort{
		{
			ContainerPort: values.Connectors.RestPort,
			Name:          "rest-api",
			Protocol:      corev1.ProtocolTCP,
		},
	}

	for _, port := range values.Service.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.Port,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	var podAntiAffinity *corev1.PodAntiAffinity
	if values.Deployment.PodAntiAffinity != nil {
		if values.Deployment.PodAntiAffinity.Type == "hard" {
			podAntiAffinity = &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
					TopologyKey: values.Deployment.PodAntiAffinity.TopologyKey,
					Namespaces:  []string{dot.Release.Namespace},
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: PodLabels(dot),
					},
				}},
			}
		} else if values.Deployment.PodAntiAffinity.Type == "soft" {
			podAntiAffinity = &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: *values.Deployment.PodAntiAffinity.Weight,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: values.Deployment.PodAntiAffinity.TopologyKey,
						Namespaces:  []string{dot.Release.Namespace},
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: PodLabels(dot),
						},
					},
				}},
			}
		} else if values.Deployment.PodAntiAffinity.Type == "custom" {
			podAntiAffinity = values.Deployment.PodAntiAffinity.Custom
		}
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   Fullname(dot),
			Labels: helmette.Merge(FullLabels(dot), values.Deployment.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:                values.Deployment.Replicas,
			ProgressDeadlineSeconds: &values.Deployment.ProgressDeadlineSeconds,
			RevisionHistoryLimit:    values.Deployment.RevisionHistoryLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: PodLabels(dot),
			},
			Strategy: values.Deployment.Strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: values.Deployment.Annotations,
					Labels:      PodLabels(dot),
				},
				Spec: corev1.PodSpec{
					// Users will not be able to set auto mount ServiceAccount token to `true` in PodSpec.
					// If user would like to mount token, then ServiceAccount should be used to allow auto
					// mounting of the ServiceAccount token (`serviceAccount.automountServiceAccountToken`
					// in the input values of connectors chart).
					AutomountServiceAccountToken:  ptr.To(false),
					TerminationGracePeriodSeconds: values.Deployment.TerminationGracePeriodSeconds,
					Affinity: &corev1.Affinity{
						NodeAffinity:    values.Deployment.NodeAffinity,
						PodAffinity:     values.Deployment.PodAffinity,
						PodAntiAffinity: podAntiAffinity,
					},
					ServiceAccountName: ServiceAccountName(dot),
					Containers: []corev1.Container{
						{
							Name:            "connectors-cluster",
							Image:           fmt.Sprintf("%s:%s", values.Image.Repository, Tag(dot)),
							ImagePullPolicy: values.Image.PullPolicy,
							SecurityContext: &values.Container.SecurityContext,
							Command:         values.Deployment.Command,
							Env:             env(&values),
							EnvFrom:         values.Deployment.ExtraEnvFrom,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/",
										Port:   intstr.FromString("rest-api"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: values.Deployment.LivenessProbe.InitialDelaySeconds,
								TimeoutSeconds:      values.Deployment.LivenessProbe.TimeoutSeconds,
								PeriodSeconds:       values.Deployment.LivenessProbe.PeriodSeconds,
								SuccessThreshold:    values.Deployment.LivenessProbe.SuccessThreshold,
								FailureThreshold:    values.Deployment.LivenessProbe.FailureThreshold,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/connectors",
										Port:   intstr.FromString("rest-api"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: values.Deployment.ReadinessProbe.InitialDelaySeconds,
								TimeoutSeconds:      values.Deployment.ReadinessProbe.TimeoutSeconds,
								PeriodSeconds:       values.Deployment.ReadinessProbe.PeriodSeconds,
								SuccessThreshold:    values.Deployment.ReadinessProbe.SuccessThreshold,
								FailureThreshold:    values.Deployment.ReadinessProbe.FailureThreshold,
							},
							Ports: ports,
							Resources: corev1.ResourceRequirements{
								Requests: values.Container.Resources.Request,
								Limits:   values.Container.Resources.Limits,
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							VolumeMounts:             volumeMountss(&values),
						},
					},
					DNSPolicy:                 corev1.DNSClusterFirst,
					RestartPolicy:             values.Deployment.RestartPolicy,
					SchedulerName:             values.Deployment.SchedulerName,
					NodeSelector:              values.Deployment.NodeSelector,
					ImagePullSecrets:          values.ImagePullSecrets,
					SecurityContext:           values.Deployment.SecurityContext,
					Tolerations:               values.Deployment.Tolerations,
					TopologySpreadConstraints: topologySpreadConstraints,
					Volumes:                   volumes(&values),
				},
			},
		},
	}
}

func env(values *Values) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  "CONNECT_CONFIGURATION",
			Value: connectorConfiguration(values),
		},
		{
			Name:  "CONNECT_ADDITIONAL_CONFIGURATION",
			Value: values.Connectors.AdditionalConfiguration,
		},
		{
			Name:  "CONNECT_BOOTSTRAP_SERVERS",
			Value: values.Connectors.BootstrapServers,
		},
	}

	if !helmette.Empty(values.Connectors.SchemaRegistryURL) {
		env = append(env, corev1.EnvVar{
			Name:  "SCHEMA_REGISTRY_URL",
			Value: values.Connectors.SchemaRegistryURL,
		})
	}

	env = append(env, corev1.EnvVar{
		Name:  "CONNECT_GC_LOG_ENABLED",
		Value: values.Container.JavaGCLogEnabled,
	}, corev1.EnvVar{
		Name:  "CONNECT_HEAP_OPTS",
		Value: fmt.Sprintf("-Xms256M -Xmx%s", values.Container.Resources.JavaMaxHeapSize),
	}, corev1.EnvVar{
		Name:  "CONNECT_LOG_LEVEL",
		Value: values.Logging.Level,
	})

	if values.Auth.SASLEnabled() {
		env = append(env, corev1.EnvVar{
			Name:  "CONNECT_SASL_USERNAME",
			Value: values.Auth.SASL.UserName,
		}, corev1.EnvVar{
			Name:  "CONNECT_SASL_MECHANISM",
			Value: values.Auth.SASL.Mechanism,
		}, corev1.EnvVar{
			Name:  "CONNECT_SASL_PASSWORD_FILE",
			Value: "rc-credentials/password",
		})
	}

	env = append(env, corev1.EnvVar{
		Name:  "CONNECT_TLS_ENABLED",
		Value: fmt.Sprintf("%v", values.Connectors.BrokerTLS.Enabled),
	})

	if !helmette.Empty(values.Connectors.BrokerTLS.CA.SecretRef) {
		ca := helmette.Default("ca.crt", values.Connectors.BrokerTLS.CA.SecretNameOverwrite)
		env = append(env, corev1.EnvVar{
			Name:  "CONNECT_TRUSTED_CERTS",
			Value: fmt.Sprintf("ca/%s", ca),
		})
	}

	if !helmette.Empty(values.Connectors.BrokerTLS.Cert.SecretRef) {
		cert := helmette.Default("tls.crt", values.Connectors.BrokerTLS.Cert.SecretNameOverwrite)
		env = append(env, corev1.EnvVar{
			Name:  "CONNECT_TLS_AUTH_CERT",
			Value: fmt.Sprintf("cert/%s", cert),
		})
	}

	if !helmette.Empty(values.Connectors.BrokerTLS.Key.SecretRef) {
		key := helmette.Default("tls.key", values.Connectors.BrokerTLS.Key.SecretNameOverwrite)
		env = append(env, corev1.EnvVar{
			Name:  "CONNECT_TLS_AUTH_KEY",
			Value: fmt.Sprintf("key/%s", key),
		})
	}

	return append(env, values.Deployment.ExtraEnv...)
}

func connectorConfiguration(values *Values) string {
	lines := []string{
		fmt.Sprintf("rest.advertised.port=%d", values.Connectors.RestPort),
		fmt.Sprintf("rest.port=%d", values.Connectors.RestPort),
		"key.converter=org.apache.kafka.connect.converters.ByteArrayConverter",
		"value.converter=org.apache.kafka.connect.converters.ByteArrayConverter",
		fmt.Sprintf("group.id=%s", values.Connectors.GroupID),
		fmt.Sprintf("offset.storage.topic=%s", values.Connectors.Storage.Topic.Offset),
		fmt.Sprintf("config.storage.topic=%s", values.Connectors.Storage.Topic.Config),
		fmt.Sprintf("status.storage.topic=%s", values.Connectors.Storage.Topic.Status),
		fmt.Sprintf("offset.storage.redpanda.remote.read=%t", values.Connectors.Storage.Remote.Read.Offset),
		fmt.Sprintf("offset.storage.redpanda.remote.write=%t", values.Connectors.Storage.Remote.Write.Offset),
		fmt.Sprintf("config.storage.redpanda.remote.read=%t", values.Connectors.Storage.Remote.Read.Config),
		fmt.Sprintf("config.storage.redpanda.remote.write=%t", values.Connectors.Storage.Remote.Write.Config),
		fmt.Sprintf("status.storage.redpanda.remote.read=%t", values.Connectors.Storage.Remote.Read.Status),
		fmt.Sprintf("status.storage.redpanda.remote.write=%t", values.Connectors.Storage.Remote.Write.Status),
		fmt.Sprintf("offset.storage.replication.factor=%d", values.Connectors.Storage.ReplicationFactor.Offset),
		fmt.Sprintf("config.storage.replication.factor=%d", values.Connectors.Storage.ReplicationFactor.Config),
		fmt.Sprintf("status.storage.replication.factor=%d", values.Connectors.Storage.ReplicationFactor.Status),
		fmt.Sprintf("producer.linger.ms=%d", values.Connectors.ProducerLingerMS),
		fmt.Sprintf("producer.batch.size=%d", values.Connectors.ProducerBatchSize),
		"config.providers=file,secretsManager,env",
		"config.providers.file.class=org.apache.kafka.common.config.provider.FileConfigProvider",
	}

	if values.Connectors.SecretManager.Enabled {
		lines = append(
			lines,
			"config.providers.secretsManager.class=com.github.jcustenborder.kafka.config.aws.SecretsManagerConfigProvider",
			fmt.Sprintf("config.providers.secretsManager.param.secret.prefix=%s%s", values.Connectors.SecretManager.ConsolePrefix, values.Connectors.SecretManager.ConnectorsPrefix),
			fmt.Sprintf("config.providers.secretsManager.param.aws.region=%s", values.Connectors.SecretManager.Region),
		)
	}

	lines = append(
		lines,
		"config.providers.env.class=org.apache.kafka.common.config.provider.EnvVarConfigProvider",
	)

	return helmette.Join("\n", lines)
}

func volumes(values *Values) []corev1.Volume {
	var volumes []corev1.Volume
	if !helmette.Empty(values.Connectors.BrokerTLS.CA.SecretRef) {
		volumes = append(volumes, corev1.Volume{
			Name: "truststore",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: ptr.To[int32](0o444),
					SecretName:  values.Connectors.BrokerTLS.CA.SecretRef,
				},
			},
		})
	}
	if !helmette.Empty(values.Connectors.BrokerTLS.Cert.SecretRef) {
		volumes = append(volumes, corev1.Volume{
			Name: "cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: ptr.To[int32](0o444),
					SecretName:  values.Connectors.BrokerTLS.Cert.SecretRef,
				},
			},
		})
	}
	if !helmette.Empty(values.Connectors.BrokerTLS.Key.SecretRef) {
		volumes = append(volumes, corev1.Volume{
			Name: "key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: ptr.To[int32](0o444),
					SecretName:  values.Connectors.BrokerTLS.Key.SecretRef,
				},
			},
		})
	}

	if values.Auth.SASLEnabled() {
		volumes = append(volumes, corev1.Volume{
			Name: "rc-credentials",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: ptr.To[int32](0o444),
					SecretName:  values.Auth.SASL.SecretRef,
				},
			},
		})
	}

	return append(volumes, values.Storage.Volume...)
}

func volumeMountss(values *Values) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	if values.Auth.SASLEnabled() {
		mounts = append(mounts, corev1.VolumeMount{
			MountPath: "/opt/kafka/connect-password/rc-credentials",
			Name:      "rc-credentials",
		})
	}

	if !helmette.Empty(values.Connectors.BrokerTLS.CA.SecretRef) {
		// The /opt/kafka/connect-certs is fixed path within Connectors
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "truststore",
			MountPath: "/opt/kafka/connect-certs/ca",
		})
	}

	if !helmette.Empty(values.Connectors.BrokerTLS.Cert.SecretRef) {
		// The /opt/kafka/connect-certs is fixed path within Connectors
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "cert",
			MountPath: "/opt/kafka/connect-certs/cert",
		})
	}

	if !helmette.Empty(values.Connectors.BrokerTLS.Key.SecretRef) {
		// The /opt/kafka/connect-certs is fixed path within Connectors
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "key",
			MountPath: "/opt/kafka/connect-certs/key",
		})
	}

	return append(mounts, values.Storage.VolumeMounts...)
}
