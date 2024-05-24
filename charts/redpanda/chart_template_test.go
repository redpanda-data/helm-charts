package redpanda_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/redpanda-data/helm-charts/pkg/valuesutil"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

type TemplateTestCase struct {
	Name       string
	Values     any
	ValuesFile string
	Assert     func(*testing.T, []byte, error)
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

	cases := CIGoldenTestCases(t)
	cases = append(cases, VersionGoldenTestsCases(t)...)
	cases = append(cases, DisableCertmanagerIntegration(t)...)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			out, err := client.Template(ctx, ".", helm.TemplateOptions{
				Name:       "redpanda",
				Values:     tc.Values,
				ValuesFile: tc.ValuesFile,
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

			tc.Assert(t, out, err)

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
				t.Logf("kube-linter error(s) found for %q: \n%s\nstderr:\n%s", tc.Name, stdout.String(), stderr.String())
			} else if errKubeLinter != nil {
				t.Logf("kube-linter error(s) found for %q: \n%s", tc.Name, errKubeLinter)
			}
			// TODO: remove comment below and the logging above once we agree to linter
			// require.NoError(t, errKubeLinter)
		})
	}
}

func CIGoldenTestCases(t *testing.T) []TemplateTestCase {
	values, err := os.ReadDir("./ci")
	require.NoError(t, err)

	cases := make([]TemplateTestCase, len(values))
	for i, f := range values {
		name := f.Name()
		cases[i] = TemplateTestCase{
			Name:       name,
			ValuesFile: "./ci/" + name,
			Assert: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				testutil.AssertGolden(t, testutil.YAML, path.Join("testdata", "ci", name+".golden"), b)
			},
		}
	}
	return cases
}

func VersionGoldenTestsCases(t *testing.T) []TemplateTestCase {
	// A collection of versions that should trigger all the gates guarded by
	// "redpanda-atleast-*" helpers.
	versions := []struct {
		Image  redpanda.PartialImage
		ErrMsg *string
	}{
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.1.0"))},
			ErrMsg: ptr.To("no longer supported"),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.2.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.3.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.3.14"))},
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v22.4.0"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image:  redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.1"))},
			ErrMsg: ptr.To("does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01."),
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.2"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.1.3"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.2.1"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v23.3.0"))},
		},
		{
			Image: redpanda.PartialImage{Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
		},
		{
			Image: redpanda.PartialImage{Repository: ptr.To("somecustomrepo"), Tag: ptr.To(redpanda.ImageTag("v24.1.0"))},
		},
	}

	// A collection of features that are protected by the various above version
	// gates.
	permutations := []redpanda.PartialValues{
		{
			Config: &redpanda.PartialConfig{
				Tunable: &redpanda.PartialTunableConfig{
					"log_segment_size_min":  100,
					"log_segment_size_max":  99999,
					"kafka_batch_max_bytes": 7777,
				},
			},
		},
		{
			Enterprise: &redpanda.PartialEnterprise{License: ptr.To("ATOTALLYVALIDLICENSE")},
		},
		{
			RackAwareness: &redpanda.PartialRackAwareness{
				Enabled:        ptr.To(true),
				NodeAnnotation: ptr.To("topology-label"),
			},
		},
	}

	var cases []TemplateTestCase
	for _, version := range versions {
		version := version
		for i, perm := range permutations {
			values, err := valuesutil.UnmarshalInto[redpanda.PartialValues](perm)
			require.NoError(t, err)

			values.Image = &version.Image

			name := fmt.Sprintf("%s-%s-%d", ptr.Deref(version.Image.Repository, "default"), *version.Image.Tag, i)

			cases = append(cases, TemplateTestCase{
				Name:   name,
				Values: values,
				Assert: func(t *testing.T, b []byte, err error) {
					if version.ErrMsg != nil {
						require.Error(t, err, "expected an error containing %q", *version.ErrMsg)
						require.Contains(t, err.Error(), *version.ErrMsg, "expected an error containing %q", *version.ErrMsg)
						return
					}
					require.NoError(t, err)
					testutil.AssertGolden(t, testutil.YAML, path.Join("testdata", "versions", name+".yaml.golden"), b)
				},
			})
		}
	}
	return cases
}

func DisableCertmanagerIntegration(t *testing.T) []TemplateTestCase {
	assertNoCerts := func(t *testing.T, b []byte, err error) {
		require.NoError(t, err)

		// Assert that no Certificate objects are in the resultant
		// objects when SecretRef is specified AND RequireClientAuth is
		// false.
		objs, err := kube.DecodeYAML(b, redpanda.Scheme)
		require.NoError(t, err)

		for _, obj := range objs {
			_, ok := obj.(*certmanagerv1.Certificate)
			// The -root-certificate is always created right now, ignore that
			// one.
			if ok && strings.HasSuffix(obj.GetName(), "-root-certificate") {
				continue
			}
			require.Falsef(t, ok, "Found unexpected Certificate %q", obj.GetName())
		}

		require.NotContains(t, b, []byte(certmanagerv1.CertificateKind))
	}

	return []TemplateTestCase{
		{
			Name: "disable-cert-manager-overriding-defaults",
			Values: valuesFromYAML(t, `
tls:
  certs:
    default:
      secretRef:
        name: some-secret
    external:
      secretRef:
        name: some-other-secret
`),
			Assert: assertNoCerts,
		},
		{
			Name: "disable-cert-manager-fully-specified",
			Values: valuesFromYAML(t, `
listeners:
  http:
    external:
      default:
        tls:
          cert: for-external
          requireClientAuth: false
    tls:
      cert: for-internal
  kafka:
    external:
      default:
        tls:
          cert: for-external
          requireClientAuth: false
    tls:
      cert: for-internal
  rpc:
    tls:
      cert: for-internal
  schemaRegistry:
    external:
      default:
        tls:
          cert: for-external
          requireClientAuth: false
    tls:
      cert: for-internal
tls:
  certs:
    default:
      enabled: false
    external:
      enabled: false
    for-external:
      secretRef:
        name: some-other-secret
    for-internal:
      secretRef:
        name: some-secret
`),
			Assert: assertNoCerts,
		},
	}
}

func valuesFromYAML(t *testing.T, values string) redpanda.PartialValues {
	// Trim newlines to help with later comparison and avoid any weirdness with
	// loading as it's likely to be written with `` strings.
	values = strings.Trim(values, "\n")

	var partialValues redpanda.PartialValues
	require.NoError(t, yaml.Unmarshal([]byte(values), &partialValues))

	out, err := yaml.Marshal(partialValues)
	require.NoError(t, err)

	// To preserve the sanity of debuggers, require that the value round trips
	// back to the same string. This should catch any typos or miss-indentations
	// that are valid YAML but invalid values.
	require.Equal(t, string(out), values+"\n", "Provided values do NOT round trip. Check for typos and ensure your keys are alphabetized. Re-marshaled values:\n%s\n", out)

	return partialValues
}
