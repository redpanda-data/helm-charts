// +gotohelm:ignore=true
package redpanda

import (
	_ "embed"
	"fmt"
	"reflect"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/charts/console"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
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

func Dot(release helmette.Release, values PartialValues, kubeConfig kube.Config) (*helmette.Dot, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, err
	}

	// NB: err1 is working around an issue in gotohelm's ASTs rewrites
	merged, err1 := helm.MergeYAMLValues("", defaultValuesYAML, valuesYaml)
	if err1 != nil {
		return nil, err1
	}

	return &helmette.Dot{
		Values:     merged,
		Chart:      ChartMeta(),
		Release:    release,
		KubeConfig: kubeConfig,
	}, nil
}

func Template(release helmette.Release, values PartialValues, kubeConfig kube.Config) ([]kube.Object, error) {
	dot, err := Dot(release, values, kubeConfig)
	if err != nil {
		return nil, err
	}

	manifests := []kube.Object{
		NodePortService(dot),
		PodDisruptionBudget(dot),
		ServiceAccount(dot),
		ServiceInternal(dot),
		ServiceMonitor(dot),
		SidecarControllersRole(dot),
		SidecarControllersRoleBinding(dot),
		StatefulSet(dot),
		PostUpgrade(dot),
		PostInstallUpgradeJob(dot),
	}

	manifests = append(manifests, asObj(ConfigMaps(dot))...)
	manifests = append(manifests, asObj(CertIssuers(dot))...)
	manifests = append(manifests, asObj(RootCAs(dot))...)
	manifests = append(manifests, asObj(ClientCerts(dot))...)
	manifests = append(manifests, asObj(ClusterRoleBindings(dot))...)
	manifests = append(manifests, asObj(ClusterRoles(dot))...)
	manifests = append(manifests, asObj(LoadBalancerServices(dot))...)
	manifests = append(manifests, asObj(Secrets(dot))...)

	j := 0
	for i := range manifests {
		// Nil unboxing issue
		if reflect.ValueOf(manifests[i]).IsNil() {
			continue
		}
		manifests[j] = manifests[i]
		j++
	}

	manifests = manifests[:j]

	v := helmette.UnmarshalInto[Values](dot.Values)
	consoleValue := helmette.UnmarshalInto[console.PartialValues](v.Console)
	if ptr.Deref(v.Console.Enabled, true) {
		if v.Console.Secret == nil {
			v.Console.Secret = &console.PartialSecretConfig{}
		}
		if !ptr.Deref(v.Console.Secret.Create, false) {
			consoleValue.Secret.Create = ptr.To(true)
			if license := GetLicenseLiteral(dot); license != "" {
				consoleValue.Secret.Enterprise = &console.PartialEnterpriseSecrets{
					License: ptr.To(license),
				}
			}
		}

		if !ptr.Deref(v.Console.ConfigMap.Create, false) {
			if consoleValue.ConfigMap == nil {
				consoleValue.ConfigMap = &console.PartialCreatable{}
			}
			consoleValue.ConfigMap.Create = ptr.To(true)
			configmap := ConsoleConfig(dot)
			if consoleValue.Console == nil {
				consoleValue.Console = &console.PartialConsole{}
			}
			consoleValue.Console.Config = helmette.UnmarshalInto[map[string]any](configmap)
		}

		if !ptr.Deref(v.Console.Deployment.Create, false) {
			if consoleValue.Deployment == nil {
				consoleValue.Deployment = &console.PartialDeploymentConfig{}
			}
			consoleValue.Deployment.Create = ptr.To(true)

			extraVolumes := []corev1.Volume{}
			extraVolumeMounts := []corev1.VolumeMount{}
			if v.Auth.IsSASLEnabled() {
				command := []string{
					"sh",
					"-c",
					strings.Join(
						[]string{
							"set -e; IFS=':' read -r KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM < <(grep \"\" $(find /mnt/users/* -print));",
							fmt.Sprintf("KAFKA_SASL_MECHANISM=${KAFKA_SASL_MECHANISM:-%s};", SASLMechanism(dot)),
							"export KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM;",
							"export KAFKA_SCHEMAREGISTRY_USERNAME=$KAFKA_SASL_USERNAME;",
							"export KAFKA_SCHEMAREGISTRY_PASSWORD=$KAFKA_SASL_PASSWORD;",
							"export REDPANDA_ADMINAPI_USERNAME=$KAFKA_SASL_USERNAME;",
							"export REDPANDA_ADMINAPI_PASSWORD=$KAFKA_SASL_PASSWORD;",
							"/app/console $@",
						}, " "),
					"--",
				}
				consoleValue.Deployment.Command = command
				extraVolumes = append(extraVolumes, corev1.Volume{
					Name: fmt.Sprintf("%s-users", Fullname(dot)),
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: v.Auth.SASL.SecretRef,
						},
					},
				})
				extraVolumeMounts = append(extraVolumeMounts, corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-users", Fullname(dot)),
					MountPath: "/mnt/users",
					ReadOnly:  true,
				})
			}

			extraEnvVars := []corev1.EnvVar{}
			if v.Listeners.Kafka.TLS.IsEnabled(&v.TLS) {
				certName := v.Listeners.Kafka.TLS.Cert
				cert := v.TLS.Certs.MustGet(certName)
				secretName := fmt.Sprintf("%s-%s-cert", Fullname(dot), certName)
				if cert.SecretRef != nil {
					secretName = cert.SecretRef.Name
				}
				if cert.CAEnabled {
					// TODO (Rafal) That could be removed as Config could be defined in ConfigMap
					extraEnvVars = append(extraEnvVars, corev1.EnvVar{
						Name:  "KAFKA_TLS_CAFILEPATH",
						Value: fmt.Sprintf("/mnt/cert/kafka/%s/ca.crt", certName),
					})
					extraVolumes = append(extraVolumes, corev1.Volume{
						Name: fmt.Sprintf("kafka-%s-cert", certName),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: ptr.To[int32](0o420),
								SecretName:  secretName,
							},
						},
					})
					extraVolumeMounts = append(extraVolumeMounts, corev1.VolumeMount{
						Name:      fmt.Sprintf("kafka-%s-cert", certName),
						MountPath: fmt.Sprintf("/mnt/cert/kafka/%s", certName),
						ReadOnly:  true,
					})
				}
			}

			if v.Listeners.SchemaRegistry.TLS.IsEnabled(&v.TLS) {
				certName := v.Listeners.SchemaRegistry.TLS.Cert
				cert := v.TLS.Certs.MustGet(certName)
				secretName := fmt.Sprintf("%s-%s-cert", Fullname(dot), certName)
				if cert.SecretRef != nil {
					secretName = cert.SecretRef.Name
				}
				if cert.CAEnabled {
					// TODO (Rafal) That could be removed as Config could be defined in ConfigMap
					extraEnvVars = append(extraEnvVars, corev1.EnvVar{
						Name:  "KAFKA_SCHEMAREGISTRY_TLS_CAFILEPATH",
						Value: fmt.Sprintf("/mnt/cert/schemaregistry/%s/ca.crt", certName),
					})
					extraVolumes = append(extraVolumes, corev1.Volume{
						Name: fmt.Sprintf("schemaregistry-%s-cert", certName),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: ptr.To[int32](0o420),
								SecretName:  secretName,
							},
						},
					})
					extraVolumeMounts = append(extraVolumeMounts, corev1.VolumeMount{
						Name:      fmt.Sprintf("schemaregistry-%s-cert", certName),
						MountPath: fmt.Sprintf("/mnt/cert/schemaregistry/%s", certName),
						ReadOnly:  true,
					})
				}
			}

			if v.Listeners.Admin.TLS.IsEnabled(&v.TLS) {
				certName := v.Listeners.Admin.TLS.Cert
				cert := v.TLS.Certs.MustGet(certName)
				secretName := fmt.Sprintf("%s-%s-cert", Fullname(dot), certName)
				if cert.SecretRef != nil {
					secretName = cert.SecretRef.Name
				}
				if cert.CAEnabled {
					extraVolumes = append(extraVolumes, corev1.Volume{
						Name: fmt.Sprintf("adminapi-%s-cert", certName),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: ptr.To[int32](0o420),
								SecretName:  secretName,
							},
						},
					})
					extraVolumeMounts = append(extraVolumeMounts, corev1.VolumeMount{
						Name:      fmt.Sprintf("adminapi-%s-cert", certName),
						MountPath: fmt.Sprintf("/mnt/cert/adminapi/%s", certName),
						ReadOnly:  true,
					})
				}
			}

			if secret := GetLicenseSecretReference(dot); secret != nil {
				consoleValue.Enterprise = &console.PartialEnterprise{
					LicenseSecretRef: &console.PartialSecretKeyRef{
						Name: ptr.To(secret.Name),
						Key:  ptr.To(secret.Key),
					},
				}
			}

			consoleValue.ExtraEnv = extraEnvVars
			consoleValue.ExtraVolumes = extraVolumes
			consoleValue.ExtraVolumeMounts = extraVolumeMounts

			consoleDot, err := console.Dot(release, consoleValue, kubeConfig)
			if err != nil {
				return nil, err
			}

			cfg := console.ConfigMap(consoleDot)
			if consoleValue.PodAnnotations == nil {
				consoleValue.PodAnnotations = map[string]string{}
			}
			consoleValue.PodAnnotations["checksum-redpanda-chart/config"] = helmette.Sha256Sum(helmette.ToYaml(cfg))
		}

		objs, err := console.Template(release, consoleValue, kubeConfig)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, objs...)
	}

	return manifests, nil
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
