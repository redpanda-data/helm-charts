// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_configmap.go.tpl
package redpanda

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func ConfigMaps(dot *helmette.Dot) []*corev1.ConfigMap {
	cms := []*corev1.ConfigMap{RedpandaConfigMap(dot), RPKProfile(dot)}
	return cms
}

func RedpandaConfigMap(dot *helmette.Dot) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Data: map[string]string{
			"bootstrap.yaml": BootstrapFile(dot),
			"redpanda.yaml":  RedpandaConfigFile(dot, true /* includeSeedServer */),
		},
	}
}

// BootstrapFile returns contents of `.bootstrap.yaml`. Keys that may be set
// via environment variables (such as tiered storage secrets) will have
// placeholders in the form of $ENVVARNAME. An init container is responsible
// for expanding said placeholders.
//
// Convention is to name envvars
// $REDPANDA_SCREAMING_CASE_CLUSTER_PROPERTY_NAME. For example,
// cloud_storage_secret_key would be $REDPANDA_CLOUD_STORAGE_SECRET_KEY.
//
// `.bootstrap.yaml` is templated and then read by both the redpanda container
// and the post install/upgrade job.
func BootstrapFile(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	bootstrap := map[string]any{
		"kafka_enable_authorization": values.Auth.IsSASLEnabled(),
		"enable_sasl":                values.Auth.IsSASLEnabled(),
		"enable_rack_awareness":      values.RackAwareness.Enabled,
		"storage_min_free_bytes":     values.Storage.StorageMinFreeBytes(),
	}

	bootstrap = helmette.Merge(bootstrap, values.AuditLogging.Translate(dot, values.Auth.IsSASLEnabled()))
	bootstrap = helmette.Merge(bootstrap, values.Logging.Translate())
	bootstrap = helmette.Merge(bootstrap, values.Config.Tunable.Translate())
	bootstrap = helmette.Merge(bootstrap, values.Config.Cluster.Translate())
	bootstrap = helmette.Merge(bootstrap, values.Auth.Translate(values.Auth.IsSASLEnabled()))
	bootstrap = helmette.Merge(bootstrap, values.Storage.GetTieredStorageConfig().Translate(&values.Storage.Tiered.CredentialsSecretRef))

	// If default_topic_replications is not set and we have at least 3 Brokers,
	// upgrade from redpanda's default of 1 to 3 so, when possible, topics are
	// HA by default.
	// See also:
	// - https://github.com/redpanda-data/helm-charts/issues/583
	// - https://github.com/redpanda-data/helm-charts/issues/1501
	if _, ok := values.Config.Cluster["default_topic_replications"]; !ok && values.Statefulset.Replicas >= 3 {
		bootstrap["default_topic_replications"] = 3
	}

	if _, ok := values.Config.Cluster["storage_min_free_bytes"]; !ok {
		bootstrap["storage_min_free_bytes"] = values.Storage.StorageMinFreeBytes()
	}

	return helmette.ToYaml(bootstrap)
}

func RedpandaConfigFile(dot *helmette.Dot, includeSeedServer bool) string {
	values := helmette.Unwrap[Values](dot.Values)

	redpanda := map[string]any{
		"empty_seed_starts_cluster": false,
	}

	if includeSeedServer {
		redpanda["seed_servers"] = values.Listeners.CreateSeedServers(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot))
	}

	redpanda = helmette.Merge(redpanda, values.Config.Node.Translate())

	configureListeners(redpanda, dot)

	redpandaYaml := map[string]any{
		"redpanda":               redpanda,
		"schema_registry":        schemaRegistry(dot),
		"schema_registry_client": kafkaClient(dot),
		"pandaproxy":             pandaProxyListener(dot),
		"pandaproxy_client":      kafkaClient(dot),
		"rpk":                    rpkNodeConfig(dot),
		"config_file":            "/etc/redpanda/redpanda.yaml",
	}

	if RedpandaAtLeast_23_3_0(dot) && values.AuditLogging.Enabled && values.Auth.IsSASLEnabled() {
		redpandaYaml["audit_log_client"] = kafkaClient(dot)
	}

	return helmette.ToYaml(redpandaYaml)
}

// RPKProfile returns a [corev1.ConfigMap] for aiding users in connecting to
// the external listeners of their redpanda cluster.
// It is meant for external consumption via NOTES.txt and is not used within
// this chart.
func RPKProfile(dot *helmette.Dot) *corev1.ConfigMap {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.External.Enabled {
		return nil
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-rpk", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Data: map[string]string{
			"profile": helmette.ToYaml(rpkProfile(dot)),
		},
	}
}

// rpkProfile generates an RPK Profile for connecting to external listeners.
// It is intended to be used by the end user via a prompt in NOTES.txt.
func rpkProfile(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := []string{}
	for i := int32(0); i < values.Statefulset.Replicas; i++ {
		brokerList = append(brokerList, fmt.Sprintf("%s:%d", advertisedHost(dot, i), int(advertisedKafkaPort(dot, i))))
	}

	adminAdvertisedList := []string{}
	for i := int32(0); i < values.Statefulset.Replicas; i++ {
		adminAdvertisedList = append(adminAdvertisedList, fmt.Sprintf("%s:%d", advertisedHost(dot, i), int(advertisedAdminPort(dot, i))))
	}

	schemaAdvertisedList := []string{}
	for i := int32(0); i < values.Statefulset.Replicas; i++ {
		schemaAdvertisedList = append(schemaAdvertisedList, fmt.Sprintf("%s:%d", advertisedHost(dot, i), int(advertisedSchemaPort(dot, i))))
	}

	kafkaTLS := rpkKafkaClientTLSConfiguration(dot)
	if _, ok := kafkaTLS["ca_file"]; ok {
		kafkaTLS["ca_file"] = "ca.crt"
	}

	adminTLS := rpkAdminAPIClientTLSConfiguration(dot)
	if _, ok := adminTLS["ca_file"]; ok {
		adminTLS["ca_file"] = "ca.crt"
	}

	schemaTLS := rpkSchemaRegistryClientTLSConfiguration(dot)
	if _, ok := schemaTLS["ca_file"]; ok {
		schemaTLS["ca_file"] = "ca.crt"
	}

	ka := map[string]any{
		"brokers": brokerList,
		"tls":     nil,
	}

	if len(kafkaTLS) > 0 {
		ka["tls"] = kafkaTLS
	}

	aa := map[string]any{
		"addresses": adminAdvertisedList,
		"tls":       nil,
	}

	if len(adminTLS) > 0 {
		aa["tls"] = adminTLS
	}

	sa := map[string]any{
		"addresses": schemaAdvertisedList,
		"tls":       nil,
	}

	if len(schemaTLS) > 0 {
		sa["tls"] = schemaTLS
	}

	result := map[string]any{
		"name":            getFirstExternalKafkaListener(dot),
		"kafka_api":       ka,
		"admin_api":       aa,
		"schema_registry": sa,
	}

	return result
}

func advertisedKafkaPort(dot *helmette.Dot, i int32) int {
	values := helmette.Unwrap[Values](dot.Values)

	externalKafkaListenerName := getFirstExternalKafkaListener(dot)

	listener := values.Listeners.Kafka.External[externalKafkaListenerName]

	port := int(values.Listeners.Kafka.Port)

	if int(listener.Port) > int(1) {
		port = int(listener.Port)
	}

	if len(listener.AdvertisedPorts) > 1 {
		port = int(listener.AdvertisedPorts[i])
	} else if len(listener.AdvertisedPorts) == 1 {
		port = int(listener.AdvertisedPorts[0])
	}

	return port
}

func advertisedAdminPort(dot *helmette.Dot, i int32) int {
	values := helmette.Unwrap[Values](dot.Values)

	keys := helmette.Keys(values.Listeners.Admin.External)

	helmette.SortAlpha(keys)

	externalAdminListenerName := helmette.First(keys)

	listener := values.Listeners.Admin.External[externalAdminListenerName.(string)]

	port := int(values.Listeners.Admin.Port)

	if int(listener.Port) > 1 {
		port = int(listener.Port)
	}

	if len(listener.AdvertisedPorts) > 1 {
		port = int(listener.AdvertisedPorts[i])
	} else if len(listener.AdvertisedPorts) == 1 {
		port = int(listener.AdvertisedPorts[0])
	}

	return port
}

func advertisedSchemaPort(dot *helmette.Dot, i int32) int {
	values := helmette.Unwrap[Values](dot.Values)

	keys := helmette.Keys(values.Listeners.SchemaRegistry.External)

	helmette.SortAlpha(keys)

	externalSchemaListenerName := helmette.First(keys)

	listener := values.Listeners.SchemaRegistry.External[externalSchemaListenerName.(string)]

	port := int(values.Listeners.SchemaRegistry.Port)

	if int(listener.Port) > 1 {
		port = int(listener.Port)
	}

	if len(listener.AdvertisedPorts) > 1 {
		port = int(listener.AdvertisedPorts[i])
	} else if len(listener.AdvertisedPorts) == 1 {
		port = int(listener.AdvertisedPorts[0])
	}

	return port
}

func advertisedHost(dot *helmette.Dot, i int32) string {
	values := helmette.Unwrap[Values](dot.Values)

	address := fmt.Sprintf("%s-%d", Fullname(dot), int(i))
	if ptr.Deref(values.External.Domain, "") != "" {
		address = fmt.Sprintf("%s.%s", address, helmette.Tpl(*values.External.Domain, dot))
	}

	if len(values.External.Addresses) <= 0 {
		return address
	}

	if len(values.External.Addresses) == 1 {
		address = values.External.Addresses[0]
	} else {
		address = values.External.Addresses[i]
	}

	if ptr.Deref(values.External.Domain, "") != "" {
		address = fmt.Sprintf("%s.%s", address, helmette.Tpl(*values.External.Domain, dot))
	}

	return address
}

func getFirstExternalKafkaListener(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	keys := helmette.Keys(values.Listeners.Kafka.External)

	helmette.SortAlpha(keys)

	return helmette.First(keys).(string)
}

func BrokerList(dot *helmette.Dot, replicas int32, port int32) []string {
	var bl []string

	for i := int32(0); i < replicas; i++ {
		bl = append(bl, fmt.Sprintf("%s-%d.%s:%d", Fullname(dot), i, InternalDomain(dot), port))
	}

	return bl
}

// https://github.com/redpanda-data/redpanda/blob/817450a480f4f2cadf66de1adc301cfaf6ccde46/src/go/rpk/pkg/config/redpanda_yaml.go#L143
func rpkNodeConfig(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := BrokerList(dot, values.Statefulset.Replicas, values.Listeners.Kafka.Port)

	var adminTLS map[string]any
	if tls := rpkAdminAPIClientTLSConfiguration(dot); len(tls) > 0 {
		adminTLS = tls
	}

	var brokerTLS map[string]any
	if tls := rpkKafkaClientTLSConfiguration(dot); len(tls) > 0 {
		brokerTLS = tls
	}

	var schemaRegistryTLS map[string]any
	if tls := rpkSchemaRegistryClientTLSConfiguration(dot); len(tls) > 0 {
		schemaRegistryTLS = tls
	}

	lockMemory, overprovisioned, flags := RedpandaAdditionalStartFlags(&values)

	result := map[string]any{
		"additional_start_flags": flags,
		"enable_memory_locking":  lockMemory,
		"overprovisioned":        overprovisioned,
		"kafka_api": map[string]any{
			"brokers": brokerList,
			"tls":     brokerTLS,
		},
		"admin_api": map[string]any{
			"addresses": values.Listeners.AdminList(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot)),
			"tls":       adminTLS,
		},
		"schema_registry": map[string]any{
			"addresses": values.Listeners.SchemaRegistryList(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot)),
			"tls":       schemaRegistryTLS,
		},
	}

	result = helmette.Merge(result, values.Tuning.Translate())
	result = helmette.Merge(result, values.Config.CreateRPKConfiguration())

	return result
}

// rpkKafkaClientTLSConfiguration returns a value suitable for use as RPK's
// "TLS" type.
// https://github.com/redpanda-data/redpanda/blob/817450a480f4f2cadf66de1adc301cfaf6ccde46/src/go/rpk/pkg/config/redpanda_yaml.go#L178
func rpkKafkaClientTLSConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	tls := values.Listeners.Kafka.TLS

	if !tls.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	result := map[string]any{
		"ca_file": tls.ServerCAPath(&values.TLS),
	}

	if tls.RequireClientAuth {
		result["cert_file"] = fmt.Sprintf("%s/%s-client/tls.crt", certificateMountPoint, Fullname(dot))
		result["key_file"] = fmt.Sprintf("%s/%s-client/tls.key", certificateMountPoint, Fullname(dot))
	}

	return result
}

// rpkAdminAPIClientTLSConfiguration returns a value suitable for use as RPK's
// "TLS" type.
// https://github.com/redpanda-data/redpanda/blob/817450a480f4f2cadf66de1adc301cfaf6ccde46/src/go/rpk/pkg/config/redpanda_yaml.go#L184
func rpkAdminAPIClientTLSConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	tls := values.Listeners.Admin.TLS

	if !tls.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	result := map[string]any{
		"ca_file": tls.ServerCAPath(&values.TLS),
	}

	if tls.RequireClientAuth {
		result["cert_file"] = fmt.Sprintf("%s/%s-client/tls.crt", certificateMountPoint, Fullname(dot))
		result["key_file"] = fmt.Sprintf("%s/%s-client/tls.key", certificateMountPoint, Fullname(dot))
	}

	return result
}

// rpkSchemaRegistryClientTLSConfiguration returns a value suitable for use as RPK's
// "TLS" type.
// https://github.com/redpanda-data/redpanda/blob/817450a480f4f2cadf66de1adc301cfaf6ccde46/src/go/rpk/pkg/config/redpanda_yaml.go#L184
func rpkSchemaRegistryClientTLSConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	tls := values.Listeners.SchemaRegistry.TLS

	if !tls.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	result := map[string]any{
		"ca_file": tls.ServerCAPath(&values.TLS),
	}

	if tls.RequireClientAuth {
		result["cert_file"] = fmt.Sprintf("%s/%s-client/tls.crt", certificateMountPoint, Fullname(dot))
		result["key_file"] = fmt.Sprintf("%s/%s-client/tls.key", certificateMountPoint, Fullname(dot))
	}

	return result
}

// kafkaClient returns the configuration for internal components of redpanda to
// connect to its own Kafka API. This is distinct from RPK's configuration for
// Kafka API interactions.
func kafkaClient(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := []map[string]any{}
	for i := int32(0); i < values.Statefulset.Replicas; i++ {
		brokerList = append(brokerList, map[string]any{
			"address": fmt.Sprintf("%s-%d.%s", Fullname(dot), i, InternalDomain(dot)),
			"port":    values.Listeners.Kafka.Port,
		})
	}

	kafkaTLS := values.Listeners.Kafka.TLS

	var brokerTLS map[string]any
	if values.Listeners.Kafka.TLS.IsEnabled(&values.TLS) {
		brokerTLS = map[string]any{
			"enabled":             true,
			"require_client_auth": kafkaTLS.RequireClientAuth,
			// NB: truststore_file here is synonymous with ca_file in the RPK
			// configuration. The difference being that redpanda does NOT read
			// the ca_file key.
			"truststore_file": kafkaTLS.ServerCAPath(&values.TLS),
		}

		if kafkaTLS.RequireClientAuth {
			brokerTLS["cert_file"] = fmt.Sprintf("%s/%s-client/tls.crt", certificateMountPoint, Fullname(dot))
			brokerTLS["key_file"] = fmt.Sprintf("%s/%s-client/tls.key", certificateMountPoint, Fullname(dot))
		}

	}

	cfg := map[string]any{
		"brokers": brokerList,
	}
	if len(brokerTLS) > 0 {
		cfg["broker_tls"] = brokerTLS
	}

	return cfg
}

func configureListeners(redpanda map[string]any, dot *helmette.Dot) {
	values := helmette.Unwrap[Values](dot.Values)

	redpanda["admin"] = values.Listeners.Admin.Listeners()
	redpanda["kafka_api"] = values.Listeners.Kafka.Listeners(&values.Auth)
	redpanda["rpc_server"] = rpcListeners(dot)

	// Backwards compatibility layer, if any of the *_tls keys are an empty
	// slice, they should instead be nil.

	redpanda["admin_api_tls"] = nil
	if tls := values.Listeners.Admin.ListenersTLS(&values.TLS); len(tls) > 0 {
		redpanda["admin_api_tls"] = tls
	}

	redpanda["kafka_api_tls"] = nil
	if tls := values.Listeners.Kafka.ListenersTLS(&values.TLS); len(tls) > 0 {
		redpanda["kafka_api_tls"] = tls
	}

	// With the exception of rpc_server_tls, it should just not be specified.
	if tls := rpcListenersTLS(dot); len(tls) > 0 {
		redpanda["rpc_server_tls"] = tls
	}
}

func pandaProxyListener(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	pandaProxy := map[string]any{}

	pandaProxy["pandaproxy_api"] = values.Listeners.HTTP.Listeners(values.Auth.IsSASLEnabled())
	pandaProxy["pandaproxy_api_tls"] = nil
	if tls := values.Listeners.HTTP.ListenersTLS(&values.TLS); len(tls) > 0 {
		pandaProxy["pandaproxy_api_tls"] = tls
	}
	return pandaProxy
}

func schemaRegistry(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	schemaReg := map[string]any{}

	schemaReg["schema_registry_api"] = values.Listeners.SchemaRegistry.Listeners(values.Auth.IsSASLEnabled())
	schemaReg["schema_registry_api_tls"] = nil
	if tls := values.Listeners.SchemaRegistry.ListenersTLS(&values.TLS); len(tls) > 0 {
		schemaReg["schema_registry_api_tls"] = tls
	}
	return schemaReg
}

func rpcListenersTLS(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	r := values.Listeners.RPC

	if !(RedpandaAtLeast_22_2_atleast_22_2_10(dot) ||
		RedpandaAtLeast_22_3_atleast_22_3_13(dot) ||
		RedpandaAtLeast_23_1_2(dot)) && (r.TLS.Enabled == nil && values.TLS.Enabled || ptr.Deref(r.TLS.Enabled, false)) {
		panic(fmt.Sprintf("Redpanda version v%s does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01.", helmette.TrimPrefix("v", Tag(dot))))
	}

	if !r.TLS.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	certName := r.TLS.Cert

	return map[string]any{
		"enabled":             true,
		"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, certName),
		"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, certName),
		"require_client_auth": r.TLS.RequireClientAuth,
		"truststore_file":     r.TLS.TrustStoreFilePath(&values.TLS),
	}
}

func rpcListeners(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	return map[string]any{
		"address": "0.0.0.0",
		"port":    values.Listeners.RPC.Port,
	}
}

// First parameter defaultTLSEnabled must come from `values.tls.enabled`.
func createInternalListenerTLSCfg(tls *TLS, internal InternalTLS) map[string]any {
	if !internal.IsEnabled(tls) {
		return map[string]any{}
	}

	return map[string]any{
		"name":                "internal",
		"enabled":             true,
		"cert_file":           fmt.Sprintf("%s/%s/tls.crt", certificateMountPoint, internal.Cert),
		"key_file":            fmt.Sprintf("%s/%s/tls.key", certificateMountPoint, internal.Cert),
		"require_client_auth": internal.RequireClientAuth,
		"truststore_file":     internal.TrustStoreFilePath(tls),
	}
}

func createInternalListenerCfg(port int32) map[string]any {
	return map[string]any{
		"name":    "internal",
		"address": "0.0.0.0",
		"port":    port,
	}
}

// RedpandaAdditionalStartFlags returns a string slice of flags suitable for use
// as `additional_start_flags`. User provided flags will override any of those
// set by default.
func RedpandaAdditionalStartFlags(values *Values) (bool, bool, []string) {
	// All `additional_start_flags` that are set by the chart.
	flags := values.Resources.GetRedpandaFlags()
	flags["--default-log-level"] = values.Logging.LogLevel

	// Unclear why this is done aside from historical reasons.
	// Legacy comment: If in developer_mode, don't set reserve-memory.
	if values.Config.Node["developer_mode"] == true {
		delete(flags, "--reserve-memory")
	}

	for key, value := range ParseCLIArgs(values.Statefulset.AdditionalRedpandaCmdFlags) {
		flags[key] = value
	}

	enabledOptions := map[string]bool{
		"true": true,
		"1":    true,
		"":     true,
	}

	// Due to a buglet in rpk, we need to set lock-memory and overprovisioned
	// via their fields in redpanda.yaml instead of additional_start_flags.
	// https://github.com/redpanda-data/helm-charts/pull/1622#issuecomment-2577922409
	lockMemory := false
	if value, ok := flags["--lock-memory"]; ok {
		lockMemory = enabledOptions[value]
		delete(flags, "--lock-memory")
	}

	overprovisioned := false
	if value, ok := flags["--overprovisioned"]; ok {
		overprovisioned = enabledOptions[value]
		delete(flags, "--overprovisioned")
	}

	// Deterministically order out list and add in values supplied flags.
	keys := helmette.Keys(flags)
	keys = helmette.SortAlpha(keys)

	var rendered []string
	for _, key := range keys {
		value := flags[key]
		// Support flags that don't have values (`--overprovisioned`) by
		// letting them be specified as key: ""
		if value == "" {
			rendered = append(rendered, key)
		} else {
			rendered = append(rendered, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return lockMemory, overprovisioned, rendered
}
