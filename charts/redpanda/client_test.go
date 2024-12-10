package redpanda_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/redpanda-data/helm-charts/charts/redpanda"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"github.com/redpanda-data/redpanda/src/go/rpk/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
	"sigs.k8s.io/yaml"
)

type Client struct {
	Ctl           *kube.Ctl
	Release       *helm.Release
	adminClients  map[string]*portForwardClient
	schemaClients map[string]*portForwardClient
	proxyClients  map[string]*portForwardClient
}

type portForwardClient struct {
	http.Client
	exposedPort int
	schema      string
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

	var cfg map[string]any
	if err := yaml.Unmarshal(out.Bytes(), &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Client) CreateTopic(ctx context.Context, topicName string) (map[string]any, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", fmt.Sprintf(`rpk topic create %s -r 3 -p 3`, topicName)},
		Stderr:  &out,
	}); err != nil {
		return nil, err
	}

	var cfg map[string]any
	if err = yaml.Unmarshal(out.Bytes(), &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
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
		return "", errors.New(eb.String())
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
		return nil, errors.New(eb.String())
	}

	var event map[string]any
	if err = json.Unmarshal(out.Bytes(), &event); err != nil {
		return nil, err
	}

	return event, nil
}

func (c *Client) GetClusterHealth(ctx context.Context) (map[string]any, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	client := c.adminClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/v1/cluster/health_overview", client.schema, client.exposedPort), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return nil, errors.New("response above 299 HTTP code")
	}

	var clusterHealth map[string]any
	if err = json.Unmarshal(body, &clusterHealth); err != nil {
		return nil, errors.WithStack(err)
	}

	return clusterHealth, nil
}

func (c *Client) GetSuperusers(ctx context.Context) ([]string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	var out, eb bytes.Buffer
	if err = c.Ctl.Exec(ctx, pod, kube.ExecOptions{
		Command: []string{"bash", "-c", `rpk cluster config get superusers`},
		Stdout:  &out,
		Stderr:  &eb,
	}); err != nil {
		return nil, err
	}

	if eb.String() != "" {
		return nil, errors.New(eb.String())
	}

	superusers := []string{}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "- ") {
			continue
		}

		superusers = append(superusers, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
	}

	return superusers, nil
}

func (c *Client) QuerySupportedFormats(ctx context.Context) ([]string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	client := c.schemaClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/schemas/types", client.schema, client.exposedPort), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return nil, errors.New("response above 299 HTTP code")
	}

	var formats []string
	if err = json.Unmarshal(body, &formats); err != nil {
		return nil, errors.WithStack(err)
	}

	return formats, nil
}

func (c *Client) RegisterSchema(ctx context.Context, schema map[string]any) (map[string]any, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	client := c.schemaClients[pod.Name]

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

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s://127.0.0.1:%d/subjects/sensor-value/versions", client.schema, client.exposedPort), bytes.NewReader(payloadStr))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return nil, errors.New("response above 299 HTTP code")
	}

	var resp map[string]any
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) RetrieveSchema(ctx context.Context, id int) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	client := c.schemaClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/schemas/ids/%d", client.schema, client.exposedPort, id), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return "", errors.New("response above 299 HTTP code")
	}

	var resp map[string]any
	if err = json.Unmarshal(body, &resp); err != nil {
		return "", err
	}

	return resp["schema"].(string), nil
}

func (c *Client) ListRegistrySubjects(ctx context.Context) ([]string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	client := c.schemaClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/subjects", client.schema, client.exposedPort), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return nil, errors.New("response above 299 HTTP code")
	}

	var resp []string
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SoftDeleteSchema(ctx context.Context, subject string, version int) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	client := c.schemaClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s://127.0.0.1:%d/subjects/%s/versions/%d", client.schema, client.exposedPort, subject, version), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	// When soft delete is called for second time the code would be 404 (Not Found)
	// and the body mentioned that subject `was soft deleted.Set permanent=true
	// to delete permanently`.
	if res.StatusCode > 299 && res.StatusCode != 404 {
		return "", errors.Newf("response above 299 HTTP code (Status Code: %d) (Body: %s)", res.StatusCode, body)
	}

	return string(body), nil
}

func (c *Client) HardDeleteSchema(ctx context.Context, subject string, version int) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	_, err = c.SoftDeleteSchema(ctx, subject, version)
	if err != nil {
		return "", errors.WithStack(err)
	}

	client := c.schemaClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s://127.0.0.1:%d/subjects/%s/versions/%d?permanent=true", client.schema, client.exposedPort, subject, version), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return "", errors.New("response above 299 HTTP code")
	}

	return string(body), nil
}

func (c *Client) ListTopics(ctx context.Context) ([]string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, err
	}

	client := c.proxyClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/topics", client.schema, client.exposedPort), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/vnd.kafka.json.v2+json")

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return nil, errors.New("response above 299 HTTP code")
	}

	var resp []string
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SendEventToTopic(ctx context.Context, records map[string]any, topicName string) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	recordsStr, err := json.Marshal(records)
	if err != nil {
		return "", err
	}

	client := c.proxyClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s://127.0.0.1:%d/topics/%s", client.schema, client.exposedPort, topicName), bytes.NewReader(recordsStr))
	if err != nil {
		return "", errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/vnd.kafka.json.v2+json")

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return "", errors.New("response above 299 HTTP code")
	}

	return string(body), nil
}

func (c *Client) RetrieveEventFromTopic(ctx context.Context, topicName string, partitionNumber int) (string, error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return "", err
	}

	client := c.proxyClients[pod.Name]

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s://127.0.0.1:%d/topics/%s/partitions/%d/records?offset=0&timeout=1000&max_bytes=100000", client.schema, client.exposedPort, topicName, partitionNumber), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}
	req.Header.Add("Accept", "application/vnd.kafka.json.v2+json")

	res, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if res.StatusCode > 299 {
		return "", errors.Newf("response above 299 HTTP code (Status Code: %d) (Body: %s)", res.StatusCode, body)
	}

	return string(body), nil
}

// ExposeRedpandaCluster will only expose ports from first (`pod-0`) kafka, Admin API,
// schema registry and HTTP proxy (aka panda proxy) ports.
//
// As future improvement function could expose all ports for each Redpanda. As possible
// returned map of Pod name to map of listener and port could be provided.
func (c *Client) ExposeRedpandaCluster(ctx context.Context, dot *helmette.Dot, out, errOut io.Writer) (func(), error) {
	pod, err := c.getStsPod(ctx, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	availablePorts, cleanup, err := c.Ctl.PortForward(ctx, pod, out, errOut)
	if err != nil {
		return cleanup, errors.WithStack(err)
	}

	if c.adminClients == nil {
		c.adminClients = make(map[string]*portForwardClient)
	}

	if c.schemaClients == nil {
		c.schemaClients = make(map[string]*portForwardClient)
	}

	if c.proxyClients == nil {
		c.proxyClients = make(map[string]*portForwardClient)
	}

	rpYaml, err := c.getRedpandaConfig(ctx)
	if err != nil {
		return cleanup, errors.WithStack(err)
	}

	values := helmette.Unwrap[redpanda.Values](dot.Values)

	defaultSecretName := fmt.Sprintf("%s-%s-%s", c.Release.Name, "default", "cert")

	secretName := defaultSecretName
	cert := values.TLS.Certs[values.Listeners.Admin.TLS.Cert]
	if ref := cert.ClientSecretRef; ref != nil {
		secretName = ref.Name
	}

	adminClient, err := c.createClient(ctx,
		getInternalPort(rpYaml.Redpanda.AdminAPI, availablePorts),
		isTLSEnabled(rpYaml.Redpanda.AdminAPITLS),
		isMutualTLSEnabled(rpYaml.Redpanda.AdminAPITLS),
		secretName)
	if err != nil {
		return cleanup, errors.WithStack(err)
	}

	c.adminClients[pod.Name] = adminClient

	secretName = defaultSecretName
	cert = values.TLS.Certs[values.Listeners.SchemaRegistry.TLS.Cert]
	if ref := cert.ClientSecretRef; ref != nil {
		secretName = ref.Name
	}

	schemaClient, err := c.createClient(ctx,
		getInternalPort(rpYaml.SchemaRegistry.SchemaRegistryAPI, availablePorts),
		isTLSEnabled(rpYaml.SchemaRegistry.SchemaRegistryAPITLS),
		isMutualTLSEnabled(rpYaml.SchemaRegistry.SchemaRegistryAPITLS),
		secretName)
	if err != nil {
		return cleanup, errors.WithStack(err)
	}

	c.schemaClients[pod.Name] = schemaClient

	secretName = defaultSecretName
	cert = values.TLS.Certs[values.Listeners.HTTP.TLS.Cert]
	if ref := cert.ClientSecretRef; ref != nil {
		secretName = ref.Name
	}

	proxyClient, err := c.createClient(ctx,
		getInternalPort(rpYaml.Pandaproxy.PandaproxyAPI, availablePorts),
		isTLSEnabled(rpYaml.Pandaproxy.PandaproxyAPITLS),
		isMutualTLSEnabled(rpYaml.Pandaproxy.PandaproxyAPITLS),
		secretName)
	if err != nil {
		return cleanup, errors.WithStack(err)
	}

	c.proxyClients[pod.Name] = proxyClient

	return cleanup, err
}

func isMutualTLSEnabled(tlsCfg []config.ServerTLS) bool {
	for _, t := range tlsCfg {
		if t.Name != "internal" || !t.Enabled {
			continue
		}
		return t.RequireClientAuth
	}
	return false
}

func isTLSEnabled(tlsCfg []config.ServerTLS) bool {
	for _, t := range tlsCfg {
		if t.Name != "internal" {
			continue
		}
		return t.Enabled
	}
	return false
}

func getInternalPort(addresses any, availablePorts []portforward.ForwardedPort) int {
	var adminListenerPort int
	switch v := addresses.(type) {
	case []config.NamedSocketAddress:
		for _, a := range v {
			if a.Name != "internal" {
				continue
			}
			adminListenerPort = a.Port
		}
	case []config.NamedAuthNSocketAddress:
		for _, a := range v {
			if a.Name != "internal" {
				continue
			}
			adminListenerPort = a.Port
		}
	}

	for _, p := range availablePorts {
		if int(p.Remote) == adminListenerPort {
			return int(p.Local)
		}
	}

	return 0
}

func (c *Client) getRedpandaConfig(ctx context.Context) (*config.RedpandaYaml, error) {
	cm, err := kube.Get[corev1.ConfigMap](ctx, c.Ctl, kube.ObjectKey{
		Name:      c.Release.Name,
		Namespace: c.Release.Namespace,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rpCfg, exist := cm.Data["redpanda.yaml"]
	if !exist {
		return nil, errors.WithStack(fmt.Errorf("redpanda.yaml not found"))
	}

	var cfg config.RedpandaYaml
	err = yaml.Unmarshal([]byte(rpCfg), &cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &cfg, nil
}

func (c *Client) createClient(ctx context.Context, port int, tlsEnabled, mTLSEnabled bool, tlsK8SSecretName string) (*portForwardClient, error) {
	if port == 0 {
		return nil, errors.New("admin internal listener port not found")
	}

	schema := "http"
	var rootCAs *x509.CertPool
	var certs []tls.Certificate
	if tlsEnabled {
		schema = "https"
		s, err := kube.Get[corev1.Secret](ctx, c.Ctl, kube.ObjectKey{
			Name:      tlsK8SSecretName,
			Namespace: c.Release.Namespace,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}

		rootCAs = x509.NewCertPool()
		ok := rootCAs.AppendCertsFromPEM(s.Data["ca.crt"])
		if !ok {
			return nil, errors.WithStack(errors.New("failed to parse CA certificate"))
		}

		if mTLSEnabled {
			cert, err := tls.X509KeyPair(s.Data["tls.crt"], s.Data["tls.key"])
			if err != nil {
				return nil, errors.WithStack(err)
			}
			certs = append(certs, cert)
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: certs,
			RootCAs:      rootCAs,
			// Available subject alternative names are defined in certs.go
			ServerName: fmt.Sprintf("%s.%s", c.Release.Name, c.Release.Namespace),
		},
		TLSHandshakeTimeout:   10 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	httpClient := http.Client{
		Transport: transport,
	}

	pfc := &portForwardClient{
		httpClient,
		port,
		schema,
	}

	return pfc, nil
}
