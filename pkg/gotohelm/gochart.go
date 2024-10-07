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
// it with the default values of this chart.
func (c *GoChart) LoadValues(values any) (helmette.Values, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	merged, err := helm.MergeYAMLValues("", c.defaultValues, valuesYaml)
	return merged, errors.WithStack(err)
}

// Dot constructs a [helmette.Dot] for this chart and any dependencies it has,
// taking into consideration the dependencies' condition.
func (c *GoChart) Dot(cfg kube.Config, release helmette.Release, values helmette.Values) (*helmette.Dot, error) {
	subcharts := map[string]*helmette.Dot{}

	for _, dep := range c.metadata.Dependencies {
		// https://github.com/helm/helm/blob/145d12f82fc7a2e39a17713340825686b661e0a1/pkg/chartutil/dependencies.go#L48
		enabled, err := values.PathValue(dep.Condition)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if asBool, ok := enabled.(bool); ok && !asBool {
			continue
		} else if !ok {
			return nil, errors.Newf("evaluating subchart %q condition %q, expected %t; got: %t (%v)", dep.Name, dep.Condition, true, enabled, enabled)
		}

		subvalues, err := values.Table(dep.Name)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		subchart, ok := c.dependencies[dep.Name]
		if !ok {
			return nil, errors.Newf("missing dependency %q", dep.Name)
		}

		subcharts[dep.Name], err = subchart.Dot(cfg, release, subvalues)
		if err != nil {
			return nil, err
		}
	}

	return &helmette.Dot{
		KubeConfig: cfg,
		Release:    release,
		Subcharts:  subcharts,
		Values:     values,
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
	loaded, err := c.LoadValues(values)
	if err != nil {
		return nil, err
	}

	dot, err := c.Dot(cfg, release, loaded)
	if err != nil {
		return nil, err
	}

	return c.render(dot)
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

	for _, dep := range c.metadata.Dependencies {
		// NB: dot.Subcharts will only contain a dependency is it's condition
		// has been met.
		subdot, ok := dot.Subcharts[dep.Name]
		if !ok {
			continue
		}

		subchart, ok := c.dependencies[dep.Name]
		if !ok {
			return nil, errors.Newf("missing dependency %q", dep.Name)
		}

		subchartManifests, err := subchart.render(subdot)
		if err != nil {
			return nil, err
		}

		manifests = append(manifests, subchartManifests...)
	}

	return manifests, nil
}
