package redpanda_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func TieredStorageStatic(t *testing.T) redpanda.PartialValues {
	license := os.Getenv("REDPANDA_LICENSE")
	if license == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	return redpanda.PartialValues{
		Config: &redpanda.PartialConfig{
			Node: redpanda.PartialNodeConfig{
				"developer_mode": true,
			},
		},
		Enterprise: &redpanda.PartialEnterprise{
			License: &license,
		},
		Storage: &redpanda.PartialStorage{
			Tiered: &redpanda.PartialTiered{
				Config: redpanda.PartialTieredStorageConfig{
					"cloud_storage_enabled":    true,
					"cloud_storage_region":     "static-region",
					"cloud_storage_bucket":     "static-bucket",
					"cloud_storage_access_key": "static-access-key",
					"cloud_storage_secret_key": "static-secret-key",
				},
			},
		},
	}
}

func TieredStorageSecret(namespace string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "tiered-storage-",
			Namespace:    namespace,
		},
		Data: map[string][]byte{
			"access": []byte("from-secret-access-key"),
			"secret": []byte("from-secret-secret-key"),
		},
	}
}

func TieredStorageSecretRefs(t *testing.T, secret *corev1.Secret) redpanda.PartialValues {
	license := os.Getenv("REDPANDA_LICENSE")
	if license == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	access := "access"
	secretKey := "secret"
	return redpanda.PartialValues{
		Config: &redpanda.PartialConfig{
			Node: redpanda.PartialNodeConfig{
				"developer_mode": true,
			},
		},
		Enterprise: &redpanda.PartialEnterprise{
			License: &license,
		},
		Storage: &redpanda.PartialStorage{
			Tiered: &redpanda.PartialTiered{
				CredentialsSecretRef: &redpanda.PartialTieredStorageCredentials{
					AccessKey: &redpanda.PartialSecretRef{Name: &secret.Name, Key: &access},
					SecretKey: &redpanda.PartialSecretRef{Name: &secret.Name, Key: &secretKey},
				},
				Config: redpanda.PartialTieredStorageConfig{
					"cloud_storage_enabled": true,
					"cloud_storage_region":  "a-region",
					"cloud_storage_bucket":  "a-bucket",
				},
			},
		},
	}
}

func TestChart(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skipping log running test...")
	}

	redpandaChart := "."

	h := helmtest.Setup(t)

	t.Run("tiered-storage-secrets", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		credsSecret, err := kube.Create(ctx, env.Ctl(), TieredStorageSecret(env.Namespace()))
		require.NoError(t, err)

		rpRelease := env.Install(redpandaChart, helm.InstallOptions{
			Values: redpanda.PartialValues{
				Config: &redpanda.PartialConfig{
					Node: redpanda.PartialNodeConfig{
						"developer_mode": true,
					},
				},
			},
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		config, err := rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, false, config["cloud_storage_enabled"])

		rpRelease = env.Upgrade(redpandaChart, rpRelease, helm.UpgradeOptions{Values: TieredStorageStatic(t)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "static-access-key", config["cloud_storage_access_key"])

		rpRelease = env.Upgrade(redpandaChart, rpRelease, helm.UpgradeOptions{Values: TieredStorageSecretRefs(t, credsSecret)})

		config, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)
		require.Equal(t, true, config["cloud_storage_enabled"])
		require.Equal(t, "from-secret-access-key", config["cloud_storage_access_key"])
	})

	t.Run("mtls-using-cert-manager", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		partial := redpanda.PartialValues{
			ClusterDomain: ptr.To("cluster.local"),
			Listeners: &redpanda.PartialListeners{
				Admin: &redpanda.PartialAdminListeners{
					TLS: &redpanda.PartialInternalTLS{
						RequireClientAuth: ptr.To(true),
					},
				},
				HTTP: &redpanda.PartialHTTPListeners{
					TLS: &redpanda.PartialInternalTLS{
						RequireClientAuth: ptr.To(true),
					},
				},
				Kafka: &redpanda.PartialKafkaListeners{
					TLS: &redpanda.PartialInternalTLS{
						RequireClientAuth: ptr.To(true),
					},
				},
				SchemaRegistry: &redpanda.PartialSchemaRegistryListeners{
					TLS: &redpanda.PartialInternalTLS{
						RequireClientAuth: ptr.To(true),
					},
				},
				RPC: &struct {
					Port *int32                       `json:"port,omitempty" jsonschema:"required"`
					TLS  *redpanda.PartialInternalTLS `json:"tls,omitempty" jsonschema:"required"`
				}{
					TLS: &redpanda.PartialInternalTLS{
						RequireClientAuth: ptr.To(true),
					},
				},
			},
		}

		rpRelease := env.Install(redpandaChart, helm.InstallOptions{
			Values: partial,
		})

		var val map[string]any
		valByte, err := os.ReadFile("values.yaml")
		require.NoError(t, err)

		require.NoError(t, yaml.Unmarshal(valByte, &val))

		partialB, err := yaml.Marshal(partial)
		require.NoError(t, err)

		var partialVal map[string]any
		require.NoError(t, yaml.Unmarshal(partialB, &partialVal))

		dot := helmette.Dot{Values: helmette.Merge(partialVal, val)}

		dot.Release.Name = rpRelease.Name
		dot.Release.Namespace = rpRelease.Namespace

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}
		_, err = rpk.ClusterConfig(ctx)
		require.NoError(t, err)

		t.Run("kafka-listener", func(t *testing.T) {
			// Test kafka
			input := "test-input"
			require.NoError(t, rpk.CreateTopic(ctx, "testTopic"))

			_, err = rpk.KafkaProduce(ctx, input, "testTopic")
			require.NoError(t, err)

			consumeOutput, err := rpk.KafkaConsume(ctx, "testTopic")
			require.NoError(t, err)
			require.Equal(t, input, consumeOutput["value"])
		})

		t.Run("admin-listener", func(t *testing.T) {
			// Test admin
			out, err := rpk.GetClusterHealth(ctx, &dot)
			require.NoError(t, err)
			require.Equal(t, true, out["is_healthy"])
		})

		t.Run("schema-registry-listener", func(t *testing.T) {
			// Test schema registry
			// Based on https://docs.redpanda.com/current/manage/schema-reg/schema-reg-api/
			formats, err := rpk.QuerySupportedFormats(ctx, &dot)
			require.NoError(t, err)
			require.Len(t, formats, 2)

			schema := map[string]any{
				"type": "record",
				"name": "sensor_sample",
				"fields": []map[string]any{
					{
						"name":        "timestamp",
						"type":        "long",
						"logicalType": "timestamp-millis",
					},
					{
						"name":        "identifier",
						"type":        "string",
						"logicalType": "uuid",
					},
					{
						"name": "value",
						"type": "long",
					},
				},
			}

			registeredID, err := rpk.RegisterSchema(ctx, &dot, schema)
			require.NoError(t, err)

			var id float64
			if idForSchema, ok := registeredID["id"]; ok {
				id = idForSchema.(float64)
			}

			schemaBytes, err := json.Marshal(schema)
			require.NoError(t, err)

			retrievedSchema, err := rpk.RetrieveSchema(ctx, &dot, int(id))
			require.NoError(t, err)
			require.JSONEq(t, string(schemaBytes), retrievedSchema)

			resp, err := rpk.ListRegistrySubjects(ctx, &dot)
			require.NoError(t, err)
			require.Equal(t, "sensor-value", resp[0])

			_, err = rpk.SoftDeleteSchema(ctx, &dot, resp[0], int(id))
			require.NoError(t, err)

			_, err = rpk.HardDeleteSchema(ctx, &dot, resp[0], int(id))
			require.NoError(t, err)
		})

		t.Run("http-proxy-listener", func(t *testing.T) {
			// Test http proxy
			// Based on https://docs.redpanda.com/current/develop/http-proxy/
			topics, err := rpk.ListTopics(ctx, &dot)
			require.NoError(t, err)
			require.Len(t, topics, 2)

			records := map[string]any{
				"records": []map[string]any{
					{
						"value":     "Redpanda",
						"partition": 0,
					},
					{
						"value":     "HTTP proxy",
						"partition": 1,
					},
					{
						"value":     "Test event",
						"partition": 2,
					},
				},
			}

			httpTestTopic := "httpTestTopic"
			require.NoError(t, rpk.CreateTopic(ctx, httpTestTopic))

			_, err = rpk.SendEventToTopic(ctx, &dot, records, httpTestTopic)
			require.NoError(t, err)
			// require.JSONEq(t, "{\"offsets\":[{\"partition\":0,\"offset\":0},{\"partition\":1,\"offset\":0},{\"partition\":2,\"offset\":0}]}", offsets)

			record, err := rpk.RetrieveEventFromTopic(ctx, &dot, httpTestTopic, 0)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("[{\"topic\":\"%s\",\"key\":null,\"value\":\"Redpanda\",\"partition\":0,\"offset\":0}]", httpTestTopic), record)

			record, err = rpk.RetrieveEventFromTopic(ctx, &dot, httpTestTopic, 1)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("[{\"topic\":\"%s\",\"key\":null,\"value\":\"HTTP proxy\",\"partition\":1,\"offset\":0}]", httpTestTopic), record)

			record, err = rpk.RetrieveEventFromTopic(ctx, &dot, httpTestTopic, 2)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("[{\"topic\":\"%s\",\"key\":null,\"value\":\"Test event\",\"partition\":2,\"offset\":0}]", httpTestTopic), record)
		})
	})

	//t.Run("mtls-using-self-created-certificates", func(t *testing.T) {
	//	ctx := testutil.Context(t)
	//
	//	env := h.Namespaced(t)
	//
	//})
}

// preTranspilerChartVersion is the latest release of the Redpanda helm chart prior to the introduction of
// ConfigMap go base implementation. It's used to verify that translated code is functionally equivalent.
const preTranspilerChartVersion = "redpanda-5.8.8.tgz"

func TestConfigMap(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Downloading Redpanda helm chart release is required as client.Template
	// function does not pass HELM_CONFIG_HOME, that prevents from downloading specific
	// Redpanda helm chart version from public helm repository.
	require.NoError(t, client.DownloadFile(ctx, "https://github.com/redpanda-data/helm-charts/releases/download/redpanda-5.8.8/redpanda-5.8.8.tgz", preTranspilerChartVersion))

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		t.Run(v.Name(), func(t *testing.T) {
			t.Parallel()

			// First generate latest released Redpanda charts manifests. From ConfigMap bootstrap,
			// redpanda node configuration and RPK profile.
			manifests, err := client.Template(ctx, filepath.Join(client.GetConfigHome(), preTranspilerChartVersion), helm.TemplateOptions{
				Name:       "redpanda",
				ValuesFile: "./ci/" + v.Name(),
				Set: []string{
					// Tests utilize some non-deterministic helpers (rng). We don't
					// really care about the stability of their output, so globally
					// disable them.
					"tests.enabled=false",
					// jwtSecret defaults to a random string. Can't have that
					// in snapshot testing so set it to a static value.
					"console.secret.login.jwtSecret=SECRETKEY",
				},
			})
			require.NoError(t, err)

			oldRedpanda, oldRPKProfile, err := getConfigMaps(manifests)
			require.NoError(t, err)

			// Now helm template will generate Redpanda configuration from local definition
			manifests, err = client.Template(ctx, ".", helm.TemplateOptions{
				Name:       "redpanda",
				ValuesFile: "./ci/" + v.Name(),
				Set: []string{
					// Tests utilize some non-deterministic helpers (rng). We don't
					// really care about the stability of their output, so globally
					// disable them.
					"tests.enabled=false",
					// jwtSecret defaults to a random string. Can't have that
					// in snapshot testing so set it to a static value.
					"console.secret.login.jwtSecret=SECRETKEY",
				},
			})
			require.NoError(t, err)

			newRedpanda, newRPKProfile, err := getConfigMaps(manifests)
			require.NoError(t, err)

			// Overprovisioned field till Redpanda chart version 5.8.8 was wrongly set to `false`
			// when CPU request value was bellow 1000 mili cores. Function `redpanda-smp`, that
			// should overwrite `overprovisioned` flag in old implementation, was not called
			// before setting `overprovisioned` flag (`{{ dig "cpu" "overprovisioned" false .Values.resources }}`).
			// redpanda-smp template - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_helpers.tpl#L187
			// redpanda-smp template invocation - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_configmap.tpl#L610
			// overprovisioned flag - https://github.com/redpanda-data/helm-charts/blob/5f287d45a3bda2763896840e505fb3de82b968b6/charts/redpanda/templates/_configmap.tpl#L607
			var newUnstructuredRedpandaConf map[string]any
			require.NoError(t, yaml.Unmarshal([]byte(newRedpanda.Data["redpanda.yaml"]), &newUnstructuredRedpandaConf))
			require.NoError(t, unstructured.SetNestedField(newUnstructuredRedpandaConf, false, "rpk", "overprovisioned"))

			require.Equal(t, getJSONObject(t, oldRedpanda.Data["redpanda.yaml"]), newUnstructuredRedpandaConf)
			require.Equal(t, getJSONObject(t, oldRedpanda.Data["bootstrap.yaml"]), getJSONObject(t, newRedpanda.Data["bootstrap.yaml"]))
			require.Equal(t, getJSONObject(t, oldRPKProfile.Data["profile"]), getJSONObject(t, newRPKProfile.Data["profile"]))
		})
	}
}

func getJSONObject(t *testing.T, input string) any {
	var output any
	require.NoError(t, yaml.Unmarshal([]byte(input), &output))
	return output
}

// getConfigMaps is parsing all manifests (resources) created by helm template
// execution. Redpanda helm chart creates 3 distinct files in ConfigMap:
// redpanda.yaml (node, tunable and cluster configuration), bootstrap.yaml
// (only cluster configuration) and profile (external connectivity rpk profile
// which is in different ConfigMap than other two).
func getConfigMaps(manifests []byte) (r *corev1.ConfigMap, rpk *corev1.ConfigMap, err error) {
	objs, err := kube.DecodeYAML(manifests, redpanda.Scheme)
	if err != nil {
		return nil, nil, err
	}

	for _, obj := range objs {
		switch obj := obj.(type) {
		case *corev1.ConfigMap:
			switch obj.Name {
			case "redpanda":
				r = obj
			case "redpanda-rpk":
				rpk = obj
			}
		}
	}

	return r, rpk, nil
}

func TestLabels(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	for _, labels := range []map[string]string{
		{"foo": "bar"},
		{"baz": "1", "quux": "2"},
		// TODO: Add a test for asserting the behavior of adding a commonLabel
		// overriding a builtin value (app.kubernetes.io/name) once the
		// expected behavior is decided.
	} {
		values := &redpanda.PartialValues{
			CommonLabels: labels,
		}

		helmValues, err := valuesutil.UnmarshalInto[helmette.Values](values)
		require.NoError(t, err)

		dot := &helmette.Dot{
			Values: helmValues,
			Chart:  redpanda.ChartMeta(),
			Release: helmette.Release{
				Name:      "redpanda",
				Namespace: "redpanda",
				Service:   "Helm",
			},
		}

		manifests, err := client.Template(ctx, ".", helm.TemplateOptions{
			Name:      dot.Release.Name,
			Namespace: dot.Release.Namespace,
			// This guarantee does not currently extend to console.
			Set: []string{"console.enabled=false"},
			// Nor does it extend to tests.
			SkipTests: true,
			Values:    values,
		})
		require.NoError(t, err)

		objs, err := kube.DecodeYAML(manifests, redpanda.Scheme)
		require.NoError(t, err)

		expectedLabels := redpanda.FullLabels(dot)
		require.Subset(t, expectedLabels, values.CommonLabels, "FullLabels does not contain CommonLabels")

		for _, obj := range objs {
			// Assert that CommonLabels is included on all top level objects.
			require.Subset(t, obj.GetLabels(), expectedLabels, "%T %q", obj, obj.GetName())

			// For other objects (replication controllers) we want to assert
			// that common labels are also included on whatever object (Pod)
			// they generate/contain a template of.
			switch obj := obj.(type) {
			case *appsv1.StatefulSet:
				expectedLabels := maps.Clone(expectedLabels)
				expectedLabels["app.kubernetes.io/component"] += "-statefulset"
				require.Subset(t, obj.Spec.Template.GetLabels(), expectedLabels, "%T/%s's %T", obj, obj.Name, obj.Spec.Template)
			}
		}
	}
}
