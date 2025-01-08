// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_poddisruptionbudget.go.tpl
package redpanda

import (
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
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
