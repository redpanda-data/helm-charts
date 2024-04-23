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
// +gotohelm:filename=_statefulset.go.tpl
package redpanda

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
)

const (
	// RedpandaContainerName is the user facing name of the redpanda container
	// in the redpanda StatefulSet. While the name of the container can
	// technically change, this is the name that is used to locate the
	// [corev1.Container] that will be smp'd into the redpanda container.
	RedpandaContainerName = "redpanda"
)

// StatefulSetRedpandaEnv returns the environment variables for the Redpanda
// container of the Redpanda Statefulset.
func StatefulSetRedpandaEnv(dot *helmette.Dot) []corev1.EnvVar {
	values := helmette.Unwrap[Values](dot.Values)

	// Ideally, this would just be a part of the strategic merge patch. While
	// we're moving the chart into go in a piecemeal fashion there isn't a "top
	// level" location to perform the merge so we're instead required to
	// Implement aspects of the SMP by hand.
	userEnv := []corev1.EnvVar{}
	for _, container := range values.Statefulset.PodTemplate.Spec.Containers {
		if container.Name == RedpandaContainerName {
			userEnv = container.Env
		}
	}

	// TODO(chrisseto): Actually implement this as a strategic merge patch.
	// EnvVar's are "last in wins" so there's not too much of a need to fully
	// implement a patch for this usecase.
	return append([]corev1.EnvVar{
		{
			Name: "SERVICE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
	}, userEnv...)
}
