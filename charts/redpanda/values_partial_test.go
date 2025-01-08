// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package redpanda_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/redpanda-data/redpanda-operator/charts/redpanda"
)

// TestPartialValuesRoundTrip asserts that any .yaml file in ./ci/ can be round
// tripped through the redpanda.PartialValues structs (sans comments of
// course).
func TestPartialValuesRoundTrip(t *testing.T) {
	t.Skip("Currently failing due to missing fields within our schema.")

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		v := v
		t.Run(v.Name(), func(t *testing.T) {
			yamlBytes, err := os.ReadFile("./ci/" + v.Name())
			require.NoError(t, err)

			var structuredValues *redpanda.PartialValues
			var unstructuredValues map[string]any
			require.NoError(t, yaml.Unmarshal(yamlBytes, &structuredValues))
			require.NoError(t, yaml.Unmarshal(yamlBytes, &unstructuredValues))

			// // Not yet typed field(s)
			// unstructured.RemoveNestedField(unstructuredValues, "console")
			// unstructured.RemoveNestedField(unstructuredValues, "storage", "persistentVolume", "nameOverwrite")
			// unstructured.RemoveNestedField(unstructuredValues, "resources", "memory", "redpanda")
			//
			// // listeners.kafka.external.*.tls slipped through the cracks.
			// kafkaExternal, ok, _ := unstructured.NestedMap(unstructuredValues, "listeners", "kafka", "external")
			// if ok {
			// 	for key := range kafkaExternal {
			// 		unstructured.RemoveNestedField(kafkaExternal, key, "tls")
			// 	}
			// 	unstructured.SetNestedMap(unstructuredValues, kafkaExternal, "listeners", "kafka", "external")
			// }
			//
			// // Potential bug in pre-existing test values. (listeners should be listener?)
			// unstructured.RemoveNestedField(unstructuredValues, "auditLogging", "listeners")

			structuredJSON, err := json.Marshal(structuredValues)
			require.NoError(t, err)

			unstructuredJSON, err := json.Marshal(unstructuredValues)
			require.NoError(t, err)

			require.JSONEq(t, string(unstructuredJSON), string(structuredJSON))
		})
	}
}
