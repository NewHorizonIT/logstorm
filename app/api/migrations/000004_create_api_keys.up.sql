CREATE TABLE api_keys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    owner_id   UUID NOT NULL,
    name       VARCHAR(100) NOT NULL,
    key_hash   TEXT NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT fk_api_keys_project
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_keys_owner
        FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_api_keys_key_hash
        UNIQUE (key_hash)
);

CREATE INDEX idx_api_keys_key_hash   ON api_keys(key_hash);
CREATE INDEX idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX idx_api_keys_owner_id   ON api_keys(owner_id);
