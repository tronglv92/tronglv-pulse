CREATE TABLE llm_calls (
    id                BIGSERIAL     PRIMARY KEY,
    tenant_id         BIGINT        NOT NULL,
    model             VARCHAR(100)  NOT NULL,
    purpose           VARCHAR(100)  NOT NULL,
    prompt_tokens     INT           NOT NULL DEFAULT 0,
    completion_tokens INT           NOT NULL DEFAULT 0,
    total_tokens      INT           NOT NULL DEFAULT 0,
    cost_usd          NUMERIC(10,6) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX idx_llm_calls_tenant_id  ON llm_calls (tenant_id);
CREATE INDEX idx_llm_calls_created_at ON llm_calls (created_at DESC);

-- L2 of the three-tier RCA cache (CLAUDE.md)
CREATE TABLE rca_cache (
    id          BIGSERIAL    PRIMARY KEY,
    fingerprint VARCHAR(64)  NOT NULL,
    summary     TEXT         NOT NULL,
    model       VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_rca_cache_fingerprint ON rca_cache (fingerprint);
