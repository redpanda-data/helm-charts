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
// +gotohelm:filename=console.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/console/backend/pkg/config"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func ConsoleConfig(dot *helmette.Dot) any {
	values := helmette.Unwrap[Values](dot.Values)

	var schemaURLs []string
	if values.Listeners.SchemaRegistry.Enabled {
		schema := "http"
		if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
			schema = "https"
		}

		for i := 0; i < values.Statefulset.Replicas; i++ {
			schemaURLs = append(schemaURLs, fmt.Sprintf("%s://%s-%d.%s:%d", schema, Fullname(dot), i, InternalDomain(dot), values.Listeners.SchemaRegistry.Port))
		}
	}

	schema := "http"
	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		schema = "https"
	}
	c := map[string]any{
		"kafka": map[string]any{
			"brokers": BrokerList(dot, values.Statefulset.Replicas, values.Listeners.Kafka.Port),
			"sasl": map[string]any{
				"enabled": values.Auth.IsSASLEnabled(),
			},
			"tls": values.Listeners.Kafka.ConsolemTLS(&values.TLS),
			"schemaRegistry": map[string]any{
				"enabled": values.Listeners.SchemaRegistry.Enabled,
				"urls":    schemaURLs,
				"tls":     values.Listeners.SchemaRegistry.ConsoleTLS(&values.TLS),
			},
		},
		"redpanda": map[string]any{
			"adminApi": map[string]any{
				"enabled": true,
				"urls": []string{
					fmt.Sprintf("%s://%s:%d", schema, InternalDomain(dot), values.Listeners.Admin.Port),
				},
				"tls": values.Listeners.Admin.ConsoleTLS(&values.TLS),
			},
		},
	}

	if values.Connectors.Enabled {
		// TODO Do not cal Dig with dot.Values as restPort that is defined in connectors helm chart is not
		// available in this function.
		// TODO Find a way to call `(include "connectors.serviceName" $connectorsValues)` template defined
		// in connectors helm chart repo.

		port := helmette.Dig(dot.Values.AsMap(), 8083, "connectors", "connectors", "restPort")
		p, ok := helmette.AsIntegral[int](port)
		if !ok {
			return c
		}

		connectorsURL := fmt.Sprintf("http://%s.%s.svc.%s:%d",
			ConnectorsFullName(dot),
			dot.Release.Namespace,
			helmette.TrimSuffix(".", values.ClusterDomain),
			p)

		c["connect"] = config.Connect{
			Enabled: values.Connectors.Enabled,
			Clusters: []config.ConnectCluster{
				{
					Name: "connectors",
					URL:  connectorsURL,
					TLS: config.ConnectClusterTLS{
						Enabled:               false,
						CaFilepath:            "",
						CertFilepath:          "",
						KeyFilepath:           "",
						InsecureSkipTLSVerify: false,
					},
					Username: "",
					Password: "",
					Token:    "",
				},
			},
		}
	}

	return helmette.Merge(values.Console.Console.Config, c)
}

func ConnectorsFullName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if helmette.Dig(dot.Values.AsMap(), "", "connectors", "connectors", "fullnameOverwrite") != "" {
		return cleanForK8s(values.Connectors.Connectors.FullnameOverwrite)
	}

	return cleanForK8s(fmt.Sprintf("%s-connectors", dot.Release.Name))
}
