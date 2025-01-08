// Copyright 2025 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// +gotohelm:filename=_configmap.go.tpl
package console

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func ConfigMap(dot *helmette.Dot) *corev1.ConfigMap {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.ConfigMap.Create {
		return nil
	}

	data := map[string]string{
		"config.yaml": fmt.Sprintf("# from .Values.console.config\n%s\n", helmette.Tpl(helmette.ToYaml(values.Console.Config), dot)),
	}

	if len(values.Console.Roles) > 0 {
		data["roles.yaml"] = helmette.Tpl(helmette.ToYaml(map[string]any{
			"roles": values.Console.Roles,
		}), dot)
	}

	if len(values.Console.RoleBindings) > 0 {
		data["role-bindings.yaml"] = helmette.Tpl(helmette.ToYaml(map[string]any{
			"roleBindings": values.Console.RoleBindings,
		}), dot)
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    Labels(dot),
			Name:      Fullname(dot),
			Namespace: dot.Release.Namespace,
		},
		Data: data,
	}
}
