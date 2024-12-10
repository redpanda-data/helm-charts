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
// +gotohelm:filename=_configmap.go.tpl
package console

import (
	"fmt"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
