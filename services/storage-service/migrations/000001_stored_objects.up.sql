-- Phase 13.1: storage object metadata (provider-independent keys only)
CREATE TABLE IF NOT EXISTS stored_objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    object_key VARCHAR(512) NOT NULL,
    storage_driver VARCHAR(32) NOT NULL,
    mime_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_stored_objects_object_key_active
    ON stored_objects (object_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_stored_objects_team_id
    ON stored_objects (team_id)
    WHERE deleted_at IS NULL;
