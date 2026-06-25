# Template Sync Schema Gap Analysis

**Date:** 2026-06-22  
**Rule:** No migrations in Phase 11.1

## Required Table: `template_sync_runs`

Phase 11.1 introduces sync run audit tracking. Production persistence requires:

```sql
CREATE TABLE template_sync_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    status VARCHAR(16) NOT NULL,
    synced_count INTEGER NOT NULL DEFAULT 0,
    created_count INTEGER NOT NULL DEFAULT 0,
    updated_count INTEGER NOT NULL DEFAULT 0,
    removed_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    meta_cursor VARCHAR(512),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_template_sync_runs_team_started
    ON template_sync_runs (team_id, started_at DESC);
CREATE INDEX idx_template_sync_runs_channel_started
    ON template_sync_runs (channel_id, started_at DESC);
```

## Current Workaround

`InMemorySyncRunRepository` — sync history and metrics survive only for process lifetime.

## Sufficient Without Migration

| Need | Existing Schema |
|------|-----------------|
| Per-template sync time | `templates.last_sync_at` |
| Drift detection | `templates.meta_template_id` + unique index |
| Idempotent upsert | `idx_templates_channel_meta_id_active` |

## Optional Future Columns

| Column | Table | Purpose |
|--------|-------|---------|
| `last_template_sync_at` | `channels` | Fast channel freshness without join |

## Recommendation

Add `template_sync_runs` migration in Phase 11.2 alongside embedded signup persistence.
