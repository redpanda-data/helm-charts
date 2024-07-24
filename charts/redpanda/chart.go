// +gotohelm:ignore=true
package redpanda

import (
	_ "embed"
	"reflect"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

var (
	// Scheme is a [runtime.Scheme] with the appropriate extensions to load all
	// objects produced by the redpanda chart.
	Scheme = runtime.NewScheme()

	//go:embed Chart.yaml
	chartYAML []byte

	//go:embed values.yaml
	defaultValuesYAML []byte

	chartMeta helmette.Chart
)

func init() {
	must(scheme.AddToScheme(Scheme))
	must(certmanagerv1.AddToScheme(Scheme))
	must(monitoringv1.AddToScheme(Scheme))

	// NB: We can't directly unmarshal into a helmette.Chart as adding json
	// tags to it breaks gotohelm.
	var chart map[string]any
	must(yaml.Unmarshal(chartYAML, &chart))

	chartMeta = helmette.Chart{
		Name:       chart["name"].(string),
		Version:    chart["version"].(string),
		AppVersion: chart["appVersion"].(string),
	}
}

// ChartMeta returns a parsed version of redpanda's Chart.yaml.
func ChartMeta() helmette.Chart {
	return chartMeta
}

func Template(release helmette.Release, values PartialValues) ([]kube.Object, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, err
	}

	// NB: err1 is working around an issue in gotohelm's ASTs rewrites
	merged, err1 := helm.MergeYAMLValues("", defaultValuesYAML, valuesYaml)
	if err1 != nil {
		return nil, err1
	}

	dot := helmette.Dot{
		Values:  merged,
		Chart:   ChartMeta(),
		Release: release,
	}

	manifests := []kube.Object{
		NodePortService(&dot),
		PodDisruptionBudget(&dot),
		ServiceAccount(&dot),
		ServiceInternal(&dot),
		ServiceMonitor(&dot),
		SidecarControllersRole(&dot),
		SidecarControllersRoleBinding(&dot),
		StatefulSet(&dot),
		PostUpgrade(&dot),
		PostInstallUpgradeJob(&dot),
	}

	manifests = append(manifests, asObj(ConfigMaps(&dot))...)
	manifests = append(manifests, asObj(CertIssuers(&dot))...)
	manifests = append(manifests, asObj(RootCAs(&dot))...)
	manifests = append(manifests, asObj(ClientCerts(&dot))...)
	manifests = append(manifests, asObj(ClusterRoleBindings(&dot))...)
	manifests = append(manifests, asObj(ClusterRoles(&dot))...)
	manifests = append(manifests, asObj(LoadBalancerServices(&dot))...)
	manifests = append(manifests, asObj(Secrets(&dot))...)

	j := 0
	for i := range manifests {
		// Nil unboxing issue
		if reflect.ValueOf(manifests[i]).IsNil() {
			continue
		}
		manifests[j] = manifests[i]
		j++
	}

	return manifests[:j], nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func asObj[T kube.Object](manifests []T) []kube.Object {
	out := make([]kube.Object, len(manifests))
	for i := range manifests {
		out[i] = manifests[i]
	}
	return out
}
