## Datadog Autodiscovery annotations

### v1

In your `values.yaml`

```yaml
statefulset:
  annotations:
    ad.datadoghq.com/redpanda.logs: '[{"source": "redpanda", "service": "redpanda_cluster"}]'
    ad.datadoghq.com/redpanda.check_names: '["openmetrics", "openmetrics"]'
    ad.datadoghq.com/redpanda.init_configs: '[{}, {}]'
    ad.datadoghq.com/redpanda.instances: |-
      [
        {
          "openmetrics_endpoint": "http://%%host%%:9644/metrics",
          "namespace": "redpanda",
          "metrics": [
            {
              "vectorized_application_uptime": "application.uptime",
              "vectorized_application_build": "application.build"
            },
            {
              "vectorized_cluster_partition_committed_offset": "cluster.partition_committed_offset",
              "vectorized_cluster_partition_end_offset": "cluster.partition_end_offset",
              "vectorized_cluster_partition_high_watermark": "cluster.partition_high_watermark",
              "vectorized_cluster_partition_last_stable_offset": "cluster.partition_last_stable_offset",
              "vectorized_cluster_partition_leader": "cluster.partition_leader",
              "vectorized_cluster_partition_leader_id": "cluster.partition_leader_id",
              "vectorized_cluster_partition_records_fetched": "cluster.partition_records_fetched",
              "vectorized_cluster_partition_records_produced": "cluster.partition_records_produced",
              "vectorized_cluster_partition_under_replicated_replicas": "cluster.partition_under_replicated_replicas"
            },
            {
              "vectorized_httpd_connections_current": "httpd.connections_current",
              "vectorized_httpd_connections": "httpd.connections",
              "vectorized_httpd_read_errors": "httpd.read_errors",
              "vectorized_httpd_reply_errors": "httpd.reply_errors",
              "vectorized_httpd_requests_served": "httpd.requests_served"
            },
            {
              "vectorized_kafka_fetch_sessions_cache_mem_usage_bytes": "kafka.fetch_sessions_cache_mem_usage_bytes",
              "vectorized_kafka_fetch_sessions_cache_sessions_count": "kafka.fetch_sessions_cache_sessions_count",
              "vectorized_kafka_latency_fetch_latency_us": "kafka.latency_fetch_latency_us",
              "vectorized_kafka_latency_produce_latency_us": "kafka.latency_produce_latency_us",
              "vectorized_kafka_rpc_active_connections": "kafka.rpc_active_connections",
              "vectorized_kafka_rpc_connection_close_errors": "kafka.rpc_connection_close_errors",
              "vectorized_kafka_rpc_connects": "kafka.rpc_connects",
              "vectorized_kafka_rpc_consumed_mem_bytes": "kafka.rpc_consumed_mem_bytes",
              "vectorized_kafka_rpc_corrupted_headers": "kafka.rpc_corrupted_headers",
              "vectorized_kafka_rpc_dispatch_handler_latency": "kafka.rpc_dispatch_handler_latency",
              "vectorized_kafka_rpc_max_service_mem_bytes": "kafka.rpc_max_service_mem_bytes",
              "vectorized_kafka_rpc_method_not_found_errors": "kafka.rpc_method_not_found_errors",
              "vectorized_kafka_rpc_received_bytes": "kafka.rpc_received_bytes",
              "vectorized_kafka_rpc_requests_blocked_memory": "kafka.rpc_requests_blocked_memory",
              "vectorized_kafka_rpc_requests_completed": "kafka.rpc_requests_completed",
              "vectorized_kafka_rpc_requests_pending": "kafka.rpc_requests_pending",
              "vectorized_kafka_rpc_sent_bytes": "kafka.rpc_sent_bytes",
              "vectorized_kafka_rpc_service_errors": "kafka.rpc_service_errors",
              "vectorized_kafka_group_offset": "kafka.group_offset"
            },
            {
              "vectorized_leader_balancer_leader_transfer_error": "leader.balancer_leader_transfer_error",
              "vectorized_leader_balancer_leader_transfer_no_improvement": "leader.balancer_leader_transfer_no_improvement",
              "vectorized_leader_balancer_leader_transfer_succeeded": "leader.balancer_leader_transfer_succeeded",
              "vectorized_leader_balancer_leader_transfer_timeout": "leader.balancer_leader_transfer_timeout"
            },
            {
              "vectorized_pandaproxy_request_latency": "pandaproxy.request_latency"
            },
            {
              "vectorized_reactor_abandoned_failed_futures": "reactor.abandoned_failed_futures",
              "vectorized_reactor_aio_bytes_read": "reactor.aio_bytes_read",
              "vectorized_reactor_aio_bytes_write": "reactor.aio_bytes_write",
              "vectorized_reactor_aio_errors": "reactor.aio_errors",
              "vectorized_reactor_aio_reads": "reactor.aio_reads",
              "vectorized_reactor_aio_writes": "reactor.aio_writes",
              "vectorized_reactor_cpp_exceptions": "reactor.cpp_exceptions",
              "vectorized_reactor_cpu_busy_ms": "reactor.cpu_busy_ms",
              "vectorized_reactor_cpu_steal_time_ms": "reactor.cpu_steal_time_ms",
              "vectorized_reactor_fstream_read_bytes": "reactor.fstream_read_bytes",
              "vectorized_reactor_fstream_read_bytes_blocked": "reactor.fstream_read_bytes_blocked",
              "vectorized_reactor_fstream_reads": "reactor.fstream_reads",
              "vectorized_reactor_fstream_reads_ahead_bytes_discarded": "reactor.fstream_reads_ahead_bytes_discarded",
              "vectorized_reactor_fstream_reads_aheads_discarded": "reactor.fstream_reads_aheads_discarded",
              "vectorized_reactor_fstream_reads_blocked": "reactor.fstream_reads_blocked",
              "vectorized_reactor_fsyncs": "reactor.fsyncs",
              "vectorized_reactor_io_threaded_fallbacks": "reactor.io_threaded_fallbacks",
              "vectorized_reactor_logging_failures": "reactor.logging_failures",
              "vectorized_reactor_polls": "reactor.polls",
              "vectorized_reactor_tasks_pending": "reactor.tasks_pending",
              "vectorized_reactor_tasks_processed": "reactor.tasks_processed",
              "vectorized_reactor_timers_pending": "reactor.timers_pending",
              "vectorized_reactor_utilization": "reactor.utilization"
            },
            {
              "vectorized_storage_compaction_backlog_controller_backlog_size": "storage.compaction_backlog_controller_backlog_size",
              "vectorized_storage_compaction_backlog_controller_error": "storage.compaction_backlog_controller_error",
              "vectorized_storage_compaction_backlog_controller_shares": "storage.compaction_backlog_controller_shares",
              "vectorized_storage_kvstore_cached_bytes": "storage.kvstore_cached_bytes",
              "vectorized_storage_kvstore_entries_fetched": "storage.kvstore_entries_fetched",
              "vectorized_storage_kvstore_entries_removed": "storage.kvstore_entries_removed",
              "vectorized_storage_kvstore_entries_written": "storage.kvstore_entries_written",
              "vectorized_storage_kvstore_key_count": "storage.kvstore_key_count",
              "vectorized_storage_kvstore_segments_rolled": "storage.kvstore_segments_rolled",
              "vectorized_storage_log_batch_parse_errors": "storage.log_batch_parse_errors",
              "vectorized_storage_log_batch_write_errors": "storage.log_batch_write_errors",
              "vectorized_storage_log_batches_read": "storage.log_batches_read",
              "vectorized_storage_log_batches_written": "storage.log_batches_written",
              "vectorized_storage_log_cache_hits": "storage.log_cache_hits",
              "vectorized_storage_log_cache_misses": "storage.log_cache_misses",
              "vectorized_storage_log_cached_batches_read": "storage.log_cached_batches_read",
              "vectorized_storage_log_cached_read_bytes": "storage.log_cached_read_bytes",
              "vectorized_storage_log_compacted_segment": "storage.log_compacted_segment",
              "vectorized_storage_log_compaction_ratio": "storage.log_compaction_ratio",
              "vectorized_storage_log_corrupted_compaction_indices": "storage.log_corrupted_compaction_indices",
              "vectorized_storage_log_log_segments_active": "storage.log_log_segments_active",
              "vectorized_storage_log_log_segments_created": "storage.log_log_segments_created",
              "vectorized_storage_log_log_segments_removed": "storage.log_log_segments_removed",
              "vectorized_storage_log_partition_size": "storage.log_partition_size",
              "vectorized_storage_log_read_bytes": "storage.log_read_bytes",
              "vectorized_storage_log_readers_added": "storage.log_readers_added",
              "vectorized_storage_log_readers_evicted": "storage.log_readers_evicted",
              "vectorized_storage_log_written_bytes": "storage.log_written_bytes"
            },
            {
              "vectorized_alien_receive_batch_queue_length": "alien.receive_batch_queue_length",
              "vectorized_alien_total_received_messages": "alien.total_received_messages",
              "vectorized_alien_total_sent_messages": "alien.total_sent_messages"
            },
            {
              "vectorized_internal_rpc_active_connections": "internal_rpc.active_connections",
              "vectorized_internal_rpc_connection_close_errors": "internal_rpc.connection_close_errors",
              "vectorized_internal_rpc_connects": "internal_rpc.connects",
              "vectorized_internal_rpc_consumed_mem_bytes": "internal_rpc.consumed_mem_bytes",
              "vectorized_internal_rpc_corrupted_headers": "internal_rpc.corrupted_headers",
              "vectorized_internal_rpc_dispatch_handler_latency": "internal_rpc.dispatch_handler_latency",
              "vectorized_internal_rpc_max_service_mem_bytes": "internal_rpc.max_service_mem_bytes",
              "vectorized_internal_rpc_method_not_found_errors": "internal_rpc.method_not_found_errors",
              "vectorized_internal_rpc_received_bytes": "internal_rpc.received_bytes",
              "vectorized_internal_rpc_requests_blocked_memory": "internal_rpc.requests_blocked_memory",
              "vectorized_internal_rpc_requests_completed": "internal_rpc.requests_completed",
              "vectorized_internal_rpc_requests_pending": "internal_rpc.requests_pending",
              "vectorized_internal_rpc_sent_bytes": "internal_rpc.sent_bytes",
              "vectorized_internal_rpc_service_errors": "internal_rpc.service_errors"
            },
            {
              "vectorized_io_queue_delay": "io_queue.delay",
              "vectorized_io_queue_queue_length": "io_queue.queue_length",
              "vectorized_io_queue_shares": "io_queue.shares",
              "vectorized_io_queue_total_bytes": "io_queue.total_bytes",
              "vectorized_io_queue_total_delay_sec": "io_queue.total_delay_sec",
              "vectorized_io_queue_total_operations": "io_queue.total_operations"
            },
            {
              "vectorized_memory_allocated_memory": "memory.allocated_memory",
              "vectorized_memory_cross_cpu_free_operations": "memory.cross_cpu_free_operations",
              "vectorized_memory_free_memory": "memory.free_memory",
              "vectorized_memory_free_operations": "memory.free_operations",
              "vectorized_memory_malloc_live_objects": "memory.malloc_live_objects",
              "vectorized_memory_malloc_operations": "memory.malloc_operations",
              "vectorized_memory_reclaims_operations": "memory.reclaims_operations",
              "vectorized_memory_total_memory": "memory.total_memory"
            },
            {
              "vectorized_raft_done_replicate_requests": "raft.done_replicate_requests",
              "vectorized_raft_group_count": "raft.group_count",
              "vectorized_raft_heartbeat_requests_errors": "raft.heartbeat_requests_errors",
              "vectorized_raft_leader_for": "raft.leader_for",
              "vectorized_raft_leadership_changes": "raft.leadership_changes",
              "vectorized_raft_log_flushes": "raft.log_flushes",
              "vectorized_raft_log_truncations": "raft.log_truncations",
              "vectorized_raft_received_append_requests": "raft.received_append_requests",
              "vectorized_raft_received_vote_requests": "raft.received_vote_requests",
              "vectorized_raft_recovery_requests_errors": "raft.recovery_requests_errors",
              "vectorized_raft_replicate_ack_all_requests": "raft.replicate_ack_all_requests",
              "vectorized_raft_replicate_ack_leader_requests": "raft.replicate_ack_leader_requests",
              "vectorized_raft_replicate_ack_none_requests": "raft.replicate_ack_none_requests",
              "vectorized_raft_replicate_request_errors": "raft.replicate_request_errors",
              "vectorized_raft_sent_vote_requests": "raft.sent_vote_requests"
            },
            {
              "vectorized_rpc_client_active_connections": "rpc_client.active_connections",
              "vectorized_rpc_client_client_correlation_errors": "rpc_client.client_correlation_errors",
              "vectorized_rpc_client_connection_errors": "rpc_client.connection_errors",
              "vectorized_rpc_client_connects": "rpc_client.connects",
              "vectorized_rpc_client_corrupted_headers": "rpc_client.corrupted_headers",
              "vectorized_rpc_client_in_bytes": "rpc_client.in_bytes",
              "vectorized_rpc_client_out_bytes": "rpc_client.out_bytes",
              "vectorized_rpc_client_read_dispatch_errors": "rpc_client.read_dispatch_errors",
              "vectorized_rpc_client_request_errors": "rpc_client.request_errors",
              "vectorized_rpc_client_request_timeouts": "rpc_client.request_timeouts",
              "vectorized_rpc_client_requests": "rpc_client.requests",
              "vectorized_rpc_client_requests_blocked_memory": "rpc_client.requests_blocked_memory",
              "vectorized_rpc_client_requests_pending": "rpc_client.requests_pending",
              "vectorized_rpc_client_server_correlation_errors": "rpc_client.server_correlation_errors"
            },
            {
              "vectorized_scheduler_queue_length": "scheduler.queue_length",
              "vectorized_scheduler_runtime_ms": "scheduler.runtime_ms",
              "vectorized_scheduler_shares": "scheduler.shares",
              "vectorized_scheduler_starvetime_ms": "scheduler.starvetime_ms",
              "vectorized_scheduler_tasks_processed": "scheduler.tasks_processed",
              "vectorized_scheduler_time_spent_on_task_quota_violations_ms": "scheduler.time_spent_on_task_quota_violations_ms",
              "vectorized_scheduler_waittime_ms": "scheduler.waittime_ms"
            },
            {
              "vectorized_stall_detector_reported": "stall.detector_reported"
            }
          ]
        },
        {
          "openmetrics_endpoint": "http://%%host%%:9644/public_metrics",
          "namespace": "redpanda",
          "metrics": [
            {"redpanda_application_uptime_seconds_total": "application.uptime_seconds_total"},
            {
              "redpanda_cloud_storage_active_segments": "cloud_storage.active_segments",
              "redpanda_cloud_storage_deleted_segments": "cloud_storage.deleted_segments",
              "redpanda_cloud_storage_errors_total": "cloud_storage.errors_total",
              "redpanda_cloud_storage_readers": "cloud_storage.readers",
              "redpanda_cloud_storage_segments": "cloud_storage.segments",
              "redpanda_cloud_storage_segments_pending_deletion": "cloud_storage.segments_pending_deletion",
              "redpanda_cloud_storage_uploaded_bytes": "cloud_storage.uploaded_bytes"
            },
            {
              "redpanda_cluster_brokers": "cluster.brokers",
              "redpanda_cluster_controller_log_limit_requests_available_rps": "cluster.controller_log_limit_requests_available_rps",
              "redpanda_cluster_controller_log_limit_requests_dropped": "cluster.controller_log_limit_requests_dropped",
              "redpanda_cluster_partition_moving_from_node": "cluster.partition_moving_from_node",
              "redpanda_cluster_partition_moving_to_node": "cluster.partition_moving_to_node",
              "redpanda_cluster_partition_node_cancelling_movements": "cluster.partition_node_cancelling_movements",
              "redpanda_cluster_partitions": "cluster.partitions",
              "redpanda_cluster_topics": "cluster.topics",
              "redpanda_cluster_unavailable_partitions": "cluster.unavailable_partitions"
            },
            {"redpanda_cpu_busy_seconds_total": "cpu.busy_seconds_total"},
            {
              "redpanda_io_queue_total_read_ops": "io_queue.total_read_ops",
              "redpanda_io_queue_total_write_ops": "io_queue.total_write_ops"
            },
            {
              "redpanda_kafka_consumer_group_committed_offset": "kafka.consumer_group_committed_offset",
              "redpanda_kafka_consumer_group_consumers": "kafka.consumer_group_consumers",
              "redpanda_kafka_consumer_group_topics": "kafka.consumer_group_topics",
              "redpanda_kafka_max_offset": "kafka.max_offset",
              "redpanda_kafka_partitions": "kafka.partitions",
              "redpanda_kafka_replicas": "kafka.replicas",
              "redpanda_kafka_request_bytes_total": "kafka.request_bytes_total",
              "redpanda_kafka_request_latency_seconds_bucket": "kafka.request_latency_seconds_bucket",
              "redpanda_kafka_request_latency_seconds_count": "kafka.request_latency_seconds_count",
              "redpanda_kafka_request_latency_seconds_sum": "kafka.request_latency_seconds_sum",
              "redpanda_kafka_under_replicated_replicas": "kafka.under_replicated_replicas"
            },
            {
              "redpanda_memory_allocated_memory": "memory.allocated_memory",
              "redpanda_memory_available_memory": "memory.available_memory",
              "redpanda_memory_available_memory_low_water_mark": "memory.available_memory_low_water_mark",
              "redpanda_memory_free_memory": "memory.free_memory"
            },
            {
              "redpanda_node_status_rpcs_received": "node.status_rpcs_received",
              "redpanda_node_status_rpcs_sent": "node.status_rpcs_sent",
              "redpanda_node_status_rpcs_timed_out": "node.status_rpcs_timed_out"
            },
            {"redpanda_raft:recovery_partition_movement_available_bandwidth": "raft_recovery.partition_movement_available_bandwidth"},
            {
              "redpanda_rest_proxy_request_errors_total": "rest_proxy.request_errors_total",
              "redpanda_rest_proxy_request_latency_seconds_bucket": "rest_proxy.request_latency_seconds_bucket",
              "redpanda_rest_proxy_request_latency_seconds_count": "rest_proxy.request_latency_seconds_count",
              "redpanda_rest_proxy_request_latency_seconds_sum": "rest_proxy.request_latency_seconds_sum"
            },
            {
              "redpanda_rpc_request_errors_total": "rpc_request.errors_total",
              "redpanda_rpc_request_latency_seconds_bucket": "rpc_request.latency_seconds_bucket",
              "redpanda_rpc_request_latency_seconds_count": "rpc_request.latency_seconds_count",
              "redpanda_rpc_request_latency_seconds_sum": "rpc_request.latency_seconds_sum"
            },
            {"redpanda_scheduler_runtime_seconds_total": "scheduler.runtime_seconds_total"},
            {
              "redpanda_schema_registry_request_errors_total": "schema_registry.request_errors_total",
              "redpanda_schema_registry_request_latency_seconds_bucket": "schema_registry.latency_seconds_bucket",
              "redpanda_schema_registry_request_latency_seconds_count": "schema_registry.latency_seconds_count",
              "redpanda_schema_registry_request_latency_seconds_sum": "schema_registry.latency_seconds_sum"
            },
            {
              "redpanda_storage_disk_free_bytes": "storage.disk_free_bytes",
              "redpanda_storage_disk_free_space_alert": "storage.free_space_alert",
              "redpanda_storage_disk_total_bytes": "storage.total_bytes"
            }
          ]
        }
      ]

```

### v2

In your `values.yaml`

```yaml
statefulset:
  annotations:
    ad.datadoghq.com/redpanda.checks: |
      {
        "openmetrics": {
          "init_config": {},
          "instances": [
            {
              "openmetrics_endpoint": "http://%%host%%:9644/metrics",
              "namespace": "redpanda",
              "max_returned_metrics": 99999,
              "request_size": 128,
              "metrics": [
                {
                  "vectorized_application_uptime": "application.uptime"
                },
                {
                  "vectorized_application_build": "application.build"
                },
                {
                  "vectorized_cluster_partition_committed_offset": "cluster.partition_committed_offset"
                },
                {
                  "vectorized_cluster_partition_end_offset": "cluster.partition_end_offset"
                },
                {
                  "vectorized_cluster_partition_high_watermark": "cluster.partition_high_watermark"
                },
                {
                  "vectorized_cluster_partition_last_stable_offset": "cluster.partition_last_stable_offset"
                },
                {
                  "vectorized_cluster_partition_leader": "cluster.partition_leader"
                },
                {
                  "vectorized_cluster_partition_leader_id": "cluster.partition_leader_id"
                },
                {
                  "vectorized_cluster_partition_records_fetched": "cluster.partition_records_fetched"
                },
                {
                  "vectorized_cluster_partition_records_produced": "cluster.partition_records_produced"
                },
                {
                  "vectorized_cluster_partition_under_replicated_replicas": "cluster.partition_under_replicated_replicas"
                },
                {
                  "vectorized_httpd_connections_current": "httpd.connections_current"
                },
                {
                  "vectorized_httpd_connections": "httpd.connections"
                },
                {
                  "vectorized_httpd_read_errors": "httpd.read_errors"
                },
                {
                  "vectorized_httpd_reply_errors": "httpd.reply_errors"
                },
                {
                  "vectorized_httpd_requests_served": "httpd.requests_served"
                },
                {
                  "vectorized_kafka_fetch_sessions_cache_mem_usage_bytes": "kafka.fetch_sessions_cache_mem_usage_bytes"
                },
                {
                  "vectorized_kafka_fetch_sessions_cache_sessions_count": "kafka.fetch_sessions_cache_sessions_count"
                },
                {
                  "vectorized_kafka_latency_fetch_latency_us": "kafka.latency_fetch_latency_us"
                },
                {
                  "vectorized_kafka_latency_produce_latency_us": "kafka.latency_produce_latency_us"
                },
                {
                  "vectorized_kafka_rpc_active_connections": "kafka.rpc_active_connections"
                },
                {
                  "vectorized_kafka_rpc_connection_close_errors": "kafka.rpc_connection_close_errors"
                },
                {
                  "vectorized_kafka_rpc_connects": "kafka.rpc_connects"
                },
                {
                  "vectorized_kafka_rpc_consumed_mem_bytes": "kafka.rpc_consumed_mem_bytes"
                },
                {
                  "vectorized_kafka_rpc_corrupted_headers": "kafka.rpc_corrupted_headers"
                },
                {
                  "vectorized_kafka_rpc_dispatch_handler_latency": "kafka.rpc_dispatch_handler_latency"
                },
                {
                  "vectorized_kafka_rpc_max_service_mem_bytes": "kafka.rpc_max_service_mem_bytes"
                },
                {
                  "vectorized_kafka_rpc_method_not_found_errors": "kafka.rpc_method_not_found_errors"
                },
                {
                  "vectorized_kafka_rpc_received_bytes": "kafka.rpc_received_bytes"
                },
                {
                  "vectorized_kafka_rpc_requests_blocked_memory": "kafka.rpc_requests_blocked_memory"
                },
                {
                  "vectorized_kafka_rpc_requests_completed": "kafka.rpc_requests_completed"
                },
                {
                  "vectorized_kafka_rpc_requests_pending": "kafka.rpc_requests_pending"
                },
                {
                  "vectorized_kafka_rpc_sent_bytes": "kafka.rpc_sent_bytes"
                },
                {
                  "vectorized_kafka_rpc_service_errors": "kafka.rpc_service_errors"
                },
                {
                  "vectorized_kafka_group_offset": "kafka.group_offset"
                },
                {
                  "vectorized_leader_balancer_leader_transfer_error": "leader.balancer_leader_transfer_error"
                },
                {
                  "vectorized_leader_balancer_leader_transfer_no_improvement": "leader.balancer_leader_transfer_no_improvement"
                },
                {
                  "vectorized_leader_balancer_leader_transfer_succeeded": "leader.balancer_leader_transfer_succeeded"
                },
                {
                  "vectorized_leader_balancer_leader_transfer_timeout": "leader.balancer_leader_transfer_timeout"
                },
                {
                  "vectorized_pandaproxy_request_latency": "pandaproxy.request_latency"
                },
                {
                  "vectorized_reactor_abandoned_failed_futures": "reactor.abandoned_failed_futures"
                },
                {
                  "vectorized_reactor_aio_bytes_read": "reactor.aio_bytes_read"
                },
                {
                  "vectorized_reactor_aio_bytes_write": "reactor.aio_bytes_write"
                },
                {
                  "vectorized_reactor_aio_errors": "reactor.aio_errors"
                },
                {
                  "vectorized_reactor_aio_reads": "reactor.aio_reads"
                },
                {
                  "vectorized_reactor_aio_writes": "reactor.aio_writes"
                },
                {
                  "vectorized_reactor_cpp_exceptions": "reactor.cpp_exceptions"
                },
                {
                  "vectorized_reactor_cpu_busy_ms": "reactor.cpu_busy_ms"
                },
                {
                  "vectorized_reactor_cpu_steal_time_ms": "reactor.cpu_steal_time_ms"
                },
                {
                  "vectorized_reactor_fstream_read_bytes": "reactor.fstream_read_bytes"
                },
                {
                  "vectorized_reactor_fstream_read_bytes_blocked": "reactor.fstream_read_bytes_blocked"
                },
                {
                  "vectorized_reactor_fstream_reads": "reactor.fstream_reads"
                },
                {
                  "vectorized_reactor_fstream_reads_ahead_bytes_discarded": "reactor.fstream_reads_ahead_bytes_discarded"
                },
                {
                  "vectorized_reactor_fstream_reads_aheads_discarded": "reactor.fstream_reads_aheads_discarded"
                },
                {
                  "vectorized_reactor_fstream_reads_blocked": "reactor.fstream_reads_blocked"
                },
                {
                  "vectorized_reactor_fsyncs": "reactor.fsyncs"
                },
                {
                  "vectorized_reactor_io_threaded_fallbacks": "reactor.io_threaded_fallbacks"
                },
                {
                  "vectorized_reactor_logging_failures": "reactor.logging_failures"
                },
                {
                  "vectorized_reactor_polls": "reactor.polls"
                },
                {
                  "vectorized_reactor_tasks_pending": "reactor.tasks_pending"
                },
                {
                  "vectorized_reactor_tasks_processed": "reactor.tasks_processed"
                },
                {
                  "vectorized_reactor_timers_pending": "reactor.timers_pending"
                },
                {
                  "vectorized_reactor_utilization": "reactor.utilization"
                },
                {
                  "vectorized_storage_compaction_backlog_controller_backlog_size": "storage.compaction_backlog_controller_backlog_size"
                },
                {
                  "vectorized_storage_compaction_backlog_controller_error": "storage.compaction_backlog_controller_error"
                },
                {
                  "vectorized_storage_compaction_backlog_controller_shares": "storage.compaction_backlog_controller_shares"
                },
                {
                  "vectorized_storage_kvstore_cached_bytes": "storage.kvstore_cached_bytes"
                },
                {
                  "vectorized_storage_kvstore_entries_fetched": "storage.kvstore_entries_fetched"
                },
                {
                  "vectorized_storage_kvstore_entries_removed": "storage.kvstore_entries_removed"
                },
                {
                  "vectorized_storage_kvstore_entries_written": "storage.kvstore_entries_written"
                },
                {
                  "vectorized_storage_kvstore_key_count": "storage.kvstore_key_count"
                },
                {
                  "vectorized_storage_kvstore_segments_rolled": "storage.kvstore_segments_rolled"
                },
                {
                  "vectorized_storage_log_batch_parse_errors": "storage.log_batch_parse_errors"
                },
                {
                  "vectorized_storage_log_batch_write_errors": "storage.log_batch_write_errors"
                },
                {
                  "vectorized_storage_log_batches_read": "storage.log_batches_read"
                },
                {
                  "vectorized_storage_log_batches_written": "storage.log_batches_written"
                },
                {
                  "vectorized_storage_log_cache_hits": "storage.log_cache_hits"
                },
                {
                  "vectorized_storage_log_cache_misses": "storage.log_cache_misses"
                },
                {
                  "vectorized_storage_log_cached_batches_read": "storage.log_cached_batches_read"
                },
                {
                  "vectorized_storage_log_cached_read_bytes": "storage.log_cached_read_bytes"
                },
                {
                  "vectorized_storage_log_compacted_segment": "storage.log_compacted_segment"
                },
                {
                  "vectorized_storage_log_compaction_ratio": "storage.log_compaction_ratio"
                },
                {
                  "vectorized_storage_log_corrupted_compaction_indices": "storage.log_corrupted_compaction_indices"
                },
                {
                  "vectorized_storage_log_log_segments_active": "storage.log_log_segments_active"
                },
                {
                  "vectorized_storage_log_log_segments_created": "storage.log_log_segments_created"
                },
                {
                  "vectorized_storage_log_log_segments_removed": "storage.log_log_segments_removed"
                },
                {
                  "vectorized_storage_log_partition_size": "storage.log_partition_size"
                },
                {
                  "vectorized_storage_log_read_bytes": "storage.log_read_bytes"
                },
                {
                  "vectorized_storage_log_readers_added": "storage.log_readers_added"
                },
                {
                  "vectorized_storage_log_readers_evicted": "storage.log_readers_evicted"
                },
                {
                  "vectorized_storage_log_written_bytes": "storage.log_written_bytes"
                },
                {
                  "vectorized_alien_receive_batch_queue_length": "alien.receive_batch_queue_length"
                },
                {
                  "vectorized_alien_total_received_messages": "alien.total_received_messages"
                },
                {
                  "vectorized_alien_total_sent_messages": "alien.total_sent_messages"
                },
                {
                  "vectorized_internal_rpc_active_connections": "internal_rpc.active_connections"
                },
                {
                  "vectorized_internal_rpc_connection_close_errors": "internal_rpc.connection_close_errors"
                },
                {
                  "vectorized_internal_rpc_connects": "internal_rpc.connects"
                },
                {
                  "vectorized_internal_rpc_consumed_mem_bytes": "internal_rpc.consumed_mem_bytes"
                },
                {
                  "vectorized_internal_rpc_corrupted_headers": "internal_rpc.corrupted_headers"
                },
                {
                  "vectorized_internal_rpc_dispatch_handler_latency": "internal_rpc.dispatch_handler_latency"
                },
                {
                  "vectorized_internal_rpc_max_service_mem_bytes": "internal_rpc.max_service_mem_bytes"
                },
                {
                  "vectorized_internal_rpc_method_not_found_errors": "internal_rpc.method_not_found_errors"
                },
                {
                  "vectorized_internal_rpc_received_bytes": "internal_rpc.received_bytes"
                },
                {
                  "vectorized_internal_rpc_requests_blocked_memory": "internal_rpc.requests_blocked_memory"
                },
                {
                  "vectorized_internal_rpc_requests_completed": "internal_rpc.requests_completed"
                },
                {
                  "vectorized_internal_rpc_requests_pending": "internal_rpc.requests_pending"
                },
                {
                  "vectorized_internal_rpc_sent_bytes": "internal_rpc.sent_bytes"
                },
                {
                  "vectorized_internal_rpc_service_errors": "internal_rpc.service_errors"
                },
                {
                  "vectorized_io_queue_delay": "io_queue.delay"
                },
                {
                  "vectorized_io_queue_queue_length": "io_queue.queue_length"
                },
                {
                  "vectorized_io_queue_shares": "io_queue.shares"
                },
                {
                  "vectorized_io_queue_total_bytes": "io_queue.total_bytes"
                },
                {
                  "vectorized_io_queue_total_delay_sec": "io_queue.total_delay_sec"
                },
                {
                  "vectorized_io_queue_total_operations": "io_queue.total_operations"
                },
                {
                  "vectorized_memory_allocated_memory": "memory.allocated_memory"
                },
                {
                  "vectorized_memory_cross_cpu_free_operations": "memory.cross_cpu_free_operations"
                },
                {
                  "vectorized_memory_free_memory": "memory.free_memory"
                },
                {
                  "vectorized_memory_free_operations": "memory.free_operations"
                },
                {
                  "vectorized_memory_malloc_live_objects": "memory.malloc_live_objects"
                },
                {
                  "vectorized_memory_malloc_operations": "memory.malloc_operations"
                },
                {
                  "vectorized_memory_reclaims_operations": "memory.reclaims_operations"
                },
                {
                  "vectorized_memory_total_memory": "memory.total_memory"
                },
                {
                  "vectorized_raft_done_replicate_requests": "raft.done_replicate_requests"
                },
                {
                  "vectorized_raft_group_count": "raft.group_count"
                },
                {
                  "vectorized_raft_heartbeat_requests_errors": "raft.heartbeat_requests_errors"
                },
                {
                  "vectorized_raft_leader_for": "raft.leader_for"
                },
                {
                  "vectorized_raft_leadership_changes": "raft.leadership_changes"
                },
                {
                  "vectorized_raft_log_flushes": "raft.log_flushes"
                },
                {
                  "vectorized_raft_log_truncations": "raft.log_truncations"
                },
                {
                  "vectorized_raft_received_append_requests": "raft.received_append_requests"
                },
                {
                  "vectorized_raft_received_vote_requests": "raft.received_vote_requests"
                },
                {
                  "vectorized_raft_recovery_requests_errors": "raft.recovery_requests_errors"
                },
                {
                  "vectorized_raft_replicate_ack_all_requests": "raft.replicate_ack_all_requests"
                },
                {
                  "vectorized_raft_replicate_ack_leader_requests": "raft.replicate_ack_leader_requests"
                },
                {
                  "vectorized_raft_replicate_ack_none_requests": "raft.replicate_ack_none_requests"
                },
                {
                  "vectorized_raft_replicate_request_errors": "raft.replicate_request_errors"
                },
                {
                  "vectorized_raft_sent_vote_requests": "raft.sent_vote_requests"
                },
                {
                  "vectorized_rpc_client_active_connections": "rpc_client.active_connections"
                },
                {
                  "vectorized_rpc_client_client_correlation_errors": "rpc_client.client_correlation_errors"
                },
                {
                  "vectorized_rpc_client_connection_errors": "rpc_client.connection_errors"
                },
                {
                  "vectorized_rpc_client_connects": "rpc_client.connects"
                },
                {
                  "vectorized_rpc_client_corrupted_headers": "rpc_client.corrupted_headers"
                },
                {
                  "vectorized_rpc_client_in_bytes": "rpc_client.in_bytes"
                },
                {
                  "vectorized_rpc_client_out_bytes": "rpc_client.out_bytes"
                },
                {
                  "vectorized_rpc_client_read_dispatch_errors": "rpc_client.read_dispatch_errors"
                },
                {
                  "vectorized_rpc_client_request_errors": "rpc_client.request_errors"
                },
                {
                  "vectorized_rpc_client_request_timeouts": "rpc_client.request_timeouts"
                },
                {
                  "vectorized_rpc_client_requests": "rpc_client.requests"
                },
                {
                  "vectorized_rpc_client_requests_blocked_memory": "rpc_client.requests_blocked_memory"
                },
                {
                  "vectorized_rpc_client_requests_pending": "rpc_client.requests_pending"
                },
                {
                  "vectorized_rpc_client_server_correlation_errors": "rpc_client.server_correlation_errors"
                },
                {
                  "vectorized_scheduler_queue_length": "scheduler.queue_length"
                },
                {
                  "vectorized_scheduler_runtime_ms": "scheduler.runtime_ms"
                },
                {
                  "vectorized_scheduler_shares": "scheduler.shares"
                },
                {
                  "vectorized_scheduler_starvetime_ms": "scheduler.starvetime_ms"
                },
                {
                  "vectorized_scheduler_tasks_processed": "scheduler.tasks_processed"
                },
                {
                  "vectorized_scheduler_time_spent_on_task_quota_violations_ms": "scheduler.time_spent_on_task_quota_violations_ms"
                },
                {
                  "vectorized_scheduler_waittime_ms": "scheduler.waittime_ms"
                },
                {
                  "vectorized_stall_detector_reported": "stall.detector_reported"
                }
              ]
            },
            {
              "openmetrics_endpoint": "http://%%host%%:9644/public_metrics",
              "namespace": "redpanda",
              "request_size": 128,
              "metrics": [
                {
                  "redpanda_application_uptime_seconds_total": "application.uptime_seconds_total"
                },
                {
                  "redpanda_cloud_storage_active_segments": "cloud_storage.active_segments"
                },
                {
                  "redpanda_cloud_storage_deleted_segments": "cloud_storage.deleted_segments"
                },
                {
                  "redpanda_cloud_storage_errors_total": "cloud_storage.errors_total"
                },
                {
                  "redpanda_cloud_storage_readers": "cloud_storage.readers"
                },
                {
                  "redpanda_cloud_storage_segments": "cloud_storage.segments"
                },
                {
                  "redpanda_cloud_storage_segments_pending_deletion": "cloud_storage.segments_pending_deletion"
                },
                {
                  "redpanda_cloud_storage_uploaded_bytes": "cloud_storage.uploaded_bytes"
                },
                {
                  "redpanda_cluster_brokers": "cluster.brokers"
                },
                {
                  "redpanda_cluster_controller_log_limit_requests_available_rps": "cluster.controller_log_limit_requests_available_rps"
                },
                {
                  "redpanda_cluster_controller_log_limit_requests_dropped": "cluster.controller_log_limit_requests_dropped"
                },
                {
                  "redpanda_cluster_partition_moving_from_node": "cluster.partition_moving_from_node"
                },
                {
                  "redpanda_cluster_partition_moving_to_node": "cluster.partition_moving_to_node"
                },
                {
                  "redpanda_cluster_partition_node_cancelling_movements": "cluster.partition_node_cancelling_movements"
                },
                {
                  "redpanda_cluster_partitions": "cluster.partitions"
                },
                {
                  "redpanda_cluster_topics": "cluster.topics"
                },
                {
                  "redpanda_cluster_unavailable_partitions": "cluster.unavailable_partitions"
                },
                {
                  "redpanda_cpu_busy_seconds_total": "cpu.busy_seconds_total"
                },
                {
                  "redpanda_io_queue_total_read_ops": "io_queue.total_read_ops"
                },
                {
                  "redpanda_io_queue_total_write_ops": "io_queue.total_write_ops"
                },
                {
                  "redpanda_kafka_consumer_group_committed_offset": "kafka.consumer_group_committed_offset"
                },
                {
                  "redpanda_kafka_consumer_group_consumers": "kafka.consumer_group_consumers"
                },
                {
                  "redpanda_kafka_consumer_group_topics": "kafka.consumer_group_topics"
                },
                {
                  "redpanda_kafka_max_offset": "kafka.max_offset"
                },
                {
                  "redpanda_kafka_partitions": "kafka.partitions"
                },
                {
                  "redpanda_kafka_replicas": "kafka.replicas"
                },
                {
                  "redpanda_kafka_request_bytes": "kafka.request_bytes"
                },
                {
                  "redpanda_kafka_request_latency_seconds": "kafka.request_latency_seconds"
                },
                {
                  "redpanda_kafka_request_latency_seconds_bucket": "kafka.request_latency_seconds_bucket"
                },
                {
                  "redpanda_kafka_request_latency_seconds_count": "kafka.request_latency_seconds_count"
                },
                {
                  "redpanda_kafka_request_latency_seconds_sum": "kafka.request_latency_seconds_sum"
                },
                {
                  "redpanda_kafka_under_replicated_replicas": "kafka.under_replicated_replicas"
                },
                {
                  "redpanda_memory_allocated_memory": "memory.allocated_memory"
                },
                {
                  "redpanda_memory_available_memory": "memory.available_memory"
                },
                {
                  "redpanda_memory_available_memory_low_water_mark": "memory.available_memory_low_water_mark"
                },
                {
                  "redpanda_memory_free_memory": "memory.free_memory"
                },
                {
                  "redpanda_node_status_rpcs_received": "node.status_rpcs_received"
                },
                {
                  "redpanda_node_status_rpcs_sent": "node.status_rpcs_sent"
                },
                {
                  "redpanda_node_status_rpcs_timed_out": "node.status_rpcs_timed_out"
                },
                {
                  "redpanda_raft:recovery_partition_movement_available_bandwidth": "raft_recovery.partition_movement_available_bandwidth"
                },
                {
                  "redpanda_rest_proxy_request_errors_total": "rest_proxy.request_errors_total"
                },
                {
                  "redpanda_rest_proxy_request_latency_seconds_bucket": "rest_proxy.request_latency_seconds_bucket"
                },
                {
                  "redpanda_rest_proxy_request_latency_seconds_count": "rest_proxy.request_latency_seconds_count"
                },
                {
                  "redpanda_rest_proxy_request_latency_seconds_sum": "rest_proxy.request_latency_seconds_sum"
                },
                {
                  "redpanda_rpc_request_errors_total": "rpc_request.errors_total"
                },
                {
                  "redpanda_rpc_request_latency_seconds_bucket": "rpc_request.latency_seconds_bucket"
                },
                {
                  "redpanda_rpc_request_latency_seconds_count": "rpc_request.latency_seconds_count"
                },
                {
                  "redpanda_rpc_request_latency_seconds_sum": "rpc_request.latency_seconds_sum"
                },
                {
                  "redpanda_scheduler_runtime_seconds_total": "scheduler.runtime_seconds_total"
                },
                {
                  "redpanda_schema_registry_request_errors_total": "schema_registry.request_errors_total"
                },
                {
                  "redpanda_schema_registry_request_latency_seconds_bucket": "schema_registry.latency_seconds_bucket"
                },
                {
                  "redpanda_schema_registry_request_latency_seconds_count": "schema_registry.latency_seconds_count"
                },
                {
                  "redpanda_schema_registry_request_latency_seconds_sum": "schema_registry.latency_seconds_sum"
                },
                {
                  "redpanda_storage_disk_free_bytes": "storage.disk_free_bytes"
                },
                {
                  "redpanda_storage_disk_free_space_alert": "storage.free_space_alert"
                },
                {
                  "redpanda_storage_disk_total_bytes": "storage.disk_total_bytes"
                }
              ]
            }
          ]
        }
      }
```
