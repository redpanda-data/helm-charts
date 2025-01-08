// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0
//go:build !generate

// +gotohelm:ignore=true
//
// Code generated by genpartial DO NOT EDIT.
package redpanda

import (
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/redpanda-operator/charts/connectors"
	"github.com/redpanda-data/redpanda-operator/charts/console"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type PartialValues struct {
	NameOverride     *string                       "json:\"nameOverride,omitempty\""
	FullnameOverride *string                       "json:\"fullnameOverride,omitempty\""
	ClusterDomain    *string                       "json:\"clusterDomain,omitempty\""
	CommonLabels     map[string]string             "json:\"commonLabels,omitempty\""
	NodeSelector     map[string]string             "json:\"nodeSelector,omitempty\""
	Affinity         *corev1.Affinity              "json:\"affinity,omitempty\" jsonschema:\"required\""
	Tolerations      []corev1.Toleration           "json:\"tolerations,omitempty\""
	Image            *PartialImage                 "json:\"image,omitempty\" jsonschema:\"required,description=Values used to define the container image to be used for Redpanda\""
	Service          *PartialService               "json:\"service,omitempty\""
	ImagePullSecrets []corev1.LocalObjectReference "json:\"imagePullSecrets,omitempty\""
	LicenseKey       *string                       "json:\"license_key,omitempty\" jsonschema:\"deprecated,pattern=^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?\\\\.(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$|^$\""
	LicenseSecretRef *PartialLicenseSecretRef      "json:\"license_secret_ref,omitempty\" jsonschema:\"deprecated\""
	AuditLogging     *PartialAuditLogging          "json:\"auditLogging,omitempty\""
	Enterprise       *PartialEnterprise            "json:\"enterprise,omitempty\""
	RackAwareness    *PartialRackAwareness         "json:\"rackAwareness,omitempty\""
	Console          *console.PartialValues        "json:\"console,omitempty\""
	Connectors       *connectors.PartialValues     "json:\"connectors,omitempty\""
	Auth             *PartialAuth                  "json:\"auth,omitempty\""
	TLS              *PartialTLS                   "json:\"tls,omitempty\""
	External         *PartialExternalConfig        "json:\"external,omitempty\""
	Logging          *PartialLogging               "json:\"logging,omitempty\""
	Monitoring       *PartialMonitoring            "json:\"monitoring,omitempty\""
	Resources        *PartialRedpandaResources     "json:\"resources,omitempty\""
	Storage          *PartialStorage               "json:\"storage,omitempty\""
	PostInstallJob   *PartialPostInstallJob        "json:\"post_install_job,omitempty\""
	Statefulset      *PartialStatefulset           "json:\"statefulset,omitempty\""
	ServiceAccount   *PartialServiceAccountCfg     "json:\"serviceAccount,omitempty\""
	RBAC             *PartialRBAC                  "json:\"rbac,omitempty\""
	Tuning           *PartialTuning                "json:\"tuning,omitempty\""
	Listeners        *PartialListeners             "json:\"listeners,omitempty\""
	Config           *PartialConfig                "json:\"config,omitempty\""
	Tests            *struct {
		Enabled *bool "json:\"enabled,omitempty\""
	} "json:\"tests,omitempty\""
	Force *bool "json:\"force,omitempty\""
}

type PartialImage struct {
	Repository *string   "json:\"repository,omitempty\" jsonschema:\"required,default=docker.redpanda.com/redpandadata/redpanda\""
	Tag        *ImageTag "json:\"tag,omitempty\" jsonschema:\"default=Chart.appVersion\""
	PullPolicy *string   "json:\"pullPolicy,omitempty\" jsonschema:\"required,pattern=^(Always|Never|IfNotPresent)$,description=The Kubernetes Pod image pull policy.\""
}

type PartialAuditLogging struct {
	Enabled                    *bool    "json:\"enabled,omitempty\""
	Listener                   *string  "json:\"listener,omitempty\""
	Partitions                 *int     "json:\"partitions,omitempty\""
	EnabledEventTypes          []string "json:\"enabledEventTypes,omitempty\""
	ExcludedTopics             []string "json:\"excludedTopics,omitempty\""
	ExcludedPrincipals         []string "json:\"excludedPrincipals,omitempty\""
	ClientMaxBufferSize        *int     "json:\"clientMaxBufferSize,omitempty\""
	QueueDrainIntervalMS       *int     "json:\"queueDrainIntervalMs,omitempty\""
	QueueMaxBufferSizeperShard *int     "json:\"queueMaxBufferSizePerShard,omitempty\""
	ReplicationFactor          *int     "json:\"replicationFactor,omitempty\""
}

type PartialEnterprise struct {
	License          *string "json:\"license,omitempty\""
	LicenseSecretRef *struct {
		Key  *string "json:\"key,omitempty\""
		Name *string "json:\"name,omitempty\""
	} "json:\"licenseSecretRef,omitempty\""
}

type PartialRackAwareness struct {
	Enabled        *bool   "json:\"enabled,omitempty\" jsonschema:\"required\""
	NodeAnnotation *string "json:\"nodeAnnotation,omitempty\" jsonschema:\"required\""
}

type PartialAuth struct {
	SASL *PartialSASLAuth "json:\"sasl,omitempty\" jsonschema:\"required\""
}

type PartialTLS struct {
	Enabled *bool             "json:\"enabled,omitempty\" jsonschema:\"required\""
	Certs   PartialTLSCertMap "json:\"certs,omitempty\" jsonschema:\"required\""
}

type PartialExternalConfig struct {
	Addresses      []string            "json:\"addresses,omitempty\""
	Annotations    map[string]string   "json:\"annotations,omitempty\""
	Domain         *string             "json:\"domain,omitempty\""
	Enabled        *bool               "json:\"enabled,omitempty\" jsonschema:\"required\""
	Type           *corev1.ServiceType "json:\"type,omitempty\" jsonschema:\"pattern=^(LoadBalancer|NodePort)$\""
	PrefixTemplate *string             "json:\"prefixTemplate,omitempty\""
	SourceRanges   []string            "json:\"sourceRanges,omitempty\""
	Service        *PartialEnableable  "json:\"service,omitempty\""
	ExternalDNS    *PartialEnableable  "json:\"externalDns,omitempty\""
}

type PartialLogging struct {
	LogLevel    *string "json:\"logLevel,omitempty\" jsonschema:\"required,pattern=^(error|warn|info|debug|trace)$\""
	UseageStats *struct {
		Enabled   *bool   "json:\"enabled,omitempty\" jsonschema:\"required\""
		ClusterID *string "json:\"clusterId,omitempty\""
	} "json:\"usageStats,omitempty\" jsonschema:\"required\""
}

type PartialMonitoring struct {
	Enabled        *bool                   "json:\"enabled,omitempty\" jsonschema:\"required\""
	ScrapeInterval *monitoringv1.Duration  "json:\"scrapeInterval,omitempty\" jsonschema:\"required\""
	Labels         map[string]string       "json:\"labels,omitempty\""
	TLSConfig      *monitoringv1.TLSConfig "json:\"tlsConfig,omitempty\""
	EnableHttp2    *bool                   "json:\"enableHttp2,omitempty\""
}

type PartialRedpandaResources struct {
	CPU *struct {
		Cores           *resource.Quantity "json:\"cores,omitempty\" jsonschema:\"required\""
		Overprovisioned *bool              "json:\"overprovisioned,omitempty\""
	} "json:\"cpu,omitempty\" jsonschema:\"required\""
	Memory *struct {
		EnableMemoryLocking *bool "json:\"enable_memory_locking,omitempty\""
		Container           *struct {
			Min *resource.Quantity "json:\"min,omitempty\""
			Max *resource.Quantity "json:\"max,omitempty\" jsonschema:\"required\""
		} "json:\"container,omitempty\" jsonschema:\"required\""
		Redpanda *struct {
			Memory        *resource.Quantity "json:\"memory,omitempty\""
			ReserveMemory *resource.Quantity "json:\"reserveMemory,omitempty\""
		} "json:\"redpanda,omitempty\""
	} "json:\"memory,omitempty\" jsonschema:\"required\""
}

type PartialStorage struct {
	HostPath         *string        "json:\"hostPath,omitempty\" jsonschema:\"required\""
	Tiered           *PartialTiered "json:\"tiered,omitempty\" jsonschema:\"required\""
	PersistentVolume *struct {
		Annotations   map[string]string  "json:\"annotations,omitempty\" jsonschema:\"required\""
		Enabled       *bool              "json:\"enabled,omitempty\" jsonschema:\"required\""
		Labels        map[string]string  "json:\"labels,omitempty\" jsonschema:\"required\""
		Size          *resource.Quantity "json:\"size,omitempty\" jsonschema:\"required\""
		StorageClass  *string            "json:\"storageClass,omitempty\" jsonschema:\"required\""
		NameOverwrite *string            "json:\"nameOverwrite,omitempty\""
	} "json:\"persistentVolume,omitempty\" jsonschema:\"required,deprecated\""
	TieredConfig                  PartialTieredStorageConfig "json:\"tieredConfig,omitempty\" jsonschema:\"deprecated\""
	TieredStorageHostPath         *string                    "json:\"tieredStorageHostPath,omitempty\" jsonschema:\"deprecated\""
	TieredStoragePersistentVolume *struct {
		Annotations  map[string]string "json:\"annotations,omitempty\" jsonschema:\"required\""
		Enabled      *bool             "json:\"enabled,omitempty\" jsonschema:\"required\""
		Labels       map[string]string "json:\"labels,omitempty\" jsonschema:\"required\""
		StorageClass *string           "json:\"storageClass,omitempty\" jsonschema:\"required\""
	} "json:\"tieredStoragePersistentVolume,omitempty\" jsonschema:\"deprecated\""
}

type PartialPostInstallJob struct {
	Resources       *corev1.ResourceRequirements "json:\"resources,omitempty\""
	Affinity        *corev1.Affinity             "json:\"affinity,omitempty\""
	Enabled         *bool                        "json:\"enabled,omitempty\""
	Labels          map[string]string            "json:\"labels,omitempty\""
	Annotations     map[string]string            "json:\"annotations,omitempty\""
	SecurityContext *corev1.SecurityContext      "json:\"securityContext,omitempty\""
	PodTemplate     *PartialPodTemplate          "json:\"podTemplate,omitempty\""
}

type PartialStatefulset struct {
	AdditionalSelectorLabels map[string]string "json:\"additionalSelectorLabels,omitempty\" jsonschema:\"required\""
	NodeAffinity             map[string]any    "json:\"nodeAffinity,omitempty\""
	Replicas                 *int32            "json:\"replicas,omitempty\" jsonschema:\"required\""
	UpdateStrategy           *struct {
		Type *string "json:\"type,omitempty\" jsonschema:\"required,pattern=^(RollingUpdate|OnDelete)$\""
	} "json:\"updateStrategy,omitempty\" jsonschema:\"required\""
	AdditionalRedpandaCmdFlags []string            "json:\"additionalRedpandaCmdFlags,omitempty\""
	Annotations                map[string]string   "json:\"annotations,omitempty\" jsonschema:\"deprecated\""
	PodTemplate                *PartialPodTemplate "json:\"podTemplate,omitempty\" jsonschema:\"required\""
	Budget                     *struct {
		MaxUnavailable *int32 "json:\"maxUnavailable,omitempty\" jsonschema:\"required\""
	} "json:\"budget,omitempty\" jsonschema:\"required\""
	StartupProbe *struct {
		InitialDelaySeconds *int32 "json:\"initialDelaySeconds,omitempty\" jsonschema:\"required\""
		FailureThreshold    *int32 "json:\"failureThreshold,omitempty\" jsonschema:\"required\""
		PeriodSeconds       *int32 "json:\"periodSeconds,omitempty\" jsonschema:\"required\""
	} "json:\"startupProbe,omitempty\" jsonschema:\"required\""
	LivenessProbe *struct {
		InitialDelaySeconds *int32 "json:\"initialDelaySeconds,omitempty\" jsonschema:\"required\""
		FailureThreshold    *int32 "json:\"failureThreshold,omitempty\" jsonschema:\"required\""
		PeriodSeconds       *int32 "json:\"periodSeconds,omitempty\" jsonschema:\"required\""
	} "json:\"livenessProbe,omitempty\" jsonschema:\"required\""
	ReadinessProbe *struct {
		InitialDelaySeconds *int32 "json:\"initialDelaySeconds,omitempty\" jsonschema:\"required\""
		FailureThreshold    *int32 "json:\"failureThreshold,omitempty\" jsonschema:\"required\""
		PeriodSeconds       *int32 "json:\"periodSeconds,omitempty\" jsonschema:\"required\""
		SuccessThreshold    *int32 "json:\"successThreshold,omitempty\""
		TimeoutSeconds      *int32 "json:\"timeoutSeconds,omitempty\""
	} "json:\"readinessProbe,omitempty\" jsonschema:\"required\""
	PodAffinity     map[string]any "json:\"podAffinity,omitempty\" jsonschema:\"required\""
	PodAntiAffinity *struct {
		TopologyKey *string        "json:\"topologyKey,omitempty\" jsonschema:\"required\""
		Type        *string        "json:\"type,omitempty\" jsonschema:\"required,pattern=^(hard|soft|custom)$\""
		Weight      *int32         "json:\"weight,omitempty\" jsonschema:\"required\""
		Custom      map[string]any "json:\"custom,omitempty\""
	} "json:\"podAntiAffinity,omitempty\" jsonschema:\"required\""
	NodeSelector                  map[string]string "json:\"nodeSelector,omitempty\" jsonschema:\"required\""
	PriorityClassName             *string           "json:\"priorityClassName,omitempty\" jsonschema:\"required\""
	TerminationGracePeriodSeconds *int64            "json:\"terminationGracePeriodSeconds,omitempty\""
	TopologySpreadConstraints     []struct {
		MaxSkew           *int32                                "json:\"maxSkew,omitempty\""
		TopologyKey       *string                               "json:\"topologyKey,omitempty\""
		WhenUnsatisfiable *corev1.UnsatisfiableConstraintAction "json:\"whenUnsatisfiable,omitempty\" jsonschema:\"pattern=^(ScheduleAnyway|DoNotSchedule)$\""
	} "json:\"topologySpreadConstraints,omitempty\" jsonschema:\"required,minItems=1\""
	Tolerations        []corev1.Toleration     "json:\"tolerations,omitempty\" jsonschema:\"required\""
	PodSecurityContext *PartialSecurityContext "json:\"podSecurityContext,omitempty\""
	SecurityContext    *PartialSecurityContext "json:\"securityContext,omitempty\" jsonschema:\"required\""
	SideCars           *struct {
		ConfigWatcher *struct {
			Enabled           *bool                   "json:\"enabled,omitempty\""
			ExtraVolumeMounts *string                 "json:\"extraVolumeMounts,omitempty\""
			Resources         map[string]any          "json:\"resources,omitempty\""
			SecurityContext   *corev1.SecurityContext "json:\"securityContext,omitempty\""
		} "json:\"configWatcher,omitempty\""
		Controllers *struct {
			Image *struct {
				Tag        *ImageTag "json:\"tag,omitempty\" jsonschema:\"required,default=Chart.appVersion\""
				Repository *string   "json:\"repository,omitempty\" jsonschema:\"required,default=docker.redpanda.com/redpandadata/redpanda-operator\""
			} "json:\"image,omitempty\""
			Enabled            *bool                   "json:\"enabled,omitempty\""
			CreateRBAC         *bool                   "json:\"createRBAC,omitempty\""
			Resources          any                     "json:\"resources,omitempty\""
			SecurityContext    *corev1.SecurityContext "json:\"securityContext,omitempty\""
			HealthProbeAddress *string                 "json:\"healthProbeAddress,omitempty\""
			MetricsAddress     *string                 "json:\"metricsAddress,omitempty\""
			PprofAddress       *string                 "json:\"pprofAddress,omitempty\""
			Run                []string                "json:\"run,omitempty\""
		} "json:\"controllers,omitempty\""
	} "json:\"sideCars,omitempty\" jsonschema:\"required\""
	ExtraVolumes      *string "json:\"extraVolumes,omitempty\""
	ExtraVolumeMounts *string "json:\"extraVolumeMounts,omitempty\""
	InitContainers    *struct {
		Configurator *struct {
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
		} "json:\"configurator,omitempty\""
		FSValidator *struct {
			Enabled           *bool          "json:\"enabled,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
			ExpectedFS        *string        "json:\"expectedFS,omitempty\""
		} "json:\"fsValidator,omitempty\""
		SetDataDirOwnership *struct {
			Enabled           *bool          "json:\"enabled,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"setDataDirOwnership,omitempty\""
		SetTieredStorageCacheDirOwnership *struct {
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"setTieredStorageCacheDirOwnership,omitempty\""
		Tuning *struct {
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"tuning,omitempty\""
		ExtraInitContainers *string "json:\"extraInitContainers,omitempty\""
	} "json:\"initContainers,omitempty\""
	InitContainerImage *struct {
		Repository *string "json:\"repository,omitempty\""
		Tag        *string "json:\"tag,omitempty\""
	} "json:\"initContainerImage,omitempty\""
}

type PartialServiceAccountCfg struct {
	Annotations                  map[string]string "json:\"annotations,omitempty\" jsonschema:\"required\""
	AutomountServiceAccountToken *bool             "json:\"automountServiceAccountToken,omitempty\""
	Create                       *bool             "json:\"create,omitempty\" jsonschema:\"required\""
	Name                         *string           "json:\"name,omitempty\" jsonschema:\"required\""
}

type PartialRBAC struct {
	Enabled     *bool             "json:\"enabled,omitempty\" jsonschema:\"required\""
	Annotations map[string]string "json:\"annotations,omitempty\" jsonschema:\"required\""
}

type PartialTuning struct {
	TuneAIOEvents   *bool   "json:\"tune_aio_events,omitempty\""
	TuneClocksource *bool   "json:\"tune_clocksource,omitempty\""
	TuneBallastFile *bool   "json:\"tune_ballast_file,omitempty\""
	BallastFilePath *string "json:\"ballast_file_path,omitempty\""
	BallastFileSize *string "json:\"ballast_file_size,omitempty\""
	WellKnownIO     *string "json:\"well_known_io,omitempty\""
}

type PartialListeners struct {
	Admin          *PartialAdminListeners          "json:\"admin,omitempty\" jsonschema:\"required\""
	HTTP           *PartialHTTPListeners           "json:\"http,omitempty\" jsonschema:\"required\""
	Kafka          *PartialKafkaListeners          "json:\"kafka,omitempty\" jsonschema:\"required\""
	SchemaRegistry *PartialSchemaRegistryListeners "json:\"schemaRegistry,omitempty\" jsonschema:\"required\""
	RPC            *struct {
		Port *int32              "json:\"port,omitempty\" jsonschema:\"required\""
		TLS  *PartialInternalTLS "json:\"tls,omitempty\" jsonschema:\"required\""
	} "json:\"rpc,omitempty\" jsonschema:\"required\""
}

type PartialConfig struct {
	Cluster              PartialClusterConfig         "json:\"cluster,omitempty\" jsonschema:\"required\""
	Node                 PartialNodeConfig            "json:\"node,omitempty\" jsonschema:\"required\""
	RPK                  map[string]any               "json:\"rpk,omitempty\""
	SchemaRegistryClient *PartialSchemaRegistryClient "json:\"schema_registry_client,omitempty\""
	PandaProxyClient     *PartialPandaProxyClient     "json:\"pandaproxy_client,omitempty\""
	Tunable              PartialTunableConfig         "json:\"tunable,omitempty\" jsonschema:\"required\""
}

type PartialService struct {
	Name     *string "json:\"name,omitempty\""
	Internal *struct {
		Annotations map[string]string "json:\"annotations,omitempty\""
	} "json:\"internal,omitempty\""
}

type PartialLicenseSecretRef struct {
	SecretName *string "json:\"secret_name,omitempty\""
	SecretKey  *string "json:\"secret_key,omitempty\""
}

type PartialTLSCertMap map[string]PartialTLSCert

type PartialEnableable struct {
	Enabled *bool "json:\"enabled,omitempty\" jsonschema:\"required\""
}

type PartialTiered struct {
	CredentialsSecretRef *PartialTieredStorageCredentials "json:\"credentialsSecretRef,omitempty\""
	Config               PartialTieredStorageConfig       "json:\"config,omitempty\""
	HostPath             *string                          "json:\"hostPath,omitempty\""
	MountType            *string                          "json:\"mountType,omitempty\" jsonschema:\"required,pattern=^(none|hostPath|emptyDir|persistentVolume)$\""
	PersistentVolume     *struct {
		Annotations   map[string]string "json:\"annotations,omitempty\" jsonschema:\"required\""
		Enabled       *bool             "json:\"enabled,omitempty\""
		Labels        map[string]string "json:\"labels,omitempty\" jsonschema:\"required\""
		NameOverwrite *string           "json:\"nameOverwrite,omitempty\""
		Size          *string           "json:\"size,omitempty\""
		StorageClass  *string           "json:\"storageClass,omitempty\" jsonschema:\"required\""
	} "json:\"persistentVolume,omitempty\""
}

type PartialTieredStorageConfig map[string]any

type PartialPodTemplate struct {
	Labels      map[string]string                      "json:\"labels,omitempty\" jsonschema:\"required\""
	Annotations map[string]string                      "json:\"annotations,omitempty\" jsonschema:\"required\""
	Spec        *applycorev1.PodSpecApplyConfiguration "json:\"spec,omitempty\""
}

type PartialSecurityContext struct {
	RunAsUser                 *int64                         "json:\"runAsUser,omitempty\""
	RunAsGroup                *int64                         "json:\"runAsGroup,omitempty\""
	AllowPrivilegeEscalation  *bool                          "json:\"allowPrivilegeEscalation,omitempty\""
	AllowPriviledgeEscalation *bool                          "json:\"allowPriviledgeEscalation,omitempty\""
	RunAsNonRoot              *bool                          "json:\"runAsNonRoot,omitempty\""
	FSGroup                   *int64                         "json:\"fsGroup,omitempty\""
	FSGroupChangePolicy       *corev1.PodFSGroupChangePolicy "json:\"fsGroupChangePolicy,omitempty\""
}

type PartialAdminListeners struct {
	External    PartialExternalListeners[PartialAdminExternal] "json:\"external,omitempty\""
	Port        *int32                                         "json:\"port,omitempty\" jsonschema:\"required\""
	AppProtocol *string                                        "json:\"appProtocol,omitempty\""
	TLS         *PartialInternalTLS                            "json:\"tls,omitempty\" jsonschema:\"required\""
}

type PartialHTTPListeners struct {
	Enabled              *bool                                         "json:\"enabled,omitempty\" jsonschema:\"required\""
	External             PartialExternalListeners[PartialHTTPExternal] "json:\"external,omitempty\""
	AuthenticationMethod *HTTPAuthenticationMethod                     "json:\"authenticationMethod,omitempty\""
	TLS                  *PartialInternalTLS                           "json:\"tls,omitempty\" jsonschema:\"required\""
	KafkaEndpoint        *string                                       "json:\"kafkaEndpoint,omitempty\" jsonschema:\"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$\""
	Port                 *int32                                        "json:\"port,omitempty\" jsonschema:\"required\""
}

type PartialKafkaListeners struct {
	AuthenticationMethod *KafkaAuthenticationMethod                     "json:\"authenticationMethod,omitempty\""
	External             PartialExternalListeners[PartialKafkaExternal] "json:\"external,omitempty\""
	TLS                  *PartialInternalTLS                            "json:\"tls,omitempty\" jsonschema:\"required\""
	Port                 *int32                                         "json:\"port,omitempty\" jsonschema:\"required\""
}

type PartialSchemaRegistryListeners struct {
	Enabled              *bool                                                   "json:\"enabled,omitempty\" jsonschema:\"required\""
	External             PartialExternalListeners[PartialSchemaRegistryExternal] "json:\"external,omitempty\""
	AuthenticationMethod *HTTPAuthenticationMethod                               "json:\"authenticationMethod,omitempty\""
	KafkaEndpoint        *string                                                 "json:\"kafkaEndpoint,omitempty\" jsonschema:\"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$\""
	Port                 *int32                                                  "json:\"port,omitempty\" jsonschema:\"required\""
	TLS                  *PartialInternalTLS                                     "json:\"tls,omitempty\" jsonschema:\"required\""
}

type PartialClusterConfig map[string]any

type PartialNodeConfig map[string]any

type PartialTunableConfig map[string]any

type PartialSASLAuth struct {
	Enabled       *bool                 "json:\"enabled,omitempty\" jsonschema:\"required\""
	Mechanism     *string               "json:\"mechanism,omitempty\""
	SecretRef     *string               "json:\"secretRef,omitempty\""
	Users         []PartialSASLUser     "json:\"users,omitempty\""
	BootstrapUser *PartialBootstrapUser "json:\"bootstrapUser,omitempty\""
}

type PartialInternalTLS struct {
	Enabled           *bool              "json:\"enabled,omitempty\""
	Cert              *string            "json:\"cert,omitempty\" jsonschema:\"required\""
	RequireClientAuth *bool              "json:\"requireClientAuth,omitempty\" jsonschema:\"required\""
	TrustStore        *PartialTrustStore "json:\"trustStore,omitempty\""
}

type PartialSchemaRegistryClient struct {
	Retries                     *int "json:\"retries,omitempty\""
	RetryBaseBackoffMS          *int "json:\"retry_base_backoff_ms,omitempty\""
	ProduceBatchRecordCount     *int "json:\"produce_batch_record_count,omitempty\""
	ProduceBatchSizeBytes       *int "json:\"produce_batch_size_bytes,omitempty\""
	ProduceBatchDelayMS         *int "json:\"produce_batch_delay_ms,omitempty\""
	ConsumerRequestTimeoutMS    *int "json:\"consumer_request_timeout_ms,omitempty\""
	ConsumerRequestMaxBytes     *int "json:\"consumer_request_max_bytes,omitempty\""
	ConsumerSessionTimeoutMS    *int "json:\"consumer_session_timeout_ms,omitempty\""
	ConsumerRebalanceTimeoutMS  *int "json:\"consumer_rebalance_timeout_ms,omitempty\""
	ConsumerHeartbeatIntervalMS *int "json:\"consumer_heartbeat_interval_ms,omitempty\""
}

type PartialPandaProxyClient struct {
	Retries                     *int "json:\"retries,omitempty\""
	RetryBaseBackoffMS          *int "json:\"retry_base_backoff_ms,omitempty\""
	ProduceBatchRecordCount     *int "json:\"produce_batch_record_count,omitempty\""
	ProduceBatchSizeBytes       *int "json:\"produce_batch_size_bytes,omitempty\""
	ProduceBatchDelayMS         *int "json:\"produce_batch_delay_ms,omitempty\""
	ConsumerRequestTimeoutMS    *int "json:\"consumer_request_timeout_ms,omitempty\""
	ConsumerRequestMaxBytes     *int "json:\"consumer_request_max_bytes,omitempty\""
	ConsumerSessionTimeoutMS    *int "json:\"consumer_session_timeout_ms,omitempty\""
	ConsumerRebalanceTimeoutMS  *int "json:\"consumer_rebalance_timeout_ms,omitempty\""
	ConsumerHeartbeatIntervalMS *int "json:\"consumer_heartbeat_interval_ms,omitempty\""
}

type PartialTLSCert struct {
	Enabled               *bool                        "json:\"enabled,omitempty\""
	CAEnabled             *bool                        "json:\"caEnabled,omitempty\" jsonschema:\"required\""
	ApplyInternalDNSNames *bool                        "json:\"applyInternalDNSNames,omitempty\""
	Duration              *string                      "json:\"duration,omitempty\" jsonschema:\"pattern=.*[smh]$\""
	IssuerRef             *cmmeta.ObjectReference      "json:\"issuerRef,omitempty\""
	SecretRef             *corev1.LocalObjectReference "json:\"secretRef,omitempty\""
	ClientSecretRef       *corev1.LocalObjectReference "json:\"clientSecretRef,omitempty\""
}

type PartialTieredStorageCredentials struct {
	AccessKey *PartialSecretRef "json:\"accessKey,omitempty\""
	SecretKey *PartialSecretRef "json:\"secretKey,omitempty\""
}

type PartialBootstrapUser struct {
	Name         *string                   "json:\"name,omitempty\""
	SecretKeyRef *corev1.SecretKeySelector "json:\"secretKeyRef,omitempty\""
	Password     *string                   "json:\"password,omitempty\""
	Mechanism    *string                   "json:\"mechanism,omitempty\" jsonschema:\"pattern=^(SCRAM-SHA-512|SCRAM-SHA-256)$\""
}

type PartialExternalListeners[T any] map[string]T

type PartialAdminExternal struct {
	Enabled         *bool               "json:\"enabled,omitempty\""
	AdvertisedPorts []int32             "json:\"advertisedPorts,omitempty\" jsonschema:\"minItems=1\""
	Port            *int32              "json:\"port,omitempty\" jsonschema:\"required\""
	NodePort        *int32              "json:\"nodePort,omitempty\""
	TLS             *PartialExternalTLS "json:\"tls,omitempty\""
}

type PartialHTTPExternal struct {
	Enabled              *bool                     "json:\"enabled,omitempty\""
	AdvertisedPorts      []int32                   "json:\"advertisedPorts,omitempty\" jsonschema:\"minItems=1\""
	Port                 *int32                    "json:\"port,omitempty\" jsonschema:\"required\""
	NodePort             *int32                    "json:\"nodePort,omitempty\""
	AuthenticationMethod *HTTPAuthenticationMethod "json:\"authenticationMethod,omitempty\""
	PrefixTemplate       *string                   "json:\"prefixTemplate,omitempty\""
	TLS                  *PartialExternalTLS       "json:\"tls,omitempty\" jsonschema:\"required\""
}

type PartialKafkaExternal struct {
	Enabled              *bool                      "json:\"enabled,omitempty\""
	AdvertisedPorts      []int32                    "json:\"advertisedPorts,omitempty\" jsonschema:\"minItems=1\""
	Port                 *int32                     "json:\"port,omitempty\" jsonschema:\"required\""
	NodePort             *int32                     "json:\"nodePort,omitempty\""
	AuthenticationMethod *KafkaAuthenticationMethod "json:\"authenticationMethod,omitempty\""
	PrefixTemplate       *string                    "json:\"prefixTemplate,omitempty\""
	TLS                  *PartialExternalTLS        "json:\"tls,omitempty\""
}

type PartialSchemaRegistryExternal struct {
	Enabled              *bool                     "json:\"enabled,omitempty\""
	AdvertisedPorts      []int32                   "json:\"advertisedPorts,omitempty\" jsonschema:\"minItems=1\""
	Port                 *int32                    "json:\"port,omitempty\""
	NodePort             *int32                    "json:\"nodePort,omitempty\""
	AuthenticationMethod *HTTPAuthenticationMethod "json:\"authenticationMethod,omitempty\""
	TLS                  *PartialExternalTLS       "json:\"tls,omitempty\""
}

type PartialSASLUser struct {
	Name      *string "json:\"name,omitempty\""
	Password  *string "json:\"password,omitempty\""
	Mechanism *string "json:\"mechanism,omitempty\" jsonschema:\"pattern=^(SCRAM-SHA-512|SCRAM-SHA-256)$\""
}

type PartialTrustStore struct {
	ConfigMapKeyRef *corev1.ConfigMapKeySelector "json:\"configMapKeyRef,omitempty\""
	SecretKeyRef    *corev1.SecretKeySelector    "json:\"secretKeyRef,omitempty\""
}

type PartialSecretRef struct {
	ConfigurationKey *string "json:\"configurationKey,omitempty\""
	Key              *string "json:\"key,omitempty\""
	Name             *string "json:\"name,omitempty\""
}

type PartialExternalTLS struct {
	Enabled           *bool              "json:\"enabled,omitempty\""
	Cert              *string            "json:\"cert,omitempty\""
	RequireClientAuth *bool              "json:\"requireClientAuth,omitempty\""
	TrustStore        *PartialTrustStore "json:\"trustStore,omitempty\""
}
