// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_values.go.tpl
package operator

import (
	_ "embed"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	Namespace = OperatorScope("Namespace")
	Cluster   = OperatorScope("Cluster")
)

var (
	//go:embed values.yaml
	DefaultValuesYAML []byte

	//go:embed values.schema.json
	ValuesSchemaJSON []byte
)

type Values struct {
	NameOverride       string                        `json:"nameOverride"`
	FullnameOverride   string                        `json:"fullnameOverride"`
	ReplicaCount       int32                         `json:"replicaCount"`
	ClusterDomain      string                        `json:"clusterDomain"`
	Image              Image                         `json:"image"`
	KubeRBACProxy      KubeRBACProxyConfig           `json:"kubeRbacProxy"`
	Configurator       Image                         `json:"configurator"`
	Config             Config                        `json:"config"`
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets"`
	LogLevel           string                        `json:"logLevel"`
	RBAC               RBAC                          `json:"rbac"`
	Webhook            Webhook                       `json:"webhook"`
	ServiceAccount     ServiceAccountConfig          `json:"serviceAccount"`
	Resources          corev1.ResourceRequirements   `json:"resources"`
	NodeSelector       map[string]string             `json:"nodeSelector"`
	Tolerations        []corev1.Toleration           `json:"tolerations"`
	Affinity           *corev1.Affinity              `json:"affinity" jsonschema:"deprecated"`
	Strategy           *appsv1.DeploymentStrategy    `json:"strategy,omitempty"`
	Annotations        map[string]string             `json:"annotations,omitempty"`
	PodAnnotations     map[string]string             `json:"podAnnotations"`
	PodLabels          map[string]string             `json:"podLabels"`
	AdditionalCmdFlags []string                      `json:"additionalCmdFlags"`
	CommonLabels       map[string]string             `json:"commonLabels"`
	Monitoring         MonitoringConfig              `json:"monitoring"`
	WebhookSecretName  string                        `json:"webhookSecretName"`
	PodTemplate        *PodTemplateSpec              `json:"podTemplate,omitempty"`
	LivenessProbe      *corev1.Probe                 `json:"livenessProbe,omitempty"`
	ReadinessProbe     *corev1.Probe                 `json:"readinessProbe,omitempty"`
	Scope              OperatorScope                 `json:"scope" jsonschema:"required,pattern=^(Namespace|Cluster)$,description=Sets the scope of the Redpanda Operator."`
}

type PodTemplateSpec struct {
	Metadata Metadata       `json:"metadata,omitempty"`
	Spec     corev1.PodSpec `json:"spec,omitempty" jsonschema:"required"`
}

type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type OperatorScope string

type Image struct {
	Repository string            `json:"repository"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy" jsonschema:"required,pattern=^(Always|Never|IfNotPresent)$,description=The Kubernetes Pod image pull policy."`
	Tag        *string           `json:"tag,omitempty"`
}

type KubeRBACProxyConfig struct {
	LogLevel int   `json:"logLevel"`
	Image    Image `json:"image"`
}

type Config struct {
	//nolint:stylecheck
	ApiVersion     string               `json:"apiVersion"`
	Kind           string               `json:"kind"`
	Health         HealthConfig         `json:"health"`
	Metrics        MetricsConfig        `json:"metrics"`
	Webhook        WebhookConfig        `json:"webhook"`
	LeaderElection LeaderElectionConfig `json:"leaderElection"`
}

type HealthConfig struct {
	HealthProbeBindAddress string `json:"healthProbeBindAddress"`
}

type MetricsConfig struct {
	BindAddress string `json:"bindAddress"`
}

type WebhookConfig struct {
	Port int `json:"port"`
}

type LeaderElectionConfig struct {
	LeaderElect  bool   `json:"leaderElect"`
	ResourceName string `json:"resourceName"`
}

type RBAC struct {
	Create                        bool `json:"create"`
	CreateAdditionalControllerCRs bool `json:"createAdditionalControllerCRs"`
	CreateRPKBundleCRs            bool `json:"createRPKBundleCRs"`
}

type Webhook struct {
	Enabled bool `json:"enabled"`
}

type ServiceAccountConfig struct {
	Annotations                  map[string]string `json:"annotations,omitempty"`
	AutomountServiceAccountToken *bool             `json:"automountServiceAccountToken,omitempty"`
	Create                       bool              `json:"create"`
	Name                         *string           `json:"name,omitempty"`
}

type MonitoringConfig struct {
	Enabled                   bool `json:"enabled"`
	DeployPrometheusKubeStack bool `json:"deployPrometheusKubeStack"`
}
