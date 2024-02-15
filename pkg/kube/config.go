package kube

import (
	"encoding/json"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type RESTConfig = rest.Config

type Config struct{}

var _ json.Marshaler = &Config{}

func ConfigFromFile(path string) (*Config, error) {
	panic("not implemented")
}

func NewConfig(kubeconfig []byte) (*Config, error) {
	panic("not implemented")
}

func (c *Config) MarshalJSON() ([]byte, error) {
	panic("not implemented")
}

func RestToConfig(cfg *rest.Config) clientcmdapi.Config {
	// Thanks to: https://github.com/kubernetes/client-go/issues/711#issuecomment-1666075787

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: cfg.CAData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: cfg.CertData,
		ClientKeyData:         cfg.KeyData,
	}

	return clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
}

var WriteToFile = clientcmd.WriteToFile
