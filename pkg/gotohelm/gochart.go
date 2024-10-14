package gotohelm

import (
	"reflect"

	"github.com/cockroachdb/errors"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/helm"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

type RenderFunc func(*helmette.Dot) []kube.Object

type GoChart struct {
	metadata      chart.Metadata
	defaultValues []byte
	renderFunc    RenderFunc
	dependencies  map[string]*GoChart
}

// MustLoad delegates to [Load] but panics upon any errors.
func MustLoad(chartYAML, defaultValuesYAML []byte, render RenderFunc, dependencies ...*GoChart) *GoChart {
	chart, err := Load(chartYAML, defaultValuesYAML, render, dependencies...)
	if err != nil {
		panic(err)
	}
	return chart
}

// Load hydrates a [GoChart] from helm YAML files and a top level [RenderFunc].
func Load(chartYAML, defaultValuesYAML []byte, render RenderFunc, dependencies ...*GoChart) (*GoChart, error) {
	var meta chart.Metadata

	if err := yaml.Unmarshal(chartYAML, &meta); err != nil {
		return nil, err
	}

	deps := map[string]*GoChart{}
	for _, dep := range dependencies {
		deps[dep.metadata.Name] = dep
	}

	return &GoChart{
		metadata:      meta,
		defaultValues: defaultValuesYAML,
		renderFunc:    render,
		dependencies:  deps,
	}, nil
}

// LoadValues coheres the provided values into a [helmette.Values] and merges
// it with the default values of this chart. Dependencies are not loaded.
func (c *GoChart) LoadValues(values any) (helmette.Values, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	merged, err := helm.MergeYAMLValues("", c.defaultValues, valuesYaml)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return merged, nil
}

func isDependencyEnabled(val helmette.Values, dep *chart.Dependency) (bool, error) {
	// https://github.com/helm/helm/blob/145d12f82fc7a2e39a17713340825686b661e0a1/pkg/chartutil/dependencies.go#L48
	if dep.Condition == "" {
		return true, nil
	}

	enabled, err := val.PathValue(dep.Condition)
	if err != nil {
		return false, errors.WithStack(err)
	}

	asBool, ok := enabled.(bool)
	if !ok {
		return false, errors.Newf("evaluating subchart %q condition %q, expected %t; got: %t (%v)", dep.Name, dep.Condition, true, enabled, enabled)
	}

	return asBool, nil
}

func mergeRootValueWithDependency(rootValues helmette.Values, dependencyValues helmette.Values, dep *chart.Dependency) (helmette.Values, error) {
	root, err := rootValues.YAML()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dependency, err := helmette.Values{dep.Name: dependencyValues}.YAML()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	merged, err := helm.MergeYAMLValues("", []byte(root), []byte(dependency))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return merged, nil
}

// Dot constructs a [helmette.Dot] for this chart and any dependencies it has,
// taking into consideration the dependencies' condition.
func (c *GoChart) Dot(cfg kube.Config, release helmette.Release, values any) (*helmette.Dot, error) {
	subcharts := map[string]*helmette.Dot{}

	loaded, err := c.LoadValues(values)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, dep := range c.metadata.Dependencies {
		subchart, ok := c.dependencies[dep.Name]
		if !ok {
			return nil, errors.Newf("missing dependency %q", dep.Name)
		}

		subvalues, err := loaded.Table(dep.Name)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// The global key is added by helm
		subvalues["global"] = struct{}{}

		// The LoadValues could be less compute intensive as Dot is recursive and LoadValues is not
		subchartDot, err := subchart.Dot(cfg, release, subvalues)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		mergedWithDep, err := mergeRootValueWithDependency(loaded, subchartDot.Values, dep)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		enabled, err := isDependencyEnabled(mergedWithDep, dep)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if !enabled {
			// When chart does not match condition then global is removed
			delete(subvalues, "global")
			continue
		}
		loaded = mergedWithDep
		subcharts[dep.Name] = subchartDot
	}

	return &helmette.Dot{
		KubeConfig: cfg,
		Release:    release,
		Subcharts:  subcharts,
		Values:     loaded,
		Chart: helmette.Chart{
			Name:       c.metadata.Name,
			Version:    c.metadata.Version,
			AppVersion: c.metadata.AppVersion,
		},
	}, nil
}

// Render is the golang equivalent of invoking `helm template/install/upgrade`
// with the exception of excluding NOTES.txt.
//
// Helm hooks are included in the returned slice, it's up to the caller
// to filter them.
func (c *GoChart) Render(cfg kube.Config, release helmette.Release, values any) ([]kube.Object, error) {
	dot, err := c.Dot(cfg, release, values)
	if err != nil {
		return nil, err
	}

	return c.render(dot)
}

// Metadata returns the parsed [chart.Metadata] describing this chart.
func (c *GoChart) Metadata() chart.Metadata {
	return c.metadata
}

// doRender is a helper to catch any panics from renderFunc and convert them to
// errors.
func (c *GoChart) doRender(dot *helmette.Dot) (_ []kube.Object, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Newf("chart execution failed: %#v", r)
		}
	}()

	manifests := c.renderFunc(dot)

	// renderFunc is expected to return nil interfaces.
	// In the helm world, these nils are filtered out by
	// _shims.render-manifests.
	j := 0
	for i := range manifests {
		// Handle the nil unboxing issue.
		if reflect.ValueOf(manifests[i]).IsNil() {
			continue
		}
		manifests[j] = manifests[i]
		j++
	}

	return manifests[:j], nil
}

func (c *GoChart) render(dot *helmette.Dot) ([]kube.Object, error) {
	manifests, err := c.doRender(dot)
	if err != nil {
		return nil, err
	}

	for _, depDot := range dot.Subcharts {
		subchart, ok := c.dependencies[depDot.Chart.Name]
		if !ok {
			return nil, errors.Newf("missing dependency %q", depDot.Chart.Name)
		}

		subchartManifests, err := subchart.render(depDot)
		if err != nil {
			return nil, err
		}

		manifests = append(manifests, subchartManifests...)
	}

	return manifests, nil
}
