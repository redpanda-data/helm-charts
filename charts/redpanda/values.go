package redpanda

type Values struct {
	NameOverride     string      `json:"nameOverride,omitempty"`
	FullnameOverride string      `json:"fullnameOverride,omitempty"`
	Config           *Config     `json:"config,omitempty"`
	Storage          *Storage    `json:"storage,omitempty"`
	Enterprise       *Enterprise `json:"enterprise,omitempty"`
}

type Enterprise struct {
	License string `json:"license,omitempty"`
}

type Config struct {
	Cluster ClusterConfig `json:"cluster,omitempty"`
}

type ClusterConfig map[string]any

type Storage struct {
	HostPath string  `json:"hostPath,omitempty"`
	Tiered   *Tiered `json:"tiered,omitempty"`
}

type SecretRef struct {
	ConfigurationKey string `json:"configurationKey,omitempty"`
	Key              string `json:"key,omitempty"`
	Name             string `json:"name,omitempty"`
}

type TieredStorageCredentials struct {
	AccessKey *SecretRef `json:"accessKey,omitempty"`
	SecretKey *SecretRef `json:"secretKey,omitempty"`
}

type Tiered struct {
	CredentialsSecretRef *TieredStorageCredentials `json:"credentialsSecretRef,omitempty"`
	Config               map[string]any            `json:"config,omitempty"`
}
