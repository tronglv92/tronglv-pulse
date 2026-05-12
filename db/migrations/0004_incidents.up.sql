CREATE TYPE incident_status AS ENUM (
    'open',
    'investigating',
    'resolving',
    'resolved',
    'closed'
);

CREATE TABLE incidents (
    id           BIGSERIAL       PRIMARY KEY,
    tenant_id    BIGINT          NOT NULL,
    title        TEXT            NOT NULL,
    description  TEXT,
    service_name VARCHAR(255)    NOT NULL,
    fingerprint  VARCHAR(64)     NOT NULL,
    status       incident_status NOT NULL DEFAULT 'open',
    severity     SMALLINT        NOT NULL DEFAULT 1,
    rca_summary  TEXT,
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_incidents_tenant_id    ON incidents (tenant_id);
CREATE INDEX idx_incidents_service_name ON incidents (service_name);
CREATE INDEX idx_incidents_status       ON incidents (status);
CREATE INDEX idx_incidents_fingerprint  ON incidents (fingerprint);
CREATE INDEX idx_incidents_created_at   ON incidents (created_at DESC);
