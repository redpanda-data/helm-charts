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
// +gotohelm:filename=_notes.go.tpl
package redpanda

import (
	"fmt"

	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
)

func Warnings(dot *helmette.Dot) []string {
	var warnings []string
	if w := cpuWarning(dot); w != "" {
		warnings = append(warnings, fmt.Sprintf(`**Warning**: %s`, w))
	}
	return warnings
}

func cpuWarning(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	coresInMillis := values.Resources.CPU.Cores.MilliValue()
	if coresInMillis < 1000 {
		return fmt.Sprintf("%dm is below the minimum recommended CPU value for Redpanda", coresInMillis)
	}
	return ""
}

func Notes(dot *helmette.Dot) []string {
	values := helmette.Unwrap[Values](dot.Values)

	anySASL := values.Auth.IsSASLEnabled()
	var notes []string
	notes = append(notes,
		``, ``, ``, ``,
		fmt.Sprintf(`Congratulations on installing %s!`, dot.Chart.Name),
		``,
		`The pods will rollout in a few seconds. To check the status:`,
		``,
		fmt.Sprintf(`  kubectl -n %s rollout status statefulset %s --watch`,
			dot.Release.Namespace,
			Fullname(dot),
		),
	)
	if values.External.Enabled && values.External.Type == corev1.ServiceTypeLoadBalancer {
		notes = append(notes,
			``,
			`If you are using the load balancer service with a cloud provider, the services will likely have automatically-generated addresses. In this scenario the advertised listeners must be updated in order for external access to work. Run the following command once Redpanda is deployed:`,
			``,
			// Yes, this really is a jsonpath string to be exposed to the user
			fmt.Sprintf(`  helm upgrade %s redpanda/redpanda --reuse-values -n %s --set $(kubectl get svc -n %s -o jsonpath='{"external.addresses={"}{ range .items[*]}{.status.loadBalancer.ingress[0].ip }{.status.loadBalancer.ingress[0].hostname}{","}{ end }{"}\n"}')`,
				Name(dot),
				dot.Release.Namespace,
				dot.Release.Namespace,
			),
		)
	}
	profiles := maps.Keys(values.Listeners.Kafka.External)
	helmette.SortAlpha(profiles)
	profileName := profiles[0]
	notes = append(notes,
		``,
		`Set up rpk for access to your external listeners:`,
	)
	profile := values.Listeners.Kafka.External[profileName]
	if TLSEnabled(dot) {
		var external string
		if profile.TLS != nil && profile.TLS.Cert != nil {
			external = *profile.TLS.Cert
		} else {
			external = values.Listeners.Kafka.TLS.Cert
		}
		notes = append(notes,
			fmt.Sprintf(`  kubectl get secret -n %s %s-%s-cert -o go-template='{{ index .data "ca.crt" | base64decode }}' > ca.crt`,
				dot.Release.Namespace,
				Fullname(dot),
				external,
			),
		)
		if values.Listeners.Kafka.TLS.RequireClientAuth || values.Listeners.Admin.TLS.RequireClientAuth {
			notes = append(notes,
				fmt.Sprintf(`  kubectl get secret -n %s %s-client -o go-template='{{ index .data "tls.crt" | base64decode }}' > tls.crt`,
					dot.Release.Namespace,
					Fullname(dot),
				),
				fmt.Sprintf(`  kubectl get secret -n %s %s-client -o go-template='{{ index .data "tls.key" | base64decode }}' > tls.key`,
					dot.Release.Namespace,
					Fullname(dot),
				),
			)
		}
	}
	notes = append(notes,
		fmt.Sprintf(`  rpk profile create --from-profile <(kubectl get configmap -n %s %s-rpk -o go-template='{{ .data.profile }}') %s`,
			dot.Release.Namespace,
			Fullname(dot),
			profileName,
		),
		``,
		`Set up dns to look up the pods on their Kubernetes Nodes. You can use this query to get the list of short-names to IP addresses. Add your external domain to the hostnames and you could test by adding these to your /etc/hosts:`,
		``,
		fmt.Sprintf(`  kubectl get pod -n %s -o custom-columns=node:.status.hostIP,name:.metadata.name --no-headers -l app.kubernetes.io/name=redpanda,app.kubernetes.io/component=redpanda-statefulset`,
			dot.Release.Namespace,
		),
	)
	if anySASL {
		notes = append(notes,
			``,
			`Set the credentials in the environment:`,
			``,
			fmt.Sprintf(`  kubectl -n %s get secret %s -o go-template="{{ range .data }}{{ . | base64decode }}{{ end }}" | IFS=: read -r %s`,
				dot.Release.Namespace,
				values.Auth.SASL.SecretRef,
				RpkSASLEnvironmentVariables(dot),
			),
			fmt.Sprintf(`  export %s`,
				RpkSASLEnvironmentVariables(dot),
			),
		)
	}
	notes = append(notes,
		``,
		`Try some sample commands:`,
	)
	if anySASL {
		notes = append(notes,
			`Create a user:`,
			``,
			fmt.Sprintf(`  %s`, RpkACLUserCreate(dot)),
			``,
			`Give the user permissions:`,
			``,
			fmt.Sprintf(`  %s`, RpkACLCreate(dot)),
		)
	}
	notes = append(notes,
		``,
		`Get the api status:`,
		``,
		fmt.Sprintf(`  %s`, RpkClusterInfo(dot)),
		``,
		`Create a topic`,
		``,
		fmt.Sprintf(`  %s`, RpkTopicCreate(dot)),
		``,
		`Describe the topic:`,
		``,
		fmt.Sprintf(`  %s`, RpkTopicDescribe(dot)),
		``,
		`Delete the topic:`,
		``,
		fmt.Sprintf(`  %s`, RpkTopicDelete(dot)),
	)

	return notes
}

// Any rpk command that's given to the user in in this file must be defined in _example-commands.tpl and tested in a test.
// These are all tested in `tests/test-kafka-sasl-status.yaml`

func RpkACLUserCreate(dot *helmette.Dot) string {
	return fmt.Sprintf(`rpk acl user create myuser --new-password changeme --mechanism %s`, SASLMechanism(dot))
}

func SASLMechanism(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Auth.SASL != nil {
		return values.Auth.SASL.Mechanism
	}
	return "SCRAM-SHA-512"
}

func RpkACLCreate(*helmette.Dot) string {
	return `rpk acl create --allow-principal 'myuser' --allow-host '*' --operation all --topic 'test-topic'`
}

func RpkClusterInfo(*helmette.Dot) string {
	return `rpk cluster info`
}

func RpkTopicCreate(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	return fmt.Sprintf(`rpk topic create test-topic -p 3 -r %d`, helmette.Min(3, int64(values.Statefulset.Replicas)))
}

func RpkTopicDescribe(*helmette.Dot) string {
	return `rpk topic describe test-topic`
}

func RpkTopicDelete(dot *helmette.Dot) string {
	return `rpk topic delete test-topic`
}

// was:   rpk sasl environment variables
//
// This will return a string with the correct environment variables to use for SASL based on the
// version of the redpanda container being used
func RpkSASLEnvironmentVariables(dot *helmette.Dot) string {
	if RedpandaAtLeast_23_2_1(dot) {
		return `RPK_USER RPK_PASS RPK_SASL_MECHANISM`
	} else {
		return `REDPANDA_SASL_USERNAME REDPANDA_SASL_PASSWORD REDPANDA_SASL_MECHANISM`
	}
}
