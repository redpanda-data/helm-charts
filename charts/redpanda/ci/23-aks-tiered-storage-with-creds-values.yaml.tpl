storage:
  persistentVolume:
    storageClass: managed-csi
  tieredConfig:
    cloud_storage_enabled: true
    cloud_storage_credentials_source: config_file
    cloud_storage_segment_max_upload_interval_sec: 1
    cloud_storage_azure_storage_account: ${TEST_STORAGE_ACCOUNT}
    cloud_storage_azure_container: ${TEST_STORAGE_CONTAINER}
    cloud_storage_azure_shared_key: ${TEST_AZURE_SHARED_KEY}
license_key: "${REDPANDA_SAMPLE_LICENSE}"

resources:
  cpu:
    cores: 400m
  memory:
    container:
      max: 2.0Gi
    redpanda:
      memory: 1Gi
      reserveMemory: 100Mi