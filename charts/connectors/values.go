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
// +gotohelm:filename=_values.go.tpl
package connectors

import (
	_ "embed"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:embed values.yaml
var DefaultValuesYAML []byte

type Values struct {
	NameOverride     string                        `json:"nameOverride"`
	FullnameOverride string                        `json:"fullnameOverride"`
	CommonLabels     map[string]string             `json:"commonLabels"`
	Tolerations      []corev1.Toleration           `json:"tolerations"`
	Image            Image                         `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
	Test             Creatable                     `json:"test"`
	Connectors       Connectors                    `json:"connectors"`
	Auth             Auth                          `json:"auth"`
	Logging          Logging                       `json:"logging"`
	Monitoring       MonitoringConfig              `json:"monitoring"`
	Container        Container                     `json:"container"`
	Deployment       DeploymentConfig              `json:"deployment"`
	Storage          Storage                       `json:"storage"`
	ServiceAccount   ServiceAccountConfig          `json:"serviceAccount"`
	Service          ServiceConfig                 `json:"service"`
	Enabled          *bool                         `json:"enabled,omitempty"`
}

type Image struct {
	Repository string            `json:"repository"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
	Tag        string            `json:"tag"`
}

type Connectors struct {
	RestPort                int32             `json:"restPort"`
	BootstrapServers        string            `json:"bootstrapServers"`
	SchemaRegistryURL       string            `json:"schemaRegistryURL"`
	AdditionalConfiguration string            `json:"additionalConfiguration"`
	SecretManager           SecretManager     `json:"secretManager"`
	ProducerBatchSize       int32             `json:"producerBatchSize"`
	ProducerLingerMS        int32             `json:"producerLingerMS"`
	Storage                 ConnectorsStorage `json:"storage"`
	GroupID                 string            `json:"groupID"`
	BrokerTLS               TLS               `json:"brokerTLS"`
}

type SecretManager struct {
	Enabled          bool   `json:"enabled"`
	Region           string `json:"region"`
	ConsolePrefix    string `json:"consolePrefix"`
	ConnectorsPrefix string `json:"connectorsPrefix"`
}

type ConnectorsStorage struct {
	ReplicationFactor struct {
		Offset int32 `json:"offset"`
		Config int32 `json:"config"`
		Status int32 `json:"status"`
	} `json:"replicationFactor"`
	Remote struct {
		Read struct {
			Offset bool `json:"offset"`
			Config bool `json:"config"`
			Status bool `json:"status"`
		} `json:"read"`
		Write struct {
			Offset bool `json:"offset"`
			Config bool `json:"config"`
			Status bool `json:"status"`
		} `json:"write"`
	} `json:"remote"`
	Topic struct {
		Offset string `json:"offset"`
		Config string `json:"config"`
		Status string `json:"status"`
	} `json:"topic"`
}

type TLS struct {
	Enabled bool `json:"enabled"`
	CA      struct {
		SecretRef           string `json:"secretRef"`
		SecretNameOverwrite string `json:"secretNameOverwrite"`
	} `json:"ca"`
	Cert struct {
		SecretRef           string `json:"secretRef"`
		SecretNameOverwrite string `json:"secretNameOverwrite"`
	} `json:"cert"`
	Key struct {
		SecretRef           string `json:"secretRef"`
		SecretNameOverwrite string `json:"secretNameOverwrite"`
	} `json:"key"`
}

type Auth struct {
	SASL struct {
		Enabled   bool   `json:"enabled"`
		Mechanism string `json:"mechanism"`
		SecretRef string `json:"secretRef"`
		UserName  string `json:"userName"`
	} `json:"sasl"`
}

func (c *Auth) SASLEnabled() bool {
	saslEnabled := !helmette.Empty(c.SASL.UserName)
	saslEnabled = saslEnabled && !helmette.Empty(c.SASL.Mechanism)
	saslEnabled = saslEnabled && !helmette.Empty(c.SASL.SecretRef)
	return saslEnabled
}

type Logging struct {
	Level string `json:"level"`
}

type MonitoringConfig struct {
	Enabled           bool                           `json:"enabled"`
	ScrapeInterval    monitoringv1.Duration          `json:"scrapeInterval"`
	Labels            map[string]string              `json:"labels"`
	Annotations       map[string]string              `json:"annotations"`
	NamespaceSelector monitoringv1.NamespaceSelector `json:"namespaceSelector"`
}

type Container struct {
	SecurityContext corev1.SecurityContext `json:"securityContext"`
	Resources       struct {
		Request         corev1.ResourceList `json:"request"`
		Limits          corev1.ResourceList `json:"limits"`
		JavaMaxHeapSize *resource.Quantity  `json:"javaMaxHeapSize"`
	} `json:"resources"`
	JavaGCLogEnabled string `json:"javaGCLogEnabled"` // XXX ugh - it ends up as an env var
}

type DeploymentConfig struct {
	Replicas      *int32                    `json:"replicas,omitempty"`
	Create        bool                      `json:"create"`
	Command       []string                  `json:"command,omitempty"`
	Strategy      appsv1.DeploymentStrategy `json:"strategy,omitempty"`
	SchedulerName string                    `json:"schedulerName"`
	Budget        struct {
		MaxUnavailable int32 `json:"maxUnavailable"`
	} `json:"budget"`
	Annotations             map[string]string      `json:"annotations"`
	LivenessProbe           *corev1.Probe          `json:"livenessProbe,omitempty"`
	ReadinessProbe          *corev1.Probe          `json:"readinessProbe,omitempty"`
	ExtraEnv                []corev1.EnvVar        `json:"extraEnv"`
	ExtraEnvFrom            []corev1.EnvFromSource `json:"extraEnvFrom"`
	ProgressDeadlineSeconds int32                  `json:"progressDeadlineSeconds"`
	RevisionHistoryLimit    *int32                 `json:"revisionHistoryLimit,omitempty"`
	PodAffinity             *corev1.PodAffinity    `json:"podAffinity,omitempty"`
	NodeAffinity            *corev1.NodeAffinity   `json:"nodeAffinity,omitempty"`
	PodAntiAffinity         *struct {
		TopologyKey string                  `json:"topologyKey"`
		Type        string                  `json:"type"`
		Weight      *int32                  `json:"weight,omitempty"`
		Custom      *corev1.PodAntiAffinity `json:"custom,omitempty"`
	} `json:"podAntiAffinity,omitempty"`
	NodeSelector                  map[string]string                 `json:"nodeSelector"`
	PriorityClassName             *string                           `json:"priorityClassName,omitempty"` // XXX uused  in original template
	Tolerations                   []corev1.Toleration               `json:"tolerations"`
	TopologySpreadConstraints     []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	SecurityContext               *corev1.PodSecurityContext        `json:"securityContext,omitempty"`
	TerminationGracePeriodSeconds *int64                            `json:"terminationGracePeriodSeconds,omitempty"`
	RestartPolicy                 corev1.RestartPolicy              `json:"restartPolicy"`
}

type Storage struct {
	Volume       []corev1.Volume      `json:"volume"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts"`
}

type ServiceAccountConfig struct {
	Annotations                  map[string]string `json:"annotations"`
	AutomountServiceAccountToken *bool             `json:"automountServiceAccountToken,omitempty"`
	Create                       bool              `json:"create"`
	Name                         string            `json:"name"`
}

type ServiceConfig struct {
	Annotations map[string]string `json:"annotations"`
	Name        string            `json:"name"`
	Ports       []struct {
		Name string `json:"name"`
		Port int32  `json:"port"`
	} `json:"ports"`
}

type Creatable struct {
	Create bool `json:"create"`
}
