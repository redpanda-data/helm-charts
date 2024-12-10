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
// +gotohelm:filename=_poddisruptionbudget.go.tpl
package redpanda

import (
	"fmt"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func PodDisruptionBudget(dot *helmette.Dot) *policyv1.PodDisruptionBudget {
	values := helmette.Unwrap[Values](dot.Values)
	budget := values.Statefulset.Budget.MaxUnavailable

	// to maintain quorum, raft cannot lose more than half its members
	minReplicas := values.Statefulset.Replicas / 2

	// the lowest we can go is 1 so allow that always
	if budget > 1 && budget > minReplicas {
		panic(fmt.Sprintf("statefulset.budget.maxUnavailable is set too high to maintain quorum: %d > %d", budget, minReplicas))
	}

	maxUnavailable := intstr.FromInt32(int32(budget))
	matchLabels := StatefulSetPodLabelsSelector(dot)
	matchLabels["redpanda.com/poddisruptionbudget"] = Fullname(dot)

	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			MaxUnavailable: &maxUnavailable,
		},
	}
}
