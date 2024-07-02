package redpanda_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
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

func (c *Client) CreateTopic(ctx context.Context, topicName string) error {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", fmt.Sprintf(`rpk topic create %s -r 3 -p 3`, topicName)},
		Stderr:  &out,
	}); err != nil {
		return err
	}

	var config map[string]any
	if err = yaml.Unmarshal(out.Bytes(), &config); err != nil {
		return err
	}

	return nil
}

func (c *Client) KafkaProduce(ctx context.Context, input, topicName string) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", fmt.Sprintf(`echo %s | rpk topic produce %s`, input, topicName)},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	return out.String(), nil
}

func (c *Client) KafkaConsume(ctx context.Context, topicName string) (map[string]any, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", fmt.Sprintf(`rpk topic consume %s -n 1`, topicName)},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var event map[string]any
	if err = json.Unmarshal(out.Bytes(), &event); err != nil {
		return nil, err
	}

	return event, nil
}

func (c *Client) GetClusterHealth(ctx context.Context, dot *helmette.Dot) (map[string]any, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var curlFlags string
	schema := "http"
	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		schema = "https"
		if values.Listeners.Admin.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s %s://%s:%d/v1/cluster/health_overview %s`, schema, redpanda.InternalDomain(dot), values.Listeners.Admin.Port, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var clusterHealth map[string]any
	if err = json.Unmarshal(out.Bytes(), &clusterHealth); err != nil {
		return nil, err
	}

	return clusterHealth, nil
}

func (c *Client) QuerySupportedFormats(ctx context.Context, dot *helmette.Dot) ([]string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var curlFlags string
	schema := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		schema = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s %s://%s:%d/schemas/types %s`, schema, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var formats []string
	if err = json.Unmarshal(out.Bytes(), &formats); err != nil {
		return nil, err
	}

	return formats, nil
}

func (c *Client) RegisterSchema(ctx context.Context, dot *helmette.Dot, schema map[string]any) (map[string]any, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	schemaStr, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"schema": string(schemaStr),
	}

	payloadStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	curlCMD := fmt.Sprintf(`curl -s -H "Content-Type: application/vnd.schemaregistry.v1+json" -X POST %s://%s:%d/subjects/sensor-value/versions -d '%s' %s`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, payloadStr, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var resp map[string]any
	if err = json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) RetrieveSchema(ctx context.Context, dot *helmette.Dot, id int) (string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s %s://%s:%d/schemas/ids/%d %s`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, id, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	var resp map[string]any
	if err = json.Unmarshal(out.Bytes(), &resp); err != nil {
		return "", err
	}

	return resp["schema"].(string), nil
}

func (c *Client) ListRegistrySubjects(ctx context.Context, dot *helmette.Dot) ([]string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s %s://%s:%d/subjects %s`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var resp []string
	if err = json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SoftDeleteSchema(ctx context.Context, dot *helmette.Dot, subject string, version int) (string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s -X DELETE %s://%s:%d/subjects/%s/versions/%d %s`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, subject, version, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	return out.String(), nil
}

func (c *Client) HardDeleteSchema(ctx context.Context, dot *helmette.Dot, subject string, version int) (string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.SchemaRegistry.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.SchemaRegistry.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s -X DELETE %s://%s:%d/subjects/%s/versions/%d %s &&`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, subject, version, curlFlags)
	curlCMD += fmt.Sprintf(`curl -s -X DELETE %s://%s:%d/subjects/%s/versions/%d?permanent=true %s`, s, redpanda.InternalDomain(dot), values.Listeners.SchemaRegistry.Port, subject, version, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	return out.String(), nil
}

func (c *Client) ListTopics(ctx context.Context, dot *helmette.Dot) ([]string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.HTTP.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.HTTP.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s %s://%s:%d/topics %s`, s, redpanda.InternalDomain(dot), values.Listeners.HTTP.Port, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, fmt.Errorf(eb.String())
	}

	var resp []string
	if err = json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SendEventToTopic(ctx context.Context, dot *helmette.Dot, records map[string]any, topicName string) (string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.HTTP.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.HTTP.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	recordsStr, err := json.Marshal(records)
	if err != nil {
		return "", err
	}

	curlCMD := fmt.Sprintf(`curl -s -X POST -H "Content-Type: application/vnd.kafka.json.v2+json" -d '%s' %s://%s:%d/topics/%s %s`, recordsStr, s, redpanda.InternalDomain(dot), values.Listeners.HTTP.Port, topicName, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	return out.String(), nil
}

func (c *Client) RetrieveEventFromTopic(ctx context.Context, dot *helmette.Dot, topicName string, partitionNumber int) (string, error) {
	values := helmette.Unwrap[redpanda.Values](dot.Values)

	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	var curlFlags string
	s := "http"
	if values.Listeners.HTTP.TLS.IsEnabled(&values.TLS) {
		s = "https"
		if values.Listeners.HTTP.TLS.RequireClientAuth {
			curlFlags = fmt.Sprintf(" --cacert /etc/tls/certs/%s-client/ca.crt --cert /etc/tls/certs/%s-client/tls.crt --key /etc/tls/certs/%s-client/tls.key", redpanda.Fullname(dot), redpanda.Fullname(dot), redpanda.Fullname(dot))
		} else {
			curlFlags = " --cacert /etc/tls/certs/default/ca.crt"
		}
	}

	curlCMD := fmt.Sprintf(`curl -s -H "Accept: application/vnd.kafka.json.v2+json" '%s://%s:%d/topics/%s/partitions/%d/records?offset=0&timeout=1000&max_bytes=100000' %s`, s, redpanda.InternalDomain(dot), values.Listeners.HTTP.Port, topicName, partitionNumber, curlFlags)

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", curlCMD},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return "", err
	}

	if eb.String() != "" {
		return "", fmt.Errorf(eb.String())
	}

	return out.String(), nil
}
