// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_chart.go.tpl
package redpanda

import (
	_ "embed"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/redpanda-data/redpanda-operator/charts/connectors"
	"github.com/redpanda-data/redpanda-operator/charts/console"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm"
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	"github.com/redpanda-data/redpanda-operator/pkg/kube"
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
	Chart = gotohelm.MustLoad(chartYAML, defaultValuesYAML, render, console.Chart, connectors.Chart)
)

// Types returns a slice containing the set of all [kube.Object] types that
// could be returned by the redpanda chart.
// +gotohelm:ignore=true
func Types() []kube.Object {
	return []kube.Object{
		&appsv1.Deployment{},
		&appsv1.StatefulSet{},
		&autoscalingv2.HorizontalPodAutoscaler{},
		&batchv1.Job{},
		&certmanagerv1.Certificate{},
		&certmanagerv1.Issuer{},
		&corev1.ConfigMap{},
		&corev1.Secret{},
		&corev1.ServiceAccount{},
		&corev1.Service{},
		&monitoringv1.PodMonitor{},
		&monitoringv1.ServiceMonitor{},
		&networkingv1.Ingress{},
		&policyv1.PodDisruptionBudget{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.ClusterRole{},
		&rbacv1.RoleBinding{},
		&rbacv1.Role{},
	}
}

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

	manifests = append(manifests, consoleChartIntegration(dot)...)

	manifests = append(manifests, connectorsChartIntegration(dot)...)

	// NB: This slice may contain nil interfaces!
	// Filtering happens elsewhere, don't call this function directly if you
	// can avoid it.
	return manifests
}
