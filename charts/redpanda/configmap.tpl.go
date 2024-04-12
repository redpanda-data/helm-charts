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
)

// RedpandaAdditionalStartFlags returns a string list of flags suitable for use
// as `additional_start_flags`. User provided flags will override any of those
// set by default.
func RedpandaAdditionalStartFlags(dot *helmette.Dot, smp, memory, reserveMemory string) []string {
	values := helmette.Unwrap[Values](dot.Values)

	// All `additional_start_flags` that are set by the chart.
	chartFlags := map[string]string{
		"smp":               smp,
		"memory":            fmt.Sprintf("%sM", memory),
		"reserve-memory":    fmt.Sprintf("%sM", reserveMemory),
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

// func RedpandaYAMLKafkaListeners(dot *helmette.Dot) []KafkaListener {
// 	values := helmette.Unwrap[Values](dot.Values)
//
// 	input := values.Listeners.Kafka
//
// 	if input.AuthenticationMethod == "" {
// 		input.AuthenticationMethod = "sasl" // ??
// 		// {{- if or (include "sasl-enabled" $root | fromJson).bool $kafkaService.authenticationMethod }}
// 		//         authentication_method: {{ default "sasl" $kafkaService.authenticationMethod }}
// 		// {{- end }}
// 	}
//
// 	internalCert, ok := values.TLS.Certs[input.TLS.Cert]
// 	if !ok {
// 		panic(fmt.Sprintf("referenced certificate not defined: %q", input.TLS.Cert))
// 	}
//
// 	listeners := []KafkaListener{
// 		{
// 			Name:                 "internal",
// 			Address:              "0.0.0.0",
// 			Port:                 values.Listeners.Kafka.Port,
// 			AuthenticationMethod: &values.Listeners.Kafka.AuthenticationMethod,
// 			TLS: KafkaListenerTLS{
// 				Enabled:           false,
// 				CertFile:          fmt.Sprintf("/etc/tls/certs/%s/tls.crt", input.TLS.Cert),
// 				KeyFile:           fmt.Sprintf("/etc/tls/certs/%s/tls.key", input.TLS.Cert),
// 				RequireClientAuth: *input.TLS.RequireClientAuth,
// 			},
// 		},
// 	}
//
// 	// This is a required field so we use the default in the redpanda debian container.
// 	defaultTrustStore := "/etc/ssl/certs/ca-certificates.crt"
//
// 	if internalCert.CAEnabled {
// 		listeners[0].TLS.TrustStoreFile = fmt.Sprintf("/etc/tls/certs/%s/ca.crt", input.TLS.Cert)
// 	} else {
// 		listeners[0].TLS.TrustStoreFile = defaultTrustStore
// 	}
//
// 	names := helmette.Keys(input.External)
// 	helmette.SortAlpha(names)
//
// 	for _, name := range names {
// 		listener := input.External[name]
//
// 		var tls KafkaListenerTLS
// 		if listener.TLS != nil && listener.TLS.Cert != "" {
// 			cert, ok := values.TLS.Certs[listener.TLS.Cert]
// 			if !ok {
// 				panic("todo")
// 			}
//
// 			tls = KafkaListenerTLS{
// 				Enabled:           *listener.Enabled,
// 				CertFile:          fmt.Sprintf("/etc/tls/certs/%s/tls.crt", listener.TLS.Cert),
// 				KeyFile:           fmt.Sprintf("/etc/tls/certs/%s/tls.key", listener.TLS.Cert),
// 				RequireClientAuth: *listener.TLS.RequireClientAuth,
// 			}
//
// 			if cert.CAEnabled {
// 				tls.TrustStoreFile = fmt.Sprintf("/etc/tls/certs/%s/ca.crt", listener.TLS.Cert)
// 			}
// 		}
//
// 		listeners = append(listeners, KafkaListener{
// 			Name:    name,
// 			Address: "0.0.0.0",
// 			Port:    listener.Port,
// 			// AdvertisedAddress: ,
// 			TLS: tls,
// 		})
// 	}
//
// 	return listeners
// }
//
// func RedpandaYAMLListenersKafkaAPI(dot *helmette.Dot) []map[string]any {
// 	kafkaListeners := RedpandaYAMLKafkaListeners(dot)
//
// 	var kafka_api []map[string]any
// 	for _, listener := range kafkaListeners {
// 		kafka_api = append(kafka_api, map[string]any{
// 			"name":                  listener.Name,
// 			"address":               listener.Address,
// 			"port":                  listener.Port,
// 			"authentication_method": listener.AuthenticationMethod,
// 		})
// 	}
//
// 	return kafka_api
// }
//
// func RedpandaYAMLListenersKafkaAPITLS(dot *helmette.Dot) []map[string]any {
// 	kafkaListeners := RedpandaYAMLKafkaListeners(dot)
//
// 	var kafka_api_tls []map[string]any
// 	for _, listener := range kafkaListeners {
// 		kafka_api_tls = append(kafka_api_tls, map[string]any{
// 			"name":             listener.Name,
// 			"enabled":          listener.TLS.Enabled,
// 			"cert_file":        listener.TLS.CertFile,
// 			"key_file":         listener.TLS.KeyFile,
// 			"trust_store_file": listener.TLS.TrustStoreFile,
// 		})
// 	}
//
// 	return kafka_api_tls
// }
