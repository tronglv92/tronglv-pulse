CREATE MATERIALIZED VIEW logs_metrics_5min
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', created_at)                AS bucket,
    tenant_id,
    service_name,
    level,
    COUNT(*)                                            AS log_count,
    COUNT(*) FILTER (WHERE level IN ('ERROR', 'FATAL')) AS error_count
FROM log_entries
GROUP BY bucket, tenant_id, service_name, level
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'logs_metrics_5min',
    start_offset      => INTERVAL '1 hour',
    end_offset        => INTERVAL '1 minute',
    schedule_interval => INTERVAL '5 minutes'
);
