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
// +gotohelm:filename=_secret.go.tpl
package console

import (
	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Secret(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Secret.Create {
		return nil
	}

	jwtSecret := values.Secret.Login.JWTSecret
	if jwtSecret == "" {
		jwtSecret = helmette.RandAlphaNum(32)
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   Fullname(dot),
			Labels: Labels(dot),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			// Set empty defaults, so that we can always mount them as env variable even if they are not used.
			// For this reason we can't use `with` to change the scope.

			// Kafka
			"kafka-sasl-password":                   ptr.Deref(values.Secret.Kafka.SASLPassword, ""),
			"kafka-protobuf-git-basicauth-password": ptr.Deref(values.Secret.Kafka.ProtobufGitBasicAuthPassword, ""),
			"kafka-sasl-aws-msk-iam-secret-key":     ptr.Deref(values.Secret.Kafka.AWSMSKIAMSecretKey, ""),
			"kafka-tls-ca":                          ptr.Deref(values.Secret.Kafka.TLSCA, ""),
			"kafka-tls-cert":                        ptr.Deref(values.Secret.Kafka.TLSCert, ""),
			"kafka-tls-key":                         ptr.Deref(values.Secret.Kafka.TLSKey, ""),
			"kafka-schema-registry-password":        ptr.Deref(values.Secret.Kafka.SchemaRegistryPassword, ""),
			"kafka-schemaregistry-tls-ca":           ptr.Deref(values.Secret.Kafka.SchemaRegistryTLSCA, ""),
			"kafka-schemaregistry-tls-cert":         ptr.Deref(values.Secret.Kafka.SchemaRegistryTLSCert, ""),
			"kafka-schemaregistry-tls-key":          ptr.Deref(values.Secret.Kafka.SchemaRegistryTLSKey, ""),

			// Login
			"login-jwt-secret":                         jwtSecret,
			"login-google-oauth-client-secret":         ptr.Deref(values.Secret.Login.Google.ClientSecret, ""),
			"login-google-groups-service-account.json": ptr.Deref(values.Secret.Login.Google.GroupsServiceAccount, ""),
			"login-github-oauth-client-secret":         ptr.Deref(values.Secret.Login.Github.ClientSecret, ""),
			"login-github-personal-access-token":       ptr.Deref(values.Secret.Login.Github.PersonalAccessToken, ""),
			"login-okta-client-secret":                 ptr.Deref(values.Secret.Login.Okta.ClientSecret, ""),
			"login-okta-directory-api-token":           ptr.Deref(values.Secret.Login.Okta.DirectoryAPIToken, ""),
			"login-oidc-client-secret":                 ptr.Deref(values.Secret.Login.OIDC.ClientSecret, ""),

			// Enterprise
			"enterprise-license": ptr.Deref(values.Secret.Enterprise.License, ""),

			// Redpanda
			"redpanda-admin-api-password": ptr.Deref(values.Secret.Redpanda.AdminAPI.Password, ""),
			"redpanda-admin-api-tls-ca":   ptr.Deref(values.Secret.Redpanda.AdminAPI.TLSCA, ""),
			"redpanda-admin-api-tls-cert": ptr.Deref(values.Secret.Redpanda.AdminAPI.TLSCert, ""),
			"redpanda-admin-api-tls-key":  ptr.Deref(values.Secret.Redpanda.AdminAPI.TLSKey, ""),
		},
	}
}
