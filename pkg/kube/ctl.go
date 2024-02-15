package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func FromEnv() (*Ctl, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	kubeConfig.ClientConfig()
	config, err := kubeConfig.ClientConfig()
	return &Ctl{config: config}, err
}

type Ctl struct {
	config *rest.Config
}

// RestConfig returns a deep copy of the *[rest.Config] used by this [Ctl].
func (c *Ctl) RestConfig() *rest.Config {
	return rest.CopyConfig(c.config)
}
