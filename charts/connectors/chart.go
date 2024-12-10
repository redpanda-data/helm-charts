// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +gotohelm:filename=_chart.go.tpl
package connectors

import (
	_ "embed"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	// Scheme is a [runtime.Scheme] with the appropriate extensions to load all
	// objects produced by the console chart.
	Scheme = runtime.NewScheme()

	//go:embed Chart.yaml
	chartYAML []byte

	//go:embed values.yaml
	defaultValuesYAML []byte

	// ChartLabel is the go version of the console helm chart.
	Chart = gotohelm.MustLoad(chartYAML, defaultValuesYAML, render)
)

// +gotohelm:ignore=true
func init() {
	must(scheme.AddToScheme(Scheme))
	must(monitoringv1.AddToScheme(Scheme))
}

// +gotohelm:ignore=true
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// render is the entrypoint to both the go and helm versions of the connectors
// helm chart.
// In helm, _shims.render-manifest is used to call and filter the output of
// this function.
// In go, this function should be call by executing [ChartLabel.Render], which will
// handle construction of [helmette.Dot], subcharting, and output filtering.
func render(dot *helmette.Dot) []kube.Object {
	manifests := []kube.Object{
		Deployment(dot),
		PodMonitor(dot),
		Service(dot),
		ServiceAccount(dot),
	}

	// NB: This slice may contain nil interfaces!
	// Filtering happens elsewhere, don't call this function directly if you
	// can avoid it.
	return manifests
}
