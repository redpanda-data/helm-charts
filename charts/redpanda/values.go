// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_values.go.tpl
package redpanda

import (
	"fmt"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/invopop/jsonschema"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/charts/connectors"
	"github.com/redpanda-data/redpanda-operator/charts/console"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

const (
	fiveGiB = 5368709120
	// That default path inside Redpanda container which is based on debian.
	defaultTruststorePath = "/etc/ssl/certs/ca-certificates.crt"

	// RedpandaContainerName is the user facing name of the redpanda container
	// in the redpanda StatefulSet. While the name of the container can
	// technically change, this is the name that is used to locate the
	// [corev1.Container] that will be smp'd into the redpanda container.
	RedpandaContainerName = "redpanda"
	// PostUpgradeContainerName is the user facing name of the post-install
	// job's container.
	PostInstallContainerName = "post-install"
	// PostUpgradeContainerName is the user facing name of the post-upgrade
	// job's container.
	PostUpgradeContainerName = "post-upgrade"
	// RedpandaControllersContainerName is the container that can perform day
	// 2 operation similarly to Redpanda operator.
	RedpandaControllersContainerName = "redpanda-controllers"

	// certificateMountPoint is a common mount point for any TLS certificate
	// defined as external truststore or as certificate that would be
	// created by cert-manager.
	certificateMountPoint = "/etc/tls/certs"
)

type MebiBytes = int64

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
	Console          console.PartialValues         `json:"console,omitempty"`
	Connectors       connectors.PartialValues      `json:"connectors"`
	Auth             Auth                          `json:"auth"`
	TLS              TLS                           `json:"tls"`
	External         ExternalConfig                `json:"external"`
	Logging          Logging                       `json:"logging"`
	Monitoring       Monitoring                    `json:"monitoring"`
	Resources        RedpandaResources             `json:"resources"`
	Storage          Storage                       `json:"storage"`
	PostInstallJob   PostInstallJob                `json:"post_install_job"`
	Statefulset      Statefulset                   `json:"statefulset"`
	ServiceAccount   ServiceAccountCfg             `json:"serviceAccount"`
	RBAC             RBAC                          `json:"rbac"`
	Tuning           Tuning                        `json:"tuning"`
	Listeners        Listeners                     `json:"listeners"`
	Config           Config                        `json:"config"`
	Tests            *struct {
		Enabled bool `json:"enabled"`
	} `json:"tests"`
	Force bool `json:"force"`
}

// +gotohelm:ignore=true
func (Values) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "license_key", "license_secret_ref")
}

// SecurityContext is a legacy mishmash of [corev1.PodSecurityContext] and
// [corev1.SecurityContext]. It's type exists for backwards compat purposes
// only.
type SecurityContext struct {
	RunAsUser                *int64 `json:"runAsUser"`
	RunAsGroup               *int64 `json:"runAsGroup"`
	AllowPrivilegeEscalation *bool  `json:"allowPrivilegeEscalation"`
	// AllowPriviledgeEscalation is typoed version of
	// [SecurityContext.AllowPrivilegeEscalation]. It's respected for backwards
	// compatibility.
	// Deprecated: Prefer AllowPrivilegeEscalation.
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

	users := []string{a.SASL.BootstrapUser.Username()}
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
	EnableHTTP2    *bool                   `json:"enableHttp2"`
}

// RedpandaResources encapsulates the calculation of the redpanda container's
// [corev1.ResourceRequirements] and parameters such as `--memory`,
// `--reserve-memory`, and `--smp`.
// This calculation supports two modes:
//
//   - Explicit mode (recommended):  Activated when `Limits` and `Requests` are
//     set. In this mode, the CLI flags are calculated directly based on the
//     provided `Limits` and `Requests`. This mode ensures predictable resource
//     allocation and is recommended for production environments. If additional
//     tuning is required, the CLI flags can be manually overridden using
//     `statefulset.additionalRedpandaCmdFlags`.
//
//   - Legacy mode (default): Used when `Limits` and `Requests` are not set.
//     In this mode, the container resources and CLI flags are calculated using
//     built-in default logic, where 80% of the container's memory is allocated
//     to Redpanda and the rest is reserved for system overhead. Legacy mode is
//     intended for backward compatibility and less controlled environments.
//
// Explicit mode offers better control and aligns with Kubernetes best
// practices. Legacy mode is a fallback for users who have not defined `Limits`
// and `Requests`.
type RedpandaResources struct {
	Limits   *corev1.ResourceList `json:"limits,omitempty"`
	Requests *corev1.ResourceList `json:"requests,omitempty"`

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

func (rr *RedpandaResources) GetResourceRequirements() corev1.ResourceRequirements {
	// If Limits and Requests are specified, use them as is.
	if rr.Limits != nil && rr.Requests != nil {
		return corev1.ResourceRequirements{
			Limits:   *rr.Limits,
			Requests: *rr.Requests,
		}
	}

	// Otherwise fallback to the historical behavior.
	reqs := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    rr.CPU.Cores,
			"memory": rr.Memory.Container.Max,
		},
	}

	if rr.Memory.Container.Min != nil {
		reqs.Requests = corev1.ResourceList{
			"cpu":    rr.CPU.Cores,
			"memory": *rr.Memory.Container.Min,
		}
	}

	return reqs
}

func (rr *RedpandaResources) GetRedpandaFlags() map[string]string {
	flags := map[string]string{
		"--reserve-memory": fmt.Sprintf("%dM", rr.reserveMemory()),
	}

	if smp := rr.smp(); smp != nil {
		flags["--smp"] = fmt.Sprintf("%d", int64(*smp))
	}

	if memory := rr.memory(); memory != nil {
		flags["--memory"] = fmt.Sprintf("%dM", int64(*memory))
	}

	// Only set lock-memory if Limits and Requests are NOT specified. It should
	// otherwise be set through additionalRedpandaCmdFlags.
	if rr.Limits == nil && rr.Requests == nil {
		flags["--lock-memory"] = fmt.Sprintf("%v", ptr.Deref(rr.Memory.EnableMemoryLocking, false))
	}

	if rr.GetOverProvisionValue() {
		flags["--overprovisioned"] = ""
	}

	return flags
}

func (rr *RedpandaResources) GetOverProvisionValue() bool {
	if rr.Limits != nil && rr.Requests != nil {
		// Get CPU prioritizing requests, falling back to limits if not
		// specified as kube-scheduler does.
		cpuReq, ok := (*rr.Requests)[corev1.ResourceCPU]
		if !ok {
			cpuReq, ok = (*rr.Limits)[corev1.ResourceCPU]
		}

		// If redpanda has been allocated less than 1 full CPU, set
		// overprovisioned to true.
		if ok && cpuReq.MilliValue() < 1000 {
			return true
		}
		return false
	}

	if rr.CPU.Cores.MilliValue() < 1000 {
		return true
	}

	return ptr.Deref(rr.CPU.Overprovisioned, false)
}

func (rr *RedpandaResources) smp() *int64 {
	if rr.Limits != nil && rr.Requests != nil {
		// Get CPU prioritizing requests, falling back to limits if not
		// specified as kube-scheduler does. This ordering also forces --smp to
		// be <= the containers CPU limits. The other way around (limits
		// fallback to requests; therefore --smp >= CPU limits) isn't useful.
		cpuReq, ok := (*rr.Requests)[corev1.ResourceCPU]
		if !ok {
			cpuReq, ok = (*rr.Limits)[corev1.ResourceCPU]
		}

		// If neither requests nor limits are set, don't set --smp.
		if !ok {
			return nil
		}

		// If CPU limits/requests are defined, set --smp to max(1, floor(cpu)).
		//
		// Due to redpanda/seastar's per core model, we can't do much with
		// fractional CPU values we need to round either up or down. Rounding
		// up would result in utilizing too much CPU from the CRI perspective
		// and cause throttling, so we round down and potentially waste some
		// quota.
		smp := cpuReq.MilliValue() / 1000
		if smp < 1 {
			smp = 1
		}
		return ptr.To(smp)
	}

	if coresInMillies := rr.CPU.Cores.MilliValue(); coresInMillies < 1000 {
		return ptr.To(int64(1))
	}
	return ptr.To(int64(rr.CPU.Cores.Value()))
}

// memory returns the amount of memory for Redpanda process. It should be
// passed to the `--memory` argument of the Redpanda process, see
// RedpandaAdditionalStartFlags and rpk redpanda start documentation.
//
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func (rr *RedpandaResources) memory() *MebiBytes {
	if rr.Limits != nil && rr.Requests != nil {
		// `--memory` will be set to something < the container's
		// resources.memory.limits value.
		// We want to allocate seastar < memory than our limit for several reasons:
		// 1. Seastar may slightly exceed this limit due to page tables and
		//    non-heap memory that's still accounted by cgroups.
		// 2. resources.limits.memory applies to the entire container. We want
		//    to keep headroom to allow exec'ing into the container and for any
		//    exec probes.
		// 3. emptyDir's storage is counted against the container's memory
		//    limits. We use these to store rendered versions of config files
		//    and therefore need to account for them.
		//    https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#memory-backed-emptydir
		// The memory reservation is done by subtracting from `--memory` and
		// always setting `--reserve-memory` to 0 rather than setting
		// `--reserve-memory`. This is an easier mental model to follow as the
		// `--reserve-memory` flag is exceptionally nuanced in practice and is
		// meant to aid seastar running on an entire VM rather than in a
		// container.

		// If either memory limit or requests are set, we take the minimum of
		// the two (relying on the invariant enforced by Kubernetes that
		// requests must be <= limits).
		memReq, ok := (*rr.Requests)[corev1.ResourceMemory]
		if !ok {
			memReq, ok = (*rr.Limits)[corev1.ResourceMemory]
		}

		// If neither requests nor limits are set, don't set --memory.
		if !ok {
			return nil
		}

		// Here we perform our memory reservation. Historically, 80% of
		// container memory was provided to Redpanda and some additional amount
		// was removed due to the usage of `--reserve-memory`. We're largely
		// blind to the lower limit of what's tolerated.
		// This calculation, therefore, is a complete split ball. We expect
		// this to change over time; ideally trending towards providing
		// redpanda more memory.
		// We intentionally err on the conservative side as we'd prefer to
		// "waste" a few megs of memory rather than risking OOM kills.
		// For simplicity, we're using a % as a static reservation would need
		// to handle weird edge cases.
		//
		// redpanda get's 90% of the container limit. (It's better than the
		// historic 80%).
		memory := int64(float64(memReq.Value()) * 0.90)

		// Cast to Membibytes.
		return ptr.To(memory / (1024 * 1024))
	}

	// Below we perform the calculations for the legacy resource mode. This
	// calculation appears to be based on an incorrect understanding of the
	// (admittedly convoluted) `--reserve-memory` seastar flag and is preserved
	// solely for backwards compatibility.
	//
	// It segments out memory for:
	// * Seastar/Redpanda (`--memory`) - .Memory.Redpanda OR 80% of memory
	// * Seastar's "subsystem" (`--reserve-memory`) - .Memory.Reserve OR 200Mi + 0.2% of memory
	// * Container processes (execing, hooks, probes, etc) - The leftovers from the above (if any)
	memory := int64(0)
	containerMemory := rr.containerMemory()

	if rpMem := rr.Memory.Redpanda; rpMem != nil && rpMem.Memory != nil {
		memory = rpMem.Memory.Value() / (1024 * 1024)
	} else {
		memory = int64(float64(containerMemory) * 0.8)
	}

	if memory == 0 {
		panic("unable to get memory value redpanda-memory")
	}

	if memory < 256 {
		panic(fmt.Sprintf("%d is below the minimum value for Redpanda", memory))
	}

	// NB: int64's are working around a bug in gotohelm's BinaryExpr detection
	// with Alias types.
	if memory+int64(rr.reserveMemory()) > containerMemory {
		panic(fmt.Sprintf("Not enough container memory for Redpanda memory values where Redpanda: %d, reserve: %d, container: %d", memory, rr.reserveMemory(), containerMemory))
	}

	return ptr.To(memory)
}

//	reserveMemory returns the amount of memory that the Redpanda process will
//	not use from the provided value in `--memory` or from the internal Redpanda
//	discovery process. It should be passed to the `--reserve-memory` argument
//	of the Redpanda process, see RedpandaAdditionalStartFlags and rpk redpanda
//	start documentation.
//
// https://docs.redpanda.com/current/reference/rpk/rpk-redpanda/rpk-redpanda-start/
func (rr *RedpandaResources) reserveMemory() MebiBytes {
	if rr.Limits != nil && rr.Requests != nil {
		// See [RedpandaResources.memory] for details here.
		return 0
	}

	// See [RedpandaResources.memory] for details here.
	if rpMem := rr.Memory.Redpanda; rpMem != nil && rpMem.ReserveMemory != nil {
		return rpMem.ReserveMemory.Value() / (1024 * 1024)
	}

	return int64(float64(rr.containerMemory())*0.002) + 200
}

// containerMemory returns either the min or max container memory values as an
// integer value of MembiBytes.
func (rr *RedpandaResources) containerMemory() MebiBytes {
	if rr.Memory.Container.Min != nil {
		return rr.Memory.Container.Min.Value() / (1024 * 1024)
	}

	return rr.Memory.Container.Max.Value() / (1024 * 1024)
}

type Storage struct {
	HostPath         string `json:"hostPath" jsonschema:"required"`
	Tiered           Tiered `json:"tiered" jsonschema:"required"`
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

// was: storage-tiered-hostpath
func (s *Storage) GetTieredStorageHostPath() string {
	hp := s.TieredStorageHostPath
	if helmette.Empty(hp) {
		hp = s.Tiered.HostPath
	}
	if helmette.Empty(hp) {
		panic(fmt.Sprintf(`storage.tiered.mountType is "%s" but storage.tiered.hostPath is empty`,
			s.Tiered.MountType,
		))
	}
	return hp
}

// TieredCacheDirectory was: tieredStorage.cacheDirectory
func (s *Storage) TieredCacheDirectory(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if dir, ok := values.Config.Node["cloud_storage_cache_directory"].(string); ok {
		return dir
	}

	// TODO: Deprecate or just remove the ability to set
	// cloud_storage_cache_directory in tiered config(s) so their reserved for
	// cluster settings only.
	tieredConfig := values.Storage.GetTieredStorageConfig()
	if dir, ok := tieredConfig["cloud_storage_cache_directory"].(string); ok {
		return dir
	}

	return "/var/lib/redpanda/data/cloud_storage_cache"
}

// TieredMountType was: storage-tiered-mountType
func (s *Storage) TieredMountType() string {
	if s.TieredStoragePersistentVolume != nil && s.TieredStoragePersistentVolume.Enabled {
		return "persistentVolume"
	}
	if !helmette.Empty(s.TieredStorageHostPath) {
		// XXX type is declared as string, but it's being used as a bool
		// This needs some care since transpilation fails with a `!= ""` check,
		// missing null values.
		return "hostPath"
	}
	return s.Tiered.MountType
}

// Storage.TieredPersistentVolumeLabels was storage-tiered-persistentVolume.labels
// support legacy storage.tieredStoragePersistentVolume
func (s *Storage) TieredPersistentVolumeLabels() map[string]string {
	if s.TieredStoragePersistentVolume != nil {
		return s.TieredStoragePersistentVolume.Labels
	}
	return s.Tiered.PersistentVolume.Labels
}

// Storage.TieredPersistentVolumeAnnotations was storage-tiered-persistentVolume.annotations
// support legacy storage.tieredStoragePersistentVolume
func (s *Storage) TieredPersistentVolumeAnnotations() map[string]string {
	if s.TieredStoragePersistentVolume != nil {
		return s.TieredStoragePersistentVolume.Annotations
	}
	return s.Tiered.PersistentVolume.Annotations
}

// storage.TieredPersistentVolumeStorageClass was storage-tiered-persistentVolume.storageClass
// support legacy storage.tieredStoragePersistentVolume
func (s *Storage) TieredPersistentVolumeStorageClass() string {
	if s.TieredStoragePersistentVolume != nil {
		return s.TieredStoragePersistentVolume.StorageClass
	}
	return s.Tiered.PersistentVolume.StorageClass
}

// +gotohelm:ignore=true
func (Storage) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "tieredConfig", "persistentVolume", "tieredStorageHostPath", "tieredStoragePersistentVolume")

	// TODO note why we do this.
	tieredConfig, _ := schema.Properties.Get("tieredConfig")
	tieredConfig.Required = []string{}
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
	Resources   *corev1.ResourceRequirements `json:"resources"`
	Affinity    corev1.Affinity              `json:"affinity"`
	Enabled     bool                         `json:"enabled"`
	Labels      map[string]string            `json:"labels"`
	Annotations map[string]string            `json:"annotations"`
	// Deprecated. Prefer [PodTemplate.Spec.SecurityContext].
	SecurityContext *corev1.SecurityContext `json:"securityContext"`
	PodTemplate     PodTemplate             `json:"podTemplate"`
}

type PodTemplate struct {
	Labels      map[string]string                      `json:"labels,omitempty" jsonschema:"required"`
	Annotations map[string]string                      `json:"annotations,omitempty" jsonschema:"required"`
	Spec        *applycorev1.PodSpecApplyConfiguration `json:"spec,omitempty"`
}

type Statefulset struct {
	AdditionalSelectorLabels map[string]string `json:"additionalSelectorLabels" jsonschema:"required"`
	NodeAffinity             map[string]any    `json:"nodeAffinity"`
	Replicas                 int32             `json:"replicas" jsonschema:"required"`
	UpdateStrategy           struct {
		Type string `json:"type" jsonschema:"required,pattern=^(RollingUpdate|OnDelete)$"`
	} `json:"updateStrategy" jsonschema:"required"`
	AdditionalRedpandaCmdFlags []string `json:"additionalRedpandaCmdFlags"`
	// Annotations are used only for `Statefulset.spec.template.metadata.annotations`. The StatefulSet does not have
	// any dedicated annotation.
	Annotations map[string]string `json:"annotations" jsonschema:"deprecated"`
	PodTemplate PodTemplate       `json:"podTemplate" jsonschema:"required"`
	Budget      struct {
		MaxUnavailable int32 `json:"maxUnavailable" jsonschema:"required"`
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
		Weight      int32          `json:"weight" jsonschema:"required"`
		Custom      map[string]any `json:"custom"`
	} `json:"podAntiAffinity" jsonschema:"required"`
	NodeSelector                  map[string]string `json:"nodeSelector" jsonschema:"required"`
	PriorityClassName             string            `json:"priorityClassName" jsonschema:"required"`
	TerminationGracePeriodSeconds int64             `json:"terminationGracePeriodSeconds"`
	TopologySpreadConstraints     []struct {
		MaxSkew           int32                                `json:"maxSkew"`
		TopologyKey       string                               `json:"topologyKey"`
		WhenUnsatisfiable corev1.UnsatisfiableConstraintAction `json:"whenUnsatisfiable" jsonschema:"pattern=^(ScheduleAnyway|DoNotSchedule)$"`
	} `json:"topologySpreadConstraints" jsonschema:"required,minItems=1"`
	Tolerations []corev1.Toleration `json:"tolerations" jsonschema:"required"`
	// Deprecated. Prefer [PodTemplate.Spec.SecurityContext].
	PodSecurityContext *SecurityContext `json:"podSecurityContext"`
	// Deprecated. Prefer [PodTemplate.Spec.Containers[*].SecurityContext].
	SecurityContext SecurityContext `json:"securityContext" jsonschema:"required"`
	SideCars        struct {
		ConfigWatcher struct {
			Enabled           bool                    `json:"enabled"`
			ExtraVolumeMounts string                  `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
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
			PprofAddress       string                  `json:"pprofAddress"`
			Run                []string                `json:"run"`
		} `json:"controllers"`
	} `json:"sideCars" jsonschema:"required"`
	ExtraVolumes      string `json:"extraVolumes"`      // XXX this is template-expanded into yaml
	ExtraVolumeMounts string `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
	InitContainers    struct {
		Configurator struct {
			ExtraVolumeMounts string         `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
			Resources         map[string]any `json:"resources"`
		} `json:"configurator"`
		FSValidator struct {
			Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
			ExpectedFS        string         `json:"expectedFS"`
		} `json:"fsValidator"`
		SetDataDirOwnership struct {
			Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
		} `json:"setDataDirOwnership"`
		SetTieredStorageCacheDirOwnership struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
		} `json:"setTieredStorageCacheDirOwnership"`
		Tuning struct {
			// Enabled           bool           `json:"enabled"`
			Resources         map[string]any `json:"resources"`
			ExtraVolumeMounts string         `json:"extraVolumeMounts"` // XXX this is template-expanded into yaml
		} `json:"tuning"`
		ExtraInitContainers string `json:"extraInitContainers"` // XXX this is template-expanded into yaml
	} `json:"initContainers"`
	InitContainerImage struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"initContainerImage"`
}

// +gotohelm:ignore=true
func (Statefulset) JSONSchemaExtend(schema *jsonschema.Schema) {
	deprecate(schema, "podSecurityContext", "securityContext")
}

type ServiceAccountCfg struct {
	Annotations                  map[string]string `json:"annotations" jsonschema:"required"`
	AutomountServiceAccountToken *bool             `json:"automountServiceAccountToken,omitempty"`
	Create                       bool              `json:"create" jsonschema:"required"`
	Name                         string            `json:"name" jsonschema:"required"`
}

type RBAC struct {
	Enabled     bool              `json:"enabled" jsonschema:"required"`
	Annotations map[string]string `json:"annotations" jsonschema:"required"`
}

type Tuning struct {
	TuneAIOEvents   bool   `json:"tune_aio_events,omitempty"`
	TuneClocksource bool   `json:"tune_clocksource,omitempty"`
	TuneBallastFile bool   `json:"tune_ballast_file,omitempty"`
	BallastFilePath string `json:"ballast_file_path,omitempty"`
	BallastFileSize string `json:"ballast_file_size,omitempty"`
	WellKnownIO     string `json:"well_known_io,omitempty"`
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

func (l *Listeners) CreateSeedServers(replicas int32, fullname, internalDomain string) []map[string]any {
	var result []map[string]any
	for i := int32(0); i < replicas; i++ {
		result = append(result, map[string]any{
			"host": map[string]any{
				"address": fmt.Sprintf("%s-%d.%s", fullname, i, internalDomain),
				"port":    l.RPC.Port,
			},
		})
	}
	return result
}

func (l *Listeners) AdminList(replicas int32, fullname, internalDomain string) []string {
	return ServerList(replicas, "", fullname, internalDomain, l.Admin.Port)
}

func (l *Listeners) SchemaRegistryList(replicas int32, fullname, internalDomain string) []string {
	return ServerList(replicas, "", fullname, internalDomain, l.SchemaRegistry.Port)
}

func ServerList(replicas int32, prefix, fullname, internalDomain string, port int32) []string {
	var result []string
	for i := int32(0); i < replicas; i++ {
		result = append(result, fmt.Sprintf("%s%s-%d.%s:%d", prefix, fullname, i, internalDomain, int(port)))
	}
	return result
}

// TrustStoreVolume returns a [corev1.Volume] containing a projected volume
// that mounts all required truststore files. If no truststores are configured,
// it returns nil.
func (l *Listeners) TrustStoreVolume(tls *TLS) *corev1.Volume {
	cmSources := map[string][]corev1.KeyToPath{}
	secretSources := map[string][]corev1.KeyToPath{}

	for _, ts := range l.TrustStores(tls) {
		projection := ts.VolumeProjection()

		if projection.Secret != nil {
			secretSources[projection.Secret.Name] = append(secretSources[projection.Secret.Name], projection.Secret.Items...)
		} else {
			cmSources[projection.ConfigMap.Name] = append(cmSources[projection.ConfigMap.Name], projection.ConfigMap.Items...)
		}
	}

	var sources []corev1.VolumeProjection

	for _, name := range helmette.SortedKeys(cmSources) {
		keys := cmSources[name]
		sources = append(sources, corev1.VolumeProjection{
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Items: dedupKeyToPaths(keys),
			},
		})
	}

	for _, name := range helmette.SortedKeys(secretSources) {
		keys := secretSources[name]
		sources = append(sources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Items: dedupKeyToPaths(keys),
			},
		})
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

func dedupKeyToPaths(items []corev1.KeyToPath) []corev1.KeyToPath {
	// NB: This logic is a non-idiomatic fashion to dance around suspected
	// limitations in gotohelm.

	seen := map[string]bool{}
	var deduped []corev1.KeyToPath

	for _, item := range items {
		if _, ok := seen[item.Key]; ok {
			continue
		}

		deduped = append(deduped, item)
		seen[item.Key] = true
	}

	return deduped
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
	ClientSecretRef       *corev1.LocalObjectReference `json:"clientSecretRef"`
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

type BootstrapUser struct {
	Name         *string                   `json:"name"`
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef"`
	Password     *string                   `json:"password"`
	Mechanism    string                    `json:"mechanism" jsonschema:"pattern=^(SCRAM-SHA-512|SCRAM-SHA-256)$"`
}

func (b *BootstrapUser) BootstrapEnvironment(fullname string) []corev1.EnvVar {
	return append(b.RpkEnvironment(fullname), corev1.EnvVar{
		Name:  "RP_BOOTSTRAP_USER",
		Value: "$(RPK_USER):$(RPK_PASS):$(RPK_SASL_MECHANISM)",
	})
}

func (b *BootstrapUser) Username() string {
	if b.Name != nil {
		return *b.Name
	}
	return "kubernetes-controller"
}

func (b *BootstrapUser) RpkEnvironment(fullname string) []corev1.EnvVar {
	return []corev1.EnvVar{{
		Name: "RPK_PASS",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: b.SecretKeySelector(fullname),
		},
	}, {
		Name:  "RPK_USER",
		Value: b.Username(),
	}, {
		Name:  "RPK_SASL_MECHANISM",
		Value: b.GetMechanism(),
	}}
}

func (b *BootstrapUser) GetMechanism() string {
	if b.Mechanism == "" {
		return "SCRAM-SHA-256"
	}
	return b.Mechanism
}

func (b *BootstrapUser) SecretKeySelector(fullname string) *corev1.SecretKeySelector {
	if b.SecretKeyRef != nil {
		return b.SecretKeyRef
	}

	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: fmt.Sprintf("%s-bootstrap-user", fullname),
		},
		Key: "password",
	}
}

type SASLUser struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Mechanism string `json:"mechanism" jsonschema:"pattern=^(SCRAM-SHA-512|SCRAM-SHA-256)$"`
}

type SASLAuth struct {
	Enabled       bool          `json:"enabled" jsonschema:"required"`
	Mechanism     string        `json:"mechanism"`
	SecretRef     string        `json:"secretRef"`
	Users         []SASLUser    `json:"users"`
	BootstrapUser BootstrapUser `json:"bootstrapUser"`
}

type TrustStore struct {
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef"`
	SecretKeyRef    *corev1.SecretKeySelector    `json:"secretKeyRef"`
}

// +gotohelm:ignore=true
func (TrustStore) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.MaxProperties = ptr.To[uint64](1)
	schema.MinProperties = ptr.To[uint64](1)
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
		return fmt.Sprintf("%s/%s/ca.crt", certificateMountPoint, t.Cert)
	}

	return defaultTruststorePath
}

// ServerCAPath returns the path on disk to a certificate that may be used to
// verify a connection with this server.
func (t *InternalTLS) ServerCAPath(tls *TLS) string {
	if tls.Certs.MustGet(t.Cert).CAEnabled {
		return fmt.Sprintf("%s/%s/ca.crt", certificateMountPoint, t.Cert)
	}
	// Strange but technically correct, if CAEnabled is false, we can't safely
	// assume that a ca.crt file will exist. So we fallback to using the
	// server's certificate itself.
	// Other options would be: failing or falling back to the container's
	// default truststore.
	return fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, t.Cert)
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
		return fmt.Sprintf("%s/%s/ca.crt", certificateMountPoint, t.GetCertName(i))
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
	External    ExternalListeners[AdminExternal] `json:"external"`
	Port        int32                            `json:"port" jsonschema:"required"`
	AppProtocol *string                          `json:"appProtocol,omitempty"`
	TLS         InternalTLS                      `json:"tls" jsonschema:"required"`
}

// ConsoleTLS is a struct that represents TLS configuration used
// in console configuration in Kafka, Schema Registry and
// Redpanda Admin API.
// For the above configuration helm chart could import struct, but
// as of the writing the struct fields tag have only `yaml` annotation.
// `sigs.k8s.io/yaml` requires `json` tags.
type ConsoleTLS struct {
	Enabled               bool   `json:"enabled"`
	CaFilepath            string `json:"caFilepath"`
	CertFilepath          string `json:"certFilepath"`
	KeyFilepath           string `json:"keyFilepath"`
	InsecureSkipTLSVerify bool   `json:"insecureSkipTlsVerify"`
}

func (l *AdminListeners) ConsoleTLS(tls *TLS) ConsoleTLS {
	t := ConsoleTLS{Enabled: l.TLS.IsEnabled(tls)}
	if !t.Enabled {
		return t
	}

	adminAPIPrefix := fmt.Sprintf("%s/%s", certificateMountPoint, l.TLS.Cert)

	// Strange but technically correct, if CAEnabled is false, we can't safely
	// assume that a ca.crt file will exist. So we fallback to using the
	// server's certificate itself.
	// Other options would be: failing or falling back to the container's
	// default truststore.
	if tls.Certs.MustGet(l.TLS.Cert).CAEnabled {
		t.CaFilepath = fmt.Sprintf("%s/ca.crt", adminAPIPrefix)
	} else {
		t.CaFilepath = fmt.Sprintf("%s/tls.crt", adminAPIPrefix)
	}

	if !l.TLS.RequireClientAuth {
		return t
	}

	t.CertFilepath = fmt.Sprintf("%s/tls.crt", adminAPIPrefix)
	t.KeyFilepath = fmt.Sprintf("%s/tls.key", adminAPIPrefix)

	return t
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
			"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, certName),
			"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, certName),
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

	for _, key := range helmette.SortedKeys(l.External) {
		lis := l.External[key]
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
			"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, certName),
			"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, certName),
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

	for _, key := range helmette.SortedKeys(l.External) {
		lis := l.External[key]
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
			"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, certName),
			"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, certName),
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

	for _, key := range helmette.SortedKeys(l.External) {
		lis := l.External[key]
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

func (l *KafkaListeners) ConsoleTLS(tls *TLS) ConsoleTLS {
	t := ConsoleTLS{Enabled: l.TLS.IsEnabled(tls)}
	if !t.Enabled {
		return t
	}

	kafkaPathPrefix := fmt.Sprintf("%s/%s", certificateMountPoint, l.TLS.Cert)

	// Strange but technically correct, if CAEnabled is false, we can't safely
	// assume that a ca.crt file will exist. So we fallback to using the
	// server's certificate itself.
	// Other options would be: failing or falling back to the container's
	// default truststore.
	if tls.Certs.MustGet(l.TLS.Cert).CAEnabled {
		t.CaFilepath = fmt.Sprintf("%s/ca.crt", kafkaPathPrefix)
	} else {
		t.CaFilepath = fmt.Sprintf("%s/tls.crt", kafkaPathPrefix)
	}

	if !l.TLS.RequireClientAuth {
		return t
	}

	t.CertFilepath = fmt.Sprintf("%s/tls.crt", kafkaPathPrefix)
	t.KeyFilepath = fmt.Sprintf("%s/tls.key", kafkaPathPrefix)

	return t
}

func (l *KafkaListeners) ConnectorsTLS(tls *TLS, fullName string) connectors.TLS {
	t := connectors.TLS{Enabled: l.TLS.IsEnabled(tls)}
	if !t.Enabled {
		return t
	}

	t.CA = struct {
		SecretRef           string `json:"secretRef"`
		SecretNameOverwrite string `json:"secretNameOverwrite"`
	}{SecretRef: fmt.Sprintf("%s-default-cert", fullName)}

	return t
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

func (l *SchemaRegistryListeners) Listeners(saslEnabled bool) []map[string]any {
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
			"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, certName),
			"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, certName),
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

	for _, key := range helmette.SortedKeys(l.External) {
		lis := l.External[key]
		if !lis.IsEnabled() || !lis.TLS.IsEnabled(&l.TLS, tls) || lis.TLS.TrustStore == nil {
			continue
		}
		tss = append(tss, lis.TLS.TrustStore)

	}

	return tss
}

func (l *SchemaRegistryListeners) ConsoleTLS(tls *TLS) ConsoleTLS {
	t := ConsoleTLS{Enabled: l.TLS.IsEnabled(tls)}
	if !t.Enabled {
		return t
	}

	schemaRegistryPrefix := fmt.Sprintf("%s/%s", certificateMountPoint, l.TLS.Cert)

	// Strange but technically correct, if CAEnabled is false, we can't safely
	// assume that a ca.crt file will exist. So we fallback to using the
	// server's certificate itself.
	// Other options would be: failing or falling back to the container's
	// default truststore.
	if tls.Certs.MustGet(l.TLS.Cert).CAEnabled {
		t.CaFilepath = fmt.Sprintf("%s/ca.crt", schemaRegistryPrefix)
	} else {
		t.CaFilepath = fmt.Sprintf("%s/tls.crt", schemaRegistryPrefix)
	}

	if !l.TLS.RequireClientAuth {
		return t
	}

	t.CertFilepath = fmt.Sprintf("%s/tls.crt", schemaRegistryPrefix)
	t.KeyFilepath = fmt.Sprintf("%s/tls.key", schemaRegistryPrefix)

	return t
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
			} else if helmette.KindIs("bool", v) {
				result[k] = v
			} else {
				result[k] = helmette.ToYaml(v)
			}
		}
	}

	return result
}

type ClusterConfig map[string]any

func (c *ClusterConfig) Translate() map[string]any {
	result := map[string]any{}

	for k, v := range *c {
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
	// ConfigurationKey is never read.
	ConfigurationKey string `json:"configurationKey"`
	Key              string `json:"key"`
	Name             string `json:"name"`
}

func (sr *SecretRef) AsSource() *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: sr.Name},
			Key:                  sr.Key,
		},
	}
}

// IsValid confirms whether EnvVarSource could be built from
// SecretRef.
func (sr *SecretRef) IsValid() bool {
	return sr != nil && !helmette.Empty(sr.Key) && !helmette.Empty(sr.Name)
}

type TieredStorageCredentials struct {
	AccessKey *SecretRef `json:"accessKey"`
	SecretKey *SecretRef `json:"secretKey"`
}

func (tsc *TieredStorageCredentials) AsEnvVars(config TieredStorageConfig) []corev1.EnvVar {
	// Environment variables will only respected if their corresponding keys
	// are not explicitly set. This is historical behavior and is largely an
	// implementation details than an explicitly choice.
	_, hasAccessKey := config["cloud_storage_access_key"]
	_, hasSecretKey := config["cloud_storage_secret_key"]
	_, hasSharedKey := config["cloud_storage_azure_shared_key"]

	var envvars []corev1.EnvVar

	if !hasAccessKey && tsc.AccessKey.IsValid() {
		envvars = append(envvars, corev1.EnvVar{
			Name:      "REDPANDA_CLOUD_STORAGE_ACCESS_KEY",
			ValueFrom: tsc.AccessKey.AsSource(),
		})
	}

	if tsc.SecretKey.IsValid() {
		if !hasSecretKey && !config.HasAzureCanaries() {
			envvars = append(envvars, corev1.EnvVar{
				Name:      "REDPANDA_CLOUD_STORAGE_SECRET_KEY",
				ValueFrom: tsc.SecretKey.AsSource(),
			})
		} else if !hasSharedKey && config.HasAzureCanaries() {
			envvars = append(envvars, corev1.EnvVar{
				Name:      "REDPANDA_CLOUD_STORAGE_AZURE_SHARED_KEY",
				ValueFrom: tsc.SecretKey.AsSource(),
			})
		}
	}

	return envvars
}

type TieredStorageConfig map[string]any

// HasAzureCanaries returns true if this configuration has keys set that would
// indicate the configuration is for a MSFT Azure environment.
//
// If true, [TieredStorageCredentials.SecretKey] should be treated as the value
// for `cloud_storage_azure_shared_key` instead of `cloud_storage_secret_key`.
func (c TieredStorageConfig) HasAzureCanaries() bool {
	_, containerExists := c["cloud_storage_azure_container"]
	_, accountExists := c["cloud_storage_azure_storage_account"]
	return containerExists && accountExists
}

func (c TieredStorageConfig) CloudStorageCacheSize() *resource.Quantity {
	value, ok := c[`cloud_storage_cache_size`]
	if !ok {
		return nil
	}
	return ptr.To(helmette.UnmarshalInto[resource.Quantity](value))
}

// Translate converts TieredStorageConfig into a map suitable for use in
// an unexpanded `.bootstrap.yaml`.
func (c TieredStorageConfig) Translate(creds *TieredStorageCredentials) map[string]any {
	// Clone ourselves as we're making changes.
	config := helmette.Merge(map[string]any{}, c)

	// For any values that can be specified as secrets and do not have explicit
	// values, inject placeholders into config which will be replaced with
	// `envsubst` in an initcontainer.
	for _, envvar := range creds.AsEnvVars(c) {
		key := helmette.Lower(envvar.Name[len("REDPANDA_"):])
		// NB: No string + string support in gotohelm.
		config[key] = fmt.Sprintf("$%s", envvar.Name)
	}

	// Expand cloud_storage_cache_size, if provided, as it can be specified as
	// a resource.Quantity.
	if size := c.CloudStorageCacheSize(); size != nil {
		config["cloud_storage_cache_size"] = size.Value()
	}

	return config
}

// +gotohelm:ignore=true
func (TieredStorageConfig) JSONSchema() *jsonschema.Schema {
	type schema struct {
		CloudStorageEnabled            bool   `json:"cloud_storage_enabled" jsonschema:"required"`
		CloudStorageAccessKey          string `json:"cloud_storage_access_key"`
		CloudStorageSecretKey          string `json:"cloud_storage_secret_key"`
		CloudStorageAPIEndpoint        string `json:"cloud_storage_api_endpoint"`
		CloudStorageAPIEndpointPort    int    `json:"cloud_storage_api_endpoint_port"`
		CloudStorageAzureADLSEndpoint  string `json:"cloud_storage_azure_adls_endpoint"`
		CloudStorageAzureADLSPort      int    `json:"cloud_storage_azure_adls_port"`
		CloudStorageBucket             string `json:"cloud_storage_bucket"`
		CloudStorageCacheCheckInterval int    `json:"cloud_storage_cache_check_interval"`
		// CloudStorageCacheDirectory is a node config property unlike
		// everything else in this struct. It should instead be set via
		// `config.node`.
		CloudStorageCacheDirectory              string            `json:"cloud_storage_cache_directory" jsonschema:"deprecated"`
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
		CloudStorageRegion                      string            `json:"cloud_storage_region"`
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
