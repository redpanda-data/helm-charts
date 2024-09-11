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
// +gotohelm:filename=_configmap.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

	kafkaTLS := rpkKafkaClientTLSConfiguration(dot)
	if _, ok := kafkaTLS["ca_file"]; ok {
		kafkaTLS["ca_file"] = "ca.crt"
	}

	adminTLS := rpkAdminAPIClientTLSConfiguration(dot)
	if _, ok := adminTLS["ca_file"]; ok {
		adminTLS["ca_file"] = "ca.crt"
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

	result := map[string]any{
		"name":      getFirstExternalKafkaListener(dot),
		"kafka_api": ka,
		"admin_api": aa,
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
		address = fmt.Sprintf("%s.%s", address, *values.External.Domain)
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

	result := map[string]any{
		"overprovisioned":        values.Resources.GetOverProvisionValue(),
		"enable_memory_locking":  ptr.Deref(values.Resources.Memory.EnableMemoryLocking, false),
		"additional_start_flags": RedpandaAdditionalStartFlags(dot, RedpandaSMP(dot)),
		"kafka_api": map[string]any{
			"brokers": brokerList,
			"tls":     brokerTLS,
		},
		"admin_api": map[string]any{
			"addresses": values.Listeners.AdminList(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot)),
			"tls":       adminTLS,
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
		result["cert_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.crt", Fullname(dot))
		result["key_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.key", Fullname(dot))
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
		result["cert_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.crt", Fullname(dot))
		result["key_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.key", Fullname(dot))
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
			brokerTLS["cert_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.crt", Fullname(dot))
			brokerTLS["key_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.key", Fullname(dot))
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
		"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
		"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
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
		"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", internal.Cert),
		"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", internal.Cert),
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

// RedpandaAdditionalStartFlags returns a string list of flags suitable for use
// as `additional_start_flags`. User provided flags will override any of those
// set by default.
func RedpandaAdditionalStartFlags(dot *helmette.Dot, smp int64) []string {
	values := helmette.Unwrap[Values](dot.Values)

	// All `additional_start_flags` that are set by the chart.
	chartFlags := map[string]string{
		"smp": fmt.Sprintf("%d", int(smp)),
		// TODO: The transpiled go template will return float64 from both RedpandaMemory and RedpandaReserveMemory
		// By wrapping return value from that function the sprintf will work as expected
		// https://github.com/redpanda-data/helm-charts/issues/1249
		"memory": fmt.Sprintf("%dM", int(RedpandaMemory(dot))),
		// TODO: The transpiled go template will return float64 from both RedpandaMemory and RedpandaReserveMemory
		// By wrapping return value from that function the sprintf will work as expected
		// https://github.com/redpanda-data/helm-charts/issues/1249
		"reserve-memory":    fmt.Sprintf("%dM", int(RedpandaReserveMemory(dot))),
		"default-log-level": values.Logging.LogLevel,
	}

	// If in developer_mode, don't set reserve-memory.
	if values.Config.Node["developer_mode"] == true {
		delete(chartFlags, "reserve-memory")
	}

	// Check to see if there are any flags overriding the defaults set by the
	// chart.
	for flag := range chartFlags {
		for _, userFlag := range values.Statefulset.AdditionalRedpandaCmdFlags {
			if helmette.RegexMatch(fmt.Sprintf("^--%s", flag), userFlag) {
				delete(chartFlags, flag)
			}
		}
	}

	// Deterministically order out list and add in values supplied flags.
	keys := helmette.Keys(chartFlags)
	helmette.SortAlpha(keys)

	flags := []string{}
	for _, key := range keys {
		flags = append(flags, fmt.Sprintf("--%s=%s", key, chartFlags[key]))
	}

	return append(flags, values.Statefulset.AdditionalRedpandaCmdFlags...)
}
