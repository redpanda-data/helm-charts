// +gotohelm:ignore=true
package redpanda

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	corev1 "k8s.io/api/core/v1"
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
	NameOverride     string            `json:"nameOverride"`
	FullnameOverride string            `json:"fullnameOverride"`
	ClusterDomain    string            `json:"clusterDomain"`
	CommonLabels     map[string]string `json:"commonLabels"`
	NodeSelector     map[string]string `json:"nodeSelector"`
	Affinity         Affinity          `json:"affinity" jsonschema:"required"`
	Tolerations      []map[string]any  `json:"tolerations"`
	Image            Image             `json:"image" jsonschema:"required,description=Values used to define the container image to be used for Redpanda"`
	Service          *Service          `json:"service"`
	// ImagePullSecrets []string `json:"imagePullSecrets"`
	LicenseKey       string            `json:"license_key" jsonschema:"deprecated,pattern=^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?\\.(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$|^$"`
	LicenseSecretRef *LicenseSecretRef `json:"license_secret_ref" jsonschema:"deprecated"`
	AuditLogging     AuditLogging      `json:"auditLogging"`
	Enterprise       Enterprise        `json:"enterprise"`
	RackAwareness    RackAwareness     `json:"rackAwareness"`
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
	ServiceAccount ServiceAccount    `json:"serviceAccount"`
	RBAC           RBAC              `json:"rbac"`
	Tuning         Tuning            `json:"tuning"`
	Listeners      Listeners         `json:"listeners"`
	Config         Config            `json:"config"`
	Tests          *struct {
		Enabled bool `json:"enabled"`
	} `json:"tests"`
}

func (Values) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "license_key", "license_secret_ref")
}

type Affinity struct {
	NodeAffinity    map[string]any `json:"nodeAffinity"`
	PodAffinity     map[string]any `json:"podAffinity"`
	PodAntiAffinity map[string]any `json:"podAntiAffinity"`
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

	// FSGroupChangePolicy string `json:"fsGroupChangePolicy" jsonschema:"pattern=^(OnRootMismatch|Always)$"`
}

type Image struct {
	Repository ImageRepository `json:"repository" jsonschema:"required,default=docker.redpanda.com/redpandadata/redpanda"`
	Tag        ImageTag        `json:"tag" jsonschema:"default=Chart.appVersion"`
	PullPolicy string          `json:"pullPolicy" jsonschema:"required,pattern=^(Always|Never|IfNotPresent)$,description=The Kubernetes Pod image pull policy."`
}

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
	ReplicationFactor          *int     `json:"replicationFactor"`
}

func (AuditLogging) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "replicationFactor", "enabledEventTypes", "excludedPrincipals", "excludedTopics")
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

type TLS struct {
	Enabled bool      `json:"enabled" jsonschema:"required"`
	Certs   TLSCertMap `json:"certs"`
}

type ExternalConfig struct {
	Addresses      []string          `json:"addresses"`
	Annotations    map[string]string `json:"annotations"`
	Domain         *string           `json:"domain"`
	Enabled        bool              `json:"enabled" jsonschema:"required"`
	Type           string            `json:"type" jsonschema:"pattern=^(LoadBalancer|NodePort)$"`
	PrefixTemplate string            `json:"prefixTemplate"`
	SourceRanges   []string          `json:"sourceRanges"`
	Service        *struct {
		Enabled bool `json:"enabled"`
	} `json:"service"`
	ExternalDNS *struct {
		Enabled bool `json:"enabled" jsonschema:"required"`
	} `json:"externalDns"`
}

type Logging struct {
	LogLevel    string `json:"logLevel" jsonschema:"required,pattern=^(error|warn|info|debug|trace)$"`
	UseageStats struct {
		Enabled bool `json:"enabled" jsonschema:"required"`
	} `json:"usageStats" jsonschema:"required"`
}

type Monitoring struct {
	Enabled        *bool             `json:"enabled" jsonschema:"required"`
	ScrapeInterval string            `json:"scrapeInterval" jsonschema:"required,pattern=.*[smh]$"`
	Labels         map[string]string `json:"labels"`
	TLSConfig      map[string]any    `json:"tlsConfig"`
}

type RedpandaResources struct {
	CPU struct {
		Cores           any  `json:"cores" jsonschema:"required,oneof_type=integer;string"`
		Overprovisioned bool `json:"overprovisioned"`
	} `json:"cpu" jsonschema:"required"`
	// Memory resources
	// For details,
	// see the [Pod resources documentation](https://docs.redpanda.com/docs/manage/kubernetes/manage-resources/#configure-memory-resources).
	Memory struct {
		// Enables memory locking.
		// For production, set to `true`.
		EnableMemoryLocking bool `json:"enable_memory_locking"`
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
			Min *MemoryAmount `json:"min"`
			// Maximum memory count for each Redpanda broker.
			// Equivalent to `resources.limits.memory`.
			// For production, use `10Gi` or greater.
			Max MemoryAmount `json:"max" jsonschema:"required"`
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
			Memory *MemoryAmount `json:"memory" jsonschema:"oneof_type=integer;string"`
			// Memory reserved for the Seastar subsystem.
			// Any value above 1Gi will provide diminishing performance benefits.
			// Equivalent to --reserve-memory.
			// For production, use 1Gi.
			ReserveMemory *MemoryAmount `json:"reserveMemory" jsonschema:"oneof_type=integer;string"`
		} `json:"redpanda"`
	} `json:"memory" jsonschema:"required"`
}

type Storage struct {
	HostPath         string  `json:"hostPath" jsonschema:"required"`
	Tiered           *Tiered `json:"tiered" jsonschema:"required"`
	PersistentVolume *struct {
		Annotations  map[string]string `json:"annotations" jsonschema:"required"`
		Enabled      bool              `json:"enabled" jsonschema:"required"`
		Labels       map[string]string `json:"labels" jsonschema:"required"`
		Size         MemoryAmount      `json:"size" jsonschema:"required"`
		StorageClass string            `json:"storageClass" jsonschema:"required"`
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

func (Storage) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "tieredConfig", "persistentVolume", "tieredStorageHostPath", "tieredStoragePersistentVolume")

	// TODO note why we do this.
	tieredConfig, _ := schema.Properties.Get("tieredConfig")
	tieredConfig.Required = []string{}
}

type PostInstallJob struct {
	Resources JobResources   `json:"resources"`
	Affinity  map[string]any `json:"affinity"`

	// Fields that are in values.yaml but not in values.schema.json.
	// Enabled     bool              `json:"enabled"`
	// Labels      map[string]string `json:"labels"`
	// Annotations map[string]string `json:"annotations"`
}

type PostUpgradeJob struct {
	Resources    JobResources   `json:"resources"`
	Affinity     map[string]any `json:"affinity"`
	ExtraEnv     any            `json:"extraEnv" jsonschema:"oneof_type=array;string"`
	ExtraEnvFrom any            `json:"extraEnvFrom" jsonschema:"oneof_type=array;string"`

	// Fields that are in values.yaml but not in values.schema.json.
	// Enabled      bool              `json:"enabled"`
	// Labels       map[string]string `json:"labels"`
	// Annotations  map[string]string `json:"annotations"`
	// BackoffLimit int               `json:"backoffLimit"`
	// ExtraEnv    []corev1.EnvVar   `json:"extraEnv"`
	// ExtraEnvFrom []corev1.EnvFromSource `json:"extraEnvFrom"`
}

type ContainerName string

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
		InitialDelaySeconds int `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int `json:"periodSeconds" jsonschema:"required"`
	} `json:"startupProbe" jsonschema:"required" jsonschema:"required"`
	LivenessProbe struct {
		InitialDelaySeconds int `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int `json:"periodSeconds" jsonschema:"required"`
	} `json:"livenessProbe" jsonschema:"required"`
	ReadinessProbe struct {
		InitialDelaySeconds int `json:"initialDelaySeconds" jsonschema:"required"`
		FailureThreshold    int `json:"failureThreshold" jsonschema:"required"`
		PeriodSeconds       int `json:"periodSeconds" jsonschema:"required"`
		// SuccessThreshold    int `json:"successThreshold"`
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
	Tolerations []any `json:"tolerations" jsonschema:"required"`
	// DEPRECATED. Not to be confused with [corev1.PodSecurityContext], this
	// field is a historical artifact that should be quickly removed.
	PodSecurityContext *SecurityContext `json:"podSecurityContext"`
	SecurityContext    SecurityContext  `json:"securityContext" jsonschema:"required"`
	SideCars           struct {
		ConfigWatcher struct {
			Enabled           bool                    `json:"enabled"`
			ExtraVolumeMounts string                  `json:"extraVolumeMounts"`
			Resources         map[string]any          `json:"resources"`
			SecurityContext   *corev1.SecurityContext `json:"securityContext"`
		} `json:"configWatcher"`
		Controllers struct {
			Image struct {
				Tag        ImageTag        `json:"tag" jsonschema:"required,default=Chart.appVersion"`
				Repository ImageRepository `json:"repository" jsonschema:"required,default=docker.redpanda.com/redpandadata/redpanda-operator"`
			} `json:"image"`
			Enabled         bool                    `json:"enabled"`
			Resources       any                     `json:"resources"`
			SecurityContext *corev1.SecurityContext `json:"securityContext"`
		} `json:"controllers"`
	} `json:"sideCars" jsonschema:"required"`
	ExtraVolumes      string `json:"extraVolumes"`
	ExtraVolumeMounts string `json:"extraVolumeMounts"`
	InitContainers    struct {
		Configurator struct {
			ExtraVolumeMounts string         `json:"extraVolumeMounts"`
			Resources         map[string]any `json:"resources"`
		} `json:"configurator"`
		FSValidator struct {
			Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"`
			ExpectedFS        string         `json:"expectedFS"`
		} `json:"fsValidator"`
		SetDataDirOwnership struct {
			Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"`
		} `json:"setDataDirOwnership"`
		SetTieredStorageCacheDirOwnership struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"`
		} `json:"setTieredStorageCacheDirOwnership"`
		Tuning struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"`
		} `json:"tuning"`
		ExtraInitContainers string `json:"extraInitContainers"`
	} `json:"initContainers"`
}

func (Statefulset) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "podSecurityContext")
}

type ServiceAccount struct {
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

type Listeners struct {
	Admin          *AdminListeners          `json:"admin" jsonschema:"required"`
	HTTP           *HTTPListeners           `json:"http" jsonschema:"required"`
	Kafka          *KafkaListeners          `json:"kafka" jsonschema:"required"`
	SchemaRegistry *SchemaRegistryListeners `json:"schemaRegistry" jsonschema:"required"`
	RPC            struct {
		Port int          `json:"port" jsonschema:"required"`
		TLS  *ExternalTLS `json:"tls" jsonschema:"required"`
	} `json:"rpc" jsonschema:"required"`
}

type Config struct {
	Cluster              ClusterConfig         `json:"cluster" jsonschema:"required"`
	Node                 NodeConfig            `json:"node" jsonschema:"required"`
	RPK                  map[string]any        `json:"rpk"`
	SchemaRegistryClient *SchemaRegistryClient `json:"schema_registry_client"`
	PandaProxyClient     *PandaProxyClient     `json:"pandaproxy_client"`
	Tunable              *TunableConfig        `json:"tunable" jsonschema:"required"`
}

type JobResources struct {
	Limits struct {
		CPU    any          `json:"cpu" jsonschema:"oneof_type=integer;string"`
		Memory MemoryAmount `json:"memory"`
	} `json:"limits"`
	Requests struct {
		CPU    any          `json:"cpu" jsonschema:"oneof_type=integer;string"`
		Memory MemoryAmount `json:"memory"`
	} `json:"requests"`
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
	// Enabled   bool   `json:"enabled"`
	CAEnabled             bool                    `json:"caEnabled" jsonschema:"required"`
	ApplyInternalDNSNames *bool                   `json:"applyInternalDNSNames"`
	Duration              string                  `json:"duration" jsonschema:"pattern=.*[smh]$"`
	IssuerRef             *cmmeta.ObjectReference `json:"issuerRef"`
	SecretRef             struct {
		Name string `json:"name"`
	} `json:"secretRef"`
}

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

func (TLSCertMap) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.PatternProperties = map[string]*jsonschema.Schema{
		"^[A-Za-z_][A-Za-z0-9_]*$": schema.AdditionalProperties,
	}
	minProps := uint64(1)
	schema.MinProperties = &minProps
	schema.AdditionalProperties = nil
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

type ExternalTLS struct {
	Cert              string `json:"cert" jsonschema:"required"`
	Enabled           bool   `json:"enabled"`
	RequireClientAuth *bool  `json:"requireClientAuth" jsonschema:"required"`
}

type AdminListeners struct {
	External ExternalListeners[AdminExternal] `json:"external"`
	Port     int                              `json:"port" jsonschema:"required"`
	TLS      *ExternalTLS                     `json:"tls" jsonschema:"required"`
}

type AdminExternal struct {
	AdvertisedPorts []int32 `json:"advertisedPorts" jsonschema:"minItems=1"`
	Enabled         *bool   `json:"enabled"`
	Port            int32   `json:"port" jsonschema:"required"`
}

type HTTPListeners struct {
	Enabled              bool                            `json:"enabled" jsonschema:"required"`
	External             ExternalListeners[HTTPExternal] `json:"external"`
	AuthenticationMethod HTTPAuthenticationMethod        `json:"authenticationMethod"`
	TLS                  *ExternalTLS                    `json:"tls" jsonschema:"required"`
	KafkaEndpoint        string                          `json:"kafkaEndpoint" jsonschema:"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$"`
	Port                 int                             `json:"port" jsonschema:"required"`
}

func (HTTPListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

type HTTPExternal struct {
	AdvertisedPorts      []int32                   `json:"advertisedPorts" jsonschema:"minItems=1"`
	Enabled              *bool                     `json:"enabled"`
	Port                 int32                     `json:"port" jsonschema:"required"`
	AuthenticationMethod *HTTPAuthenticationMethod `json:"authenticationMethod"`
	PrefixTemplate       *string                    `json:"prefixTemplate"`
	TLS                  *ExternalTLS              `json:"tls" jsonschema:"required"`
}

func (HTTPExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
	// TODO document me. Legacy matching needs to be removed in a minor bump.
	tls, _ := schema.Properties.Get("tls")
	tls.Required = []string{}
	schema.Required = []string{"port"}
}

type KafkaListeners struct {
	AuthenticationMethod KafkaAuthenticationMethod        `json:"authenticationMethod"`
	External             ExternalListeners[KafkaExternal] `json:"external"`
	TLS                  *ExternalTLS                     `json:"tls" jsonschema:"required"`
	Port                 int                              `json:"port" jsonschema:"required"`
}

func (KafkaListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

type KafkaExternal struct {
	AdvertisedPorts      []int32                    `json:"advertisedPorts" jsonschema:"minItems=1"`
	Enabled              *bool                      `json:"enabled"`
	Port                 int32                      `json:"port" jsonschema:"required"`
	AuthenticationMethod *KafkaAuthenticationMethod `json:"authenticationMethod"`
	PrefixTemplate       *string                     `json:"prefixTemplate"`
}

func (KafkaExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

type SchemaRegistryListeners struct {
	Enabled              bool                                      `json:"enabled" jsonschema:"required"`
	External             ExternalListeners[SchemaRegistryExternal] `json:"external"`
	AuthenticationMethod *HTTPAuthenticationMethod                 `json:"authenticationMethod"`
	KafkaEndpoint        string                                    `json:"kafkaEndpoint" jsonschema:"required,pattern=^[A-Za-z_-][A-Za-z0-9_-]*$"`
	Port                 int                                       `json:"port" jsonschema:"required"`
	TLS                  *ExternalTLS                              `json:"tls" jsonschema:"required"`
}

func (SchemaRegistryListeners) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
}

type SchemaRegistryExternal struct {
	AdvertisedPorts      []int32                   `json:"advertisedPorts" jsonschema:"minItems=1"`
	Enabled              *bool                     `json:"enabled"`
	Port                 int32                     `json:"port"`
	AuthenticationMethod *HTTPAuthenticationMethod `json:"authenticationMethod"`
	TLS                  *ExternalTLS              `json:"tls"`
}

func (SchemaRegistryExternal) JSONSchemaExtend(schema *jsonschema.Schema) {
	makeNullable(schema, "authenticationMethod")
	// TODO this as well
	tls, _ := schema.Properties.Get("tls")
	tls.Required = []string{}
}

type TunableConfig map[string]any

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

type NodeConfig map[string]any

type ClusterConfig map[string]any

type SecretRef struct {
	ConfigurationKey string `json:"configurationKey"`
	Key              string `json:"key"`
	Name             string `json:"name"`
}

type TieredStorageCredentials struct {
	ConfigurationKey string     `json:"configurationKey" jsonschema:"deprecated"`
	Key              string     `json:"key" jsonschema:"deprecated"`
	Name             string     `json:"name" jsonschema:"deprecated"`
	AccessKey        *SecretRef `json:"accessKey"`
	SecretKey        *SecretRef `json:"secretKey"`
}

func (tsc TieredStorageCredentials) IsAccessKeyReferenceValid() bool {
	return tsc.AccessKey != nil && tsc.AccessKey.Name != "" && tsc.AccessKey.Key != ""
}

func (tsc TieredStorageCredentials) IsSecretKeyReferenceValid() bool {
	return tsc.SecretKey != nil && tsc.SecretKey.Name != "" && tsc.SecretKey.Key != ""
}

func (TieredStorageCredentials) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "configurationKey", "key", "name")
}

type TieredStorageConfig map[string]any

func (tsc TieredStorageConfig) IsTieredStorageEnabled() bool {
	if b, ok := tsc["cloud_storage_enabled"]; ok && b.(bool) {
		return true
	}
	return false
}

func (TieredStorageConfig) JSONSchema() *jsonschema.Schema {
	type schema struct {
		CloudStorageEnabled                     bool   `json:"cloud_storage_enabled" jsonschema:"required"`
		CloudStorageAccessKey                   string `json:"cloud_storage_access_key"`
		CloudStorageSecretKey                   string `json:"cloud_storage_secret_key"`
		CloudStorageAPIEndpoint                 string `json:"cloud_storage_api_endpoint"`
		CloudStorageAPIEndpointPort             int    `json:"cloud_storage_api_endpoint_port"`
		CloudStorageAzureADLSEndpoint           string `json:"cloud_storage_azure_adls_endpoint"`
		CloudStorageAzureADLSPort               int    `json:"cloud_storage_azure_adls_port"`
		CloudStorageBucket                      string `json:"cloud_storage_bucket" jsonschema:"required"`
		CloudStorageCacheCheckInterval          int    `json:"cloud_storage_cache_check_interval"`
		CloudStorageCacheDirectory              string `json:"cloud_storage_cache_directory"`
		CloudStorageCacheSize                   any    `json:"cloud_storage_cache_size" jsonschema:"oneof_type=integer;string"`
		CloudStorageCredentialsSource           string `json:"cloud_storage_credentials_source" jsonschema:"pattern=^(config_file|aws_instance_metadata|sts|gcp_instance_metadata)$"`
		CloudStorageDisableTLS                  bool   `json:"cloud_storage_disable_tls"`
		CloudStorageEnableRemoteRead            bool   `json:"cloud_storage_enable_remote_read"`
		CloudStorageEnableRemoteWrite           bool   `json:"cloud_storage_enable_remote_write"`
		CloudStorageInitialBackoffMS            int    `json:"cloud_storage_initial_backoff_ms"`
		CloudStorageManifestUploadTimeoutMS     int    `json:"cloud_storage_manifest_upload_timeout_ms"`
		CloudStorageMaxConnectionIdleTimeMS     int    `json:"cloud_storage_max_connection_idle_time_ms"`
		CloudStorageMaxConnections              int    `json:"cloud_storage_max_connections"`
		CloudStorageReconciliationIntervalMS    int    `json:"cloud_storage_reconciliation_interval_ms"`
		CloudStorageRegion                      string `json:"cloud_storage_region" jsonschema:"required"`
		CloudStorageSegmentMaxUploadIntervalSec int    `json:"cloud_storage_segment_max_upload_interval_sec"`
		CloudStorageSegmentUploadTimeoutMS      int    `json:"cloud_storage_segment_upload_timeout_ms"`
		CloudStorageTrustFile                   string `json:"cloud_storage_trust_file"`
		CloudStorageUploadCtrlDCoeff            int    `json:"cloud_storage_upload_ctrl_d_coeff"`
		CloudStorageUploadCtrlMaxShares         int    `json:"cloud_storage_upload_ctrl_max_shares"`
		CloudStorageUploadCtrlMinShares         int    `json:"cloud_storage_upload_ctrl_min_shares"`
		CloudStorageUploadCtrlPCoeff            int    `json:"cloud_storage_upload_ctrl_p_coeff"`
		CloudStorageUploadCtrlUpdateIntervalMS  int    `json:"cloud_storage_upload_ctrl_update_interval_ms"`
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
