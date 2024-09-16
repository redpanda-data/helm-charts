package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type (
	Config     = clientcmdapi.Config
	RESTConfig = rest.Config
)

var WriteToFile = clientcmd.WriteToFile

func RestToConfig(cfg *rest.Config) clientcmdapi.Config {
	// Thanks to: https://github.com/kubernetes/client-go/issues/711#issuecomment-1666075787

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthority:     cfg.CAFile,
		CertificateAuthorityData: cfg.CAData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		Token:                 cfg.BearerToken,
		TokenFile:             cfg.BearerTokenFile,
		ClientCertificateData: cfg.CertData,
		ClientCertificate:     cfg.CertFile,
		ClientKeyData:         cfg.KeyData,
		ClientKey:             cfg.KeyFile,
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

func ConfigToRest(cfg Config) (*RESTConfig, error) {
	clientConfig := clientcmd.NewNonInteractiveClientConfig(cfg, cfg.CurrentContext, &clientcmd.ConfigOverrides{}, nil)
	return clientConfig.ClientConfig()
}
