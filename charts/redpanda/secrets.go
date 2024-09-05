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
// +gotohelm:filename=_secrets.go.tpl
package redpanda

import (
	"fmt"
	"strings"

	"github.com/redpanda-data/helm-charts/pkg/gotohelm/helmette"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Secrets(dot *helmette.Dot) []*corev1.Secret {
	var secrets []*corev1.Secret
	secrets = append(secrets, SecretSTSLifecycle(dot))
	if saslUsers := SecretSASLUsers(dot); saslUsers != nil {
		secrets = append(secrets, saslUsers)
	}
	if configWatcher := SecretConfigWatcher(dot); configWatcher != nil {
		secrets = append(secrets, configWatcher)
	}
	secrets = append(secrets, SecretConfigurator(dot))
	if fsValidator := SecretFSValidator(dot); fsValidator != nil {
		secrets = append(secrets, fsValidator)
	}
	if bootstrapUser := SecretBootstrapUser(dot); bootstrapUser != nil {
		secrets = append(secrets, bootstrapUser)
	}
	return secrets
}

func SecretSTSLifecycle(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sts-lifecycle", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}
	adminCurlFlags := adminTLSCurlFlags(dot)
	secret.StringData["common.sh"] = helmette.Join("\n", []string{
		`#!/usr/bin/env bash`,
		``,
		`# the SERVICE_NAME comes from the metadata.name of the pod, essentially the POD_NAME`,
		fmt.Sprintf(`CURL_URL="%s"`, adminInternalURL(dot)),
		``,
		`# commands used throughout`,
		fmt.Sprintf(`CURL_NODE_ID_CMD="curl --silent --fail %s ${CURL_URL}/v1/node_config"`, adminCurlFlags),
		``,
		`CURL_MAINTENANCE_DELETE_CMD_PREFIX='curl -X DELETE --silent -o /dev/null -w "%{http_code}"'`,
		`CURL_MAINTENANCE_PUT_CMD_PREFIX='curl -X PUT --silent -o /dev/null -w "%{http_code}"'`,
		fmt.Sprintf(`CURL_MAINTENANCE_GET_CMD="curl -X GET --silent %s ${CURL_URL}/v1/maintenance"`, adminCurlFlags),
	})

	postStartSh := []string{
		`#!/usr/bin/env bash`,
		`# This code should be similar if not exactly the same as that found in the panda-operator, see`,
		`# https://github.com/redpanda-data/redpanda/blob/e51d5b7f2ef76d5160ca01b8c7a8cf07593d29b6/src/go/k8s/pkg/resources/secret.go`,
		``,
		`# path below should match the path defined on the statefulset`,
		`source /var/lifecycle/common.sh`,
		``,
		`postStartHook () {`,
		`  set -x`,
		``,
		`  touch /tmp/postStartHookStarted`,
		``,
		`  until NODE_ID=$(${CURL_NODE_ID_CMD} | grep -o '\"node_id\":[^,}]*' | grep -o '[^: ]*$'); do`,
		`      sleep 0.5`,
		`  done`,
		``,
		`  echo "Clearing maintenance mode on node ${NODE_ID}"`,
		fmt.Sprintf(`  CURL_MAINTENANCE_DELETE_CMD="${CURL_MAINTENANCE_DELETE_CMD_PREFIX} %s ${CURL_URL}/v1/brokers/${NODE_ID}/maintenance"`, adminCurlFlags),
		`  # a 400 here would mean not in maintenance mode`,
		`  until [ "${status:-}" = '"200"' ] || [ "${status:-}" = '"400"' ]; do`,
		`      status=$(${CURL_MAINTENANCE_DELETE_CMD})`,
		`      sleep 0.5`,
		`  done`,
	}
	if values.Auth.SASL.Enabled && values.Auth.SASL.SecretRef != "" {
		postStartSh = append(postStartSh,
			`  # Setup and export SASL bootstrap-user`,
			`  IFS=":" read -r USER_NAME PASSWORD MECHANISM < <(grep "" $(find /etc/secrets/users/* -print))`,
			fmt.Sprintf(`  MECHANISM=${MECHANISM:-%s}`, helmette.Dig(dot.Values.AsMap(), "SCRAM-SHA-512", "auth", "sasl", "mechanism")),
			`  rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} || true`,
		)
	}
	postStartSh = append(postStartSh,
		``,

		`  touch /tmp/postStartHookFinished`,
		`}`,
		``,
		`postStartHook`,
		`true`,
	)
	secret.StringData["postStart.sh"] = helmette.Join("\n", postStartSh)

	preStopSh := []string{
		`#!/usr/bin/env bash`,
		`# This code should be similar if not exactly the same as that found in the panda-operator, see`,
		`# https://github.com/redpanda-data/redpanda/blob/e51d5b7f2ef76d5160ca01b8c7a8cf07593d29b6/src/go/k8s/pkg/resources/secret.go`,
		``,
		`touch /tmp/preStopHookStarted`,
		``,
		`# path below should match the path defined on the statefulset`,
		`source /var/lifecycle/common.sh`,
		``,
		`set -x`,
		``,
		`preStopHook () {`,
		`  until NODE_ID=$(${CURL_NODE_ID_CMD} | grep -o '\"node_id\":[^,}]*' | grep -o '[^: ]*$'); do`,
		`      sleep 0.5`,
		`  done`,
		``,
		`  echo "Setting maintenance mode on node ${NODE_ID}"`,
		fmt.Sprintf(`  CURL_MAINTENANCE_PUT_CMD="${CURL_MAINTENANCE_PUT_CMD_PREFIX} %s ${CURL_URL}/v1/brokers/${NODE_ID}/maintenance"`, adminCurlFlags),
		`  until [ "${status:-}" = '"200"' ]; do`,
		`      status=$(${CURL_MAINTENANCE_PUT_CMD})`,
		`      sleep 0.5`,
		`  done`,
		``,
		`  until [ "${finished:-}" = "true" ] || [ "${draining:-}" = "false" ]; do`,
		`      res=$(${CURL_MAINTENANCE_GET_CMD})`,
		`      finished=$(echo $res | grep -o '\"finished\":[^,}]*' | grep -o '[^: ]*$')`,
		`      draining=$(echo $res | grep -o '\"draining\":[^,}]*' | grep -o '[^: ]*$')`,
		`      sleep 0.5`,
		`  done`,
		``,
		`  touch /tmp/preStopHookFinished`,
		`}`,
	}
	if values.Statefulset.Replicas > 2 && !helmette.Dig(values.Config.Node, false, "recovery_mode_enabled").(bool) {
		preStopSh = append(preStopSh,
			`preStopHook`,
		)
	} else {
		preStopSh = append(preStopSh,
			`touch /tmp/preStopHookFinished`,
			`echo "Not enough replicas or in recovery mode, cannot put a broker into maintenance mode."`,
		)
	}
	preStopSh = append(preStopSh,
		`true`,
	)
	secret.StringData["preStop.sh"] = helmette.Join("\n", preStopSh)
	return secret
}

func SecretSASLUsers(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Auth.SASL.SecretRef != "" && values.Auth.SASL.Enabled && len(values.Auth.SASL.Users) > 0 {
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      values.Auth.SASL.SecretRef,
				Namespace: dot.Release.Namespace,
				Labels:    FullLabels(dot),
			},
			Type:       corev1.SecretTypeOpaque,
			StringData: map[string]string{},
		}
		usersTxt := []string{}
		// Working around lack of support for += or strings.Join at the moment
		for _, user := range values.Auth.SASL.Users {
			if helmette.Empty(user.Mechanism) {
				usersTxt = append(usersTxt, fmt.Sprintf("%s:%s", user.Name, user.Password))
			} else {
				usersTxt = append(usersTxt, fmt.Sprintf("%s:%s:%s", user.Name, user.Password, user.Mechanism))
			}
		}
		secret.StringData["users.txt"] = helmette.Join("\n", usersTxt)
		return secret
	} else if values.Auth.SASL.Enabled && values.Auth.SASL.SecretRef == "" {
		panic("auth.sasl.secretRef cannot be empty when auth.sasl.enabled=true")
	} else {
		// XXX no secret generated when enabled, we have a secret ref, but we have no users
		return nil
	}
}

func SecretBootstrapUser(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)
	if !values.Auth.SASL.Enabled || values.Auth.SASL.BootstrapUser.SecretKeyRef != nil {
		return nil
	}

	password := helmette.RandAlphaNum(32)

	userPassword := values.Auth.SASL.BootstrapUser.Password
	if userPassword != nil {
		password = *userPassword
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-bootstrap-user", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"password": password,
		},
	}
}

func SecretConfigWatcher(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Statefulset.SideCars.ConfigWatcher.Enabled {
		return nil
	}

	sasl := values.Auth.SASL
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config-watcher", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}

	var saslUserSh []string
	saslUserSh = append(saslUserSh,
		`#!/usr/bin/env bash`,
		``,
		`trap 'error_handler $? $LINENO' ERR`,
		``,
		`error_handler() {`,
		`  echo "Error: ($1) occurred at line $2"`,
		`}`,
		``,
		`set -e`,
		``,
		`# rpk cluster health can exit non-zero if it's unable to dial brokers. This`,
		`# can happen for many reasons but we never want this script to crash as it`,
		`# would take down yet another broker and make a bad situation worse.`,
		`# Instead, just wait for the command to eventually exit zero.`,
		`echo "Waiting for cluster to be ready"`,
		`until rpk cluster health --watch --exit-when-healthy; do`,
		`  echo "rpk cluster health failed. Waiting 5 seconds before trying again..."`,
		`  sleep 5`,
		`done`,
	)
	if sasl.Enabled && sasl.SecretRef != "" {
		saslUserSh = append(saslUserSh,
			`while true; do`,
			`  echo "RUNNING: Monitoring and Updating SASL users"`,
			`  USERS_DIR="/etc/secrets/users"`,
			``,
			`  new_users_list(){`,
			`    LIST=$1`,
			`    NEW_USER=$2`,
			`    if [[ -n "${LIST}" ]]; then`,
			`      LIST="${NEW_USER},${LIST}"`,
			`    else`,
			`      LIST="${NEW_USER}"`,
			`    fi`,
			``,
			`    echo "${LIST}"`,
			`  }`,
			``,
			`  process_users() {`,
			`    USERS_DIR=${1-"/etc/secrets/users"}`,
			`    USERS_FILE=$(find ${USERS_DIR}/* -print)`,
			`    USERS_LIST=""`,
			`    READ_LIST_SUCCESS=0`,
			`    # Read line by line, handle a missing EOL at the end of file`,
			`    while read p || [ -n "$p" ] ; do`,
			`      IFS=":" read -r USER_NAME PASSWORD MECHANISM <<< $p`,
			`      # Do not process empty lines`,
			`      if [ -z "$USER_NAME" ]; then`,
			`        continue`,
			`      fi`,
			`      if [[ "${USER_NAME// /}" != "$USER_NAME" ]]; then`,
			`        continue`,
			`      fi`,
			`      echo "Creating user ${USER_NAME}..."`,
			fmt.Sprintf(`      MECHANISM=${MECHANISM:-%s}`, helmette.Dig(dot.Values.AsMap(), "SCRAM-SHA-512", "auth", "sasl", "mechanism")),
			`      creation_result=$(rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} 2>&1) && creation_result_exit_code=$? || creation_result_exit_code=$?  # On a non-success exit code`,
			`      if [[ $creation_result_exit_code -ne 0 ]]; then`,
			`        # Check if the stderr contains "User already exists"`,
			`        # this error occurs when password has changed`,
			`        if [[ $creation_result == *"User already exists"* ]]; then`,
			`          echo "Update user ${USER_NAME}"`,
			`          # we will try to update by first deleting`,
			`          deletion_result=$(rpk acl user delete ${USER_NAME} 2>&1) && deletion_result_exit_code=$? || deletion_result_exit_code=$?`,
			`          if [[ $deletion_result_exit_code -ne 0 ]]; then`,
			`            echo "deletion of user ${USER_NAME} failed: ${deletion_result}"`,
			`            READ_LIST_SUCCESS=1`,
			`            break`,
			`          fi`,
			`          # Now we update the user`,
			`          update_result=$(rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} 2>&1) && update_result_exit_code=$? || update_result_exit_code=$?  # On a non-success exit code`,
			`          if [[ $update_result_exit_code -ne 0 ]]; then`,
			`            echo "updating user ${USER_NAME} failed: ${update_result}"`,
			`            READ_LIST_SUCCESS=1`,
			`            break`,
			`          else`,
			`            echo "Updated user ${USER_NAME}..."`,
			`            USERS_LIST=$(new_users_list "${USERS_LIST}" "${USER_NAME}")`,
			`          fi`,
			`        else`,
			`          # Another error occurred, so output the original message and exit code`,
			`          echo "error creating user ${USER_NAME}: ${creation_result}"`,
			`          READ_LIST_SUCCESS=1`,
			`          break`,
			`        fi`,
			`      # On a success, the user was created so output that`,
			`      else`,
			`        echo "Created user ${USER_NAME}..."`,
			`        USERS_LIST=$(new_users_list "${USERS_LIST}" "${USER_NAME}")`,
			`      fi`,
			`    done < $USERS_FILE`,
			``,
			`    if [[ -n "${USERS_LIST}" && ${READ_LIST_SUCCESS} ]]; then`,
			`      echo "Setting superusers configurations with users [${USERS_LIST}]"`,
			`      superuser_result=$(rpk cluster config set superusers [${USERS_LIST}] 2>&1) && superuser_result_exit_code=$? || superuser_result_exit_code=$?`,
			`      if [[ $superuser_result_exit_code -ne 0 ]]; then`,
			`          echo "Setting superusers configurations failed: ${superuser_result}"`,
			`      else`,
			`          echo "Completed setting superusers configurations"`,
			`      fi`,
			`    fi`,
			`  }`,
			``,
			`  # before we do anything ensure we have the bootstrap user`,
			`  echo "Ensuring bootstrap user ${RPK_USER}..."`,
			`  creation_result=$(rpk acl user create ${USER_NAME} --password=${RPK_PASS} --mechanism ${RPK_SASL_MECHANISM} 2>&1) && creation_result_exit_code=$? || creation_result_exit_code=$?  # On a non-success exit code`,
			`  if [[ $creation_result_exit_code -ne 0 ]]; then`,
			`    if [[ $creation_result == *"User already exists"* ]]; then`,
			`      echo "Bootstrap user already created"`,
			`    else`,
			`      echo "error creating user ${RPK_USER}: ${creation_result}"`,
			`    fi`,
			`  fi`,
			``,
			`  # first time processing`,
			`  process_users $USERS_DIR`,
			``,
			`  # subsequent changes detected here`,
			`  # watching delete_self as documented in https://ahmet.im/blog/kubernetes-inotify/`,
			`  USERS_FILE=$(find ${USERS_DIR}/* -print)`,
			`  while RES=$(inotifywait -q -e delete_self ${USERS_FILE}); do`,
			`    process_users $USERS_DIR`,
			`  done`,
			`done`,
		)
	} else {
		saslUserSh = append(saslUserSh,
			`echo "Nothing to do. Sleeping..."`,
			`sleep infinity`,
		)
	}

	secret.StringData["sasl-user.sh"] = helmette.Join("\n", saslUserSh)
	return secret
}

func SecretFSValidator(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Statefulset.InitContainers.FSValidator.Enabled {
		return nil
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-fs-validator", Fullname(dot)[:49]),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}

	secret.StringData["fsValidator.sh"] = `set -e
EXPECTED_FS_TYPE=$1

DATA_DIR="/var/lib/redpanda/data"
TEST_FILE="testfile"

echo "checking data directory exist..."
if [ ! -d "${DATA_DIR}" ]; then
  echo "data directory does not exists, exiting"
  exit 1
fi

echo "checking filesystem type..."
FS_TYPE=$(df -T $DATA_DIR  | tail -n +2 | awk '{print $2}')

if [ "${FS_TYPE}" != "${EXPECTED_FS_TYPE}" ]; then
  echo "file system found to be ${FS_TYPE} when expected ${EXPECTED_FS_TYPE}"
  exit 1
fi

echo "checking if able to create a test file..."

touch ${DATA_DIR}/${TEST_FILE}
result=$(touch ${DATA_DIR}/${TEST_FILE} 2> /dev/null; echo $?)
if [ "${result}" != "0" ]; then
  echo "could not write testfile, may not have write permission"
  exit 1
fi

echo "checking if able to delete a test file..."

result=$(rm ${DATA_DIR}/${TEST_FILE} 2> /dev/null; echo $?)
if [ "${result}" != "0" ]; then
  echo "could not delete testfile"
  exit 1
fi

echo "passed"`
	return secret
}

func SecretConfigurator(dot *helmette.Dot) *corev1.Secret {
	values := helmette.Unwrap[Values](dot.Values)

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%.51s-configurator", Fullname(dot)),
			Namespace: dot.Release.Namespace,
			Labels:    FullLabels(dot),
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}
	configuratorSh := []string{}
	configuratorSh = append(configuratorSh,
		`set -xe`,
		`SERVICE_NAME=$1`,
		`KUBERNETES_NODE_NAME=$2`,
		`POD_ORDINAL=${SERVICE_NAME##*-}`,
		"BROKER_INDEX=`expr $POD_ORDINAL + 1`", // POSIX sh-safe
		``,
		`CONFIG=/etc/redpanda/redpanda.yaml`,
		``,
		`# Setup config files`,
		`cp /tmp/base-config/redpanda.yaml "${CONFIG}"`,
		`cp /tmp/base-config/bootstrap.yaml /etc/redpanda/.bootstrap.yaml`,
	)
	if !RedpandaAtLeast_22_3_0(dot) {
		configuratorSh = append(configuratorSh,
			``,
			`# Configure bootstrap`,
			`## Not used for Redpanda v22.3.0+`,
			`rpk --config "${CONFIG}" redpanda config set redpanda.node_id "${POD_ORDINAL}"`,
			`if [ "${POD_ORDINAL}" = "0" ]; then`,
			`	rpk --config "${CONFIG}" redpanda config set redpanda.seed_servers '[]' --format yaml`,
			`fi`,
		)
	}

	kafkaSnippet := secretConfiguratorKafkaConfig(dot)
	configuratorSh = append(configuratorSh, kafkaSnippet...)

	httpSnippet := secretConfiguratorHTTPConfig(dot)
	configuratorSh = append(configuratorSh, httpSnippet...)

	if RedpandaAtLeast_22_3_0(dot) && values.RackAwareness.Enabled {
		configuratorSh = append(configuratorSh,
			``,
			`# Configure Rack Awareness`,
			`set +x`,
			fmt.Sprintf(`RACK=$(curl --silent --cacert /run/secrets/kubernetes.io/serviceaccount/ca.crt --fail -H 'Authorization: Bearer '$(cat /run/secrets/kubernetes.io/serviceaccount/token) "https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT_HTTPS}/api/v1/nodes/${KUBERNETES_NODE_NAME}?pretty=true" | grep %s | grep -v '\"key\":' | sed 's/.*": "\([^"]\+\).*/\1/')`,
				helmette.SQuote(helmette.Quote(values.RackAwareness.NodeAnnotation)),
			),
			`set -x`,
			`rpk --config "$CONFIG" redpanda config set redpanda.rack "${RACK}"`,
		)
	}
	secret.StringData["configurator.sh"] = helmette.Join("\n", configuratorSh)
	return secret
}

func secretConfiguratorKafkaConfig(dot *helmette.Dot) []string {
	values := helmette.Unwrap[Values](dot.Values)

	internalAdvertiseAddress := fmt.Sprintf("%s.%s", "${SERVICE_NAME}", InternalDomain(dot))

	var snippet []string

	// Handle kafka listener
	listenerName := "kafka"
	listenerAdvertisedName := listenerName
	redpandaConfigPart := "redpanda"
	snippet = append(snippet,
		``,
		fmt.Sprintf(`LISTENER=%s`, helmette.Quote(helmette.ToJSON(map[string]any{
			"name":    "internal",
			"address": internalAdvertiseAddress,
			"port":    values.Listeners.Kafka.Port,
		}))),
		fmt.Sprintf(`rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[0] "$LISTENER"`,
			redpandaConfigPart,
			listenerAdvertisedName,
		),
	)
	if len(values.Listeners.Kafka.External) > 0 {
		externalCounter := 0
		for externalName, externalVals := range values.Listeners.Kafka.External {
			externalCounter = externalCounter + 1
			snippet = append(snippet,
				``,
				fmt.Sprintf(`ADVERTISED_%s_ADDRESSES=()`, helmette.Upper(listenerName)),
			)
			for _, replicaIndex := range helmette.Until(int(values.Statefulset.Replicas)) {
				// advertised-port for kafka
				port := externalVals.Port // This is always defined for kafka
				if len(externalVals.AdvertisedPorts) > 0 {
					if len(externalVals.AdvertisedPorts) == 1 {
						port = externalVals.AdvertisedPorts[0]
					} else {
						port = externalVals.AdvertisedPorts[replicaIndex]
					}
				}

				host := advertisedHostJSON(dot, externalName, port, replicaIndex)
				// XXX: the original code used the stringified `host` value as a template
				// for re-expansion; however it was impossible to make this work usefully,
				/// even with the original yaml template.
				address := helmette.ToJSON(host)
				prefixTemplate := ptr.Deref(externalVals.PrefixTemplate, "")
				if prefixTemplate == "" {
					// Required because the values might not specify this, it'll ensur we see "" if it's missing.
					prefixTemplate = helmette.Default("", values.External.PrefixTemplate)
				}
				snippet = append(snippet,
					``,
					fmt.Sprintf(`PREFIX_TEMPLATE=%s`, helmette.Quote(prefixTemplate)),
					fmt.Sprintf(`ADVERTISED_%s_ADDRESSES+=(%s)`,
						helmette.Upper(listenerName),
						helmette.Quote(address),
					),
				)
			}

			snippet = append(snippet,
				``,
				fmt.Sprintf(`rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[%d] "${ADVERTISED_%s_ADDRESSES[$POD_ORDINAL]}"`,
					redpandaConfigPart,
					listenerAdvertisedName,
					externalCounter,
					helmette.Upper(listenerName),
				),
			)
		}
	}

	return snippet
}

func secretConfiguratorHTTPConfig(dot *helmette.Dot) []string {
	values := helmette.Unwrap[Values](dot.Values)

	internalAdvertiseAddress := fmt.Sprintf("%s.%s", "${SERVICE_NAME}", InternalDomain(dot))

	var snippet []string

	// Handle kafka listener
	listenerName := "http"
	listenerAdvertisedName := "pandaproxy"
	redpandaConfigPart := "pandaproxy"
	snippet = append(snippet,
		``,
		fmt.Sprintf(`LISTENER=%s`, helmette.Quote(helmette.ToJSON(map[string]any{
			"name":    "internal",
			"address": internalAdvertiseAddress,
			"port":    values.Listeners.HTTP.Port,
		}))),
		fmt.Sprintf(`rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[0] "$LISTENER"`,
			redpandaConfigPart,
			listenerAdvertisedName,
		),
	)
	if len(values.Listeners.HTTP.External) > 0 {
		externalCounter := 0
		for externalName, externalVals := range values.Listeners.HTTP.External {
			externalCounter = externalCounter + 1
			snippet = append(snippet,
				``,
				fmt.Sprintf(`ADVERTISED_%s_ADDRESSES=()`, helmette.Upper(listenerName)),
			)
			for _, replicaIndex := range helmette.Until(int(values.Statefulset.Replicas)) {
				// advertised-port for kafka
				port := externalVals.Port // This is always defined for kafka
				if len(externalVals.AdvertisedPorts) > 0 {
					if len(externalVals.AdvertisedPorts) == 1 {
						port = externalVals.AdvertisedPorts[0]
					} else {
						port = externalVals.AdvertisedPorts[replicaIndex]
					}
				}

				host := advertisedHostJSON(dot, externalName, port, replicaIndex)
				// XXX: the original code used the stringified `host` value as a template
				// for re-expansion; however it was impossible to make this work usefully,
				/// even with the original yaml template.
				address := helmette.ToJSON(host)

				prefixTemplate := ptr.Deref(externalVals.PrefixTemplate, "")
				if prefixTemplate == "" {
					// Required because the values might not specify this, it'll ensur we see "" if it's missing.
					prefixTemplate = helmette.Default("", values.External.PrefixTemplate)
				}
				snippet = append(snippet,
					``,
					fmt.Sprintf(`PREFIX_TEMPLATE=%s`, helmette.Quote(prefixTemplate)),
					fmt.Sprintf(`ADVERTISED_%s_ADDRESSES+=(%s)`,
						helmette.Upper(listenerName),
						helmette.Quote(address),
					),
				)
			}

			snippet = append(snippet,
				``,
				fmt.Sprintf(`rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[%d] "${ADVERTISED_%s_ADDRESSES[$POD_ORDINAL]}"`,
					redpandaConfigPart,
					listenerAdvertisedName,
					externalCounter,
					helmette.Upper(listenerName),
				),
			)
		}
	}

	return snippet
}

// The following from _helpers.tpm

func adminTLSCurlFlags(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if !values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		return ""
	}
	path := fmt.Sprintf("/etc/tls/certs/%s", values.Listeners.Admin.TLS.Cert)
	if values.Listeners.Admin.TLS.RequireClientAuth {
		return fmt.Sprintf("--cacert %s/ca.crt --cert %s/tls.crt --key %s/tls.key", path, path, path)
	}
	// XXX fix up a bug in the template
	return fmt.Sprintf("--cacert %s/ca.crt", path)
}

func externalAdvertiseAddress(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	eaa := "${SERVICE_NAME}"
	externalDomainTemplate := ptr.Deref(values.External.Domain, "")
	expanded := helmette.Tpl(externalDomainTemplate, dot)
	if !helmette.Empty(expanded) {
		eaa = fmt.Sprintf("%s.%s", "${SERVICE_NAME}", expanded)
	}
	return eaa
}

// was advertised-host
func advertisedHostJSON(dot *helmette.Dot, externalName string, port int32, replicaIndex int) map[string]any {
	values := helmette.Unwrap[Values](dot.Values)

	host := map[string]any{
		"name":    externalName,
		"address": externalAdvertiseAddress(dot),
		"port":    port,
	}
	if len(values.External.Addresses) > 0 {
		address := ""
		if len(values.External.Addresses) > 1 {
			address = values.External.Addresses[replicaIndex]
		} else {
			address = values.External.Addresses[0]
		}
		if domain := ptr.Deref(values.External.Domain, ""); domain != "" {
			host = map[string]any{
				"name":    externalName,
				"address": fmt.Sprintf("%s.%s", address, domain),
				"port":    port,
			}
		} else {
			host = map[string]any{
				"name":    externalName,
				"address": address,
				"port":    port,
			}
		}
	}
	return host
}

// adminInternalHTTPProtocol was admin-http-protocol
func adminInternalHTTPProtocol(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	if values.Listeners.Admin.TLS.IsEnabled(&values.TLS) {
		return "https"
	}
	return "http"
}

// Additional helpers

func adminInternalURL(dot *helmette.Dot) string {
	values := helmette.Unwrap[Values](dot.Values)

	return fmt.Sprintf("%s://%s.%s.%s.svc.%s:%d",
		adminInternalHTTPProtocol(dot),
		`${SERVICE_NAME}`,
		ServiceName(dot),
		dot.Release.Namespace,
		strings.TrimSuffix(values.ClusterDomain, "."),
		values.Listeners.Admin.Port,
	)
}
