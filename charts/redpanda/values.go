// +gotohelm:filename=values.go.tpl
package redpanda

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/invopop/jsonschema"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

const (
	fiveGiB = 5368709120
	// That default path inside Redpanda container which is based on debian.
	defaultTruststorePath = "/etc/ssl/certs/ca-certificates.crt"
)

// values.go contains a collection of go structs that (loosely) map to
// values.yaml and are used for generating values.schema.json. Commented out
// struct fields are fields that are valid in the eyes of values.yaml but are
// not present in the hand written jsonschema. While the migration to a
// generated jsonschema is underway, there will be a variety of hacks,
// one-offs, and anonymous structs all aimed at minimizing the diff between the
// handwritten schema and the now generated one. Over time these oddities will
// be smoothed out and removed. Eventually, values.yaml will be generated from
// the Values struct as well to ensure that nothing can ever get out of sync.

type Values struct {
	NameOverride     string                        `json:"nameOverride"`
	FullnameOverride string                        `json:"fullnameOverride"`
	ClusterDomain    string                        `json:"clusterDomain"`
	CommonLabels     map[string]string             `json:"commonLabels"`
	NodeSelector     map[string]string             `json:"nodeSelector"`
	Affinity         corev1.Affinity               `json:"affinity" jsonschema:"required"`
	Tolerations      []corev1.Toleration           `json:"tolerations"`
	Image            Image                         `json:"image" jsonschema:"required,description=Values used to define the container image to be used for Redpanda"`
	Service          *Service                      `json:"service"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
	LicenseKey       string                        `json:"license_key" jsonschema:"deprecated,pattern=^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?\\.(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$|^$"`
	LicenseSecretRef *LicenseSecretRef             `json:"license_secret_ref" jsonschema:"deprecated"`
	AuditLogging     AuditLogging                  `json:"auditLogging"`
	Enterprise       Enterprise                    `json:"enterprise"`
	RackAwareness    RackAwareness                 `json:"rackAwareness"`
	// Console          any      `json:"console"`
	// Connectors       any      `json:"connectors"`
	Auth           Auth              `json:"auth"`
	TLS            TLS               `json:"tls"`
	External       ExternalConfig    `json:"external"`
	Logging        Logging           `json:"logging"`
	Monitoring     Monitoring        `json:"monitoring"`
	Resources      RedpandaResources `json:"resources"`
	Storage        Storage           `json:"storage"`
	PostInstallJob PostInstallJob    `json:"post_install_job"`
	PostUpgradeJob PostUpgradeJob    `json:"post_upgrade_job"`
	Statefulset    Statefulset       `json:"statefulset"`
	ServiceAccount ServiceAccountCfg `json:"serviceAccount"`
	RBAC           RBAC              `json:"rbac"`
	Tuning         Tuning            `json:"tuning"`
	Listeners      Listeners         `json:"listeners"`
	Config         Config            `json:"config"`
	Tests          *struct {
		Enabled bool `json:"enabled"`
	} `json:"tests"`
}

// +gotohelm:ignore=true
func (Values) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "license_key", "license_secret_ref")
}

// SecurityContext is a legacy mishmash of [corev1.PodSecurityContext] and
// [corev1.SecurityContext]. It's type exists for backwards compat purposes
// only.
type SecurityContext struct {
	RunAsUser                 *int64                         `json:"runAsUser"`
	RunAsGroup                *int64                         `json:"runAsGroup"`
	AllowPriviledgeEscalation *bool                          `json:"allowPriviledgeEscalation"`
	RunAsNonRoot              *bool                          `json:"runAsNonRoot"`
	FSGroup                   *int64                         `json:"fsGroup"`
	FSGroupChangePolicy       *corev1.PodFSGroupChangePolicy `json:"fsGroupChangePolicy"`
}

type Image struct {
	Repository string   `json:"repository" jsonschema:"required,default=docker.redpanda.com/redpandadata/redpanda"`
	Tag        ImageTag `json:"tag" jsonschema:"default=Chart.appVersion"`
	PullPolicy string   `json:"pullPolicy" jsonschema:"required,pattern=^(Always|Never|IfNotPresent)$,description=The Kubernetes Pod image pull policy."`
}

// +gotohelm:ignore=true
func (Image) JSONSchemaExtend(schema *jsonschema.Schema) {
	tag, _ := schema.Properties.Get("tag")
	repo, _ := schema.Properties.Get("repository")

	tag.Description = "The container image tag. Use the Redpanda release version. Must be a valid semver prefixed with a 'v'."
	repo.Description = "container image repository"
}

type Service struct {
	Name     *string `json:"name"`
	Internal struct {
		Annotations map[string]string `json:"annotations"`
	} `json:"internal"`
}

type LicenseSecretRef struct {
	SecretName string `json:"secret_name"`
	SecretKey  string `json:"secret_key"`
}

type AuditLogging struct {
	Enabled                    bool     `json:"enabled"`
	Listener                   string   `json:"listener"`
	Partitions                 int      `json:"partitions"`
	EnabledEventTypes          []string `json:"enabledEventTypes"`
	ExcludedTopics             []string `json:"excludedTopics"`
	ExcludedPrincipals         []string `json:"excludedPrincipals"`
	ClientMaxBufferSize        int      `json:"clientMaxBufferSize"`
	QueueDrainIntervalMS       int      `json:"queueDrainIntervalMs"`
	QueueMaxBufferSizeperShard int      `json:"queueMaxBufferSizePerShard"`
	ReplicationFactor          int      `json:"replicationFactor"`
}

// +gotohelm:ignore=true
func (AuditLogging) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "replicationFactor", "enabledEventTypes", "excludedPrincipals", "excludedTopics")
}

func (a *AuditLogging) Translate(dot *helmette.Dot, isSASLEnabled bool) map[string]any {
	result := map[string]any{}

	if !RedpandaAtLeast_23_3_0(dot) {
		return result
	}

	enabled := a.Enabled && isSASLEnabled
	result["audit_enabled"] = enabled
	if !enabled {
		return result
	}

	if int(a.ClientMaxBufferSize) != 16777216 {
		result["audit_client_max_buffer_size"] = a.ClientMaxBufferSize
	}

	if int(a.QueueDrainIntervalMS) != 500 {
		result["audit_queue_drain_interval_ms"] = a.QueueDrainIntervalMS
	}

	if int(a.QueueMaxBufferSizeperShard) != 1048576 {
		result["audit_queue_max_buffer_size_per_shard"] = a.QueueMaxBufferSizeperShard
	}

	if int(a.Partitions) != 12 {
		result["audit_log_num_partitions"] = a.Partitions
	}

	if a.ReplicationFactor != 0 {
		result["audit_log_replication_factor"] = a.ReplicationFactor
	}

	if len(a.EnabledEventTypes) > 0 {
		result["audit_enabled_event_types"] = a.EnabledEventTypes
	}

	if len(a.ExcludedTopics) > 0 {
		result["audit_excluded_topics"] = a.ExcludedTopics
	}

	if len(a.ExcludedPrincipals) > 0 {
		result["audit_excluded_principals"] = a.ExcludedPrincipals
	}

	return result
}

type Enterprise struct {
	License          string `json:"license"`
	LicenseSecretRef *struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"licenseSecretRef"`
}

type RackAwareness struct {
	Enabled        bool   `json:"enabled" jsonschema:"required"`
	NodeAnnotation string `json:"nodeAnnotation" jsonschema:"required"`
}

type Auth struct {
	SASL *SASLAuth `json:"sasl" jsonschema:"required"`
}

func (a *Auth) IsSASLEnabled() bool {
	if a.SASL == nil {
		return false
	}

	return a.SASL.Enabled
}

func (a *Auth) Translate(isSASLEnabled bool) map[string]any {
	if !isSASLEnabled {
		return nil
	}

	if len(a.SASL.Users) == 0 {
		return nil
	}

	users := []string{}
	for _, u := range a.SASL.Users {
		users = append(users, u.Name)
	}

	return map[string]any{
		"superusers": users,
	}
}

type TLS struct {
	Enabled bool       `json:"enabled" jsonschema:"required"`
	Certs   TLSCertMap `json:"certs" jsonschema:"required"`
}

type ExternalConfig struct {
	Addresses      []string           `json:"addresses"`
	Annotations    map[string]string  `json:"annotations"`
	Domain         *string            `json:"domain"`
	Enabled        bool               `json:"enabled" jsonschema:"required"`
	Type           corev1.ServiceType `json:"type" jsonschema:"pattern=^(LoadBalancer|NodePort)$"`
	PrefixTemplate string             `json:"prefixTemplate"`
	SourceRanges   []string           `json:"sourceRanges"`
	Service        Enableable         `json:"service"`
	ExternalDNS    *Enableable        `json:"externalDns"`
}

type Enableable struct {
	Enabled bool `json:"enabled" jsonschema:"required"`
}

type Logging struct {
	LogLevel    string `json:"logLevel" jsonschema:"required,pattern=^(error|warn|info|debug|trace)$"`
	UseageStats struct {
		Enabled   bool    `json:"enabled" jsonschema:"required"`
		ClusterID *string `json:"clusterId"`
	} `json:"usageStats" jsonschema:"required"`
}

func (l *Logging) Translate() map[string]any {
	result := map[string]any{}

	if clusterID := ptr.Deref(l.UseageStats.ClusterID, ""); clusterID != "" {
		result["cluster_id"] = clusterID
	}

	return result
}

type Monitoring struct {
	Enabled        bool                    `json:"enabled" jsonschema:"required"`
	ScrapeInterval monitoringv1.Duration   `json:"scrapeInterval" jsonschema:"required"`
	Labels         map[string]string       `json:"labels"`
	TLSConfig      *monitoringv1.TLSConfig `json:"tlsConfig"`
	EnableHttp2    *bool                   `json:"enableHttp2"`
}

type RedpandaResources struct {
	CPU struct {
		Cores           resource.Quantity `json:"cores" jsonschema:"required"`
		Overprovisioned *bool             `json:"overprovisioned"`
	} `json:"cpu" jsonschema:"required"`
	// Memory resources
	// For details,
	// see the [Pod resources documentation](https://docs.redpanda.com/docs/manage/kubernetes/manage-resources/#configure-memory-resources).
	Memory struct {
		// Enables memory locking.
		// For production, set to `true`.
		EnableMemoryLocking *bool `json:"enable_memory_locking"`
		// It is recommended to have at least 2Gi of memory per core for the Redpanda binary.
		// This memory is taken from the total memory given to each container.
		// The Helm chart allocates 80% of the container's memory to Redpanda, leaving the rest for
		// the Seastar subsystem (reserveMemory) and other container processes.
		// So at least 2.5Gi per core is recommended in order to ensure Redpanda has a full 2Gi.
		//
		// These values affect `--memory` and `--reserve-memory` flags passed to Redpanda and the memory
		// requests/limits in the StatefulSet.
		// Valid suffixes: k, M, G, T, P, E, Ki, Mi, Gi, Ti, Pi, Ei
		// Suffixes are defined as International System of units (http://physics.nist.gov/cuu/Units/binary.html).
		// To create `Guaranteed` Pod QoS for Redpanda brokers, provide both container max and min values for the container.
		// For details, see
		// https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/#create-a-pod-that-gets-assigned-a-qos-class-of-guaranteed
		// * Every container in the Pod must have a memory limit and a memory request.
		// * For every container in the Pod, the memory limit must equal the memory request.
		Container struct {
			// Minimum memory count for each Redpanda broker.
			// If omitted, the `min` value is equal to the `max` value (requested resources defaults to limits).
			// This setting is equivalent to `resources.requests.memory`.
			// For production, use 10Gi or greater.
			Min *resource.Quantity `json:"min"`
			// Maximum memory count for each Redpanda broker.
			// Equivalent to `resources.limits.memory`.
			// For production, use `10Gi` or greater.
			Max resource.Quantity `json:"max" jsonschema:"required"`
		} `json:"container" jsonschema:"required"`
		// This optional `redpanda` object allows you to specify the memory size for both the Redpanda
		// process and the underlying reserved memory used by Seastar.
		// This section is omitted by default, and memory sizes are calculated automatically
		// based on container memory.
		// Uncommenting this section and setting memory and reserveMemory values will disable
		// automatic calculation.
		//
		// If you are setting the following values manually, keep in mind the following guidelines.
		// Getting this wrong may lead to performance issues, instability, and loss of data:
		// The amount of memory to allocate to a container is determined by the sum of three values:
		// 1. Redpanda (at least 2Gi per core, ~80% of the container's total memory)
		// 2. Seastar subsystem (200Mi * 0.2% of the container's total memory, 200Mi < x < 1Gi)
		// 3. Other container processes (whatever small amount remains)
		Redpanda *struct {
			// Memory for the Redpanda process.
			// This must be lower than the container's memory (resources.memory.container.min if provided, otherwise
			// resources.memory.container.max).
			// Equivalent to --memory.
			// For production, use 8Gi or greater.
			Memory *resource.Quantity `json:"memory"`
			// Memory reserved for the Seastar subsystem.
			// Any value above 1Gi will provide diminishing performance benefits.
			// Equivalent to --reserve-memory.
			// For production, use 1Gi.
			ReserveMemory *resource.Quantity `json:"reserveMemory"`
		} `json:"redpanda"`
	} `json:"memory" jsonschema:"required"`
}

func (rr *RedpandaResources) GetOverProvisionValue() bool {
	if rr.CPU.Cores.MilliValue() < 1000 {
		return true
	}

	return ptr.Deref(rr.CPU.Overprovisioned, false)
}

type Storage struct {
	HostPath         string  `json:"hostPath" jsonschema:"required"`
	Tiered           *Tiered `json:"tiered" jsonschema:"required"`
	PersistentVolume *struct {
		Annotations   map[string]string `json:"annotations" jsonschema:"required"`
		Enabled       bool              `json:"enabled" jsonschema:"required"`
		Labels        map[string]string `json:"labels" jsonschema:"required"`
		Size          resource.Quantity `json:"size" jsonschema:"required"`
		StorageClass  string            `json:"storageClass" jsonschema:"required"`
		NameOverwrite string            `json:"nameOverwrite"`
	} `json:"persistentVolume" jsonschema:"required,deprecated"`
	TieredConfig                  TieredStorageConfig `json:"tieredConfig" jsonschema:"deprecated"`
	TieredStorageHostPath         string              `json:"tieredStorageHostPath" jsonschema:"deprecated"`
	TieredStoragePersistentVolume *struct {
		Annotations  map[string]string `json:"annotations" jsonschema:"required"`
		Enabled      bool              `json:"enabled" jsonschema:"required"`
		Labels       map[string]string `json:"labels" jsonschema:"required"`
		StorageClass string            `json:"storageClass" jsonschema:"required"`
	} `json:"tieredStoragePersistentVolume" jsonschema:"deprecated"`
}

func (s *Storage) IsTieredStorageEnabled() bool {
	conf := s.GetTieredStorageConfig()

	b, ok := conf["cloud_storage_enabled"]
	return ok && b.(bool)
}

func (s *Storage) GetTieredStorageConfig() TieredStorageConfig {
	if len(s.TieredConfig) > 0 {
		return s.TieredConfig
	}

	return s.Tiered.Config
}

// +gotohelm:ignore=true
func (Storage) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "tieredConfig", "persistentVolume", "tieredStorageHostPath", "tieredStoragePersistentVolume")

	// TODO note why we do this.
	tieredConfig, _ := schema.Properties.Get("tieredConfig")
	tieredConfig.Required = []string{}
}

func (s *Storage) Translate() map[string]any {
	result := map[string]any{}

	if !s.IsTieredStorageEnabled() {
		return result
	}

	tieredStorageConfig := s.GetTieredStorageConfig()
	for k, v := range tieredStorageConfig {
		if v == nil || helmette.Empty(v) {
			continue
		}

		if k == "cloud_storage_cache_size" && v != nil {
			result[k] = fmt.Sprintf("%d", helmette.UnmarshalInto[*resource.Quantity](v).Value())
			continue
		}

		if str, ok := v.(string); ok {
			result[k] = str
		} else if b, ok := v.(bool); ok {
			result[k] = b
		} else if f, isFloat := helmette.AsNumeric(v); isFloat {
			result[k] = int(f)
		} else {
			result[k] = helmette.MustToJSON(v)
		}
	}

	return result
}

func (s *Storage) StorageMinFreeBytes() int64 {
	if s.PersistentVolume != nil && !s.PersistentVolume.Enabled {
		// Five GiB literal
		return fiveGiB
	}

	minimumFreeBytes := float64(s.PersistentVolume.Size.Value()) * 0.05
	return helmette.Min(fiveGiB, int64(minimumFreeBytes))
}

type PostInstallJob struct {
	Resources       *corev1.ResourceRequirements `json:"resources"`
	Affinity        corev1.Affinity              `json:"affinity"`
	Enabled         bool                         `json:"enabled"`
	Labels          map[string]string            `json:"labels"`
	Annotations     map[string]string            `json:"annotations"`
	SecurityContext *corev1.SecurityContext      `json:"securityContext"`
}

type PostUpgradeJob struct {
	Resources       corev1.ResourceRequirements `json:"resources"`
	Affinity        corev1.Affinity             `json:"affinity"`
	Enabled         bool                        `json:"enabled"`
	Labels          map[string]string           `json:"labels"`
	Annotations     map[string]string           `json:"annotations"`
	BackoffLimit    *int32                      `json:"backoffLimit"`
	ExtraEnv        []corev1.EnvVar             `json:"extraEnv"`
	ExtraEnvFrom    []corev1.EnvFromSource      `json:"extraEnvFrom"`
	SecurityContext *corev1.SecurityContext     `json:"securityContext"`
}

type ContainerName string

// +gotohelm:ignore=true
func (ContainerName) JSONSchemaExtend(s *jsonschema.Schema) {
	s.Enum = append(s.Enum, RedpandaContainerName)
}

type Container struct {
	Name ContainerName   `json:"name" jsonschema:"required"`
	Env  []corev1.EnvVar `json:"env" jsonschema:"required"`
}

// PodSpec is a subset of [corev1.PodSpec] that will be merged into the objects
// constructed by this helm chart via means of a [strategic merge
// patch](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment).
// NOTE: At the time of writing, merging is manually implemented for each
// field. Ideally, a more generally applicable solution could be used.
type PodSpec struct {
	Containers []Container `json:"containers" jsonschema:"required"`
}

type PodTemplate struct {
	Labels      map[string]string `json:"labels" jsonschema:"required"`
	Annotations map[string]string `json:"annotations" jsonschema:"required"`
	Spec        PodSpec           `json:"spec" jsonschema:"required"`
}

type Statefulset struct {
	AdditionalSelectorLabels map[string]string `json:"additionalSelectorLabels" jsonschema:"required"`
	NodeAffinity             map[string]any    `json:"nodeAffinity"`
	Replicas                 int               `json:"replicas" jsonschema:"required"`
	UpdateStrategy           struct {
		Type string `json:"type" jsonschema:"required,pattern=^(RollingUpdate|OnDelete)$"`
	} `json:"updateStrategy" jsonschema:"required"`
	AdditionalRedpandaCmdFlags []string `json:"additionalRedpandaCmdFlags"`
	// Annotations are used only for `Statefulset.spec.template.metadata.annotations`. The StatefulSet does not have
	// any dedicated annotation.
	Annotations map[string]string `json:"annotations" jsonschema:"deprecated"`
	PodTemplate PodTemplate       `json:"podTemplate" jsonschema:"required"`
	Budget      struct {
		MaxUnavailable int `json:"maxUnavailable" jsonschema:"required"`
	} `json:"budget" jsonschema:"required"`
	StartupProbe struct {
		InitialDelaySeconds int32 `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int32 `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int32 `json:"periodSeconds" jsonschema:"required"`
	} `json:"startupProbe" jsonschema:"required"`
	LivenessProbe struct {
		InitialDelaySeconds int32 `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int32 `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int32 `json:"periodSeconds" jsonschema:"required"`
	} `json:"livenessProbe" jsonschema:"required"`
	ReadinessProbe struct {
		InitialDelaySeconds int32 `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int32 `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int32 `json:"periodSeconds" jsonschema:"required"`
		SuccessThreshold    int32 `json:"successThreshold"`
		TimeoutSeconds      int32 `json:"timeoutSeconds"`
	} `json:"readinessProbe" jsonschema:"required"`
	PodAffinity     map[string]any `json:"podAffinity" jsonschema:"required"`
	PodAntiAffinity struct {
		TopologyKey string         `json:"topologyKey" jsonschema:"required"`
		Type        string         `json:"type" jsonschema:"required,pattern=^(hard|soft|custom)$"`
		Weight      int            `json:"weight" jsonschema:"required"`
		Custom      map[string]any `json:"custom"`
	} `json:"podAntiAffinity" jsonschema:"required"`
	NodeSelector                  map[string]string `json:"nodeSelector" jsonschema:"required"`
	PriorityClassName             string            `json:"priorityClassName" jsonschema:"required"`
	TerminationGracePeriodSeconds int               `json:"terminationGracePeriodSeconds"`
	TopologySpreadConstraints     []struct {
		MaxSkew           int    `json:"maxSkew"`
		TopologyKey       string `json:"topologyKey"`
		WhenUnsatisfiable string `json:"whenUnsatisfiable" jsonschema:"pattern=^(ScheduleAnyway|DoNotSchedule)$"`
	} `json:"topologySpreadConstraints" jsonschema:"required,minItems=1"`
	Tolerations []corev1.Toleration `json:"tolerations" jsonschema:"required"`
	// DEPRECATED. Not to be confused with [corev1.PodSecurityContext], this
	// field is a historical artifact that should be quickly removed.
	PodSecurityContext *SecurityContext `json:"podSecurityContext"`
	SecurityContext    SecurityContext  `json:"securityContext" jsonschema:"required"`
	SideCars           struct {
		ConfigWatcher struct {
			Enabled           bool                    `json:"enabled"`
			ExtraVolumeMounts []corev1.VolumeMount    `json:"extraVolumeMounts"`
			Resources         map[string]any          `json:"resources"`
			SecurityContext   *corev1.SecurityContext `json:"securityContext"`
		} `json:"configWatcher"`
		Controllers struct {
			Image struct {
				Tag        ImageTag `json:"tag" jsonschema:"required,default=Chart.appVersion"`
				Repository string   `json:"repository" jsonschema:"required,default=docker.redpanda.com/redpandadata/redpanda-operator"`
			} `json:"image"`
			Enabled            bool                    `json:"enabled"`
			CreateRBAC         bool                    `json:"createRBAC"`
			Resources          any                     `json:"resources"`
			SecurityContext    *corev1.SecurityContext `json:"securityContext"`
			HealthProbeAddress string                  `json:"healthProbeAddress"`
			MetricsAddress     string                  `json:"metricsAddress"`
			Run                []string                `json:"run"`
		} `json:"controllers"`
	} `json:"sideCars" jsonschema:"required"`
	ExtraVolumes      []corev1.Volume      `json:"extraVolumes"`
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
	InitContainers    struct {
		Configurator struct {
			ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
			Resources         map[string]any       `json:"resources"`
		} `json:"configurator"`
		FSValidator struct {
			Enabled           bool                 `json:"enabled"`
			Resources         map[string]any       `json:"resources"`
			ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
			ExpectedFS        string               `json:"expectedFS"`
		} `json:"fsValidator"`
		SetDataDirOwnership struct {
			Enabled           bool                 `json:"enabled"`
			Resources         map[string]any       `json:"resources"`
			ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
		} `json:"setDataDirOwnership"`
		SetTieredStorageCacheDirOwnership struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any       `json:"resources"`
			ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
		} `json:"setTieredStorageCacheDirOwnership"`
		Tuning struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any       `json:"resources"`
			ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts"`
		} `json:"tuning"`
		ExtraInitContainers []corev1.Container `json:"extraInitContainers"`
	} `json:"initContainers"`
	InitContainerImage struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"initContainerImage"`
}

// +gotohelm:ignore=true
func (Statefulset) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "podSecurityContext")
}

type ServiceAccountCfg struct {
	Create      bool              `json:"create" jsonschema:"required"`
	Name        string            `json:"name" jsonschema:"required"`
	Annotations map[string]string `json:"annotations" jsonschema:"required"`
}

type RBAC struct {
	Enabled     bool              `json:"enabled" jsonschema:"required"`
	Annotations map[string]string `json:"annotations" jsonschema:"required"`
}

type Tuning struct {
	TuneAIOEvents   bool   `json:"tune_aio_events"`
	TuneClocksource bool   `json:"tune_clocksource"`
	TuneBallastFile bool   `json:"tune_ballast_file"`
	BallastFilePath string `json:"ballast_file_path"`
	BallastFileSize string `json:"ballast_file_size"`
	WellKnownIO     string `json:"well_known_io"`
}

func (t *Tuning) Translate() map[string]any {
	result := map[string]any{}

	s := helmette.ToJSON(t)
	tune := helmette.FromJSON(s)
	m, ok := tune.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	for k, v := range m {
		result[k] = v
	}

	return result
}

type Listeners struct {
	Admin          AdminListeners          `json:"admin" jsonschema:"required"`
	HTTP           HTTPListeners           `json:"http" jsonschema:"required"`
	Kafka          KafkaListeners          `json:"kafka" jsonschema:"required"`
	SchemaRegistry SchemaRegistryListeners `json:"schemaRegistry" jsonschema:"required"`
	RPC            struct {
		Port int32       `json:"port" jsonschema:"required"`
		TLS  InternalTLS `json:"tls" jsonschema:"required"`
	} `json:"rpc" jsonschema:"required"`
}

func (l *Listeners) CreateSeedServers(replicas int, fullname, internalDomain string) []map[string]any {
	var result []map[string]any
	for i := 0; i < replicas; i++ {
		result = append(result, map[string]any{
			"host": map[string]any{
				"address": fmt.Sprintf("%s-%d.%s", fullname, i, internalDomain),
				"port":    l.RPC.Port,
			},
		})
	}
	return result
}

func (l *Listeners) AdminList(replicas int, fullname, internalDomain string) []string {
	var result []string
	for i := 0; i < replicas; i++ {
		result = append(result, fmt.Sprintf("%s-%d.%s:%d", fullname, i, internalDomain, int(l.Admin.Port)))
	}
	return result
}

// TrustStoreVolume returns a [corev1.Volume] containing a projected volume
// that mounts all required truststore files. If no truststores are configured,
// it returns nil.
func (l *Listeners) TrustStoreVolume(tls *TLS) *corev1.Volume {
	var sources []corev1.VolumeProjection
	for _, ts := range l.TrustStores(tls) {
		sources = append(sources, ts.VolumeProjection())
	}

	if len(sources) < 1 {
		return nil
	}

	return &corev1.Volume{
		Name: "truststores",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: sources,
			},
		},
	}
}

// TrustStores returns an aggregate slice of all "active" [TrustStore]s across
// all listeners.
func (l *Listeners) TrustStores(tls *TLS) []*TrustStore {
	tss := l.Kafka.TrustStores(tls)
	tss = append(tss, l.Admin.TrustStores(tls)...)
	tss = append(tss, l.HTTP.TrustStores(tls)...)
	tss = append(tss, l.SchemaRegistry.TrustStores(tls)...)
	return tss
}

type Config struct {
	Cluster              ClusterConfig         `json:"cluster" jsonschema:"required"`
	Node                 NodeConfig            `json:"node" jsonschema:"required"`
	RPK                  map[string]any        `json:"rpk"`
	SchemaRegistryClient *SchemaRegistryClient `json:"schema_registry_client"`
	PandaProxyClient     *PandaProxyClient     `json:"pandaproxy_client"`
	Tunable              TunableConfig         `json:"tunable" jsonschema:"required"`
}

func (c *Config) CreateRPKConfiguration() map[string]any {
	result := map[string]any{}

	for k, v := range c.RPK {
		result[k] = v
	}

	return result
}

type SchemaRegistryClient struct {
	Retries                     int `json:"retries"`
	RetryBaseBackoffMS          int `json:"retry_base_backoff_ms"`
	ProduceBatchRecordCount     int `json:"produce_batch_record_count"`
	ProduceBatchSizeBytes       int `json:"produce_batch_size_bytes"`
	ProduceBatchDelayMS         int `json:"produce_batch_delay_ms"`
	ConsumerRequestTimeoutMS    int `json:"consumer_request_timeout_ms"`
	ConsumerRequestMaxBytes     int `json:"consumer_request_max_bytes"`
	ConsumerSessionTimeoutMS    int `json:"consumer_session_timeout_ms"`
	ConsumerRebalanceTimeoutMS  int `json:"consumer_rebalance_timeout_ms"`
	ConsumerHeartbeatIntervalMS int `json:"consumer_heartbeat_interval_ms"`
}

type PandaProxyClient struct {
	Retries                     int `json:"retries"`
	RetryBaseBackoffMS          int `json:"retry_base_backoff_ms"`
	ProduceBatchRecordCount     int `json:"produce_batch_record_count"`
	ProduceBatchSizeBytes       int `json:"produce_batch_size_bytes"`
	ProduceBatchDelayMS         int `json:"produce_batch_delay_ms"`
	ConsumerRequestTimeoutMS    int `json:"consumer_request_timeout_ms"`
	ConsumerRequestMaxBytes     int `json:"consumer_request_max_bytes"`
	ConsumerSessionTimeoutMS    int `json:"consumer_session_timeout_ms"`
	ConsumerRebalanceTimeoutMS  int `json:"consumer_rebalance_timeout_ms"`
	ConsumerHeartbeatIntervalMS int `json:"consumer_heartbeat_interval_ms"`
}

type TLSCert struct {
	// Enabled should be interpreted as `true` if not set.
	Enabled               *bool                        `json:"enabled"`
	CAEnabled             bool                         `json:"caEnabled" jsonschema:"required"`
	ApplyInternalDNSNames *bool                        `json:"applyInternalDNSNames"`
	Duration              string                       `json:"duration" jsonschema:"pattern=.*[smh]$"`
	IssuerRef             *cmmeta.ObjectReference      `json:"issuerRef"`
	SecretRef             *corev1.LocalObjectReference `json:"secretRef"`
}

// +gotohelm:ignore=true
func (TLSCert) JSONSchemaExtend(schema *jsonschema.Schema) {
	// An object reference could allow anything but we want to require that the
	// reference is to either a ClusterIssuer or Issuer.
	ref, _ := schema.Properties.Get("issuerRef")
	refKind, _ := ref.Properties.Get("kind")
	refKind.Enum = []any{
		certmanagerv1.ClusterIssuerKind,
		certmanagerv1.IssuerKind,
	}
}

type TLSCertMap map[string]TLSCert

// +gotohelm:ignore=true
func (TLSCertMap) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.PatternProperties = map[string]*jsonschema.Schema{
		"^[A-Za-z_][A-Za-z0-9_]*$": schema.AdditionalProperties,
	}
	minProps := uint64(1)
	schema.MinProperties = &minProps
	schema.AdditionalProperties = nil
}

func (m TLSCertMap) MustGet(name string) *TLSCert {
	cert, ok := m[name]
	if !ok {
		panic(fmt.Sprintf("Certificate %q referenced, but not found in the tls.certs map", name))
	}
	return &cert
}

type SASLUser struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Mechanism string `json:"mechanism" jsonschema:"pattern=^(SCRAM-SHA-512|SCRAM-SHA-256)$"`
}

type SASLAuth struct {
	Enabled   bool       `json:"enabled" jsonschema:"required"`
	Mechanism string     `json:"mechanism"`
	SecretRef string     `json:"secretRef"`
	Users     []SASLUser `json:"users"`
}

type TrustStore struct {
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef"`
	SecretKeyRef    *corev1.SecretKeySelector    `json:"secretKeyRef"`
}

func (t *TrustStore) TrustStoreFilePath() string {
	return fmt.Sprintf("%s/%s", TrustStoreMountPath, t.RelativePath())
}

func (t *TrustStore) RelativePath() string {
	if t.ConfigMapKeyRef != nil {
		return fmt.Sprintf("configmaps/%s-%s", t.ConfigMapKeyRef.Name, t.ConfigMapKeyRef.Key)
	}
	return fmt.Sprintf("secrets/%s-%s", t.SecretKeyRef.Name, t.SecretKeyRef.Key)
}

func (t *TrustStore) VolumeProjection() corev1.VolumeProjection {
	if t.ConfigMapKeyRef != nil {
		return corev1.VolumeProjection{
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.ConfigMapKeyRef.Name,
				},
				Items: []corev1.KeyToPath{{
					Key:  t.ConfigMapKeyRef.Key,
					Path: t.RelativePath(),
				}},
			},
		}
	}
	return corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: t.SecretKeyRef.Name,
			},
			Items: []corev1.KeyToPath{{
				Key:  t.SecretKeyRef.Key,
				Path: t.RelativePath(),
			}},
		},
	}
}

// InternalTLS is the TLS configuration for "internal" listeners. Internal
// listeners all have default values specified within values.yaml which allows
// us to be more strict about the schema here.
// TODO Unify this struct with ExternalTLS and/or remove the concept of
// internal and external listeners all together.
type InternalTLS struct {
	Enabled           *bool       `json:"enabled"`
	Cert              string      `json:"cert" jsonschema:"required"`
	RequireClientAuth bool        `json:"requireClientAuth" jsonschema:"required"`
	TrustStore        *TrustStore `json:"trustStore"`
}

// IsEnabled reports the value of [InternalTLS.Enabled], falling back to
// [TLS.Enabled] if not specified.
func (t *InternalTLS) IsEnabled(tls *TLS) bool {
	// Default Enabled to the value of the global TLS struct.
	return ptr.Deref(t.Enabled, tls.Enabled) && t.Cert != ""
}

func (t *InternalTLS) TrustStoreFilePath(tls *TLS) string {
	if t.TrustStore != nil {
		return t.TrustStore.TrustStoreFilePath()
	}

	if tls.Certs.MustGet(t.Cert).CAEnabled {
		return fmt.Sprintf("/etc/tls/certs/%s/ca.crt", t.Cert)
	}

	return defaultTruststorePath
}

// ExternalTLS is the TLS configuration associated with a given "external"
// listener. The schema is more loose than InternalTLS. All fields have default
// values but are interpreted differently depending on their context (IE kafka
// vs schemaRegistry) tread lightly.
type ExternalTLS struct {
	// Enabled, when `false`, indicates that this struct should treated as if
	// it was not specified. If `nil`, defaults to [InternalTLS.Enabled].
	// Prefer to use `IsEnabled` rather than checking this field directly.
	Enabled           *bool       `json:"enabled"`
	Cert              *string     `json:"cert"`
	RequireClientAuth *bool       `json:"requireClientAuth"`
	TrustStore        *TrustStore `json:"trustStore"`
}

func (t *ExternalTLS) GetCert(i *InternalTLS, tls *TLS) *TLSCert {
	return tls.Certs.MustGet(t.GetCertName(i))
}

func (t *ExternalTLS) GetCertName(i *InternalTLS) string {
	return ptr.Deref(t.Cert, i.Cert)
}

func (t *ExternalTLS) TrustStoreFilePath(i *InternalTLS, tls *TLS) string {
	if t.TrustStore != nil {
		return t.TrustStore.TrustStoreFilePath()
	}

	if t.GetCert(i, tls).CAEnabled {
		return fmt.Sprintf("/etc/tls/certs/%s/ca.crt", t.GetCertName(i))
	}

	return defaultTruststorePath
}

// IsEnabled reports the value of [ExternalTLS.Enabled], falling back to
// [InternalTLS.IsEnabled] if not specified.
func (t *ExternalTLS) IsEnabled(i *InternalTLS, tls *TLS) bool {
	// If t is nil, interpret Enabled as false.
	if t == nil {
		return false
	}
	return t.GetCertName(i) != "" && ptr.Deref(t.Enabled, i.IsEnabled(tls))
}

type AdminListeners struct {
	External ExternalListeners[AdminExternal] `json:"external"`
	Port     int32                            `json:"port" jsonschema:"required"`
	TLS      InternalTLS                      `json:"tls" jsonschema:"required"`
}

func (l *AdminListeners) Listeners() []map[string]any {
	admin := []map[string]any{
		createInternalListenerCfg(l.Port),
	}

	for k, lis := range l.External {
		if !lis.IsEnabled() {
			continue
		}

		admin = append(admin, map[string]any{
			"name":    k,
			"port":    lis.Port,
			"address": "0.0.0.0",
		})
	}
	return admin
}

func (l *AdminListeners) ListenersTLS(tls *TLS) []map[string]any {
	admin := []map[string]any{}

	internal := createInternalListenerTLSCfg(tls, l.TLS)
	if len(internal) > 0 {
		admin = append(admin, internal)
	}

	for k, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) {
			continue
		}

		certName := lis.TLS.GetCertName(&l.TLS)

		admin = append(admin, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(lis.TLS.RequireClientAuth, false),
			"truststore_file":     lis.TLS.TrustStoreFilePath(&l.TLS, tls),
		})
	}
	return admin
}

// TrustStores returns a slice of all configured and enabled [TrustStore]s on
// both internal and external listeners.
func (l *AdminListeners) TrustStores(tls *TLS) []*TrustStore {
	tss := []*TrustStore{}

	if l.TLS.IsEnabled(tls) && l.TLS.TrustStore != nil {
		tss = append(tss, l.TLS.TrustStore)
	}

	for _, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

type AdminExternal struct {
	// Enabled indicates if this listener is enabled. If not specified,
	// defaults to the value of [ExternalConfig.Enabled].
	Enabled         *bool        `json:"enabled"`
	AdvertisedPorts []int32      `json:"advertisedPorts" jsonschema:"minItems=1"`
	Port            int32        `json:"port" jsonschema:"required"`
	NodePort        *int32       `json:"nodePort"`
	TLS             *ExternalTLS `json:"tls"`
}

func (l *AdminExternal) IsEnabled() bool {
	return ptr.Deref(l.Enabled, true) && l.Port > 0
}

type HTTPListeners struct {
	Enabled              bool                            `json:"enabled" jsonschema:"required"`
	External             ExternalListeners[HTTPExternal] `json:"external"`
	AuthenticationMethod *HTTPAuthenticationMethod       `json:"authenticationMethod"`
	TLS                  InternalTLS                     `json:"tls" jsonschema:"required"`
	KafkaEndpoint        string                          `json:"kafkaEndpoint" jsonschema:"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$"`
	Port                 int32                           `json:"port" jsonschema:"required"`
}

// +gotohelm:ignore=true
func (HTTPListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

func (l *HTTPListeners) Listeners(saslEnabled bool) []map[string]any {
	internal := createInternalListenerCfg(l.Port)

	if saslEnabled {
		internal["authentication_method"] = "http_basic"
	}

	if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
		internal["authentication_method"] = am
	}

	result := []map[string]any{
		internal,
	}

	for k, l := range l.External {
		if !l.IsEnabled() {
			continue
		}

		listener := map[string]any{
			"name":    k,
			"port":    l.Port,
			"address": "0.0.0.0",
		}

		if saslEnabled {
			listener["authentication_method"] = "http_basic"
		}

		if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
			listener["authentication_method"] = am
		}

		result = append(result, listener)
	}

	return result
}

func (l *HTTPListeners) ListenersTLS(tls *TLS) []map[string]any {
	pp := []map[string]any{}

	internal := createInternalListenerTLSCfg(tls, l.TLS)
	if len(internal) > 0 {
		pp = append(pp, internal)
	}

	for k, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) {
			continue
		}

		certName := lis.TLS.GetCertName(&l.TLS)

		pp = append(pp, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(lis.TLS.RequireClientAuth, false),
			"truststore_file":     lis.TLS.TrustStoreFilePath(&l.TLS, tls),
		})
	}
	return pp
}

// TrustStores returns a slice of all configured and enabled [TrustStore]s on
// both internal and external listeners.
func (l *HTTPListeners) TrustStores(tls *TLS) []*TrustStore {
	var tss []*TrustStore

	if l.TLS.IsEnabled(tls) && l.TLS.TrustStore != nil {
		tss = append(tss, l.TLS.TrustStore)
	}

	for _, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

type HTTPExternal struct {
	// Enabled indicates if this listener is enabled. If not specified,
	// defaults to the value of [ExternalConfig.Enabled].
	Enabled              *bool                     `json:"enabled"`
	AdvertisedPorts      []int32                   `json:"advertisedPorts" jsonschema:"minItems=1"`
	Port                 int32                     `json:"port" jsonschema:"required"`
	NodePort             *int32                    `json:"nodePort"`
	AuthenticationMethod *HTTPAuthenticationMethod `json:"authenticationMethod"`
	PrefixTemplate       *string                   `json:"prefixTemplate"`
	TLS                  *ExternalTLS              `json:"tls" jsonschema:"required"`
}

func (l *HTTPExternal) IsEnabled() bool {
	return ptr.Deref(l.Enabled, true) && l.Port > 0
}

// +gotohelm:ignore=true
func (HTTPExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
	// TODO document me. Legacy matching needs to be removed in a minor bump.
	tls, _ := schema.Properties.Get("tls")
	tls.Required = []string{}
	schema.Required = []string{"port"}
}

type KafkaListeners struct {
	AuthenticationMethod *KafkaAuthenticationMethod       `json:"authenticationMethod"`
	External             ExternalListeners[KafkaExternal] `json:"external"`
	TLS                  InternalTLS                      `json:"tls" jsonschema:"required"`
	Port                 int32                            `json:"port" jsonschema:"required"`
}

// +gotohelm:ignore=true
func (KafkaListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

// Listeners returns a slice of maps suitable for use as the value of
// `kafka_api` in a redpanda.yml file.
func (l *KafkaListeners) Listeners(auth *Auth) []map[string]any {
	internal := createInternalListenerCfg(l.Port)

	if auth.IsSASLEnabled() {
		internal["authentication_method"] = "sasl"
	}

	if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
		internal["authentication_method"] = am
	}

	kafka := []map[string]any{
		internal,
	}

	for k, l := range l.External {
		if !l.IsEnabled() {
			continue
		}

		listener := map[string]any{
			"name":    k,
			"port":    l.Port,
			"address": "0.0.0.0",
		}

		if auth.IsSASLEnabled() {
			listener["authentication_method"] = "sasl"
		}

		if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
			listener["authentication_method"] = am
		}

		kafka = append(kafka, listener)
	}

	return kafka
}

// ListenersTLS returns a slice of maps suitable for use as the value of
// `kafka_api_tls` in a redpanda.yml file.
func (l *KafkaListeners) ListenersTLS(tls *TLS) []map[string]any {
	kafka := []map[string]any{}

	internal := createInternalListenerTLSCfg(tls, l.TLS)
	if len(internal) > 0 {
		kafka = append(kafka, internal)
	}

	for k, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) {
			continue
		}

		certName := lis.TLS.GetCertName(&l.TLS)

		kafka = append(kafka, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(lis.TLS.RequireClientAuth, false),
			"truststore_file":     lis.TLS.TrustStoreFilePath(&l.TLS, tls),
		})
	}
	return kafka
}

// TrustStores returns a slice of all configured and enabled [TrustStore]s on
// both internal and external listeners.
func (l *KafkaListeners) TrustStores(tls *TLS) []*TrustStore {
	var tss []*TrustStore

	if l.TLS.IsEnabled(tls) && l.TLS.TrustStore != nil {
		tss = append(tss, l.TLS.TrustStore)
	}

	for _, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

type KafkaExternal struct {
	// Enabled indicates if this listener is enabled. If not specified,
	// defaults to the value of [ExternalConfig.Enabled].
	Enabled         *bool   `json:"enabled"`
	AdvertisedPorts []int32 `json:"advertisedPorts" jsonschema:"minItems=1"`
	Port            int32   `json:"port" jsonschema:"required"`
	// TODO CHECK NODE PORT USAGE
	NodePort             *int32                     `json:"nodePort"`
	AuthenticationMethod *KafkaAuthenticationMethod `json:"authenticationMethod"`
	PrefixTemplate       *string                    `json:"prefixTemplate"`
	TLS                  *ExternalTLS               `json:"tls"`
}

func (l *KafkaExternal) IsEnabled() bool {
	return ptr.Deref(l.Enabled, true) && l.Port > 0
}

// +gotohelm:ignore=true
func (KafkaExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

type SchemaRegistryListeners struct {
	// Enabled indicates if this listener is enabled. If not specified,
	// defaults to the value of [ExternalConfig.Enabled].
	Enabled              bool                                      `json:"enabled" jsonschema:"required"`
	External             ExternalListeners[SchemaRegistryExternal] `json:"external"`
	AuthenticationMethod *HTTPAuthenticationMethod                 `json:"authenticationMethod"`
	KafkaEndpoint        string                                    `json:"kafkaEndpoint" jsonschema:"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$"`
	Port                 int32                                     `json:"port" jsonschema:"required"`
	TLS                  InternalTLS                               `json:"tls" jsonschema:"required"`
}

// +gotohelm:ignore=true
func (SchemaRegistryListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

func (sr *SchemaRegistryListeners) Listeners(saslEnabled bool) []map[string]any {
	internal := createInternalListenerCfg(sr.Port)

	if saslEnabled {
		internal["authentication_method"] = "http_basic"
	}

	if am := ptr.Deref(sr.AuthenticationMethod, ""); am != "" {
		internal["authentication_method"] = am
	}

	result := []map[string]any{
		internal,
	}

	for k, l := range sr.External {
		if !l.IsEnabled() {
			continue
		}

		listener := map[string]any{
			"name":    k,
			"port":    l.Port,
			"address": "0.0.0.0",
		}

		if saslEnabled {
			listener["authentication_method"] = "http_basic"
		}

		if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
			listener["authentication_method"] = am
		}

		result = append(result, listener)
	}

	return result
}

func (l *SchemaRegistryListeners) ListenersTLS(tls *TLS) []map[string]any {
	listeners := []map[string]any{}

	internal := createInternalListenerTLSCfg(tls, l.TLS)
	if len(internal) > 0 {
		listeners = append(listeners, internal)
	}

	for k, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) {
			continue
		}

		certName := lis.TLS.GetCertName(&l.TLS)

		listeners = append(listeners, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(lis.TLS.RequireClientAuth, false),
			"truststore_file":     lis.TLS.TrustStoreFilePath(&l.TLS, tls),
		})
	}
	return listeners
}

// TrustStores returns a slice of all configured and enabled [TrustStore]s on
// both internal and external listeners.
func (l *SchemaRegistryListeners) TrustStores(tls *TLS) []*TrustStore {
	var tss []*TrustStore

	if l.TLS.IsEnabled(tls) && l.TLS.TrustStore != nil {
		tss = append(tss, l.TLS.TrustStore)
	}

	for _, lis := range l.External {
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

type SchemaRegistryExternal struct {
	// Enabled indicates if this listener is enabled. If not specified,
	// defaults to the value of [ExternalConfig.Enabled].
	Enabled              *bool                     `json:"enabled"`
	AdvertisedPorts      []int32                   `json:"advertisedPorts" jsonschema:"minItems=1"`
	Port                 int32                     `json:"port"`
	NodePort             *int32                    `json:"nodePort"`
	AuthenticationMethod *HTTPAuthenticationMethod `json:"authenticationMethod"`
	TLS                  *ExternalTLS              `json:"tls"`
}

func (l *SchemaRegistryExternal) IsEnabled() bool {
	return ptr.Deref(l.Enabled, true) && l.Port > 0
}

// +gotohelm:ignore=true
func (SchemaRegistryExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
	// TODO this as well
	tls, _ := schema.Properties.Get("tls")
	tls.Required = []string{}
}

type TunableConfig map[string]any

// +gotohelm:ignore=true
func (TunableConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.AdditionalProperties = jsonschema.TrueSchema
	schema.Properties = orderedmap.New[string, *jsonschema.Schema]()
	schema.Properties.Set("log_retention_ms", &jsonschema.Schema{
		Type: "integer",
	})
	schema.Properties.Set("group_initial_rebalance_delay", &jsonschema.Schema{
		Type: "integer",
	})
}

func (c *TunableConfig) Translate() map[string]any {
	if c == nil {
		return nil
	}

	result := map[string]any{}

	for k, v := range *c {
		if !helmette.Empty(v) {
			result[k] = v
		}
	}
	return result
}

type NodeConfig map[string]any

func (c *NodeConfig) Translate() map[string]any {
	result := map[string]any{}

	for k, v := range *c {
		if !helmette.Empty(v) {
			if _, ok := helmette.AsNumeric(v); ok {
				result[k] = v
			} else {
				result[k] = helmette.ToYaml(v)
			}
		}
	}

	return result
}

type ClusterConfig map[string]any

func (c *ClusterConfig) Translate(replicas int, skipDefaultTopic bool) map[string]any {
	result := map[string]any{}

	for k, v := range *c {
		if k == "default_topic_replications" && !skipDefaultTopic {
			r := int(replicas)
			input := int(r)
			if num, ok := helmette.AsIntegral[int](v); ok {
				input = num
			}

			if f, ok := helmette.AsNumeric(v); ok {
				input = int(f)
			}

			result[k] = helmette.Min(int64(input), int64(r+(r%2)-1))
			continue
		}

		if b, ok := v.(bool); ok {
			result[k] = b
			continue
		}

		if !helmette.Empty(v) {
			result[k] = v
		}
	}

	return result
}

type SecretRef struct {
	ConfigurationKey string `json:"configurationKey"`
	Key              string `json:"key"`
	Name             string `json:"name"`
}

// IsValid confirms whether EnvVarSource could be built from
// SecretRef.
func (sr *SecretRef) IsValid() bool {
	return sr != nil && !helmette.Empty(sr.Key) && !helmette.Empty(sr.Name)
}

type TieredStorageCredentials struct {
	ConfigurationKey string     `json:"configurationKey" jsonschema:"deprecated"`
	Key              string     `json:"key" jsonschema:"deprecated"`
	Name             string     `json:"name" jsonschema:"deprecated"`
	AccessKey        *SecretRef `json:"accessKey"`
	SecretKey        *SecretRef `json:"secretKey"`
}

func (tsc *TieredStorageCredentials) IsAccessKeyReferenceValid() bool {
	return tsc.AccessKey != nil && tsc.AccessKey.Name != "" && tsc.AccessKey.Key != ""
}

func (tsc *TieredStorageCredentials) IsSecretKeyReferenceValid() bool {
	return tsc.SecretKey != nil && tsc.SecretKey.Name != "" && tsc.SecretKey.Key != ""
}

// +gotohelm:ignore=true
func (TieredStorageCredentials) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "configurationKey", "key", "name")
}

type TieredStorageConfig map[string]any

// +gotohelm:ignore=true
func (TieredStorageConfig) JSONSchema() *jsonschema.Schema {
	type schema struct {
		CloudStorageEnabled                     bool              `json:"cloud_storage_enabled" jsonschema:"required"`
		CloudStorageAccessKey                   string            `json:"cloud_storage_access_key"`
		CloudStorageSecretKey                   string            `json:"cloud_storage_secret_key"`
		CloudStorageAPIEndpoint                 string            `json:"cloud_storage_api_endpoint"`
		CloudStorageAPIEndpointPort             int               `json:"cloud_storage_api_endpoint_port"`
		CloudStorageAzureADLSEndpoint           string            `json:"cloud_storage_azure_adls_endpoint"`
		CloudStorageAzureADLSPort               int               `json:"cloud_storage_azure_adls_port"`
		CloudStorageBucket                      string            `json:"cloud_storage_bucket" jsonschema:"required"`
		CloudStorageCacheCheckInterval          int               `json:"cloud_storage_cache_check_interval"`
		CloudStorageCacheDirectory              string            `json:"cloud_storage_cache_directory"`
		CloudStorageCacheSize                   *ResourceQuantity `json:"cloud_storage_cache_size"`
		CloudStorageCredentialsSource           string            `json:"cloud_storage_credentials_source" jsonschema:"pattern=^(config_file|aws_instance_metadata|sts|gcp_instance_metadata)$"`
		CloudStorageDisableTLS                  bool              `json:"cloud_storage_disable_tls"`
		CloudStorageEnableRemoteRead            bool              `json:"cloud_storage_enable_remote_read"`
		CloudStorageEnableRemoteWrite           bool              `json:"cloud_storage_enable_remote_write"`
		CloudStorageInitialBackoffMS            int               `json:"cloud_storage_initial_backoff_ms"`
		CloudStorageManifestUploadTimeoutMS     int               `json:"cloud_storage_manifest_upload_timeout_ms"`
		CloudStorageMaxConnectionIdleTimeMS     int               `json:"cloud_storage_max_connection_idle_time_ms"`
		CloudStorageMaxConnections              int               `json:"cloud_storage_max_connections"`
		CloudStorageReconciliationIntervalMS    int               `json:"cloud_storage_reconciliation_interval_ms"`
		CloudStorageRegion                      string            `json:"cloud_storage_region" jsonschema:"required"`
		CloudStorageSegmentMaxUploadIntervalSec int               `json:"cloud_storage_segment_max_upload_interval_sec"`
		CloudStorageSegmentUploadTimeoutMS      int               `json:"cloud_storage_segment_upload_timeout_ms"`
		CloudStorageTrustFile                   string            `json:"cloud_storage_trust_file"`
		CloudStorageUploadCtrlDCoeff            int               `json:"cloud_storage_upload_ctrl_d_coeff"`
		CloudStorageUploadCtrlMaxShares         int               `json:"cloud_storage_upload_ctrl_max_shares"`
		CloudStorageUploadCtrlMinShares         int               `json:"cloud_storage_upload_ctrl_min_shares"`
		CloudStorageUploadCtrlPCoeff            int               `json:"cloud_storage_upload_ctrl_p_coeff"`
		CloudStorageUploadCtrlUpdateIntervalMS  int               `json:"cloud_storage_upload_ctrl_update_interval_ms"`
	}

	r := &jsonschema.Reflector{
		Anonymous: true,
		// Set for backwards compat.
		ExpandedStruct: true,
		// Set for backwards compat.
		DoNotReference: true,
		// Set for backwards compat.
		AllowAdditionalProperties: true,
		// Set because explicit behavior is much better.
		RequiredFromJSONSchemaTags: true,
	}

	s := r.Reflect(&schema{})
	s.Version = ""
	return s
}

type Tiered struct {
	CredentialsSecretRef TieredStorageCredentials `json:"credentialsSecretRef"`
	Config               TieredStorageConfig      `json:"config"`
	HostPath             string                   `json:"hostPath"`
	MountType            string                   `json:"mountType" jsonschema:"required,pattern=^(none|hostPath|emptyDir|persistentVolume)$"`
	PersistentVolume     struct {
		Annotations   map[string]string `json:"annotations" jsonschema:"required"`
		Enabled       bool              `json:"enabled"`
		Labels        map[string]string `json:"labels" jsonschema:"required"`
		NameOverwrite string            `json:"nameOverwrite"`
		Size          string            `json:"size"`
		StorageClass  string            `json:"storageClass" jsonschema:"required"`
	} `json:"persistentVolume"`
}
