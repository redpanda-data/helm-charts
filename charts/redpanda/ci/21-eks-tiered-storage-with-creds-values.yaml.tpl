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
  tieredConfig:
    cloud_storage_enabled: true
    cloud_storage_credentials_source: config_file
    cloud_storage_access_key: "${AWS_ACCESS_KEY_ID}"
    cloud_storage_secret_key: "${AWS_SECRET_ACCESS_KEY}"
    cloud_storage_region: "${AWS_REGION}"
    cloud_storage_bucket: "${TEST_BUCKET}"
    cloud_storage_segment_max_upload_interval_sec: 1
license_key: "${REDPANDA_SAMPLE_LICENSE}"