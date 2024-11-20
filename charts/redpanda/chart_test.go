package redpanda_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/redpanda-data/helm-charts/charts/connectors"
	"github.com/redpanda-data/helm-charts/charts/console"
	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/tlsgeneration"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
	"go.uber.org/zap/zaptest"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestChart(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skipping log running test...")
	}

	log := zaptest.NewLogger(t)
	w := &zapio.Writer{Log: log, Level: zapcore.InfoLevel}
	wErr := &zapio.Writer{Log: log, Level: zapcore.ErrorLevel}

	redpandaChart := "."

	h := helmtest.Setup(t)

	t.Run("mtls-using-cert-manager", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		partial := mTLSValuesUsingCertManager()

		rpRelease := env.Install(ctx, redpandaChart, helm.InstallOptions{
			Values: partial,
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		dot := &helmette.Dot{
			Values:  *helmette.UnmarshalInto[*helmette.Values](partial),
			Release: helmette.Release{Name: rpRelease.Name, Namespace: rpRelease.Namespace},
			Chart: helmette.Chart{
				Name: "redpanda",
			},
		}

		cleanup, err := rpk.ExposeRedpandaCluster(ctx, dot, w, wErr)
		if cleanup != nil {
			t.Cleanup(cleanup)
		}
		require.NoError(t, err)

		assert.NoErrorf(t, kafkaListenerTest(ctx, rpk), "Kafka listener sub test failed")
		assert.NoErrorf(t, adminListenerTest(ctx, rpk), "Admin listener sub test failed")
		schemaBytes, retrievedSchema, err := schemaRegistryListenerTest(ctx, rpk)
		assert.JSONEq(t, string(schemaBytes), retrievedSchema)
		assert.NoErrorf(t, err, "Schema Registry listener sub test failed")
		assert.NoErrorf(t, httpProxyListenerTest(ctx, rpk), "HTTP Proxy listener sub test failed")
	})

	t.Run("mtls-using-self-created-certificates", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		serverTLSSecretName := "server-tls-secret"
		clientTLSSecretName := "client-tls-secret"

		partial := mTLSValuesWithProvidedCerts(serverTLSSecretName, clientTLSSecretName)

		r, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1799999999))
		require.NoError(t, err)

		chartReleaseName := fmt.Sprintf("chart-%d", r.Int64())
		ca, sPublic, sPrivate, cPublic, cPrivate, err := tlsgeneration.ClientServerCertificate(chartReleaseName, env.Namespace())
		require.NoError(t, err)

		s := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serverTLSSecretName,
				Namespace: env.Namespace(),
			},
			Data: map[string][]byte{
				"ca.crt":  ca,
				"tls.crt": sPublic,
				"tls.key": sPrivate,
			},
		}
		_, err = kube.Create[corev1.Secret](ctx, env.Ctl(), s)
		require.NoError(t, err)

		c := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clientTLSSecretName,
				Namespace: env.Namespace(),
			},
			Data: map[string][]byte{
				"ca.crt":  ca,
				"tls.crt": cPublic,
				"tls.key": cPrivate,
			},
		}
		_, err = kube.Create[corev1.Secret](ctx, env.Ctl(), c)
		require.NoError(t, err)

		rpRelease := env.Install(ctx, redpandaChart, helm.InstallOptions{
			Values:    partial,
			Name:      chartReleaseName,
			Namespace: env.Namespace(),
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		dot := &helmette.Dot{
			Values:  *helmette.UnmarshalInto[*helmette.Values](partial),
			Release: helmette.Release{Name: rpRelease.Name, Namespace: rpRelease.Namespace},
			Chart: helmette.Chart{
				Name: "redpanda",
			},
		}

		cleanup, err := rpk.ExposeRedpandaCluster(ctx, dot, w, wErr)
		if cleanup != nil {
			t.Cleanup(cleanup)
		}
		require.NoError(t, err)

		assert.NoErrorf(t, kafkaListenerTest(ctx, rpk), "Kafka listener sub test failed")
		assert.NoErrorf(t, adminListenerTest(ctx, rpk), "Admin listener sub test failed")
		schemaBytes, retrievedSchema, err := schemaRegistryListenerTest(ctx, rpk)
		assert.NoErrorf(t, err, "Schema Registry listener sub test failed")
		assert.JSONEq(t, string(schemaBytes), retrievedSchema)
		assert.NoErrorf(t, httpProxyListenerTest(ctx, rpk), "HTTP Proxy listener sub test failed")
	})

	t.Run("admin api auth required", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		partial := redpanda.PartialValues{
			External:      &redpanda.PartialExternalConfig{Enabled: ptr.To(false)},
			ClusterDomain: ptr.To("cluster.local"),
			Config: &redpanda.PartialConfig{
				Cluster: redpanda.PartialClusterConfig{
					"admin_api_require_auth": true,
				},
			},
			Auth: &redpanda.PartialAuth{
				SASL: &redpanda.PartialSASLAuth{
					Enabled: ptr.To(true),
					Users: []redpanda.PartialSASLUser{{
						Name:      ptr.To("superuser"),
						Password:  ptr.To("superpassword"),
						Mechanism: ptr.To("SCRAM-SHA-512"),
					}},
				},
			},
		}

		r, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1799999999))
		require.NoError(t, err)

		chartReleaseName := fmt.Sprintf("chart-%d", r.Int64())
		rpRelease := env.Install(ctx, redpandaChart, helm.InstallOptions{
			Values:    partial,
			Name:      chartReleaseName,
			Namespace: env.Namespace(),
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		dot := &helmette.Dot{
			Values:  *helmette.UnmarshalInto[*helmette.Values](partial),
			Release: helmette.Release{Name: rpRelease.Name, Namespace: rpRelease.Namespace},
			Chart: helmette.Chart{
				Name: "redpanda",
			},
		}

		cleanup, err := rpk.ExposeRedpandaCluster(ctx, dot, w, wErr)
		if cleanup != nil {
			t.Cleanup(cleanup)
		}
		require.NoError(t, err)

		assert.NoErrorf(t, kafkaListenerTest(ctx, rpk), "Kafka listener sub test failed")
		assert.NoErrorf(t, adminListenerTest(ctx, rpk), "Admin listener sub test failed")
		assert.NoErrorf(t, superuserTest(ctx, rpk, "superuser", "kubernetes-controller"), "Superuser sub test failed")
	})

	t.Run("admin api auth required - pre-existing secret", func(t *testing.T) {
		ctx := testutil.Context(t)

		env := h.Namespaced(t)

		env.Ctl().Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-secret",
				Namespace: env.Namespace(),
			},
			StringData: map[string]string{
				"users.txt": "superuser:superpassword:SCRAM-SHA-512",
			},
		})

		partial := redpanda.PartialValues{
			External:      &redpanda.PartialExternalConfig{Enabled: ptr.To(false)},
			ClusterDomain: ptr.To("cluster.local"),
			Config: &redpanda.PartialConfig{
				Cluster: redpanda.PartialClusterConfig{
					"admin_api_require_auth": true,
				},
			},
			Auth: &redpanda.PartialAuth{
				SASL: &redpanda.PartialSASLAuth{
					Enabled:   ptr.To(true),
					SecretRef: ptr.To("my-secret"),
				},
			},
		}

		r, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1799999999))
		require.NoError(t, err)

		chartReleaseName := fmt.Sprintf("chart-%d", r.Int64())
		rpRelease := env.Install(ctx, redpandaChart, helm.InstallOptions{
			Values:    partial,
			Name:      chartReleaseName,
			Namespace: env.Namespace(),
		})

		rpk := Client{Ctl: env.Ctl(), Release: &rpRelease}

		dot := &helmette.Dot{
			Values:  *helmette.UnmarshalInto[*helmette.Values](partial),
			Release: helmette.Release{Name: rpRelease.Name, Namespace: rpRelease.Namespace},
			Chart: helmette.Chart{
				Name: "redpanda",
			},
		}

		cleanup, err := rpk.ExposeRedpandaCluster(ctx, dot, w, wErr)
		if cleanup != nil {
			t.Cleanup(cleanup)
		}
		require.NoError(t, err)

		assert.NoErrorf(t, kafkaListenerTest(ctx, rpk), "Kafka listener sub test failed")
		assert.NoErrorf(t, adminListenerTest(ctx, rpk), "Admin listener sub test failed")
		assert.NoErrorf(t, superuserTest(ctx, rpk, "superuser", "kubernetes-controller"), "Superuser sub test failed")
	})
}

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

func kafkaListenerTest(ctx context.Context, rpk Client) error {
	input := "test-input"
	topicName := "testTopic"
	_, err := rpk.CreateTopic(ctx, topicName)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = rpk.KafkaProduce(ctx, input, topicName)
	if err != nil {
		return errors.WithStack(err)
	}

	consumeOutput, err := rpk.KafkaConsume(ctx, topicName)
	if err != nil {
		return errors.WithStack(err)
	}

	if input != consumeOutput["value"] {
		return fmt.Errorf("expected value %s, got %s", input, consumeOutput["value"])
	}

	return nil
}

func adminListenerTest(ctx context.Context, rpk Client) error {
	deadline := time.After(1 * time.Minute)
	for {
		select {
		case <-time.Tick(5 * time.Second):
			out, err := rpk.GetClusterHealth(ctx)
			if err != nil {
				continue
			}

			if out["is_healthy"].(bool) {
				return nil
			}
		case <-deadline:
			return fmt.Errorf("deadline exceeded")
		case <-ctx.Done():
			return fmt.Errorf("context deadline exceeded")
		}
	}
}

func superuserTest(ctx context.Context, rpk Client, superusers ...string) error {
	deadline := time.After(1 * time.Minute)
	for {
		select {
		case <-time.Tick(5 * time.Second):
			configuredSuperusers, err := rpk.GetSuperusers(ctx)
			if err != nil {
				continue
			}

			if equalElements(configuredSuperusers, superusers) {
				return nil
			}
		case <-deadline:
			return fmt.Errorf("deadline exceeded")
		case <-ctx.Done():
			return fmt.Errorf("context deadline exceeded")
		}
	}
}

func equalElements[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[T]int)
	for _, val := range a {
		counts[val]++
	}

	for _, val := range b {
		if counts[val] == 0 {
			return false
		}
		counts[val]--
	}

	return true
}

func schemaRegistryListenerTest(ctx context.Context, rpk Client) ([]byte, string, error) {
	// Test schema registry
	// Based on https://docs.redpanda.com/current/manage/schema-reg/schema-reg-api/
	formats, err := rpk.QuerySupportedFormats(ctx)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	// There is JSON, PROTOBUF and AVRO formats
	if len(formats) != 3 {
		return nil, "", fmt.Errorf("expected 2 supported formats, got %d", len(formats))
	}

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

	registeredID, err := rpk.RegisterSchema(ctx, schema)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	var id float64
	if idForSchema, ok := registeredID["id"]; ok {
		id = idForSchema.(float64)
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	retrievedSchema, err := rpk.RetrieveSchema(ctx, int(id))
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	resp, err := rpk.ListRegistrySubjects(ctx)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}
	if resp[0] != "sensor-value" {
		return nil, "", fmt.Errorf("expected sensor-value %s, got %s", resp[0], registeredID["id"])
	}

	_, err = rpk.SoftDeleteSchema(ctx, resp[0], int(id))
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	_, err = rpk.HardDeleteSchema(ctx, resp[0], int(id))
	if err != nil {
		return nil, "", errors.WithStack(err)
	}

	return schemaBytes, retrievedSchema, nil
}

type HTTPResponse []struct {
	Topic     string  `json:"topic"`
	Key       *string `json:"key"`
	Value     string  `json:"value"`
	Partition int     `json:"partition"`
	Offset    int     `json:"offset"`
}

func httpProxyListenerTest(ctx context.Context, rpk Client) error {
	// Test http proxy
	// Based on https://docs.redpanda.com/current/develop/http-proxy/
	_, err := rpk.ListTopics(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

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
	_, err = rpk.CreateTopic(ctx, httpTestTopic)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = rpk.SendEventToTopic(ctx, records, httpTestTopic)
	if err != nil {
		return errors.WithStack(err)
	}

	time.Sleep(time.Second * 5)

	record, err := rpk.RetrieveEventFromTopic(ctx, httpTestTopic, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	expectedRecord := HTTPResponse{
		{
			Topic:     httpTestTopic,
			Key:       nil,
			Value:     "Redpanda",
			Partition: 0,
			Offset:    0,
		},
	}

	b, err := json.Marshal(&expectedRecord)
	if err != nil {
		return errors.WithStack(err)
	}

	if string(b) != record {
		return fmt.Errorf("expected record %s, got %s", string(b), record)
	}

	record, err = rpk.RetrieveEventFromTopic(ctx, httpTestTopic, 1)
	if err != nil {
		return errors.WithStack(err)
	}

	expectedRecord = HTTPResponse{
		{
			Topic:     httpTestTopic,
			Key:       nil,
			Value:     "HTTP proxy",
			Partition: 1,
			Offset:    0,
		},
	}

	b, err = json.Marshal(&expectedRecord)
	if err != nil {
		return errors.WithStack(err)
	}

	if string(b) != record {
		return fmt.Errorf("expected record %s, got %s", string(b), record)
	}

	record, err = rpk.RetrieveEventFromTopic(ctx, httpTestTopic, 2)
	if err != nil {
		return errors.WithStack(err)
	}

	expectedRecord = HTTPResponse{
		{
			Topic:     httpTestTopic,
			Key:       nil,
			Value:     "Test event",
			Partition: 2,
			Offset:    0,
		},
	}

	b, err = json.Marshal(&expectedRecord)
	if err != nil {
		return errors.WithStack(err)
	}

	if string(b) != record {
		return fmt.Errorf("expected record %s, got %s", string(b), record)
	}

	return nil
}

func mTLSValuesUsingCertManager() redpanda.PartialValues {
	return redpanda.PartialValues{
		External:      &redpanda.PartialExternalConfig{Enabled: ptr.To(false)},
		ClusterDomain: ptr.To("cluster.local"),
		Listeners: &redpanda.PartialListeners{
			Admin: &redpanda.PartialAdminListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialAdminExternal]{
				//	"default": redpanda.PartialAdminExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
				},
			},
			HTTP: &redpanda.PartialHTTPListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialHTTPExternal]{
				//	"default": redpanda.PartialHTTPExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
				},
			},
			Kafka: &redpanda.PartialKafkaListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialKafkaExternal]{
				//	"default": redpanda.PartialKafkaExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
				},
			},
			SchemaRegistry: &redpanda.PartialSchemaRegistryListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialSchemaRegistryExternal]{
				//	"default": redpanda.PartialSchemaRegistryExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
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
}

func mTLSValuesWithProvidedCerts(serverTLSSecretName, clientTLSSecretName string) redpanda.PartialValues {
	return redpanda.PartialValues{
		External:      &redpanda.PartialExternalConfig{Enabled: ptr.To(false)},
		ClusterDomain: ptr.To("cluster.local"),
		TLS: &redpanda.PartialTLS{
			Enabled: ptr.To(true),
			Certs: redpanda.PartialTLSCertMap{
				"provided": redpanda.PartialTLSCert{
					Enabled:         ptr.To(true),
					CAEnabled:       ptr.To(true),
					SecretRef:       &corev1.LocalObjectReference{Name: serverTLSSecretName},
					ClientSecretRef: &corev1.LocalObjectReference{Name: clientTLSSecretName},
				},
				"default": redpanda.PartialTLSCert{Enabled: ptr.To(false)},
			},
		},
		Listeners: &redpanda.PartialListeners{
			Admin: &redpanda.PartialAdminListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialAdminExternal]{
				//	"default": redpanda.PartialAdminExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
					Cert:              ptr.To("provided"),
				},
			},
			HTTP: &redpanda.PartialHTTPListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialHTTPExternal]{
				//	"default": redpanda.PartialHTTPExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
					Cert:              ptr.To("provided"),
				},
			},
			Kafka: &redpanda.PartialKafkaListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialKafkaExternal]{
				//	"default": redpanda.PartialKafkaExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
					Cert:              ptr.To("provided"),
				},
			},
			SchemaRegistry: &redpanda.PartialSchemaRegistryListeners{
				//External: redpanda.PartialExternalListeners[redpanda.PartialSchemaRegistryExternal]{
				//	"default": redpanda.PartialSchemaRegistryExternal{Enabled: ptr.To(false), Port: ptr.To(int32(0))},
				//},
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
					Cert:              ptr.To("provided"),
				},
			},
			RPC: &struct {
				Port *int32                       `json:"port,omitempty" jsonschema:"required"`
				TLS  *redpanda.PartialInternalTLS `json:"tls,omitempty" jsonschema:"required"`
			}{
				TLS: &redpanda.PartialInternalTLS{
					RequireClientAuth: ptr.To(true),
					Cert:              ptr.To("provided"),
				},
			},
		},
	}
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
			// This guarantee does not currently extend to console.
			Console: &console.PartialValues{Enabled: ptr.To(false)},
			// Nor connectors.
			Connectors: &connectors.PartialValues{Enabled: ptr.To(false)},
		}

		helmValues, err := redpanda.Chart.LoadValues(values)
		require.NoError(t, err)

		dot, err := redpanda.Chart.Dot(kube.Config{}, helmette.Release{
			Name:      "redpanda",
			Namespace: "redpanda",
			Service:   "Helm",
		}, helmValues)
		require.NoError(t, err)

		manifests, err := client.Template(ctx, ".", helm.TemplateOptions{
			Name:      dot.Release.Name,
			Namespace: dot.Release.Namespace,
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

func TestGoHelmEquivalence(t *testing.T) {
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// TODO: Add additional cases for better coverage. Generating random inputs
	// generally results in invalid inputs.
	values := redpanda.PartialValues{
		Enterprise: &redpanda.PartialEnterprise{License: ptr.To("LICENSE_PLACEHOLDER")},
		External: &redpanda.PartialExternalConfig{
			// include, required and tpl are not yet implemented in gotohelm package
			Domain:         ptr.To("{{ trunc 4 .Values.external.prefixTemplate | lower | repeat 3 }}-testing "),
			Type:           ptr.To(corev1.ServiceTypeLoadBalancer),
			PrefixTemplate: ptr.To("$POD_ORDINAL-XYZ-$(echo -n $HOST_IP_ADDRESS | sha256sum | head -c 7)"),
			ExternalDNS:    &redpanda.PartialEnableable{Enabled: ptr.To(true)},
		},
		Statefulset: &redpanda.PartialStatefulset{
			ExtraVolumeMounts: ptr.To(`- name: test-extra-volume
  mountPath: {{ upper "/fake/lifecycle" }}`),
			ExtraVolumes: ptr.To(`- name: test-extra-volume
  secret:
    secretName: {{ trunc 5 .Values.enterprise.license }}-sts-lifecycle
    defaultMode: 0774`),
			InitContainers: GetInitContainer(),
		},
	}

	// We're not interested in tests, console, or connectors so always disable
	// those.
	values.Tests = &struct {
		Enabled *bool "json:\"enabled,omitempty\""
	}{
		Enabled: ptr.To(false),
	}

	values.Console = &console.PartialValues{
		Enabled: ptr.To(true),
		Ingress: &console.PartialIngressConfig{
			Enabled: ptr.To(true),
		},
		Secret: &console.PartialSecretConfig{
			Login: &console.PartialLoginSecrets{
				JWTSecret: ptr.To("JWT_PLACEHOLDER"),
			},
		},
		Tests: &console.PartialEnableable{Enabled: ptr.To(false)},
		// ServiceAccount and AutomountServiceAccountToken could be removed after Console helm chart release
		// Currently there is difference between dependency Console Deployment and ServiceAccount
		ServiceAccount: &console.PartialServiceAccountConfig{
			AutomountServiceAccountToken: ptr.To(false),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
	values.Connectors = &connectors.PartialValues{Enabled: ptr.To(false), Test: &connectors.PartialCreatable{Create: ptr.To(false)}}

	goObjs, err := redpanda.Chart.Render(kube.Config{}, helmette.Release{
		Name:      "gotohelm",
		Namespace: "mynamespace",
		Service:   "Helm",
	}, values)
	require.NoError(t, err)

	rendered, err := client.Template(context.Background(), ".", helm.TemplateOptions{
		Name:      "gotohelm",
		Namespace: "mynamespace",
		Values:    values,
	})
	require.NoError(t, err)

	helmObjs, err := kube.DecodeYAML(rendered, redpanda.Scheme)
	require.NoError(t, err)

	slices.SortStableFunc(helmObjs, func(a, b kube.Object) int {
		aStr := fmt.Sprintf("%s/%s/%s", a.GetObjectKind().GroupVersionKind().String(), a.GetNamespace(), a.GetName())
		bStr := fmt.Sprintf("%s/%s/%s", b.GetObjectKind().GroupVersionKind().String(), b.GetNamespace(), b.GetName())
		return strings.Compare(aStr, bStr)
	})

	slices.SortStableFunc(goObjs, func(a, b kube.Object) int {
		aStr := fmt.Sprintf("%s/%s/%s", a.GetObjectKind().GroupVersionKind().String(), a.GetNamespace(), a.GetName())
		bStr := fmt.Sprintf("%s/%s/%s", b.GetObjectKind().GroupVersionKind().String(), b.GetNamespace(), b.GetName())
		return strings.Compare(aStr, bStr)
	})

	const stsIdx = 14

	// resource.Quantity is a special object. To Ensure they compare correctly,
	// we'll round trip it through JSON so the internal representations will
	// match (assuming the values are actually equal).
	goObjs[stsIdx].(*appsv1.StatefulSet).Spec.Template.Spec.Containers[0].Resources, err = valuesutil.UnmarshalInto[corev1.ResourceRequirements](goObjs[stsIdx].(*appsv1.StatefulSet).Spec.Template.Spec.Containers[0].Resources)
	require.NoError(t, err)

	helmObjs[stsIdx].(*appsv1.StatefulSet).Spec.Template.Spec.Containers[0].Resources, err = valuesutil.UnmarshalInto[corev1.ResourceRequirements](helmObjs[stsIdx].(*appsv1.StatefulSet).Spec.Template.Spec.Containers[0].Resources)
	require.NoError(t, err)

	assert.Equal(t, len(helmObjs), len(goObjs))

	// Iterate and compare instead of a single comparison for better error
	// messages. Some divergences will fail an Equal check on slices but not
	// report which element(s) aren't equal.
	for i := range helmObjs {
		assert.Equal(t, helmObjs[i], goObjs[i])
	}
}

func GetInitContainer() *struct {
	Configurator *struct {
		ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		Resources         map[string]any "json:\"resources,omitempty\""
	} "json:\"configurator,omitempty\""
	FSValidator *struct {
		Enabled           *bool          "json:\"enabled,omitempty\""
		Resources         map[string]any "json:\"resources,omitempty\""
		ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		ExpectedFS        *string        "json:\"expectedFS,omitempty\""
	} "json:\"fsValidator,omitempty\""
	SetDataDirOwnership *struct {
		Enabled           *bool          "json:\"enabled,omitempty\""
		Resources         map[string]any "json:\"resources,omitempty\""
		ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
	} "json:\"setDataDirOwnership,omitempty\""
	SetTieredStorageCacheDirOwnership *struct {
		Resources         map[string]any "json:\"resources,omitempty\""
		ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
	} "json:\"setTieredStorageCacheDirOwnership,omitempty\""
	Tuning *struct {
		Resources         map[string]any "json:\"resources,omitempty\""
		ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
	} "json:\"tuning,omitempty\""
	ExtraInitContainers *string "json:\"extraInitContainers,omitempty\""
} {
	return &struct {
		Configurator *struct {
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
		} "json:\"configurator,omitempty\""
		FSValidator *struct {
			Enabled           *bool          "json:\"enabled,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
			ExpectedFS        *string        "json:\"expectedFS,omitempty\""
		} "json:\"fsValidator,omitempty\""
		SetDataDirOwnership *struct {
			Enabled           *bool          "json:\"enabled,omitempty\""
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"setDataDirOwnership,omitempty\""
		SetTieredStorageCacheDirOwnership *struct {
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"setTieredStorageCacheDirOwnership,omitempty\""
		Tuning *struct {
			Resources         map[string]any "json:\"resources,omitempty\""
			ExtraVolumeMounts *string        "json:\"extraVolumeMounts,omitempty\""
		} "json:\"tuning,omitempty\""
		ExtraInitContainers *string "json:\"extraInitContainers,omitempty\""
	}{
		ExtraInitContainers: ptr.To(`- name: "test-init-container"
  image: "mintel/docker-alpine-bash-curl-jq:latest"
  command: [ "/bin/bash", "-c" ]
  args:
    - |
      set -xe
      echo "Hello World!"`),
	}
}
