package redpanda_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"reflect"
	"slices"
	"testing"
	"testing/quick"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/helm/helmtest"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TieredStorageStatic(t *testing.T) redpanda.PartialValues {
	license := os.Getenv("REDPANDA_LICENSE")
	if license == "" {
		t.Skipf("$REDPANDA_LICENSE is not set")
	}

	return redpanda.PartialValues{
		Config: &redpanda.PartialConfig{
			Node: &redpanda.PartialNodeConfig{
				"developer_mode": true,
			},
		},
		Enterprise: &redpanda.PartialEnterprise{
			License: &license,
		},
		Storage: &redpanda.PartialStorage{
			Tiered: &redpanda.PartialTiered{
				Config: &redpanda.PartialTieredStorageConfig{
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
			Node: &redpanda.PartialNodeConfig{
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
				Config: &redpanda.PartialTieredStorageConfig{
					"cloud_storage_enabled": true,
					"cloud_storage_region":  "a-region",
					"cloud_storage_bucket":  "a-bucket",
				},
			},
		},
	}
}

// TestSchemaBackwardCompat asserts that all values.schema.json are backwards
// compatible by one minor version **provided that no deprecated fields are
// specified**.
func TestSchemaBackwardCompat(t *testing.T) {
	schemaFiles, err := os.ReadDir("testdata/schemas")
	require.NoError(t, err)

	var schemas []*jsonschema.Schema
	var names []string

	// TODO at somepoint we'll have to figure out how to sort these.
	for _, schemaFile := range schemaFiles {
		schemaBytes, err := os.ReadFile(path.Join("testdata/schemas", schemaFile.Name()))
		require.NoError(t, err)

		var data map[string]any
		require.NoError(t, json.Unmarshal(schemaBytes, &data))

		schema, err := valuesutil.UnmarshalInto[*jsonschema.Schema](fixSchema(data))
		require.NoError(t, err)

		schemas = append(schemas, schema)
		names = append(names, schemaFile.Name()[:len(schemaFile.Name())-len(".schema.json")])
	}

	// Inject the most recent chart schema as HEAD to check for any breaking
	// changes across patch versions or new minor versions.
	schemas = append(schemas, redpanda.JSONSchema())
	names = append(names, "HEAD")

	for i := 0; i < len(names)-1; i++ {
		fromName := names[i]
		toName := names[i+1]
		fromSchema := schemas[i]
		toSchema := schemas[i+1]

		t.Run(fmt.Sprintf("%s to %s", fromName, toName), func(t *testing.T) {
			quick.Check(func(values map[string]any, schema *jsonschema.Schema) bool {
				return valuesutil.Validate(schema, values) == nil
			}, &quick.Config{
				Values: func(v []reflect.Value, r *rand.Rand) {
					// Generate a valid value from the previous schema.
					v[0] = reflect.ValueOf(valuesutil.Generate(r, fromSchema))
					// Validate it against the next schema.
					v[1] = reflect.ValueOf(toSchema)
				},
			})
		})
	}
}

func TestTemplateProperties(t *testing.T) {
	// - Setting statefulset.replicas > 1000 causes timeouts.
	// - Not setting advertisedPorts causes out of index errors.
	t.Skip("Currently finding too many failures")

	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	f := func(values *redpanda.PartialValues) error {
		// Helm template can hang on some values. We don't want that to stall
		// tests nor would we want customers to experience that. Cut it off
		// with a deadline.
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, err := client.Template(ctx, ".", helm.TemplateOptions{
			Name:   "redpanda",
			Values: values,
			Set: []string{
				"tests.enabled=false",
			},
		})
		return err
	}

	err = quick.Check(func(values *redpanda.PartialValues) bool {
		return f(values) == nil
	}, &quick.Config{
		Values: func(args []reflect.Value, rng *rand.Rand) {
			values := valuesutil.Generate(rng, redpanda.JSONSchema())
			partial, _ := valuesutil.UnmarshalInto[redpanda.PartialValues](values)
			FixPartialCerts(&partial)
			args[0] = reflect.ValueOf(&partial)
		},
	})
	require.NoError(t, err)
}

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	require.NoError(t, client.RepoAdd(ctx, "redpanda", "https://charts.redpanda.com"))
	require.NoError(t, client.DependencyBuild(ctx, "."), "failed to refresh helm dependencies")

	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	for _, v := range values {
		v := v
		t.Run(v.Name(), func(t *testing.T) {
			t.Parallel()

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
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

			// kube-lint template file
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			inputYaml := bytes.NewBuffer(out)

			cmd := exec.CommandContext(ctx, "kube-linter", "lint", "-", "--format", "json")
			cmd.Stdin = inputYaml
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			errKubeLinter := cmd.Run()
			if errKubeLinter != nil && len(stderr.String()) > 0 {
				t.Logf("kube-linter error(s) found for %q: \n%s\nstderr:\n%s", v.Name(), stdout.String(), stderr.String())
			} else if errKubeLinter != nil {
				t.Logf("kube-linter error(s) found for %q: \n%s", v.Name(), errKubeLinter)
			}
			// TODO: remove comment below and the logging above once we agree to linter
			// require.NoError(t, errKubeLinter)

			testutil.AssertGolden(t, testutil.YAML, "./testdata/"+v.Name()+".golden", out)
		})
	}
}

func TestChart(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skipping log running test...")
	}

	redpandaChart := "."

	env := helmtest.Setup(t).Namespaced(t)

	t.Run("tiered-storage-secrets", func(t *testing.T) {
		ctx := testutil.Context(t)

		credsSecret, err := kube.Create(ctx, env.Ctl(), TieredStorageSecret(env.Namespace()))
		require.NoError(t, err)

		rpRelease := env.Install(redpandaChart, helm.InstallOptions{
			Values: redpanda.PartialValues{
				Config: &redpanda.PartialConfig{
					Node: &redpanda.PartialNodeConfig{
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
}

// FixPartialCerts is a helper for tests utilizing valuesutil.Generate. There's
// no way to constraint values in a JSON schema to those that exist as keys of
// another object (our TLS certs). FixPartialCerts will traverse a
// [redpanda.PartialValues] and retroactively create certificates if they don't
// exist to ensure validity.
func FixPartialCerts(values *redpanda.PartialValues) {
	if values.TLS == nil {
		values.TLS = &redpanda.PartialTLS{Enabled: ptr.To(true)}
	}

	if values.TLS.Certs == nil {
		values.TLS.Certs = &redpanda.PartialTLSCertMap{}
	}

	expectedCerts := []*redpanda.PartialExternalTLS{
		values.Listeners.Admin.TLS,
		values.Listeners.Kafka.TLS,
		values.Listeners.HTTP.TLS,
	}

	for _, cert := range expectedCerts {
		if cert == nil {
			continue
		}

		if _, ok := (*values.TLS.Certs)[*cert.Cert]; ok {
			continue
		}

		(*values.TLS.Certs)[*cert.Cert] = redpanda.PartialTLSCert{CAEnabled: ptr.To(false)}
	}
}

// fixSchema fixes minor issues with older hand written jsonschemas and expands
// arrays in "type" to oneOf's as our jsonschema library requires type to be a
// string.
// NB: Dynamic fixing was elected to ensure that it's easy to repopulate
// testdata/schema with raw git commands.
func fixSchema(schema map[string]any) map[string]any {
	for _, propKey := range []string{"properties", "patternProperties", "additionalProperties"} {
		if props, ok := schema[propKey].(map[string]any); ok {
			for key, value := range props {
				if asMap, ok := value.(map[string]any); ok {
					props[key] = fixSchema(asMap)
				}
			}
		}
	}

	if items, ok := schema["items"].(map[string]any); ok {
		schema["items"] = fixSchema(items)
	}

	if typeArr, ok := schema["type"].([]any); ok {
		var oneOf []map[string]any
		for _, t := range typeArr {
			clone := maps.Clone(schema)
			clone["type"] = t
			oneOf = append(oneOf, clone)
		}
		schema = map[string]any{"oneOf": oneOf}
	}

	// Rename parameters to properties.
	if props, ok := schema["parameters"]; ok {
		schema["properties"] = props
		delete(schema, "parameters")
	}

	// Add missing object types.
	_, hasType := schema["type"]
	_, hasProps := schema["properties"]
	if !hasType && hasProps {
		schema["type"] = "object"
	}

	// Remove any unspecified properties from required.
	if required, ok := schema["required"].([]any); ok {
		for i := 0; i < len(required); i++ {
			key := required[i].(string)
			_, hasProp := schema["properties"].(map[string]any)[key]
			if !hasProp {
				required = slices.Delete(required, i, i+1)
				i--
			}
		}
		schema["required"] = required
	}

	return schema
}
