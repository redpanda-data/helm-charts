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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func ConfigMap(dot *helmette.Dot) []corev1.ConfigMap {
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
	bootstrap = helmette.Merge(bootstrap, values.Config.Cluster.Translate(values.Statefulset.Replicas, false))
	bootstrap = helmette.Merge(bootstrap, values.Auth.Translate(values.Auth.IsSASLEnabled()))

	// configureUsageStats(bootstrap, dot)
	// configureTunable(bootstrap, dot)
	// configureClusterConfiguration(bootstrap, dot, false)
	// configureSuperUsers(bootstrap, dot)

	redpanda := map[string]any{
		"kafka_enable_authorization": values.Auth.IsSASLEnabled(),
		"enable_sasl":                values.Auth.IsSASLEnabled(),
		"empty_seed_starts_cluster":  false,
		"storage_min_free_bytes":     values.Storage.StorageMinFreeBytes(),
		"seed_servers":               values.Listeners.CreateSeedServers(values.Statefulset.Replicas, Fullname(dot), InternalDomain(dot)),
	}

	redpanda = helmette.Merge(redpanda, values.AuditLogging.Translate(dot, values.Auth.IsSASLEnabled()))
	redpanda = helmette.Merge(redpanda, values.Logging.Translate())
	redpanda = helmette.Merge(redpanda, values.Config.Tunable.Translate())
	redpanda = helmette.Merge(redpanda, values.Config.Cluster.Translate(values.Statefulset.Replicas, true))
	redpanda = helmette.Merge(redpanda, values.Auth.Translate(values.Auth.IsSASLEnabled()))
	redpanda = helmette.Merge(redpanda, values.Config.Node.Translate())

	// configureUsageStats(redpanda, dot)
	// configureTunable(redpanda, dot)
	// configureClusterConfiguration(redpanda, dot, true)
	// configureNodeConfiguration(redpanda, dot)
	configureListeners(redpanda, dot)
	// configureSuperUsers(redpanda, dot)

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

	cms := []corev1.ConfigMap{
		{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      Fullname(dot),
				Namespace: dot.Release.Namespace,
				Labels:    FullLabels(dot),
			},
			Data: map[string]string{
				"bootstrap.yaml": helmette.ToYaml(bootstrap),
				"redpanda.yaml":  helmette.ToYaml(redpandaYaml),
			},
		},
	}

	if values.External.Enabled {
		cms = append(cms, corev1.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("%s-rpk", Fullname(dot)),
				Namespace: dot.Release.Namespace,
				Labels:    FullLabels(dot),
			},
			Data: map[string]string{
				"profile": helmette.ToYaml(rpkProfile(dot)),
			},
		})
	}
	return cms
}

func rpkProfile(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := []string{}
	for i := 0; i < values.Statefulset.Replicas; i++ {
		brokerList = append(brokerList, fmt.Sprintf("%s:%d", advertisedHost(dot, i), int(advertisedKafkaPort(dot, i))))
	}

	adminAdvertisedList := []string{}
	for i := 0; i < values.Statefulset.Replicas; i++ {
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
		"name":      getFistExternalKafkaListener(dot),
		"kafka_api": ka,
		"admin_api": aa,
	}

	return result
}

func advertisedKafkaPort(dot *helmette.Dot, i int) int {
	values := helmette.Unwrap[Values](dot.Values)

	externalKafkaListenerName := getFistExternalKafkaListener(dot)

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

func advertisedAdminPort(dot *helmette.Dot, i int) int {
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

func advertisedHost(dot *helmette.Dot, i int) string {
	values := helmette.Unwrap[Values](dot.Values)

	address := fmt.Sprintf("%s-%d", Fullname(dot), int(i))
	if ptr.Deref(values.External.Domain, "") != "" {
		// TODO: TPL DOES NOT WORK. Maybe tpl call could be removed
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

func getFistExternalKafkaListener(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	keys := helmette.Keys(values.Listeners.Kafka.External)

	helmette.SortAlpha(keys)

	return helmette.First(keys).(string)
}

func rpkConfiguration(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	brokerList := []string{}
	r := values.Statefulset.Replicas
	for i := 0; i < r; i++ {
		brokerList = append(brokerList, fmt.Sprintf("%s-%d.%s:%d", Fullname(dot), i, InternalDomain(dot), int(values.Listeners.Kafka.Port)))
	}

	var adminTLS map[string]any
	if tls := adminTLSConfiguration(dot); len(tls) > 0 {
		adminTLS = tls
	}

	var brokerTLS map[string]any
	if tls := brokersTLSConfiguration(dot); len(tls) > 0 {
		brokerTLS = tls
	}

	result := map[string]any{
		"overprovisioned":        ptr.Deref(values.Resources.CPU.Overprovisioned, false),
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
	certName := values.Listeners.Kafka.TLS.Cert

	if cert, ok := values.TLS.Certs[certName]; ok && cert.CAEnabled {
		result["truststore_file"] = fmt.Sprintf("/etc/tls/certs/%s/ca.crt", values.Listeners.Kafka.TLS.Cert)
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

	certName := values.Listeners.Admin.TLS.Cert

	if cert, ok := values.TLS.Certs[certName]; ok && cert.CAEnabled {
		result["truststore_file"] = fmt.Sprintf("/etc/tls/certs/%s/ca.crt", values.Listeners.Admin.TLS.Cert)
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
	for i := 0; i < values.Statefulset.Replicas; i++ {
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
			"truststore_file":     getCertificate(&values.TLS.Certs, kafkaTLS.Cert),
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
	redpanda["admin"] = adminListeners(dot)
	redpanda["admin_api_tls"] = nil
	tls := adminListenersTLS(dot)
	if len(tls) > 0 {
		redpanda["admin_api_tls"] = tls
	}
	redpanda["kafka_api"] = kafkaListeners(dot)
	redpanda["kafka_api_tls"] = nil
	tls = kafkaListenersTLS(dot)
	if len(tls) > 0 {
		redpanda["kafka_api_tls"] = tls
	}
	redpanda["rpc_server"] = rpcListeners(dot)
	rpcTLS := rpcListenersTLS(dot)
	if len(rpcTLS) > 0 {
		redpanda["rpc_server_tls"] = rpcTLS
	}
}

func pandaProxyListener(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	pandaProxy := map[string]any{}

	pandaProxy["pandaproxy_api"] = values.Listeners.HTTP.Listeners(values.Auth.IsSASLEnabled())
	tls := pandaProxyListenersTLS(dot)
	pandaProxy["pandaproxy_api_tls"] = nil
	if len(tls) > 0 {
		pandaProxy["pandaproxy_api_tls"] = tls
	}
	return pandaProxy
}

func pandaProxyListenersTLS(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	pp := []map[string]any{}

	internal := createInternalListenerTLSCfg(&values.TLS, values.Listeners.HTTP.TLS)
	if len(internal) > 0 {
		pp = append(pp, internal)
	}

	for k, l := range values.Listeners.HTTP.External {
		if !l.IsEnabled() || !l.TLS.IsEnabled(&values.Listeners.HTTP.TLS, &values.TLS) {
			continue
		}

		certName := l.TLS.GetCertName(&values.Listeners.HTTP.TLS)

		pp = append(pp, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(l.TLS.RequireClientAuth, false),
			"truststore_file":     getCertificate(&values.TLS.Certs, certName),
		})
	}
	return pp
}

func schemaRegistry(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	schemaReg := map[string]any{}

	schemaReg["schema_registry_api"] = values.Listeners.SchemaRegistry.Listeners(values.Auth.IsSASLEnabled())
	tls := schemaRegistryListenersTLS(dot)
	schemaReg["schema_registry_api_tls"] = nil
	if len(tls) > 0 {
		schemaReg["schema_registry_api_tls"] = tls
	}
	return schemaReg
}

func schemaRegistryListenersTLS(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	sr := []map[string]any{}

	internal := createInternalListenerTLSCfg(&values.TLS, values.Listeners.SchemaRegistry.TLS)
	if len(internal) > 0 {
		sr = append(sr, internal)
	}

	for k, l := range values.Listeners.SchemaRegistry.External {
		if !l.IsEnabled() || !l.TLS.IsEnabled(&values.Listeners.SchemaRegistry.TLS, &values.TLS) {
			continue
		}

		certName := l.TLS.GetCertName(&values.Listeners.SchemaRegistry.TLS)

		sr = append(sr, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(l.TLS.RequireClientAuth, false),
			"truststore_file":     getCertificate(&values.TLS.Certs, certName),
		})
	}
	return sr
}

func rpcListenersTLS(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	r := values.Listeners.RPC

	if !r.TLS.IsEnabled(&values.TLS) {
		return map[string]any{}
	}

	certName := r.TLS.Cert

	return map[string]any{
		"enabled":             true,
		"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
		"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
		"require_client_auth": r.TLS.RequireClientAuth,
		"truststore_file":     getCertificate(&values.TLS.Certs, certName),
	}
}

func rpcListeners(dot *helmette.Dot) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	return map[string]any{
		"address": "0.0.0.0",
		"port":    values.Listeners.RPC.Port,
	}
}

func kafkaListenersTLS(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	kafka := []map[string]any{}

	internal := createInternalListenerTLSCfg(&values.TLS, values.Listeners.Kafka.TLS)
	if len(internal) > 0 {
		kafka = append(kafka, internal)
	}

	for k, l := range values.Listeners.Kafka.External {
		if !l.IsEnabled() || !l.TLS.IsEnabled(&values.Listeners.Kafka.TLS, &values.TLS) {
			continue
		}

		certName := l.TLS.GetCertName(&values.Listeners.Kafka.TLS)

		kafka = append(kafka, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(l.TLS.RequireClientAuth, false),
			"truststore_file":     getCertificate(&values.TLS.Certs, certName),
		})
	}
	return kafka
}

func getCertificate(certs *TLSCertMap, certName string) string {
	defaultTruststorePath := "/etc/ssl/certs/ca-certificates.crt"
	// TLSCertMap is not defined inside each listener as TLSCertMap can be shared
	// between each listener.
	if certs == nil {
		panic("TLS map is not defined")
	}
	if certName == "" {
		return defaultTruststorePath
	}
	c := *certs
	// TLSCert can overwrite ca/truststore path in the configuration
	if crt, ok := c[certName]; ok && crt.CAEnabled {
		return fmt.Sprintf("/etc/tls/certs/%s/ca.crt", certName)
	} else if !ok {
		panic(fmt.Sprintf("Certificate name reference (%s) defined in listener, but not found in the tls.certs map", certName))
	}
	return defaultTruststorePath
}

func kafkaListeners(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	kf := values.Listeners.Kafka
	internal := createInternalListenerCfg(values.Listeners.Kafka.Port)

	if values.Auth.IsSASLEnabled() {
		internal["authentication_method"] = "sasl"
	}

	if am := ptr.Deref(kf.AuthenticationMethod, ""); am != "" {
		internal["authentication_method"] = am
	}

	kafka := []map[string]any{
		internal,
	}

	for k, l := range kf.External {
		if !l.IsEnabled() {
			continue
		}

		listener := map[string]any{
			"name":    k,
			"port":    l.Port,
			"address": "0.0.0.0",
		}

		if values.Auth.IsSASLEnabled() {
			listener["authentication_method"] = "sasl"
		}

		if am := ptr.Deref(l.AuthenticationMethod, ""); am != "" {
			listener["authentication_method"] = am
		}

		kafka = append(kafka, listener)
	}
	return kafka
}

func adminListenersTLS(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	admin := []map[string]any{}

	internal := createInternalListenerTLSCfg(&values.TLS, values.Listeners.Admin.TLS)
	if len(internal) > 0 {
		admin = append(admin, internal)
	}

	for k, l := range values.Listeners.Admin.External {
		if !l.IsEnabled() || !l.TLS.IsEnabled(&values.Listeners.Admin.TLS, &values.TLS) {
			continue
		}

		certName := l.TLS.GetCertName(&values.Listeners.Admin.TLS)

		admin = append(admin, map[string]any{
			"name":                k,
			"enabled":             true,
			"cert_file":           fmt.Sprintf("/etc/tls/certs/%s/tls.crt", certName),
			"key_file":            fmt.Sprintf("/etc/tls/certs/%s/tls.key", certName),
			"require_client_auth": ptr.Deref(l.TLS.RequireClientAuth, false),
			"truststore_file":     getCertificate(&values.TLS.Certs, certName),
		})
	}
	return admin
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
		"truststore_file":     getCertificate(&tls.Certs, internal.Cert),
	}
}

func createInternalListenerCfg(port int) map[string]any {
	return map[string]any{
		"name":    "internal",
		"address": "0.0.0.0",
		"port":    port,
	}
}

func adminListeners(dot *helmette.Dot) []map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	admin := []map[string]any{
		createInternalListenerCfg(values.Listeners.Admin.Port),
	}
	for k, l := range values.Listeners.Admin.External {
		if !l.IsEnabled() {
			continue
		}

		admin = append(admin, map[string]any{
			"name":    k,
			"port":    l.Port,
			"address": "0.0.0.0",
		})
	}
	return admin
}

// RedpandaAdditionalStartFlags returns a string list of flags suitable for use
// as `additional_start_flags`. User provided flags will override any of those
// set by default.
func RedpandaAdditionalStartFlags(dot *helmette.Dot, smp int) []string {
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
