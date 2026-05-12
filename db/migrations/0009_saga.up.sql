-- status: 0=started 1=completed 2=compensating 3=failed
CREATE TABLE saga_instances (
    id           BIGSERIAL    PRIMARY KEY,
    saga_type    VARCHAR(100) NOT NULL,
    status       SMALLINT     NOT NULL DEFAULT 0,
    payload      JSONB        NOT NULL,
    current_step VARCHAR(100),
    error        TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_saga_instances_status     ON saga_instances (status);
CREATE INDEX idx_saga_instances_saga_type  ON saga_instances (saga_type);
CREATE INDEX idx_saga_instances_created_at ON saga_instances (created_at DESC);
