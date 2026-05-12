-- status: 0=pending 1=published 2=failed
CREATE TABLE outbox_events (
    id           BIGSERIAL    PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type   VARCHAR(100) NOT NULL,
    topic        VARCHAR(255) NOT NULL,
    payload      JSONB        NOT NULL,
    status       SMALLINT     NOT NULL DEFAULT 0,
    retry_count  SMALLINT     NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_outbox_events_status       ON outbox_events (status);
CREATE INDEX idx_outbox_events_created_at   ON outbox_events (created_at ASC);
CREATE INDEX idx_outbox_events_aggregate_id ON outbox_events (aggregate_id);
