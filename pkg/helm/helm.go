package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type RawYAML []byte

type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Chart struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

type Release struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Revision   int       `json:"revision"`
	DeployedAt time.Time `json:"deployedAt"`
	Status     string    `json:"status"`
	Chart      string    `json:"chart"`
	AppVersion string    `json:"app_version"`
}

// Client is a sandboxed programmatic API for the `helm` CLI.
//
// It leverages an isolated HELM_CONFIG_HOME directory to keep operation
// hermetic but shares a global cache to keep network chatter to a minimum. See
// `helm env` for more details.
type Client struct {
	env        []string
	configHome string
}

type Options struct {
	ConfigHome string
	KubeConfig *rest.Config
}

func (o *Options) asEnv() ([]string, error) {
	if o.ConfigHome == "" {
		var err error
		o.ConfigHome, err = os.MkdirTemp(os.TempDir(), "go-helm-client")
		if err != nil {
			return nil, err
		}
	}

	kubeConfigPath := "/dev/null"
	if o.KubeConfig != nil {
		kubeConfigPath = path.Join(o.ConfigHome, "kubeconfig")
		if err := kube.WriteToFile(kube.RestToConfig(o.KubeConfig), kubeConfigPath); err != nil {
			return nil, err
		}
	}

	return []string{
		fmt.Sprintf("KUBECONFIG=%s", kubeConfigPath),
		fmt.Sprintf("HELM_CONFIG_HOME=%s", path.Join(o.ConfigHome, "helm-config")),
	}, nil
}

// New creates a new helm client.
func New(opts Options) (*Client, error) {
	// Clone the host environment.
	env, err := opts.asEnv()
	if err != nil {
		return nil, err
	}

	return &Client{
		configHome: opts.ConfigHome,
		env:        append(os.Environ(), env...),
	}, nil
}

func (c *Client) List(ctx context.Context) ([]Release, error) {
	stdout, _, err := c.runHelm(ctx, "list", "-A", "--output=json")
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := json.Unmarshal(stdout, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (c *Client) Get(ctx context.Context, namespace, name string) (Release, error) {
	stdout, _, err := c.runHelm(ctx, "get", "metadata", name, "--output=json", "--namespace", namespace)
	if err != nil {
		return Release{}, err
	}

	var release Release
	if err := json.Unmarshal(stdout, &release); err != nil {
		return Release{}, err
	}
	return release, nil
}

func (c *Client) ShowValues(ctx context.Context, chart string, values any) error {
	stdout, _, err := c.runHelm(ctx, "show", "values", chart)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(stdout, &values)
}

func (c *Client) GetValues(ctx context.Context, release *Release, values any) error {
	stdout, _, err := c.runHelm(ctx, "get", "values", release.Name, "--output=json", "--namespace", release.Namespace)
	if err != nil {
		return err
	}
	return json.Unmarshal(stdout, &values)
}

type InstallOptions struct {
	CreateNamespace bool     `flag:"create-namespace"`
	Name            string   `flag:"-"`
	Namespace       string   `flag:"namespace"`
	Values          any      `flag:"-"`
	Version         string   `flag:"version"`
	NoWait          bool     `flag:"wait"`
	NoWaitForJobs   bool     `flag:"wait-for-jobs"`
	GenerateName    bool     `flag:"generate-name"`
	ValuesFile      string   `flag:"values"`
	Set             []string `flag:"set"`
}

func (c *Client) Install(ctx context.Context, chart string, opts InstallOptions) (Release, error) {
	if opts.Name == "" {
		opts.GenerateName = true
	}

	if opts.Values != nil {
		var err error
		opts.ValuesFile, err = c.writeValues(opts.Values)
		if err != nil {
			return Release{}, err
		}
	}

	args := []string{"install", chart, "--output=json"}
	args = append(args, ToFlags(opts)...)

	if opts.Name != "" {
		args = slices.Insert(args, 1, opts.Name)
	}

	stdout, _, err := c.runHelm(ctx, args...)
	if err != nil {
		return Release{}, err
	}

	// TODO(chrisseto): The result of `helm install` appears to be its own
	// unique type. The closest equivalent is `helm get all` but that can't be
	// output as JSON.
	// For now, we scrape out the name and use `helm get metadata` to return
	// consistent information.
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return Release{}, err
	}

	return c.Get(ctx, opts.Namespace, result["name"].(string))
}

type TemplateOptions struct {
	Name         string   `flag:"-"`
	Namespace    string   `flag:"namespace"`
	Values       any      `flag:"-"`
	Version      string   `flag:"version"`
	GenerateName bool     `flag:"generate-name"`
	ValuesFile   string   `flag:"values"`
	Set          []string `flag:"set"`
}

func (c *Client) Template(ctx context.Context, chart string, opts TemplateOptions) ([]byte, error) {
	if opts.Name == "" {
		opts.GenerateName = true
	}

	if opts.Values != nil {
		var err error
		opts.ValuesFile, err = c.writeValues(opts.Values)
		if err != nil {
			return nil, err
		}
	}

	// TODO figure out how to remove kube-version.
	args := []string{"template", chart, "--dry-run=client", "--kube-version=v1.21.0", "--debug"}
	args = append(args, ToFlags(opts)...)

	if opts.Name != "" {
		args = slices.Insert(args, 1, opts.Name)
	}

	stdout, _, err := c.runHelm(ctx, args...)
	if err != nil {
		return nil, err
	}

	return stdout, nil
}

type UpgradeOptions struct {
	CreateNamespace bool     `flag:"create-namespace"`
	Install         bool     `flag:"install"`
	Namespace       string   `flag:"namespace"`
	Version         string   `flag:"version"`
	NoWait          bool     `flag:"wait"`
	NoWaitForJobs   bool     `flag:"wait-for-jobs"`
	ReuseValues     bool     `flag:"reuse-values"`
	Values          any      `flag:"-"`
	ValuesFile      string   `flag:"values"`
	Set             []string `flag:"set"`
}

func (c *Client) Upgrade(ctx context.Context, release, chart string, opts UpgradeOptions) (Release, error) {
	if opts.Values != nil {
		var err error
		opts.ValuesFile, err = c.writeValues(opts.Values)
		if err != nil {
			return Release{}, err
		}
	}

	args := []string{"upgrade", release, chart, "--output=json"}
	args = append(args, ToFlags(opts)...)

	stdout, _, err := c.runHelm(ctx, args...)
	if err != nil {
		return Release{}, err
	}

	// TODO(chrisseto): The result of `helm install` appears to be its own
	// unique type. The closest equivalent is `helm get all` but that can't be
	// output as JSON.
	// For now, we scrape out the name and use `helm get metadata` to return
	// consistent information.
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return Release{}, err
	}

	return c.Get(ctx, opts.Namespace, result["name"].(string))
}

func (c *Client) Test(ctx context.Context, release Release) error {
	stdout, _, err := c.runHelm(ctx, "test", release.Name, "--namespace", release.Namespace, "--logs")
	return errors.Wrapf(err, "stdout: %s", stdout)
}

func (c *Client) RepoList(ctx context.Context) ([]Repo, error) {
	out, _, err := c.runHelm(ctx, "repo", "list", "--output=json")
	if err != nil {
		return nil, err
	}

	var repos []Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (c *Client) RepoAdd(ctx context.Context, name, url string) error {
	_, _, err := c.runHelm(ctx, "repo", "add", name, url)
	return err
}

func (c *Client) Search(ctx context.Context, keyword string) ([]Chart, error) {
	out, _, err := c.runHelm(ctx, "search", "repo", keyword, "--output=json")
	if err != nil {
		return nil, err
	}

	var charts []Chart
	if err := json.Unmarshal(out, &charts); err != nil {
		return nil, err
	}
	return charts, nil
}

func (c *Client) runHelm(ctx context.Context, args ...string) ([]byte, []byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	log.Printf("Executing: %#v", strings.Join(append([]string{"helm"}, args...), " "))
	cmd := exec.CommandContext(ctx, "helm", args...)

	cmd.Env = c.env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), errors.Wrapf(err, "stderr: %s", stderr.String())
}

// writeValues writes a helm values file to a unique file in HELM_CONFIG_HOME
// and returns the path to the written file.
func (c *Client) writeValues(values any) (string, error) {
	valuesFile, err := os.CreateTemp(c.configHome, "values-*.yaml")
	if err != nil {
		return "", err
	}

	valueBytes, ok := values.(RawYAML)
	if !ok {
		valueBytes, err = yaml.Marshal(values)
		if err != nil {
			return "", err
		}
	}

	if _, err := valuesFile.Write(valueBytes); err != nil {
		return "", err
	}

	if err := valuesFile.Close(); err != nil {
		return "", err
	}

	return valuesFile.Name(), nil
}
