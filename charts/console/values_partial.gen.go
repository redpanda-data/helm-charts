//go:build !generate

// +gotohelm:ignore=true
//
// Code generated by genpartial DO NOT EDIT.
package console

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type PartialValues struct {
	ReplicaCount                 *int                              "json:\"replicaCount,omitempty\""
	Image                        *PartialImage                     "json:\"image,omitempty\""
	ImagePullSecrets             []corev1.LocalObjectReference     "json:\"imagePullSecrets,omitempty\""
	NameOverride                 *string                           "json:\"nameOverride,omitempty\""
	FullnameOverride             *string                           "json:\"fullnameOverride,omitempty\""
	AutomountServiceAccountToken *bool                             "json:\"automountServiceAccountToken,omitempty\""
	ServiceAccount               *PartialServiceAccountConfig      "json:\"serviceAccount,omitempty\""
	CommonLabels                 map[string]string                 "json:\"commonLabels,omitempty\""
	Annotations                  map[string]string                 "json:\"annotations,omitempty\""
	PodAnnotations               map[string]string                 "json:\"podAnnotations,omitempty\""
	PodLabels                    map[string]string                 "json:\"podLabels,omitempty\""
	PodSecurityContext           *corev1.PodSecurityContext        "json:\"podSecurityContext,omitempty\""
	SecurityContext              *corev1.SecurityContext           "json:\"securityContext,omitempty\""
	Service                      *PartialServiceConfig             "json:\"service,omitempty\""
	Ingress                      *PartialIngressConfig             "json:\"ingress,omitempty\""
	Resources                    *corev1.ResourceRequirements      "json:\"resources,omitempty\""
	Autoscaling                  *PartialAutoScaling               "json:\"autoscaling,omitempty\""
	NodeSelector                 map[string]string                 "json:\"nodeSelector,omitempty\""
	Tolerations                  []corev1.Toleration               "json:\"tolerations,omitempty\""
	Affinity                     *corev1.Affinity                  "json:\"affinity,omitempty\""
	TopologySpreadConstraints    []corev1.TopologySpreadConstraint "json:\"topologySpreadConstraints,omitempty\""
	PriorityClassName            *string                           "json:\"priorityClassName,omitempty\""
	Console                      *PartialConsole                   "json:\"console,omitempty\""
	ExtraEnv                     []corev1.EnvVar                   "json:\"extraEnv,omitempty\""
	ExtraEnvFrom                 []corev1.EnvFromSource            "json:\"extraEnvFrom,omitempty\""
	ExtraVolumes                 []corev1.Volume                   "json:\"extraVolumes,omitempty\""
	ExtraVolumeMounts            []corev1.VolumeMount              "json:\"extraVolumeMounts,omitempty\""
	ExtraContainers              []corev1.Container                "json:\"extraContainers,omitempty\""
	InitContainers               *PartialInitContainers            "json:\"initContainers,omitempty\""
	SecretMounts                 []PartialSecretMount              "json:\"secretMounts,omitempty\""
	Secret                       *PartialSecretConfig              "json:\"secret,omitempty\""
	Enterprise                   *PartialEnterprise                "json:\"enterprise,omitempty\""
	LivenessProbe                *corev1.Probe                     "json:\"livenessProbe,omitempty\""
	ReadinessProbe               *corev1.Probe                     "json:\"readinessProbe,omitempty\""
	ConfigMap                    *PartialCreatable                 "json:\"configmap,omitempty\""
	Deployment                   *PartialCreatable                 "json:\"deployment,omitempty\""
	Strategy                     *appsv1.DeploymentStrategy        "json:\"strategy,omitempty\""
	Tests                        *PartialEnableable                "json:\"tests,omitempty\""
}

type PartialImage struct {
	Registry   *string            "json:\"registry,omitempty\""
	Repository *string            "json:\"repository,omitempty\""
	PullPolicy *corev1.PullPolicy "json:\"pullPolicy,omitempty\""
	Tag        *string            "json:\"tag,omitempty\""
}

type PartialServiceAccountConfig struct {
	Create                       *bool             "json:\"create,omitempty\""
	AutomountServiceAccountToken *bool             "json:\"automountServiceAccountToken,omitempty\""
	Annotations                  map[string]string "json:\"annotations,omitempty\""
	Name                         *string           "json:\"name,omitempty\""
}

type PartialServiceConfig struct {
	Type        *corev1.ServiceType "json:\"type,omitempty\""
	Port        *int32              "json:\"port,omitempty\""
	NodePort    *int32              "json:\"nodePort,omitempty\""
	TargetPort  *int32              "json:\"targetPort,omitempty\""
	Annotations map[string]string   "json:\"annotations,omitempty\""
}

type PartialIngressConfig struct {
	Enabled     *bool                     "json:\"enabled,omitempty\""
	ClassName   *string                   "json:\"className,omitempty\""
	Annotations map[string]string         "json:\"annotations,omitempty\""
	Hosts       []PartialIngressHost      "json:\"hosts,omitempty\""
	TLS         []networkingv1.IngressTLS "json:\"tls,omitempty\""
}

type PartialAutoScaling struct {
	Enabled                           *bool  "json:\"enabled,omitempty\""
	MinReplicas                       *int32 "json:\"minReplicas,omitempty\""
	MaxReplicas                       *int32 "json:\"maxReplicas,omitempty\""
	TargetCPUUtilizationPercentage    *int32 "json:\"targetCPUUtilizationPercentage,omitempty\""
	TargetMemoryUtilizationPercentage *int32 "json:\"targetMemoryUtilizationPercentage,omitempty\""
}

type PartialConsole struct {
	Config       any              "json:\"config,omitempty\""
	Roles        []map[string]any "json:\"roles,omitempty\""
	RoleBindings []map[string]any "json:\"roleBindings,omitempty\""
}

type PartialInitContainers struct {
	ExtraInitContainers *string "json:\"extraInitContainers,omitempty\""
}

type PartialSecretConfig struct {
	Create     *bool                     "json:\"create,omitempty\""
	Kafka      *PartialKafkaSecrets      "json:\"kafka,omitempty\""
	Login      *PartialLoginSecrets      "json:\"login,omitempty\""
	Enterprise *PartialEnterpriseSecrets "json:\"enterprise,omitempty\""
	Redpanda   *PartialRedpandaSecrets   "json:\"redpanda,omitempty\""
}

type PartialEnterprise struct {
	LicenseSecretRef *PartialSecretKeyRef "json:\"licenseSecretRef,omitempty\""
}

type PartialCreatable struct {
	Create *bool "json:\"create,omitempty\""
}

type PartialEnableable struct {
	Enabled *bool "json:\"enabled,omitempty\""
}

type PartialSecretMount struct {
	Name        *string "json:\"name,omitempty\""
	SecretName  *string "json:\"secretName,omitempty\""
	Path        *string "json:\"path,omitempty\""
	DefaultMode *int    "json:\"defaultMode,omitempty\""
}

type PartialKafkaSecrets struct {
	SASLPassword                 *string "json:\"saslPassword,omitempty\""
	AWSMSKIAMSecretKey           *string "json:\"awsMskIamSecretKey,omitempty\""
	TLSCA                        *string "json:\"tlsCa,omitempty\""
	TLSCert                      *string "json:\"tlsCert,omitempty\""
	TLSKey                       *string "json:\"tlsKey,omitempty\""
	TLSPassphrase                *string "json:\"tlsPassphrase,omitempty\""
	SchemaRegistryPassword       *string "json:\"schemaRegistryPassword,omitempty\""
	SchemaRegistryTLSCA          *string "json:\"schemaRegistryTlsCa,omitempty\""
	SchemaRegistryTLSCert        *string "json:\"schemaRegistryTlsCert,omitempty\""
	SchemaRegistryTLSKey         *string "json:\"schemaRegistryTlsKey,omitempty\""
	ProtobufGitBasicAuthPassword *string "json:\"protobufGitBasicAuthPassword,omitempty\""
}

type PartialLoginSecrets struct {
	JWTSecret *string                    "json:\"jwtSecret,omitempty\""
	Google    *PartialGoogleLoginSecrets "json:\"google,omitempty\""
	Github    *PartialGithubLoginSecrets "json:\"github,omitempty\""
	Okta      *PartialOktaLoginSecrets   "json:\"okta,omitempty\""
	OIDC      *PartialOIDCLoginSecrets   "json:\"oidc,omitempty\""
}

type PartialEnterpriseSecrets struct {
	License *string "json:\"License,omitempty\""
}

type PartialRedpandaSecrets struct {
	AdminAPI *PartialRedpandaAdminAPISecrets "json:\"adminApi,omitempty\""
}

type PartialSecretKeyRef struct {
	Name *string "json:\"name,omitempty\""
	Key  *string "json:\"key,omitempty\""
}

type PartialIngressHost struct {
	Host  *string              "json:\"host,omitempty\""
	Paths []PartialIngressPath "json:\"paths,omitempty\""
}

type PartialGoogleLoginSecrets struct {
	ClientSecret         *string "json:\"clientSecret,omitempty\""
	GroupsServiceAccount *string "json:\"groupsServiceAccount,omitempty\""
}

type PartialGithubLoginSecrets struct {
	ClientSecret        *string "json:\"clientSecret,omitempty\""
	PersonalAccessToken *string "json:\"personalAccessToken,omitempty\""
}

type PartialOktaLoginSecrets struct {
	ClientSecret      *string "json:\"clientSecret,omitempty\""
	DirectoryAPIToken *string "json:\"directoryApiToken,omitempty\""
}

type PartialOIDCLoginSecrets struct {
	ClientSecret *string "json:\"clientSecret,omitempty\""
}

type PartialRedpandaAdminAPISecrets struct {
	Password *string "json:\"password,omitempty\""
	TLSCA    *string "json:\"tlsCa,omitempty\""
	TLSCert  *string "json:\"tlsCert,omitempty\""
	TLSKey   *string "json:\"tlsKey,omitempty\""
}

type PartialIngressPath struct {
	Path     *string                "json:\"path,omitempty\""
	PathType *networkingv1.PathType "json:\"pathType,omitempty\""
}
