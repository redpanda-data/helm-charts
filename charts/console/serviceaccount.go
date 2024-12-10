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
// +gotohelm:filename=_serviceaccount.go.tpl
package console

import (
	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Create the name of the service account to use
func ServiceAccountName(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.ServiceAccount.Create {
		if values.ServiceAccount.Name != "" {
			return values.ServiceAccount.Name
		}
		return Fullname(dot)
	}

	return helmette.Default("default", values.ServiceAccount.Name)
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
		AutomountServiceAccountToken: ptr.To(values.ServiceAccount.AutomountServiceAccountToken),
	}
}
