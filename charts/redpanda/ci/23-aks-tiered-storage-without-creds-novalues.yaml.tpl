# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
---
storage:
  persistentVolume:
    storageClass: managed-csi
  tiered:
    persistentVolume:
      storageClass: managed-csi
    config:
      cloud_storage_enabled: true
      cloud_storage_credentials_source: config_file
      cloud_storage_segment_max_upload_interval_sec: 1
      cloud_storage_azure_storage_account: ${TEST_STORAGE_ACCOUNT}
      cloud_storage_azure_container: ${TEST_STORAGE_CONTAINER}
      cloud_storage_azure_shared_key: ${TEST_AZURE_SHARED_KEY}

resources:
  cpu:
    cores: 400m
  memory:
    container:
      max: 2.0Gi
    redpanda:
      memory: 1Gi
      reserveMemory: 100Mi

console:
  # Until https://github.com/redpanda-data/console-enterprise/pull/256 is released the console
  # test named `test-license-with-console.yaml` needs to work with unreleased Redpanda Console version.
  image:
    registry: redpandadata
    repository: console-unstable
    tag: master-8a51854
