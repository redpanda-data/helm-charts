// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_serviceaccount.go.tpl
package operator

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

// Create the name of the service account to use
func ServiceAccountName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.ServiceAccount.Create {
		if ptr.Deref(values.ServiceAccount.Name, "") != "" {
			return *values.ServiceAccount.Name
		}
		return Fullname(dot)
	}

	return helmette.Default(Fullname(dot), *values.ServiceAccount.Name)
}

func ServiceAccount(dot *helmette.Dot) *corev1.ServiceAccount {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.ServiceAccount.Create {
		return nil
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceAccountName(dot),
			Labels:      Labels(dot),
			Namespace:   dot.Release.Namespace,
			Annotations: values.ServiceAccount.Annotations,
		},
		AutomountServiceAccountToken: values.ServiceAccount.AutomountServiceAccountToken,
	}
}
