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
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
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

// ChartLock is a helm lock file for dependencies.
type ChartLock struct {
	// Generated is the date the lock file was last generated.
	Generated time.Time `json:"generated"`
	// Digest is a hash of the dependencies in Chart.yaml.
	Digest string `json:"digest"`
	// Dependencies is the list of dependencies that this lock file has locked.
	Dependencies []*Dependency `json:"dependencies"`
}

// Dependency describes a chart upon which another chart depends.
type Dependency struct {
	// Name is the name of the dependency.
	//
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name"`
	// Version is the version (range) of this chart.
	//
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty"`
	// The URL to the repository.
	//
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	ImportValues []interface{} `json:"import-values,omitempty"`
	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty"`
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
	config     *action.Configuration
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

	registryClient, err := registry.NewClient(registry.ClientOptDebug(true),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptPlainHTTP())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Client{
		config: &action.Configuration{
			// NB: Currently, only `helm template` is "in process" as opposed
			// to sub processing everything. If anything else is migrated to
			// use this configuration, we'll probably want to use the secret
			// storage engine for compatibilities with the `helm` CLI.
			Releases:       storage.Init(driver.NewMemory()),
			RegistryClient: registryClient,
		},
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
	SkipTests    bool     `flag:"skip-tests"`
}

func (c *Client) Template(ctx context.Context, chart string, opts TemplateOptions) ([]byte, error) {
	// NOTE: Unlike other methods, Template calls into helm directly. This is
	// to minimize any potential overhead from go/helm/cobra's start up time
	// and allow us to be much more aggressive with writing tests through
	// Template.
	// TODO: Support IsUpgrade and the like and find a nice way to inject a
	// fake KubeClient.
	client := action.NewInstall(c.config)

	// Options taken, more or less, directly from helm. This make the template
	// command not interact with a Kube APIServer.
	// https://github.com/helm/helm/blob/51a07e7e78cba05126a84c0d890851d7490d2e20/cmd/helm/template.go#L89-L94
	client.APIVersions = chartutil.DefaultVersionSet
	client.ClientOnly = true
	client.DryRun = true
	client.GenerateName = opts.GenerateName
	client.Namespace = opts.Namespace
	client.ReleaseName = opts.Name
	client.Replace = true // Skips "name" checks.

	if client.Namespace == "" {
		client.Namespace = "default"
	}

	// TODO figure out how to remove this. Without it helm complains about K8s
	// compat issues.
	client.KubeVersion = &chartutil.KubeVersion{Version: "v1.21.0", Minor: "21", Major: "1"}

	releaseName, chart, err := client.NameAndChart([]string{chart})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Strange but helm does exactly this. The `client` handles figuring out if
	// it needs to generate a name from .ReleaseName and returns it instead of
	// storing it. It also reads the release name from .ReleaseName so we have
	// to set it ourselves.
	client.ReleaseName = releaseName

	previous := os.Getenv("HELM_CONFIG_HOME")
	os.Setenv("HELM_CONFIG_HOME", path.Join(c.configHome, "helm-config"))
	chart, err = client.ChartPathOptions.LocateChart(chart, &cli.EnvSettings{
		RegistryConfig:   envOr("HELM_REGISTRY_CONFIG", helmpath.ConfigPath("registry/config.json")),
		RepositoryConfig: envOr("HELM_REPOSITORY_CONFIG", helmpath.ConfigPath("repositories.yaml")),
		RepositoryCache:  envOr("HELM_REPOSITORY_CACHE", helmpath.CachePath("repository")),
		Debug:            true,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	os.Setenv("HELM_CONFIG_HOME", previous)

	loadedChart, err := loader.Load(chart)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := action.CheckDependencies(loadedChart, loadedChart.Metadata.Dependencies); err != nil {
		return nil, errors.WithStack(err)
	}

	vOpts := values.Options{Values: opts.Set}

	if opts.Values != nil {
		valuesFile, err := c.writeValues(opts.Values)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		vOpts.ValueFiles = append(vOpts.ValueFiles, valuesFile)
	}

	if opts.ValuesFile != "" {
		vOpts.ValueFiles = append(vOpts.ValueFiles, opts.ValuesFile)
	}

	values, err := vOpts.MergeValues(nil /* getter.Providers that's not used unless a URL is provided. */)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rel, err := client.RunWithContext(ctx, loadedChart, values)
	if err != nil {
		return nil, err
	}

	manifest := bytes.NewBuffer([]byte(rel.Manifest))

	// Hooks are not included in .Manifest and need to be injected into our
	// output. We copy helm's convention of adding a "Source:" header to each
	// file.
	for _, hook := range rel.Hooks {
		if opts.SkipTests && slices.Contains(hook.Events, release.HookTest) {
			continue
		}
		fmt.Fprintf(manifest, "---\n# Source: %s\n%s\n", hook.Path, hook.Manifest)
	}

	return manifest.Bytes(), nil
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
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

func (c *Client) RepoUpdate(ctx context.Context) error {
	_, _, err := c.runHelm(ctx, "repo", "update")
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

func (c *Client) DependencyBuild(ctx context.Context, chartDir string) error {
	_, _, err := c.runHelmInDir(ctx, chartDir, "dep", "build")
	return err
}

func (c *Client) runHelm(ctx context.Context, args ...string) ([]byte, []byte, error) {
	// NB: an empty string will cause os/exec to use it's default of the
	// working directory of the calling process.
	return c.runHelmInDir(ctx, "", args...)
}

func (c *Client) runHelmInDir(ctx context.Context, dir string, args ...string) ([]byte, []byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	log.Printf("Executing: %#v", strings.Join(append([]string{"helm"}, args...), " "))
	cmd := exec.CommandContext(ctx, "helm", args...)

	cmd.Dir = dir
	cmd.Env = c.env
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

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

func GetChartLock(filePath string) (ChartLock, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return ChartLock{}, err
	}

	cl := ChartLock{}
	err = yaml.Unmarshal(b, &cl)
	if err != nil {
		return ChartLock{}, err
	}

	return cl, nil
}

func UpdateChartLock(chartLock ChartLock, filepath string) error {
	b, err := yaml.Marshal(chartLock)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, b, 0o644)
}
