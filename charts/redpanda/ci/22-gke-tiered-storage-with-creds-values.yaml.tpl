storage:
  tieredConfig:
    cloud_storage_enabled: true
    cloud_storage_api_endpoint: storage.googleapis.com
    cloud_storage_credentials_source: config_file
    cloud_storage_region: "US-WEST1"
    cloud_storage_bucket: "${TEST_BUCKET}"
    cloud_storage_segment_max_upload_interval_sec: 1
license_key: "${REDPANDA_SAMPLE_LICENSE}"


resources:
  cpu:
    cores: 1
  memory:
    container:
      max: 1Gi
