SELECT remove_continuous_aggregate_policy('logs_metrics_5min', if_exists => true);
DROP MATERIALIZED VIEW IF EXISTS logs_metrics_5min;
