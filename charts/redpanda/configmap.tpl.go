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
