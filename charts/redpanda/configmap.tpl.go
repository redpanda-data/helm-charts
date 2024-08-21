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
	cms := []*corev1.ConfigMap{RedpandaConfigMap(dot, true)}
	cms = append(cms, RPKProfile(dot)...)
	return cms
}

func ConfigMapsWithoutSeedServer(dot *helmette.Dot) []*corev1.ConfigMap {
	cms := []*corev1.ConfigMap{RedpandaConfigMap(dot, false)}
	cms = append(cms, RPKProfile(dot)...)
	return cms
}

func RedpandaConfigMap(dot *helmette.Dot, includeSeedServer bool) *corev1.ConfigMap {
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
			"redpanda.yaml":  RedpandaConfigFile(dot, includeSeedServer),
		},
	}
}

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
	bootstrap = helmette.Merge(bootstrap, values.Config.Cluster.Translate(values.Statefulset.Replicas, false, false))
	bootstrap = helmette.Merge(bootstrap, values.Auth.Translate(values.Auth.IsSASLEnabled()))

	return helmette.ToYaml(bootstrap)
}

func RedpandaConfigFile(dot *helmette.Dot, includeSeedServer bool) string {
	values := helmette.Unwrap[Values](dot.Values)

	redpanda := map[string]any{
		"kafka_enable_authorization": values.Auth.IsSASLEnabled(),
		"enable_sasl":                values.Auth.IsSASLEnabled(),
		"empty_seed_starts_cluster":  false,
		"storage_min_free_bytes":     values.Storage.StorageMinFreeBytes(),
	}

	if includeSeedServer {
		redpanda["seed_servers"] = values.Listeners.CreateSeedServers(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot))
	}

	redpanda = helmette.Merge(redpanda, values.AuditLogging.Translate(dot, values.Auth.IsSASLEnabled()))
	redpanda = helmette.Merge(redpanda, values.Logging.Translate())
	redpanda = helmette.Merge(redpanda, values.Config.Tunable.Translate())
	redpanda = helmette.Merge(redpanda, values.Config.Cluster.Translate(values.Statefulset.Replicas, true, true))
	redpanda = helmette.Merge(redpanda, values.Auth.Translate(values.Auth.IsSASLEnabled()))
	redpanda = helmette.Merge(redpanda, values.Config.Node.Translate())

	configureListeners(redpanda, dot)

	redpandaYaml := map[string]any{
		"redpanda":               redpanda,
		"schema_registry":        schemaRegistry(dot),
		"schema_registry_client": kafkaClient(dot),
		"pandaproxy":             pandaProxyListener(dot),
		"pandaproxy_client":      kafkaClient(dot),
		"rpk":                    rpkConfiguration(dot),
		"config_file":            "/etc/redpanda/redpanda.yaml",
	}

	if RedpandaAtLeast_23_3_0(dot) && values.AuditLogging.Enabled && values.Auth.IsSASLEnabled() {
		redpandaYaml["audit_log_client"] = kafkaClient(dot)
	}

	redpandaYaml = helmette.Merge(redpandaYaml, values.Storage.Translate())

	return helmette.ToYaml(redpandaYaml)
}

func RPKProfile(dot *helmette.Dot) []*corev1.ConfigMap {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.External.Enabled {
		return nil
	}

	return []*corev1.ConfigMap{
		{
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
		},
	}
}

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

	kafkaTLS := brokersTLSConfiguration(dot)
	if _, ok := kafkaTLS["truststore_file"]; ok {
		kafkaTLS["ca_file"] = "ca.crt"
		delete(kafkaTLS, "truststore_file")
	}

	adminTLS := adminTLSConfiguration(dot)
	if _, ok := adminTLS["truststore_file"]; ok {
		adminTLS["ca_file"] = "ca.crt"
		delete(adminTLS, "truststore_file")
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

func rpkConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := BrokerList(dot, values.Statefulset.Replicas, values.Listeners.Kafka.Port)

	var adminTLS map[string]any
	if tls := adminTLSConfiguration(dot); len(tls) > 0 {
		adminTLS = tls
	}

	var brokerTLS map[string]any
	if tls := brokersTLSConfiguration(dot); len(tls) > 0 {
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

func brokersTLSConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Listeners.Kafka.TLS.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	result := map[string]any{}

	if truststore := values.Listeners.Kafka.TLS.TrustStoreFilePath(&values.TLS); truststore != defaultTruststorePath {
		result["truststore_file"] = truststore
	}

	if values.Listeners.Kafka.TLS.RequireClientAuth {
		result["cert_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.crt", Fullname(dot))
		result["key_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.key", Fullname(dot))

	}

	return result
}

func adminTLSConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	result := map[string]any{}
	if !values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		return result
	}

	if truststore := values.Listeners.Admin.TLS.TrustStoreFilePath(&values.TLS); truststore != defaultTruststorePath {
		result["truststore_file"] = truststore
	}

	if values.Listeners.Admin.TLS.RequireClientAuth {
		result["cert_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.crt", Fullname(dot))
		result["key_file"] = fmt.Sprintf("/etc/tls/certs/%s-client/tls.key", Fullname(dot))

	}

	return result
}

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
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", kafkaTLS.Cert),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", kafkaTLS.Cert),
			"require_client_auth": kafkaTLS.RequireClientAuth,
			"truststore_file":     kafkaTLS.TrustStoreFilePath(&values.TLS),
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
