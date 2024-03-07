package redpanda_test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type Client struct {
	Ctl     *kube.Ctl
	Release *helm.Release
}

func (c *Client) getStsPod(ctx context.Context, ordinal int) (*corev1.Pod, error) {
	return kube.Get[corev1.Pod](ctx, c.Ctl, kube.ObjectKey{
		Name:      fmt.Sprintf("%s-%d", c.Release.Name, ordinal),
		Namespace: c.Release.Namespace,
	})
}

func (c *Client) ClusterConfig(ctx context.Context) (redpanda.ClusterConfig, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", `rpk cluster config export -f /dev/stderr`},
		Stderr:  &out,
	}); err != nil {
		return nil, err
	}

	var config map[string]any
	if err := yaml.Unmarshal(out.Bytes(), &config); err != nil {
		return nil, err
	}

	return config, nil
}
