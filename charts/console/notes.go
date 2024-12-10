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
package console

import (
	"fmt"

	"github.com/redpanda-data/redpanda-operator/pkg/gotohelm/helmette"
)

func Notes(dot *helmette.Dot) []string {
	values := helmette.Unwrap[Values](dot.Values)

	commands := []string{
		`1. Get the application URL by running these commands:`,
	}
	if values.Ingress.Enabled {
		scheme := "http"
		if len(values.Ingress.TLS) > 0 {
			scheme = "https"
		}
		for _, host := range values.Ingress.Hosts {
			for _, path := range host.Paths {
				commands = append(commands, fmt.Sprintf("%s://%s%s", scheme, host.Host, path.Path))
			}
		}
	} else if helmette.Contains("NodePort", string(values.Service.Type)) {
		commands = append(
			commands,
			fmt.Sprintf(`  export NODE_PORT=$(kubectl get --namespace %s -o jsonpath="{.spec.ports[0].nodePort}" services %s)`, dot.Release.Namespace, Fullname(dot)),
			fmt.Sprintf(`  export NODE_IP=$(kubectl get nodes --namespace %s -o jsonpath="{.items[0].status.addresses[0].address}")`, dot.Release.Namespace),
			"  echo http://$NODE_IP:$NODE_PORT",
		)
	} else if helmette.Contains("NodePort", string(values.Service.Type)) {
		commands = append(
			commands,
			`    NOTE: It may take a few minutes for the LoadBalancer IP to be available.`,
			fmt.Sprintf(`          You can watch the status of by running 'kubectl get --namespace %s svc -w %s'`, dot.Release.Namespace, Fullname(dot)),
			fmt.Sprintf(`  export SERVICE_IP=$(kubectl get svc --namespace %s %s --template "{{ range (index .status.loadBalancer.ingress 0) }}{{.}}{{ end }}")`, dot.Release.Namespace, Fullname(dot)),
			fmt.Sprintf(`  echo http://$SERVICE_IP:%d`, values.Service.Port),
		)
	} else if helmette.Contains("ClusterIP", string(values.Service.Type)) {
		commands = append(
			commands,
			fmt.Sprintf(`  export POD_NAME=$(kubectl get pods --namespace %s -l "app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s" -o jsonpath="{.items[0].metadata.name}")`, dot.Release.Namespace, Name(dot), dot.Release.Name),
			fmt.Sprintf(`  export CONTAINER_PORT=$(kubectl get pod --namespace %s $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")`, dot.Release.Namespace),
			`  echo "Visit http://127.0.0.1:8080 to use your application"`,
			fmt.Sprintf(`  kubectl --namespace %s port-forward $POD_NAME 8080:$CONTAINER_PORT`, dot.Release.Namespace),
		)
	}

	return commands
}
