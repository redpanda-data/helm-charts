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
package redpanda

import (
	_ "embed"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm"
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	"github.com/redpanda-data/helm-charts/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	// Scheme is a [runtime.Scheme] with the appropriate extensions to load all
	// objects produced by the redpanda chart.
	Scheme = runtime.NewScheme()

	//go:embed Chart.yaml
	chartYAML []byte

	//go:embed values.yaml
	defaultValuesYAML []byte

	// Chart is the go version of the redpanda helm chart.
	Chart = gotohelm.MustLoad(chartYAML, defaultValuesYAML, render)
)

// +gotohelm:ignore=true
func init() {
	must(scheme.AddToScheme(Scheme))
	must(certmanagerv1.AddToScheme(Scheme))
	must(monitoringv1.AddToScheme(Scheme))
}

// +gotohelm:ignore=true
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// render is the entrypoint to both the go and helm versions of the redpanda
// helm chart.
// In helm, _shims.render-manifest is used to call and filter the output of
// this function.
// In go, this function should be call by executing [Chart.Render], which will
// handle construction of [helmette.Dot], subcharting, and output filtering.
func render(dot *helmette.Dot) []kube.Object {
	manifests := []kube.Object{
		NodePortService(dot),
		PodDisruptionBudget(dot),
		ServiceAccount(dot),
		ServiceInternal(dot),
		ServiceMonitor(dot),
		SidecarControllersRole(dot),
		SidecarControllersRoleBinding(dot),
		StatefulSet(dot),
		PostInstallUpgradeJob(dot),
	}

	// NB: gotohelm doesn't currently have a way to handle casting from
	// []Instance -> []Interface as doing so generally requires some go
	// helpers.
	// Instead, it's easiest (though painful to read and write) to iterate over
	// all functions that return slices and append them one at a time.

	for _, obj := range ConfigMaps(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range CertIssuers(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range RootCAs(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range ClientCerts(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range ClusterRoleBindings(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range ClusterRoles(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range LoadBalancerServices(dot) {
		manifests = append(manifests, obj)
	}

	for _, obj := range Secrets(dot) {
		manifests = append(manifests, obj)
	}

	// NB: This slice may contain nil interfaces!
	// Filtering happens elsewhere, don't call this function directly if you
	// can avoid it.
	return manifests
}
