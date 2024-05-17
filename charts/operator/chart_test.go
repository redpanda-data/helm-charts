package operator

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/helm-charts/pkg/testutil"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
)

func TestTemplate(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	// Chart deps are kept within ./charts as a tgz archive, which is git
	// ignored. Helm dep build will ensure that ./charts is in sync with
	// Chart.lock, which is tracked by git.
	require.NoError(t, client.RepoAdd(ctx, "kube-prometheus-stack", "https://prometheus-community.github.io/helm-charts"))
	require.NoError(t, client.DependencyBuild(ctx, "."), "failed to refresh helm dependencies")

	scheme := runtime.NewScheme()
	require.NoError(t, clientscheme.AddToScheme(scheme))
	require.NoError(t, certmanagerv1.AddToScheme(scheme))

	testCases := []struct {
		Name   string
		Values any
		Assert func(*testing.T, []kube.Object)
	}{
		{
			Name:   "defaults",
			Values: map[string]any{},
			Assert: func(t *testing.T, objs []kube.Object) {
				hasCrds := slices.ContainsFunc(objs, func(obj kube.Object) bool {
					_, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
					return ok
				})
				require.False(t, hasCrds, "default values should NOT include CRDs")
			},
		},
		{
			Name: "all-the-rbac",
			Values: map[string]any{
				"rbac": map[string]any{
					"createAdditionalControllerCRs": true,
					"createRPKBundleCRs":            true,
				},
			},
		},
		{
			Name: "no-rbac",
			Values: map[string]any{
				"rbac": map[string]any{
					"create": false,
				},
			},
		},
		{
			Name: "webhook",
			Values: map[string]any{
				"webhook": map[string]any{
					"enabled": true,
				},
			},
		},
		{
			Name: "scope",
			Values: map[string]any{
				"scope": "Cluster",
				"webhook": map[string]any{
					"enabled": true,
				},
			},
		},
	}

	for _, tc := range testCases {
		out, err := client.Template(ctx, ".", helm.TemplateOptions{
			Name:   "redpanda-operator",
			Values: tc.Values,
		})
		require.NoError(t, err)

		testutil.AssertGolden(t, testutil.YAML, fmt.Sprintf("testdata/template-%s.golden", tc.Name), out)

		manifests, err := kube.DecodeYAML(out, scheme)
		require.NoError(t, err)

		if tc.Assert != nil {
			tc.Assert(t, manifests)
		}
	}
}

func TestTemplateRBAC(t *testing.T) {
	ctx := testutil.Context(t)
	client, err := helm.New(helm.Options{ConfigHome: testutil.TempDir(t)})
	require.NoError(t, err)

	scheme := runtime.NewScheme()
	require.NoError(t, apiextensionsv1.AddToScheme(scheme))
	require.NoError(t, certmanagerv1.AddToScheme(scheme))
	require.NoError(t, clientscheme.AddToScheme(scheme))

	testCases := []struct {
		Name   string
		Values any
	}{
		{
			Name: "namespace-scoped",
			Values: map[string]any{
				"scope": "Namespace",
				"rbac": map[string]any{
					"create":                        true,
					"createAdditionalControllerCRs": true,
					"createRPKBundleCRs":            true,
				},
			},
		},
		{
			Name: "cluster-scoped",
			Values: map[string]any{
				"scope": "Cluster",
				"webhook": map[string]any{
					"enabled": true,
				},
				"rbac": map[string]any{
					"create":                        true,
					"createAdditionalControllerCRs": true,
					"createRPKBundleCRs":            true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			out, err := client.Template(ctx, ".", helm.TemplateOptions{
				Name:   "redpanda-operator",
				Values: tc.Values,
			})
			require.NoError(t, err)

			manifests, err := kube.DecodeYAML(out, scheme)
			require.NoError(t, err)

			i := -1
			for j, obj := range manifests {
				switch obj.(type) {
				case *rbacv1.Role, *rbacv1.ClusterRole, *rbacv1.RoleBinding, *rbacv1.ClusterRoleBinding:
					i++
					manifests[i] = manifests[j]
				default:
					continue
				}
			}

			manifests = manifests[:i:i]

			slices.SortStableFunc(manifests, func(a, b kube.Object) int {
				return strings.Compare(a.GetName(), b.GetName())
			})

			encoded, err := kube.EncodeYAML(scheme, manifests...)
			testutil.AssertGolden(t, testutil.YAML, fmt.Sprintf("testdata/rbac-%s.golden", tc.Name), encoded)
		})
	}
}
