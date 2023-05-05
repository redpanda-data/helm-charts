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