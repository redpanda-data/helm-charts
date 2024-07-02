{{- /* Generated from "secrets.go" */ -}}

{{- define "redpanda.Secrets" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $secrets := (coalesce nil) -}}
{{- $secrets = (concat (default (list ) $secrets) (list (get (fromJson (include "redpanda.SecretSTSLifecycle" (dict "a" (list $dot) ))) "r"))) -}}
{{- $saslUsers_1 := (get (fromJson (include "redpanda.SecretSASLUsers" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $saslUsers_1 (coalesce nil)) -}}
{{- $secrets = (concat (default (list ) $secrets) (list $saslUsers_1)) -}}
{{- end -}}
{{- $configWatcher_2 := (get (fromJson (include "redpanda.SecretConfigWatcher" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $configWatcher_2 (coalesce nil)) -}}
{{- $secrets = (concat (default (list ) $secrets) (list $configWatcher_2)) -}}
{{- end -}}
{{- $secrets = (concat (default (list ) $secrets) (list (get (fromJson (include "redpanda.SecretConfigurator" (dict "a" (list $dot) ))) "r"))) -}}
{{- $fsValidator_3 := (get (fromJson (include "redpanda.SecretFSValidator" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $fsValidator_3 (coalesce nil)) -}}
{{- $secrets = (concat (default (list ) $secrets) (list $fsValidator_3)) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secrets) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretSTSLifecycle" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $secret := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-sts-lifecycle" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "type" "Opaque" "stringData" (dict ) )) -}}
{{- $adminCurlFlags := (get (fromJson (include "redpanda.adminTLSCurlFlags" (dict "a" (list $dot) ))) "r") -}}
{{- $_ := (set $secret.stringData "common.sh" (join "\n" (list `#!/usr/bin/env bash` `` `# the SERVICE_NAME comes from the metadata.name of the pod, essentially the POD_NAME` (printf `CURL_URL="%s"` (get (fromJson (include "redpanda.adminInternalURL" (dict "a" (list $dot) ))) "r")) `` `# commands used throughout` (printf `CURL_NODE_ID_CMD="curl --silent --fail %s ${CURL_URL}/v1/node_config"` $adminCurlFlags) `` `CURL_MAINTENANCE_DELETE_CMD_PREFIX='curl -X DELETE --silent -o /dev/null -w "%{http_code}"'` `CURL_MAINTENANCE_PUT_CMD_PREFIX='curl -X PUT --silent -o /dev/null -w "%{http_code}"'` (printf `CURL_MAINTENANCE_GET_CMD="curl -X GET --silent %s ${CURL_URL}/v1/maintenance"` $adminCurlFlags)))) -}}
{{- $postStartSh := (list `#!/usr/bin/env bash` `# This code should be similar if not exactly the same as that found in the panda-operator, see` `# https://github.com/redpanda-data/redpanda/blob/e51d5b7f2ef76d5160ca01b8c7a8cf07593d29b6/src/go/k8s/pkg/resources/secret.go` `` `# path below should match the path defined on the statefulset` `source /var/lifecycle/common.sh` `` `postStartHook () {` `  set -x` `` `  touch /tmp/postStartHookStarted` `` `  until NODE_ID=$(${CURL_NODE_ID_CMD} | grep -o '\"node_id\":[^,}]*' | grep -o '[^: ]*$'); do` `      sleep 0.5` `  done` `` `  echo "Clearing maintenance mode on node ${NODE_ID}"` (printf `  CURL_MAINTENANCE_DELETE_CMD="${CURL_MAINTENANCE_DELETE_CMD_PREFIX} %s ${CURL_URL}/v1/brokers/${NODE_ID}/maintenance"` $adminCurlFlags) `  # a 400 here would mean not in maintenance mode` `  until [ "${status:-}" = '"200"' ] || [ "${status:-}" = '"400"' ]; do` `      status=$(${CURL_MAINTENANCE_DELETE_CMD})` `      sleep 0.5` `  done`) -}}
{{- if (and $values.auth.sasl.enabled (ne $values.auth.sasl.secretRef "")) -}}
{{- $postStartSh = (concat (default (list ) $postStartSh) (list `  # Setup and export SASL bootstrap-user` `  IFS=":" read -r USER_NAME PASSWORD MECHANISM < <(grep "" $(find /etc/secrets/users/* -print))` (printf `  MECHANISM=${MECHANISM:-%s}` (dig "auth" "sasl" "mechanism" "SCRAM-SHA-512" $dot.Values.AsMap)) `  rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} || true`)) -}}
{{- end -}}
{{- $postStartSh = (concat (default (list ) $postStartSh) (list `` `  touch /tmp/postStartHookFinished` `}` `` `postStartHook` `true`)) -}}
{{- $_ := (set $secret.stringData "postStart.sh" (join "\n" $postStartSh)) -}}
{{- $preStopSh := (list `#!/usr/bin/env bash` `# This code should be similar if not exactly the same as that found in the panda-operator, see` `# https://github.com/redpanda-data/redpanda/blob/e51d5b7f2ef76d5160ca01b8c7a8cf07593d29b6/src/go/k8s/pkg/resources/secret.go` `` `touch /tmp/preStopHookStarted` `` `# path below should match the path defined on the statefulset` `source /var/lifecycle/common.sh` `` `set -x` `` `preStopHook () {` `  until NODE_ID=$(${CURL_NODE_ID_CMD} | grep -o '\"node_id\":[^,}]*' | grep -o '[^: ]*$'); do` `      sleep 0.5` `  done` `` `  echo "Setting maintenance mode on node ${NODE_ID}"` (printf `  CURL_MAINTENANCE_PUT_CMD="${CURL_MAINTENANCE_PUT_CMD_PREFIX} %s ${CURL_URL}/v1/brokers/${NODE_ID}/maintenance"` $adminCurlFlags) `  until [ "${status:-}" = '"200"' ]; do` `      status=$(${CURL_MAINTENANCE_PUT_CMD})` `      sleep 0.5` `  done` `` `  until [ "${finished:-}" = "true" ] || [ "${draining:-}" = "false" ]; do` `      res=$(${CURL_MAINTENANCE_GET_CMD})` `      finished=$(echo $res | grep -o '\"finished\":[^,}]*' | grep -o '[^: ]*$')` `      draining=$(echo $res | grep -o '\"draining\":[^,}]*' | grep -o '[^: ]*$')` `      sleep 0.5` `  done` `` `  touch /tmp/preStopHookFinished` `}`) -}}
{{- if (and (gt ($values.statefulset.replicas | int) (2 | int)) (not (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" (dig "recovery_mode_enabled" false $values.config.node)) ))) "r"))) -}}
{{- $preStopSh = (concat (default (list ) $preStopSh) (list `preStopHook`)) -}}
{{- else -}}
{{- $preStopSh = (concat (default (list ) $preStopSh) (list `touch /tmp/preStopHookFinished` `echo "Not enough replicas or in recovery mode, cannot put a broker into maintenance mode."`)) -}}
{{- end -}}
{{- $preStopSh = (concat (default (list ) $preStopSh) (list `true`)) -}}
{{- $_ := (set $secret.stringData "preStop.sh" (join "\n" $preStopSh)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secret) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretSASLUsers" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (and (and (ne $values.auth.sasl.secretRef "") $values.auth.sasl.enabled) (gt ((get (fromJson (include "_shims.len" (dict "a" (list $values.auth.sasl.users) ))) "r") | int) (0 | int))) -}}
{{- $secret := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $values.auth.sasl.secretRef "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "type" "Opaque" "stringData" (dict ) )) -}}
{{- $usersTxt := (list ) -}}
{{- range $_, $user := $values.auth.sasl.users -}}
{{- if (ne $user.mechanism "") -}}
{{- $usersTxt = (concat (default (list ) $usersTxt) (list (printf "%s:%s:%s" $user.name $user.password $user.mechanism))) -}}
{{- else -}}
{{- $usersTxt = (concat (default (list ) $usersTxt) (list (printf "%s:%s" $user.name $user.password))) -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $secret.stringData "users.txt" (join "\n" $usersTxt)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secret) | toJson -}}
{{- break -}}
{{- else -}}{{- if (and $values.auth.sasl.enabled (eq $values.auth.sasl.secretRef "")) -}}
{{- $_ := (fail "auth.sasl.secretRef cannot be empty when auth.sasl.enabled=true") -}}
{{- else -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretConfigWatcher" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.statefulset.sideCars.configWatcher.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $sasl := $values.auth.sasl -}}
{{- $secret := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-config-watcher" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "type" "Opaque" "stringData" (dict ) )) -}}
{{- $saslUserSh := (coalesce nil) -}}
{{- $saslUserSh = (concat (default (list ) $saslUserSh) (list `#!/usr/bin/env bash` `` `trap 'error_handler $? $LINENO' ERR` `` `error_handler() {` `  echo "Error: ($1) occurred at line $2"` `}` `` `set -e` `` `# rpk cluster health can exit non-zero if it's unable to dial brokers. This` `# can happen for many reasons but we never want this script to crash as it` `# would take down yet another broker and make a bad situation worse.` `# Instead, just wait for the command to eventually exit zero.` `echo "Waiting for cluster to be ready"` `until rpk cluster health --watch --exit-when-healthy; do` `  echo "rpk cluster health failed. Waiting 5 seconds before trying again..."` `  sleep 5` `done`)) -}}
{{- if (and $sasl.enabled (ne $sasl.secretRef "")) -}}
{{- $saslUserSh = (concat (default (list ) $saslUserSh) (list `while true; do` `  echo "RUNNING: Monitoring and Updating SASL users"` `  USERS_DIR="/etc/secrets/users"` `` `  new_users_list(){` `    LIST=$1` `    NEW_USER=$2` `    if [[ -n "${LIST}" ]]; then` `      LIST="${NEW_USER},${LIST}"` `    else` `      LIST="${NEW_USER}"` `    fi` `` `    echo "${LIST}"` `  }` `` `  process_users() {` `    USERS_DIR=${1-"/etc/secrets/users"}` `    USERS_FILE=$(find ${USERS_DIR}/* -print)` `    USERS_LIST=""` `    READ_LIST_SUCCESS=0` `    # Read line by line, handle a missing EOL at the end of file` `    while read p || [ -n "$p" ] ; do` `      IFS=":" read -r USER_NAME PASSWORD MECHANISM <<< $p` `      # Do not process empty lines` `      if [ -z "$USER_NAME" ]; then` `        continue` `      fi` `      if [[ "${USER_NAME// /}" != "$USER_NAME" ]]; then` `        continue` `      fi` `      echo "Creating user ${USER_NAME}..."` (printf `      MECHANISM=${MECHANISM:-%s}` (dig "auth" "sasl" "mechanism" "SCRAM-SHA-512" $dot.Values.AsMap)) `      creation_result=$(rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} 2>&1) && creation_result_exit_code=$? || creation_result_exit_code=$?  # On a non-success exit code` `      if [[ $creation_result_exit_code -ne 0 ]]; then` `        # Check if the stderr contains "User already exists"` `        # this error occurs when password has changed` `        if [[ $creation_result == *"User already exists"* ]]; then` `          echo "Update user ${USER_NAME}"` `          # we will try to update by first deleting` `          deletion_result=$(rpk acl user delete ${USER_NAME} 2>&1) && deletion_result_exit_code=$? || deletion_result_exit_code=$?` `          if [[ $deletion_result_exit_code -ne 0 ]]; then` `            echo "deletion of user ${USER_NAME} failed: ${deletion_result}"` `            READ_LIST_SUCCESS=1` `            break` `          fi` `          # Now we update the user` `          update_result=$(rpk acl user create ${USER_NAME} --password=${PASSWORD} --mechanism ${MECHANISM} 2>&1) && update_result_exit_code=$? || update_result_exit_code=$?  # On a non-success exit code` `          if [[ $update_result_exit_code -ne 0 ]]; then` `            echo "updating user ${USER_NAME} failed: ${update_result}"` `            READ_LIST_SUCCESS=1` `            break` `          else` `            echo "Updated user ${USER_NAME}..."` `            USERS_LIST=$(new_users_list "${USERS_LIST}" "${USER_NAME}")` `          fi` `        else` `          # Another error occurred, so output the original message and exit code` `          echo "error creating user ${USER_NAME}: ${creation_result}"` `          READ_LIST_SUCCESS=1` `          break` `        fi` `      # On a success, the user was created so output that` `      else` `        echo "Created user ${USER_NAME}..."` `        USERS_LIST=$(new_users_list "${USERS_LIST}" "${USER_NAME}")` `      fi` `    done < $USERS_FILE` `` `    if [[ -n "${USERS_LIST}" && ${READ_LIST_SUCCESS} ]]; then` `      echo "Setting superusers configurations with users [${USERS_LIST}]"` `      superuser_result=$(rpk cluster config set superusers [${USERS_LIST}] 2>&1) && superuser_result_exit_code=$? || superuser_result_exit_code=$?` `      if [[ $superuser_result_exit_code -ne 0 ]]; then` `          echo "Setting superusers configurations failed: ${superuser_result}"` `      else` `          echo "Completed setting superusers configurations"` `      fi` `    fi` `  }` `` `  # first time processing` `  process_users $USERS_DIR` `` `  # subsequent changes detected here` `  # watching delete_self as documented in https://ahmet.im/blog/kubernetes-inotify/` `  USERS_FILE=$(find ${USERS_DIR}/* -print)` `  while RES=$(inotifywait -q -e delete_self ${USERS_FILE}); do` `    process_users $USERS_DIR` `  done` `done`)) -}}
{{- else -}}
{{- $saslUserSh = (concat (default (list ) $saslUserSh) (list `echo "Nothing to do. Sleeping..."` `sleep infinity`)) -}}
{{- end -}}
{{- $_ := (set $secret.stringData "sasl-user.sh" (join "\n" $saslUserSh)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secret) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretFSValidator" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.statefulset.initContainers.fsValidator.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $secret := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-fs-validator" (substr 0 (49 | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "type" "Opaque" "stringData" (dict ) )) -}}
{{- $_ := (set $secret.stringData "fsValidator.sh" `set -e
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

echo "passed"`) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secret) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretConfigurator" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $secret := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-configurator" (substr 0 (51 | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "type" "Opaque" "stringData" (dict ) )) -}}
{{- $configuratorSh := (list ) -}}
{{- $configuratorSh = (concat (default (list ) $configuratorSh) (list `set -xe` `SERVICE_NAME=$1` `KUBERNETES_NODE_NAME=$2` `POD_ORDINAL=${SERVICE_NAME##*-}` "BROKER_INDEX=`expr $POD_ORDINAL + 1`" `` `CONFIG=/etc/redpanda/redpanda.yaml` `` `# Setup config files` `cp /tmp/base-config/redpanda.yaml "${CONFIG}"` `cp /tmp/base-config/bootstrap.yaml /etc/redpanda/.bootstrap.yaml`)) -}}
{{- if (not (get (fromJson (include "redpanda.RedpandaAtLeast_22_3_0" (dict "a" (list $dot) ))) "r")) -}}
{{- $configuratorSh = (concat (default (list ) $configuratorSh) (list `` `# Configure bootstrap` `## Not used for Redpanda v22.3.0+` `rpk --config "${CONFIG}" redpanda config set redpanda.node_id "${POD_ORDINAL}"` `if [ "${POD_ORDINAL}" = "0" ]; then` `	rpk --config "${CONFIG}" redpanda config set redpanda.seed_servers '[]' --format yaml` `fi`)) -}}
{{- end -}}
{{- $kafkaSnippet := (get (fromJson (include "redpanda.secretConfiguratorKafkaConfig" (dict "a" (list $dot) ))) "r") -}}
{{- $configuratorSh = (concat (default (list ) $configuratorSh) (default (list ) $kafkaSnippet)) -}}
{{- $httpSnippet := (get (fromJson (include "redpanda.secretConfiguratorHTTPConfig" (dict "a" (list $dot) ))) "r") -}}
{{- $configuratorSh = (concat (default (list ) $configuratorSh) (default (list ) $httpSnippet)) -}}
{{- if (and (get (fromJson (include "redpanda.RedpandaAtLeast_22_3_0" (dict "a" (list $dot) ))) "r") $values.rackAwareness.enabled) -}}
{{- $configuratorSh = (concat (default (list ) $configuratorSh) (list `` `# Configure Rack Awareness` `set +x` (printf `RACK=$(curl --silent --cacert /run/secrets/kubernetes.io/serviceaccount/ca.crt --fail -H 'Authorization: Bearer '$(cat /run/secrets/kubernetes.io/serviceaccount/token) "https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT_HTTPS}/api/v1/nodes/${KUBERNETES_NODE_NAME}?pretty=true" | grep %s | grep -v '\"key\":' | sed 's/.*": "\([^"]\+\).*/\1/')` (squote (quote $values.rackAwareness.nodeAnnotation))) `set -x` `rpk --config "$CONFIG" redpanda config set redpanda.rack "${RACK}"`)) -}}
{{- end -}}
{{- $_ := (set $secret.stringData "configurator.sh" (join "\n" $configuratorSh)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $secret) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.secretConfiguratorKafkaConfig" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $internalAdvertiseAddress := (printf "%s.%s" "${SERVICE_NAME}" (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) -}}
{{- $snippet := (coalesce nil) -}}
{{- $listenerName := "kafka" -}}
{{- $listenerAdvertisedName := $listenerName -}}
{{- $redpandaConfigPart := "redpanda" -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `LISTENER=%s` (quote (toJson (dict "name" "internal" "address" $internalAdvertiseAddress "port" ($values.listeners.kafka.port | int) )))) (printf `rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[0] "$LISTENER"` $redpandaConfigPart $listenerAdvertisedName))) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $values.listeners.kafka.external) ))) "r") | int) (0 | int)) -}}
{{- $externalCounter := (0 | int) -}}
{{- range $externalName, $externalVals := $values.listeners.kafka.external -}}
{{- $externalCounter = ((add $externalCounter (1 | int)) | int) -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `ADVERTISED_%s_ADDRESSES=()` (upper $listenerName)))) -}}
{{- range $_, $replicaIndex := (until ($values.statefulset.replicas | int)) -}}
{{- $port := ($externalVals.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $externalVals.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $externalVals.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = (index $externalVals.advertisedPorts (0 | int)) -}}
{{- else -}}
{{- $port = (index $externalVals.advertisedPorts $replicaIndex) -}}
{{- end -}}
{{- end -}}
{{- $host := (get (fromJson (include "redpanda.advertisedHostJSON" (dict "a" (list $dot $externalName $port $replicaIndex) ))) "r") -}}
{{- $address := (toJson $host) -}}
{{- $prefixTemplate := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $externalVals.prefixTemplate "") ))) "r") -}}
{{- if (eq $prefixTemplate "") -}}
{{- $prefixTemplate = (default "" $values.external.prefixTemplate) -}}
{{- end -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `PREFIX_TEMPLATE=%s` (quote $prefixTemplate)) (printf `ADVERTISED_%s_ADDRESSES+=(%s)` (upper $listenerName) (quote $address)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[%d] "${ADVERTISED_%s_ADDRESSES[$POD_ORDINAL]}"` $redpandaConfigPart $listenerAdvertisedName $externalCounter (upper $listenerName)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $snippet) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.secretConfiguratorHTTPConfig" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $internalAdvertiseAddress := (printf "%s.%s" "${SERVICE_NAME}" (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) -}}
{{- $snippet := (coalesce nil) -}}
{{- $listenerName := "http" -}}
{{- $listenerAdvertisedName := "pandaproxy" -}}
{{- $redpandaConfigPart := "pandaproxy" -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `LISTENER=%s` (quote (toJson (dict "name" "internal" "address" $internalAdvertiseAddress "port" ($values.listeners.http.port | int) )))) (printf `rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[0] "$LISTENER"` $redpandaConfigPart $listenerAdvertisedName))) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $values.listeners.http.external) ))) "r") | int) (0 | int)) -}}
{{- $externalCounter := (0 | int) -}}
{{- range $externalName, $externalVals := $values.listeners.http.external -}}
{{- $externalCounter = ((add $externalCounter (1 | int)) | int) -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `ADVERTISED_%s_ADDRESSES=()` (upper $listenerName)))) -}}
{{- range $_, $replicaIndex := (until ($values.statefulset.replicas | int)) -}}
{{- $port := ($externalVals.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $externalVals.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $externalVals.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = (index $externalVals.advertisedPorts (0 | int)) -}}
{{- else -}}
{{- $port = (index $externalVals.advertisedPorts $replicaIndex) -}}
{{- end -}}
{{- end -}}
{{- $host := (get (fromJson (include "redpanda.advertisedHostJSON" (dict "a" (list $dot $externalName $port $replicaIndex) ))) "r") -}}
{{- $address := (toJson $host) -}}
{{- $prefixTemplate := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $externalVals.prefixTemplate "") ))) "r") -}}
{{- if (eq $prefixTemplate "") -}}
{{- $prefixTemplate = (default "" $values.external.prefixTemplate) -}}
{{- end -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `PREFIX_TEMPLATE=%s` (quote $prefixTemplate)) (printf `ADVERTISED_%s_ADDRESSES+=(%s)` (upper $listenerName) (quote $address)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $snippet = (concat (default (list ) $snippet) (list `` (printf `rpk redpanda config --config "$CONFIG" set %s.advertised_%s_api[%d] "${ADVERTISED_%s_ADDRESSES[$POD_ORDINAL]}"` $redpandaConfigPart $listenerAdvertisedName $externalCounter (upper $listenerName)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $snippet) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminTLSCurlFlags" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.admin.tls $values.tls) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" "") | toJson -}}
{{- break -}}
{{- end -}}
{{- $path := (printf "/etc/tls/certs/%s" $values.listeners.admin.tls.cert) -}}
{{- if $values.listeners.admin.tls.requireClientAuth -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "--cacert %s/ca.crt --cert %s/tls.crt --key %s/tls.key" $path $path $path)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "--cacert %s/ca.crt" $path)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.externalAdvertiseAddress" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $eaa := "${SERVICE_NAME}" -}}
{{- $externalDomainTemplate := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.external.domain "") ))) "r") -}}
{{- $expanded := (tpl $externalDomainTemplate $dot) -}}
{{- if (not (empty $expanded)) -}}
{{- $eaa = (printf "%s.%s" "${SERVICE_NAME}" $expanded) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $eaa) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.advertisedHostJSON" -}}
{{- $dot := (index .a 0) -}}
{{- $externalName := (index .a 1) -}}
{{- $port := (index .a 2) -}}
{{- $replicaIndex := (index .a 3) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $host := (dict "name" $externalName "address" (get (fromJson (include "redpanda.externalAdvertiseAddress" (dict "a" (list $dot) ))) "r") "port" $port ) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $values.external.addresses) ))) "r") | int) (0 | int)) -}}
{{- $address := "" -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $values.external.addresses) ))) "r") | int) (1 | int)) -}}
{{- $address = (index $values.external.addresses $replicaIndex) -}}
{{- else -}}
{{- $address = (index $values.external.addresses (0 | int)) -}}
{{- end -}}
{{- $domain_4 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.external.domain "") ))) "r") -}}
{{- if (ne $domain_4 "") -}}
{{- $host = (dict "name" $externalName "address" (printf "%s.%s" $address $domain_4) "port" $port ) -}}
{{- else -}}
{{- $host = (dict "name" $externalName "address" $address "port" $port ) -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $host) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminInternalHTTPProtocol" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.admin.tls $values.tls) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" "https") | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "http") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminInternalURL" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s://%s.%s.%s.svc.%s:%d" (get (fromJson (include "redpanda.adminInternalHTTPProtocol" (dict "a" (list $dot) ))) "r") `${SERVICE_NAME}` (get (fromJson (include "redpanda.ServiceName" (dict "a" (list $dot) ))) "r") $dot.Release.Namespace (trimSuffix "." $values.clusterDomain) ($values.listeners.admin.port | int))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

