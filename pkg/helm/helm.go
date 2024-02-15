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
	"strconv"
	"time"

	"github.com/redpanda-data/helm-charts/pkg/kube"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

const helmTimestampFormat = `2006-01-02 15:04:05.999999999 -0700 MST`

// Time is a wrapper around [time.Time] to match Helm's JSON time format.
type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(t.Time.Format(helmTimestampFormat))), nil
}

func (t *Time) UnmarshalJSON(in []byte) error {
	raw, err := strconv.Unquote(string(in))
	if err != nil {
		return err
	}
	t.Time, err = time.Parse(helmTimestampFormat, raw)
	return err
}

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
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    Time   `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

// Client is a sandboxed programmatic API for the `helm` CLI.
//
// It leverages an isolated HELM_CONFIG_HOME directory to keep operation
// hermetic but shares a global cache to keep network chatter to a minimum. See
// `helm env` for more details.
type Client struct {
	env []string
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
		env: append(os.Environ(), env...),
	}, nil
}

type InstallOptions struct {
	CreateNamespace bool
	Name            string
	Namespace       string
	Values          map[string]any
	Version         string
	NoWait          bool
	NoWaitForJobs   bool
}

func (o *InstallOptions) asFlags() ([]string, error) {
	valuesFile, err := os.CreateTemp(os.TempDir(), "helm-values")
	if err != nil {
		return nil, err
	}

	valuesBytes, err := yaml.Marshal(o.Values)
	if err != nil {
		return nil, err
	}

	if _, err := valuesFile.Write(valuesBytes); err != nil {
		return nil, err
	}

	if err := valuesFile.Close(); err != nil {
		return nil, err
	}

	flags := []string{
		fmt.Sprintf("--namespace=%s", o.Namespace),
		fmt.Sprintf("--values=%s", valuesFile.Name()),
	}

	if o.CreateNamespace {
		flags = append(flags, "--create-namespace")
	}

	if o.Name == "" {
		flags = append(flags, "--generate-name")
	}

	if o.Version != "" {
		flags = append(flags, fmt.Sprintf("--version=%s", o.Version))
	}

	if !o.NoWait {
		flags = append(flags, "--wait")
	}

	if !o.NoWaitForJobs {
		flags = append(flags, "--wait-for-jobs")
	}

	return flags, nil
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

func (c *Client) Install(ctx context.Context, chart string, opts InstallOptions) error {
	flags, err := opts.asFlags()
	if err != nil {
		return err
	}

	args := []string{"install", chart, "--output=json"}
	args = append(args, flags...)

	if opts.Name != "" {
		args = slices.Insert(args, 1, opts.Name)
	}

	stdout, _, err := c.runHelm(ctx, args...)
	if err != nil {
		return err
	}

	// Primarily this is an assertion that we're seeing valid JSON in the
	// output. In the future, this value will be returned (or used to generate
	// a return value).
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return err
	}

	return nil
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

	log.Printf("Executing: %#v", append([]string{"helm"}, args...))
	cmd := exec.CommandContext(ctx, "helm", args...)

	cmd.Env = c.env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.Bytes(), stderr.Bytes(), err
}
