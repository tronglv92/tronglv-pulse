CREATE TABLE error_embeddings (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL,
    fingerprint VARCHAR(64)  NOT NULL,
    embedding   vector(1536) NOT NULL,
    model       VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_error_embeddings_tenant_id   ON error_embeddings (tenant_id);
CREATE INDEX idx_error_embeddings_fingerprint ON error_embeddings (fingerprint);
CREATE UNIQUE INDEX idx_error_embeddings_fp_tenant
    ON error_embeddings (tenant_id, fingerprint) WHERE deleted_at IS NULL;
CREATE INDEX idx_error_embeddings_vector
    ON error_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
