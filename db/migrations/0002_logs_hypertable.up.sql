CREATE TABLE log_entries (
    id           BIGSERIAL    NOT NULL,
    tenant_id    BIGINT       NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    level        VARCHAR(20)  NOT NULL,
    message      TEXT         NOT NULL,
    fingerprint  VARCHAR(64)  NOT NULL,
    metadata     JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
);
SELECT create_hypertable('log_entries', 'created_at', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_log_entries_tenant_id    ON log_entries (tenant_id, created_at DESC);
CREATE INDEX idx_log_entries_service_name ON log_entries (service_name, created_at DESC);
CREATE INDEX idx_log_entries_level        ON log_entries (level, created_at DESC);
CREATE INDEX idx_log_entries_fingerprint  ON log_entries (fingerprint);
CREATE INDEX idx_log_entries_message_trgm ON log_entries USING gin (message gin_trgm_ops);

CREATE TABLE anomaly_events (
    id           BIGSERIAL        NOT NULL,
    tenant_id    BIGINT           NOT NULL,
    service_name VARCHAR(255)     NOT NULL,
    fingerprint  VARCHAR(64)      NOT NULL,
    z_score      DOUBLE PRECISION NOT NULL,
    window_size  INT              NOT NULL,
    log_count    INT              NOT NULL,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
);
SELECT create_hypertable('anomaly_events', 'created_at', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_anomaly_events_tenant_id    ON anomaly_events (tenant_id, created_at DESC);
CREATE INDEX idx_anomaly_events_fingerprint  ON anomaly_events (fingerprint);
CREATE INDEX idx_anomaly_events_service_name ON anomaly_events (service_name, created_at DESC);
